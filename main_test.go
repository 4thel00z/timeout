package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var binary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "timeout-test")
	if err != nil {
		panic(err)
	}
	binary = filepath.Join(dir, "timeout")
	build := exec.Command("go", "build", "-o", binary, ".")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		panic(err)
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

type outcome struct {
	code    int
	stderr  string
	elapsed time.Duration
}

func runTimeout(t *testing.T, args ...string) outcome {
	t.Helper()
	cmd := exec.Command(binary, args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		t.Fatalf("running timeout: %v", err)
	}
	return outcome{code: cmd.ProcessState.ExitCode(), stderr: stderr.String(), elapsed: elapsed}
}

func TestExitCodes(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want int
	}{
		{"finishes in time", []string{"5", "true"}, 0},
		{"passes exit status through", []string{"5", "sh", "-c", "exit 3"}, 3},
		{"times out", []string{"0.2", "sleep", "5"}, exitTimedOut},
		{"zero disables timeout", []string{"0", "sh", "-c", "sleep 0.3; exit 4"}, 4},
		{"kill signal", []string{"-s", "KILL", "0.2", "sleep", "5"}, 137},
		{"preserve status", []string{"--preserve-status", "0.2", "sleep", "5"}, 143},
		{"kill after ignored term", []string{"-k", "0.3", "0.2", "sh", "-c", "trap '' TERM; sleep 5"}, 137},
		{"command not found", []string{"5", "definitely-not-a-command-xyz"}, exitNotFound},
		{"not executable", []string{"5", "/etc/hosts"}, exitCannotInvoke},
		{"missing operand", []string{"5"}, exitFailure},
		{"bad duration", []string{"soon", "true"}, exitFailure},
		{"bad signal", []string{"-s", "NOPE", "1", "true"}, exitFailure},
		{"help", []string{"--help"}, 0},
		{"version", []string{"--version"}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := runTimeout(t, tc.args...)
			if got.code != tc.want {
				t.Fatalf("exit code = %d, want %d (stderr: %s)", got.code, tc.want, got.stderr)
			}
		})
	}
}

func TestTimeoutStopsEarly(t *testing.T) {
	got := runTimeout(t, "0.2", "sleep", "10")
	if got.elapsed > 3*time.Second {
		t.Fatalf("timeout took %v, expected it to stop the command after about 200ms", got.elapsed)
	}
}

func TestSignalsWholeProcessGroup(t *testing.T) {
	got := runTimeout(t, "0.2", "sh", "-c", "sleep 10 & sleep 10; wait")
	if got.code != exitTimedOut {
		t.Fatalf("exit code = %d, want %d", got.code, exitTimedOut)
	}
	if got.elapsed > 3*time.Second {
		t.Fatalf("grandchildren kept timeout alive for %v", got.elapsed)
	}
}

func TestVerboseReportsSignal(t *testing.T) {
	got := runTimeout(t, "-v", "0.2", "sleep", "5")
	if !strings.Contains(got.stderr, "sending signal TERM to command 'sleep'") {
		t.Fatalf("verbose output missing signal report: %q", got.stderr)
	}
}
