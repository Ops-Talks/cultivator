package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
)

type Fields map[string]any

type Logger struct {
	format string
	out    io.Writer
	errOut io.Writer
	mu     sync.Mutex
}

func New(format string, out io.Writer, errOut io.Writer) *Logger {
	if format == "" {
		format = "text"
	}

	return &Logger{
		format: format,
		out:    out,
		errOut: errOut,
	}
}

func (l *Logger) Info(message string, fields Fields) {
	l.write("info", l.out, message, fields)
}

func (l *Logger) Error(message string, fields Fields) {
	l.write("error", l.errOut, message, fields)
}

func (l *Logger) write(level string, writer io.Writer, message string, fields Fields) {
	if writer == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.format == "json" {
		payload := map[string]any{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"level":     level,
			"message":   message,
		}

		for key, value := range fields {
			payload[key] = value
		}

		encoded, err := json.Marshal(payload)
		if err != nil {
			if _, writeErr := fmt.Fprintf(writer, "{\"level\":\"error\",\"message\":\"failed to encode log entry\"}\n"); writeErr != nil {
				return
			}
			return
		}

		if _, writeErr := fmt.Fprintf(writer, "%s\n", encoded); writeErr != nil {
			return
		}
		return
	}

	pairs := make([]string, 0, len(fields))
	for key := range fields {
		pairs = append(pairs, key)
	}

	sort.Strings(pairs)
	formatted := make([]string, 0, len(pairs))
	for _, key := range pairs {
		formatted = append(formatted, fmt.Sprintf("%s=%v", key, fields[key]))
	}

	line := strings.TrimSpace(fmt.Sprintf("[%s] %s %s", strings.ToUpper(level), message, strings.Join(formatted, " ")))
	if _, writeErr := fmt.Fprintf(writer, "%s\n", line); writeErr != nil {
		return
	}
}
