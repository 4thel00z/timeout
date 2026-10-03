package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var suffixUnits = map[string]time.Duration{
	"s": time.Second,
	"m": time.Minute,
	"h": time.Hour,
	"d": 24 * time.Hour,
}

var signalsByName = map[string]syscall.Signal{
	"HUP":    syscall.SIGHUP,
	"INT":    syscall.SIGINT,
	"QUIT":   syscall.SIGQUIT,
	"ILL":    syscall.SIGILL,
	"TRAP":   syscall.SIGTRAP,
	"ABRT":   syscall.SIGABRT,
	"BUS":    syscall.SIGBUS,
	"FPE":    syscall.SIGFPE,
	"KILL":   syscall.SIGKILL,
	"USR1":   syscall.SIGUSR1,
	"SEGV":   syscall.SIGSEGV,
	"USR2":   syscall.SIGUSR2,
	"PIPE":   syscall.SIGPIPE,
	"ALRM":   syscall.SIGALRM,
	"TERM":   syscall.SIGTERM,
	"CHLD":   syscall.SIGCHLD,
	"CONT":   syscall.SIGCONT,
	"STOP":   syscall.SIGSTOP,
	"TSTP":   syscall.SIGTSTP,
	"TTIN":   syscall.SIGTTIN,
	"TTOU":   syscall.SIGTTOU,
	"URG":    syscall.SIGURG,
	"XCPU":   syscall.SIGXCPU,
	"XFSZ":   syscall.SIGXFSZ,
	"VTALRM": syscall.SIGVTALRM,
	"PROF":   syscall.SIGPROF,
	"WINCH":  syscall.SIGWINCH,
	"IO":     syscall.SIGIO,
	"SYS":    syscall.SIGSYS,
}

func parseDuration(text string) (time.Duration, error) {
	invalid := fmt.Errorf("invalid time interval '%s'", text)
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return 0, invalid
	}
	number, unit := splitSuffix(trimmed)
	value, err := strconv.ParseFloat(number, 64)
	if err != nil {
		d, err := time.ParseDuration(trimmed)
		if err != nil || d < 0 {
			return 0, invalid
		}
		return d, nil
	}
	if value < 0 || math.IsNaN(value) {
		return 0, invalid
	}
	scaled := math.Ceil(value * float64(unit))
	if scaled >= math.MaxInt64 {
		return time.Duration(math.MaxInt64), nil
	}
	return time.Duration(scaled), nil
}

func splitSuffix(text string) (string, time.Duration) {
	unit, ok := suffixUnits[text[len(text)-1:]]
	if !ok {
		return text, time.Second
	}
	return text[:len(text)-1], unit
}

func parseSignal(text string) (syscall.Signal, error) {
	if number, err := strconv.Atoi(text); err == nil {
		if number <= 0 || number >= 32 {
			return 0, fmt.Errorf("invalid signal '%s'", text)
		}
		return syscall.Signal(number), nil
	}
	name := strings.TrimPrefix(strings.ToUpper(text), "SIG")
	sig, ok := signalsByName[name]
	if !ok {
		return 0, fmt.Errorf("invalid signal '%s'", text)
	}
	return sig, nil
}

func signalName(sig syscall.Signal) string {
	for name, candidate := range signalsByName {
		if candidate == sig {
			return name
		}
	}
	return strconv.Itoa(int(sig))
}
