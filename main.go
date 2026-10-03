// Command timeout runs a command and stops it if it is still running after a time limit.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

const (
	exitTimedOut     = 124
	exitFailure      = 125
	exitCannotInvoke = 126
	exitNotFound     = 127
	exitSignalBase   = 128
)

var version = "dev"

var forwardedSignals = []os.Signal{
	syscall.SIGINT,
	syscall.SIGTERM,
	syscall.SIGHUP,
	syscall.SIGQUIT,
	syscall.SIGUSR1,
	syscall.SIGUSR2,
}

type options struct {
	duration       time.Duration
	killAfter      time.Duration
	signal         syscall.Signal
	preserveStatus bool
	foreground     bool
	verbose        bool
	command        []string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

func run(args []string, stderr io.Writer) int {
	opts, err := parseOptions(args, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "timeout: %v\n", err)
		fmt.Fprintln(stderr, "Try 'timeout --help' for more information.")
		return exitFailure
	}
	return supervise(opts, stderr)
}

func parseOptions(args []string, stderr io.Writer) (options, error) {
	var (
		opts        options
		signalName  string
		killAfter   string
		showVersion bool
	)
	flags := flag.NewFlagSet("timeout", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { printUsage(flags.Output()) }
	for _, name := range []string{"s", "signal"} {
		flags.StringVar(&signalName, name, "TERM", "")
	}
	for _, name := range []string{"k", "kill-after"} {
		flags.StringVar(&killAfter, name, "0", "")
	}
	for _, name := range []string{"v", "verbose"} {
		flags.BoolVar(&opts.verbose, name, false, "")
	}
	flags.BoolVar(&opts.preserveStatus, "preserve-status", false, "")
	flags.BoolVar(&opts.foreground, "foreground", false, "")
	flags.BoolVar(&showVersion, "version", false, "")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if showVersion {
		fmt.Fprintf(flags.Output(), "timeout %s\n", version)
		return options{}, flag.ErrHelp
	}
	if flags.NArg() < 2 {
		return options{}, errors.New("missing operand")
	}

	var err error
	if opts.signal, err = parseSignal(signalName); err != nil {
		return options{}, err
	}
	if opts.killAfter, err = parseDuration(killAfter); err != nil {
		return options{}, err
	}
	if opts.duration, err = parseDuration(flags.Arg(0)); err != nil {
		return options{}, err
	}
	opts.command = flags.Args()[1:]
	return opts, nil
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, `Usage: timeout [OPTION] DURATION COMMAND [ARG]...
Start COMMAND, and kill it if still running after DURATION.

Options:
  -s, --signal SIGNAL     signal to send on timeout (name or number, default TERM)
  -k, --kill-after DUR    also send KILL if COMMAND is still running DUR after the first signal
  -v, --verbose           report every signal sent on stderr
      --preserve-status   exit with COMMAND's status even when it times out
      --foreground        keep COMMAND in the foreground process group (TTY access, children not signalled)
      --version           print the version and exit
  -h, --help              print this help and exit

DURATION is a number with an optional suffix: s (seconds, default), m (minutes),
h (hours) or d (days). Go durations such as 1m30s or 250ms work too. 0 disables the timeout.

Exit status:
  124  COMMAND timed out and --preserve-status was not given
  125  timeout itself failed
  126  COMMAND was found but could not be invoked
  127  COMMAND was not found
  137  COMMAND (or timeout) was sent KILL
  else the exit status of COMMAND
`)
}

func supervise(opts options, stderr io.Writer) int {
	cmd := exec.Command(opts.command[0], opts.command[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if !opts.foreground {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	incoming := make(chan os.Signal, 1)
	signal.Notify(incoming, forwardedSignals...)
	defer signal.Stop(incoming)

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(stderr, "timeout: failed to run command '%s': %v\n", opts.command[0], err)
		return startFailureCode(err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	deadline := afterOrNever(opts.duration)
	var killDeadline <-chan time.Time
	timedOut := false
	send := func(sig syscall.Signal) {
		if opts.verbose {
			fmt.Fprintf(stderr, "timeout: sending signal %s to command '%s'\n", signalName(sig), opts.command[0])
		}
		deliver(cmd.Process.Pid, sig, opts.foreground)
		if !opts.foreground && sig != syscall.SIGKILL && sig != syscall.SIGCONT {
			deliver(cmd.Process.Pid, syscall.SIGCONT, false)
		}
	}
	armKill := func() {
		if opts.killAfter <= 0 || killDeadline != nil {
			return
		}
		killDeadline = time.After(opts.killAfter)
	}

	for {
		select {
		case <-done:
			return exitStatus(cmd.ProcessState, timedOut, opts.preserveStatus)
		case sig := <-incoming:
			send(sig.(syscall.Signal))
			armKill()
		case <-deadline:
			timedOut = true
			deadline = nil
			send(opts.signal)
			armKill()
		case <-killDeadline:
			killDeadline = nil
			send(syscall.SIGKILL)
		}
	}
}

func afterOrNever(d time.Duration) <-chan time.Time {
	if d <= 0 {
		return nil
	}
	return time.After(d)
}

func deliver(pid int, sig syscall.Signal, foreground bool) {
	if foreground {
		_ = syscall.Kill(pid, sig)
		return
	}
	_ = syscall.Kill(-pid, sig)
}

func startFailureCode(err error) int {
	if errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist) {
		return exitNotFound
	}
	return exitCannotInvoke
}

func exitStatus(state *os.ProcessState, timedOut, preserveStatus bool) int {
	status, ok := state.Sys().(syscall.WaitStatus)
	if !ok {
		return exitFailure
	}
	if status.Signaled() && status.Signal() == syscall.SIGKILL {
		return exitSignalBase + int(syscall.SIGKILL)
	}
	if timedOut && !preserveStatus {
		return exitTimedOut
	}
	if status.Signaled() {
		return exitSignalBase + int(status.Signal())
	}
	return status.ExitStatus()
}
