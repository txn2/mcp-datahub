package client

import (
	"fmt"
	"log"
	"os"
)

// Logger defines the interface for debug logging.
// Compatible with slog.Logger and other structured loggers.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// NopLogger is a logger that discards all output.
type NopLogger struct{}

// Debug implements Logger.
func (NopLogger) Debug(_ string, _ ...any) {}

// Info implements Logger.
func (NopLogger) Info(_ string, _ ...any) {}

// Warn implements Logger.
func (NopLogger) Warn(_ string, _ ...any) {}

// Error implements Logger.
func (NopLogger) Error(_ string, _ ...any) {}

// StdLogger wraps the standard library logger.
type StdLogger struct {
	logger *log.Logger
	debug  bool
}

// NewStdLogger creates a new StdLogger.
// If debug is true, Debug messages are logged; otherwise they are discarded.
func NewStdLogger(debug bool) *StdLogger {
	return &StdLogger{
		logger: log.New(os.Stderr, "[datahub] ", log.LstdFlags),
		debug:  debug,
	}
}

// Debug implements Logger.
func (l *StdLogger) Debug(msg string, args ...any) {
	if l.debug {
		l.log("DEBUG", msg, args...)
	}
}

// Info implements Logger.
func (l *StdLogger) Info(msg string, args ...any) {
	l.log("INFO", msg, args...)
}

// Warn implements Logger.
func (l *StdLogger) Warn(msg string, args ...any) {
	l.log("WARN", msg, args...)
}

// Error implements Logger.
func (l *StdLogger) Error(msg string, args ...any) {
	l.log("ERROR", msg, args...)
}

func (l *StdLogger) log(level, msg string, args ...any) {
	if len(args) == 0 {
		l.logger.Printf("%s: %s", level, msg)
		return
	}

	// Format key-value pairs
	pairs := make([]string, 0, len(args)/2)
	for i := 0; i < len(args)-1; i += 2 {
		key, ok := args[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", args[i])
		}
		pairs = append(pairs, fmt.Sprintf("%s=%v", key, args[i+1]))
	}

	if len(pairs) > 0 {
		l.logger.Printf("%s: %s [%s]", level, msg, joinStrings(pairs, " "))
	} else {
		l.logger.Printf("%s: %s", level, msg)
	}
}

// joinStrings joins strings with a separator (avoiding strings import for this simple case).
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}
