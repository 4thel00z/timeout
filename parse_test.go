package main

import (
	"syscall"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	cases := map[string]time.Duration{
		"0":     0,
		"10":    10 * time.Second,
		"1.5":   1500 * time.Millisecond,
		"2s":    2 * time.Second,
		"0.5m":  30 * time.Second,
		"1h":    time.Hour,
		"1d":    24 * time.Hour,
		"250ms": 250 * time.Millisecond,
		"1m30s": 90 * time.Second,
	}
	for input, want := range cases {
		got, err := parseDuration(input)
		if err != nil {
			t.Errorf("parseDuration(%q) error: %v", input, err)
			continue
		}
		if got != want {
			t.Errorf("parseDuration(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestParseDurationRejects(t *testing.T) {
	for _, input := range []string{"", "-1", "abc", "1x", "-2m", "NaN"} {
		if _, err := parseDuration(input); err == nil {
			t.Errorf("parseDuration(%q) accepted invalid input", input)
		}
	}
}

func TestParseSignal(t *testing.T) {
	cases := map[string]syscall.Signal{
		"TERM":    syscall.SIGTERM,
		"sigkill": syscall.SIGKILL,
		"SIGINT":  syscall.SIGINT,
		"hup":     syscall.SIGHUP,
		"9":       syscall.SIGKILL,
		"15":      syscall.SIGTERM,
	}
	for input, want := range cases {
		got, err := parseSignal(input)
		if err != nil {
			t.Errorf("parseSignal(%q) error: %v", input, err)
			continue
		}
		if got != want {
			t.Errorf("parseSignal(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestParseSignalRejects(t *testing.T) {
	for _, input := range []string{"", "0", "64", "-1", "NOPE", "SIG"} {
		if _, err := parseSignal(input); err == nil {
			t.Errorf("parseSignal(%q) accepted invalid input", input)
		}
	}
}

func TestSignalName(t *testing.T) {
	if got := signalName(syscall.SIGKILL); got != "KILL" {
		t.Errorf("signalName(SIGKILL) = %q, want KILL", got)
	}
}
