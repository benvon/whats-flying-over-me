package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestLoggerLevels(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	// Test all log levels
	testCases := []struct {
		level   Level
		message string
		fields  map[string]interface{}
	}{
		{LevelDebug, "debug message", map[string]interface{}{"key": "value"}},
		{LevelInfo, "info message", map[string]interface{}{"key": "value"}},
		{LevelWarn, "warn message", map[string]interface{}{"key": "value"}},
		{LevelErr, "error message", map[string]interface{}{"key": "value"}},
		{LevelCritical, "critical message", map[string]interface{}{"key": "value"}},
	}

	for _, tc := range testCases {
		switch tc.level {
		case LevelDebug:
			Debug(tc.message, tc.fields)
		case LevelInfo:
			Info(tc.message, tc.fields)
		case LevelWarn:
			Warn(tc.message, tc.fields)
		case LevelErr:
			Err(tc.message, tc.fields)
		case LevelCritical:
			Critical(tc.message, tc.fields)
		}
	}

	// Close the write end and read the output
	w.Close()
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != len(testCases) {
		t.Errorf("expected %d log lines, got %d", len(testCases), len(lines))
	}

	for i, line := range lines {
		if line == "" {
			continue
		}

		var entry logEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("failed to unmarshal log line %d: %v", i, err)
			continue
		}

		expected := testCases[i]
		if entry.Level != expected.level {
			t.Errorf("line %d: expected level %s, got %s", i, expected.level, entry.Level)
		}
		if entry.Message != expected.message {
			t.Errorf("line %d: expected message %s, got %s", i, expected.message, entry.Message)
		}
		if entry.Fields["key"] != "value" {
			t.Errorf("line %d: expected field value 'value', got %v", i, entry.Fields["key"])
		}
	}
}

func TestLoggerWithoutFields(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = oldStdout
	defer func() {
		os.Stdout = oldStdout
	}()

	// Test logging without fields
	Info("simple message", nil)

	// Close the write end and read the output
	w.Close()
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 1 {
		t.Errorf("expected 1 log line, got %d", len(lines))
	}

	if lines[0] == "" {
		return
	}

	var entry logEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("failed to unmarshal log line: %v", err)
	}

	if entry.Level != LevelInfo {
		t.Errorf("expected level %s, got %s", LevelInfo, entry.Level)
	}
	if entry.Message != "simple message" {
		t.Errorf("expected message 'simple message', got %s", entry.Message)
	}
	if entry.Fields == nil {
		t.Error("expected fields to be empty map, got nil")
	}
}
