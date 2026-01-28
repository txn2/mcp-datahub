package client

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestNopLogger(t *testing.T) {
	// NopLogger should not panic and should do nothing
	l := NopLogger{}
	l.Debug("test", "key", "value")
	l.Info("test", "key", "value")
	l.Warn("test", "key", "value")
	l.Error("test", "key", "value")
}

func TestNewStdLogger(t *testing.T) {
	l := NewStdLogger(true)
	if l == nil {
		t.Fatal("NewStdLogger returned nil")
	}
	if !l.debug {
		t.Error("expected debug to be true")
	}

	l2 := NewStdLogger(false)
	if l2.debug {
		t.Error("expected debug to be false")
	}
}

func TestStdLoggerDebugEnabled(t *testing.T) {
	var buf bytes.Buffer
	l := &StdLogger{
		logger: log.New(&buf, "[test] ", 0),
		debug:  true,
	}

	l.Debug("debug message", "key", "value")
	output := buf.String()
	if !strings.Contains(output, "DEBUG") {
		t.Errorf("expected DEBUG in output, got: %s", output)
	}
	if !strings.Contains(output, "debug message") {
		t.Errorf("expected 'debug message' in output, got: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("expected 'key=value' in output, got: %s", output)
	}
}

func TestStdLoggerDebugDisabled(t *testing.T) {
	var buf bytes.Buffer
	l := &StdLogger{
		logger: log.New(&buf, "[test] ", 0),
		debug:  false,
	}

	l.Debug("debug message", "key", "value")
	if buf.Len() > 0 {
		t.Errorf("expected no output when debug disabled, got: %s", buf.String())
	}
}

func TestStdLoggerInfo(t *testing.T) {
	var buf bytes.Buffer
	l := &StdLogger{
		logger: log.New(&buf, "[test] ", 0),
		debug:  false,
	}

	l.Info("info message")
	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("expected INFO in output, got: %s", output)
	}
	if !strings.Contains(output, "info message") {
		t.Errorf("expected 'info message' in output, got: %s", output)
	}
}

func TestStdLoggerWarn(t *testing.T) {
	var buf bytes.Buffer
	l := &StdLogger{
		logger: log.New(&buf, "[test] ", 0),
		debug:  false,
	}

	l.Warn("warn message", "count", 42)
	output := buf.String()
	if !strings.Contains(output, "WARN") {
		t.Errorf("expected WARN in output, got: %s", output)
	}
	if !strings.Contains(output, "count=42") {
		t.Errorf("expected 'count=42' in output, got: %s", output)
	}
}

func TestStdLoggerError(t *testing.T) {
	var buf bytes.Buffer
	l := &StdLogger{
		logger: log.New(&buf, "[test] ", 0),
		debug:  false,
	}

	l.Error("error message", "err", "something failed")
	output := buf.String()
	if !strings.Contains(output, "ERROR") {
		t.Errorf("expected ERROR in output, got: %s", output)
	}
}

func TestStdLoggerNonStringKey(t *testing.T) {
	var buf bytes.Buffer
	l := &StdLogger{
		logger: log.New(&buf, "[test] ", 0),
		debug:  true,
	}

	// Pass a non-string key (int)
	l.Debug("test", 123, "value")
	output := buf.String()
	if !strings.Contains(output, "123=value") {
		t.Errorf("expected '123=value' in output, got: %s", output)
	}
}

func TestStdLoggerOddArgs(t *testing.T) {
	var buf bytes.Buffer
	l := &StdLogger{
		logger: log.New(&buf, "[test] ", 0),
		debug:  true,
	}

	// Odd number of args - last one should be ignored
	l.Debug("test", "key")
	output := buf.String()
	// Should still log the message
	if !strings.Contains(output, "DEBUG") {
		t.Errorf("expected DEBUG in output, got: %s", output)
	}
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name     string
		strs     []string
		sep      string
		expected string
	}{
		{
			name:     "empty slice",
			strs:     []string{},
			sep:      " ",
			expected: "",
		},
		{
			name:     "single element",
			strs:     []string{"a"},
			sep:      " ",
			expected: "a",
		},
		{
			name:     "multiple elements",
			strs:     []string{"a", "b", "c"},
			sep:      " ",
			expected: "a b c",
		},
		{
			name:     "comma separator",
			strs:     []string{"x", "y"},
			sep:      ",",
			expected: "x,y",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinStrings(tt.strs, tt.sep)
			if result != tt.expected {
				t.Errorf("joinStrings(%v, %q) = %q, want %q", tt.strs, tt.sep, result, tt.expected)
			}
		})
	}
}
