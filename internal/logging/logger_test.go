package logging

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

func Test_Info_text(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	l := New(LevelInfo, &out, &bytes.Buffer{})
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

func Test_Error_somethingFailed(t *testing.T) {
	t.Parallel()

	var errOut bytes.Buffer
	l := New(LevelInfo, &bytes.Buffer{}, &errOut)
	l.Error("something failed", Fields{"code": 42})

	line := errOut.String()
	if !strings.Contains(line, "[ERROR]") {
		t.Errorf("expected [ERROR] in output, got %q", line)
	}
	if !strings.Contains(line, "something failed") {
		t.Errorf("expected message in output, got %q", line)
	}
}

func Test_write_nilWriter(t *testing.T) {
	t.Parallel()

	l := New(LevelInfo, nil, nil)
	l.Info("no output", Fields{})
	l.Error("no output", Fields{})
}

func Test_write_emptyFields(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	l := New(LevelInfo, &out, &bytes.Buffer{})
	l.Info("clean", Fields{})

	line := strings.TrimSpace(out.String())
	if !strings.HasSuffix(line, "clean") {
		t.Errorf("expected line to end with message, got %q", line)
	}
}

func Test_write_fieldsSorted(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	l := New(LevelInfo, &out, &bytes.Buffer{})
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

func Test_write_concurrent(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	l := New(LevelInfo, &out, &bytes.Buffer{})

	const goroutines = 20
	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			l.Info("concurrent", Fields{"n": "v"})
		})
	}
	wg.Wait()

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != goroutines {
		t.Errorf("expected %d lines, got %d", goroutines, len(lines))
	}
}

func Test_Output_unstructured(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	l := New(LevelInfo, &out, &bytes.Buffer{})
	output := "Plan: 5 to add, 2 to change, 1 to destroy"
	l.Output(output)

	line := strings.TrimSpace(out.String())
	if line != output {
		t.Errorf("expected unstructured output %q, got %q", output, line)
	}
}

func Test_Output_multiLine(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	l := New(LevelInfo, &out, &bytes.Buffer{})
	output := "Line 1\nLine 2\nLine 3"
	l.Output(output)

	result := out.String()
	if !strings.Contains(result, "Line 1") || !strings.Contains(result, "Line 2") || !strings.Contains(result, "Line 3") {
		t.Errorf("expected all lines in output, got %q", result)
	}
}

func Test_Output_empty(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	l := New(LevelInfo, &out, &bytes.Buffer{})
	l.Output("")

	if out.Len() != 0 {
		t.Errorf("expected no output for empty string, got %q", out.String())
	}
}

func Test_Output_nilWriter(t *testing.T) {
	t.Parallel()

	l := New(LevelInfo, nil, &bytes.Buffer{})
	l.Output("test") // Should not panic
}

func Test_write_levelFiltering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		minLevel   Level
		logFn      func(l *Logger)
		wantOut    bool
		wantErrOut bool
	}{
		{
			name:     "debug suppressed at info level",
			minLevel: LevelInfo,
			logFn:    func(l *Logger) { l.Debug("dbg", Fields{}) },
			wantOut:  false,
		},
		{
			name:     "debug emitted at debug level",
			minLevel: LevelDebug,
			logFn:    func(l *Logger) { l.Debug("dbg", Fields{}) },
			wantOut:  true,
		},
		{
			name:     "info emitted at info level",
			minLevel: LevelInfo,
			logFn:    func(l *Logger) { l.Info("msg", Fields{}) },
			wantOut:  true,
		},
		{
			name:     "info suppressed at warning level",
			minLevel: LevelWarning,
			logFn:    func(l *Logger) { l.Info("msg", Fields{}) },
			wantOut:  false,
		},
		{
			name:       "error suppressed at nothing — always emitted",
			minLevel:   LevelError,
			logFn:      func(l *Logger) { l.Error("err", Fields{}) },
			wantErrOut: true,
		},
		{
			name:       "error suppressed below its level",
			minLevel:   LevelError,
			logFn:      func(l *Logger) { l.Warning("warn", Fields{}) },
			wantOut:    false,
			wantErrOut: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var out, errOut bytes.Buffer
			l := New(tc.minLevel, &out, &errOut)
			tc.logFn(l)

			if tc.wantOut && out.Len() == 0 {
				t.Errorf("expected output in out, got nothing")
			}
			if !tc.wantOut && out.Len() > 0 {
				t.Errorf("expected no output in out, got %q", out.String())
			}
			if tc.wantErrOut && errOut.Len() == 0 {
				t.Errorf("expected output in errOut, got nothing")
			}
			if !tc.wantErrOut && errOut.Len() > 0 {
				t.Errorf("expected no output in errOut, got %q", errOut.String())
			}
		})
	}
}

