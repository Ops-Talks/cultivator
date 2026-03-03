package logging

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
)

// Level represents the severity of a log message.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarning
	LevelError
)

// String returns the lowercase label for the level, used in log output.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarning:
		return "warning"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}

// ParseLevel converts a string to a Level. Accepted values (case-insensitive):
// "debug", "info", "warning", "error". Returns an error for unknown values.
func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warning", "warn":
		return LevelWarning, nil
	case "error":
		return LevelError, nil
	default:
		return LevelInfo, fmt.Errorf("unknown log level %q: must be one of debug, info, warning, error", s)
	}
}

// Fields holds structured key-value pairs attached to a log entry.
type Fields map[string]any

// Logger is a thread-safe structured logger with configurable minimum level.
type Logger struct {
	level  Level
	out    io.Writer
	errOut io.Writer
	mu     sync.Mutex
}

// New creates a Logger that only emits messages at or above the given level.
func New(level Level, out io.Writer, errOut io.Writer) *Logger {
	return &Logger{
		level:  level,
		out:    out,
		errOut: errOut,
	}
}

// Debug logs a message at DEBUG level.
func (l *Logger) Debug(message string, fields Fields) {
	l.write(LevelDebug, l.out, message, fields)
}

// Info logs a message at INFO level.
func (l *Logger) Info(message string, fields Fields) {
	l.write(LevelInfo, l.out, message, fields)
}

// Warning logs a message at WARNING level.
func (l *Logger) Warning(message string, fields Fields) {
	l.write(LevelWarning, l.out, message, fields)
}

// Error logs a message at ERROR level.
func (l *Logger) Error(message string, fields Fields) {
	l.write(LevelError, l.errOut, message, fields)
}

// Output logs unstructured, multi-line output (e.g., terragrunt/terraform output).
// Preserves formatting and structure without any encoding.
func (l *Logger) Output(output string) {
	if l.out == nil || strings.TrimSpace(output) == "" {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if _, writeErr := fmt.Fprintf(l.out, "%s\n", output); writeErr != nil {
		return
	}
}

func (l *Logger) write(level Level, writer io.Writer, message string, fields Fields) {
	if level < l.level {
		return
	}

	if writer == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	pairs := make([]string, 0, len(fields))
	for key := range fields {
		pairs = append(pairs, key)
	}

	sort.Strings(pairs)
	formatted := make([]string, 0, len(pairs))
	for _, key := range pairs {
		formatted = append(formatted, fmt.Sprintf("%s=%v", key, fields[key]))
	}

	line := strings.TrimSpace(fmt.Sprintf("[%s] %s %s", strings.ToUpper(level.String()), message, strings.Join(formatted, " ")))
	if _, writeErr := fmt.Fprintf(writer, "%s\n", line); writeErr != nil {
		return
	}
}
