package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
	"testing"
)

func TestNew_defaultsToTextFormat(t *testing.T) {
	t.Parallel()

	l := New("", &bytes.Buffer{}, &bytes.Buffer{})
	if l.format != "text" {
		t.Errorf("expected format text, got %q", l.format)
	}
}

func TestLogger_InfoText(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	l := New("text", &out, &bytes.Buffer{})
	l.Info("hello world", Fields{"key": "value"})

	line := out.String()
	if !strings.Contains(line, "[INFO]") {
		t.Errorf("expected [INFO] in output, got %q", line)
	}
	if !strings.Contains(line, "hello world") {
		t.Errorf("expected message in output, got %q", line)
	}
	if !strings.Contains(line, "key=value") {
		t.Errorf("expected key=value in output, got %q", line)
	}
}

func TestLogger_ErrorText(t *testing.T) {
	t.Parallel()

	var errOut bytes.Buffer
	l := New("text", &bytes.Buffer{}, &errOut)
	l.Error("something failed", Fields{"code": 42})

	line := errOut.String()
	if !strings.Contains(line, "[ERROR]") {
		t.Errorf("expected [ERROR] in output, got %q", line)
	}
	if !strings.Contains(line, "something failed") {
		t.Errorf("expected message in output, got %q", line)
	}
}

func TestLogger_InfoJSON(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	l := New("json", &out, &bytes.Buffer{})
	l.Info("json message", Fields{"env": "prod"})

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("output is not valid JSON: %v — got %q", err, out.String())
	}
	if payload["level"] != "info" {
		t.Errorf("expected level info, got %v", payload["level"])
	}
	if payload["message"] != "json message" {
		t.Errorf("expected message json message, got %v", payload["message"])
	}
	if payload["env"] != "prod" {
		t.Errorf("expected env prod, got %v", payload["env"])
	}
	if _, ok := payload["timestamp"]; !ok {
		t.Error("expected timestamp field in JSON output")
	}
}

func TestLogger_ErrorJSON(t *testing.T) {
	t.Parallel()

	var errOut bytes.Buffer
	l := New("json", &bytes.Buffer{}, &errOut)
	l.Error("fail", Fields{})

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(errOut.String())), &payload); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if payload["level"] != "error" {
		t.Errorf("expected level error, got %v", payload["level"])
	}
}

func TestLogger_NilWriter(t *testing.T) {
	t.Parallel()

	l := New("text", nil, nil)
	l.Info("no output", Fields{})
	l.Error("no output", Fields{})
}

func TestLogger_EmptyFields(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	l := New("text", &out, &bytes.Buffer{})
	l.Info("clean", Fields{})

	line := strings.TrimSpace(out.String())
	if !strings.HasSuffix(line, "clean") {
		t.Errorf("expected line to end with message, got %q", line)
	}
}

func TestLogger_FieldsSorted(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	l := New("text", &out, &bytes.Buffer{})
	l.Info("order", Fields{"z": "last", "a": "first", "m": "mid"})

	line := out.String()
	aPos := strings.Index(line, "a=first")
	mPos := strings.Index(line, "m=mid")
	zPos := strings.Index(line, "z=last")
	if aPos < 0 || mPos < 0 || zPos < 0 {
		t.Fatalf("expected all fields in output, got %q", line)
	}
	if aPos >= mPos || mPos >= zPos {
		t.Errorf("expected fields sorted alphabetically, got %q", line)
	}
}

func TestLogger_ConcurrentWrites(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	l := New("text", &out, &bytes.Buffer{})

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			l.Info("concurrent", Fields{"n": "v"})
		}()
	}
	wg.Wait()

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != goroutines {
		t.Errorf("expected %d lines, got %d", goroutines, len(lines))
	}
}