func Test_Debug_text(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	l := New(LevelDebug, &out, &bytes.Buffer{})
	l.Debug("debug message", Fields{"trace": "id-1"})

	line := out.String()
	if !strings.Contains(line, "[DEBUG]") {
		t.Errorf("expected [DEBUG] in output, got %q", line)
	}
	if !strings.Contains(line, "debug message") {
		t.Errorf("expected message in output, got %q", line)
	}
	if !strings.Contains(line, "trace=id-1") {
		t.Errorf("expected field in output, got %q", line)
	}
}

func Test_Warning_text(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	l := New(LevelWarning, &out, &bytes.Buffer{})
	l.Warning("something degraded", Fields{"component": "runner"})

	line := out.String()
	if !strings.Contains(line, "[WARNING]") {
		t.Errorf("expected [WARNING] in output, got %q", line)
	}
	if !strings.Contains(line, "something degraded") {
		t.Errorf("expected message in output, got %q", line)
	}
}

func Test_ParseLevel_validValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warning", LevelWarning},
		{"warn", LevelWarning},
		{"WARNING", LevelWarning},
		{"error", LevelError},
		{"ERROR", LevelError},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got, err := ParseLevel(tc.input)
			if err != nil {
				t.Fatalf("ParseLevel(%q) unexpected error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("ParseLevel(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func Test_ParseLevel_invalidValue(t *testing.T) {
	t.Parallel()

	_, err := ParseLevel("verbose")
	if err == nil {
		t.Error("expected error for unknown level, got nil")
	}
}

func Test_ParseLevel_emptyString(t *testing.T) {
	t.Parallel()

	_, err := ParseLevel("")
	if err == nil {
		t.Error("expected error for empty string, got nil")
	}
}

func Test_LogSummaryTable(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	l := New(LevelInfo, &out, &bytes.Buffer{})

	rows := []SummaryRow{
		{Module: "vpc", Command: "plan", Status: "SUCCESS", Duration: "1s", Notes: ""},
		{Module: "db", Command: "plan", Status: "FAILURE", Duration: "500ms", Notes: "error"},
	}

	l.LogSummaryTable(rows, "1.5s")

	result := out.String()
	// tablewriter by default upper-cases headers and footers
	expectedStrings := []string{
		"MODULE", "COMMAND", "STATUS", "DURATION", "NOTES",
		"vpc", "plan", "SUCCESS", "1s",
		"db", "FAILURE", "500ms", "error",
		"TOTAL RUNTIME", "1.5S",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(result, s) {
			t.Errorf("expected summary table to contain %q, but it didn't. Result:\n%s", s, result)
		}
	}
}

func Test_Level_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		level Level
		want  string
	}{
		{LevelDebug, "debug"},
		{LevelInfo, "info"},
		{LevelWarning, "warning"},
		{LevelError, "error"},
		{Level(99), "unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()

			if got := tc.level.String(); got != tc.want {
				t.Errorf("Level.String() = %q, want %q", got, tc.want)
			}
		})
	}
}
