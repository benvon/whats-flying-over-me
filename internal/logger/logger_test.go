package logger

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

func capture(t *testing.T, f func()) map[string]interface{} {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	f()
	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}
	os.Stdout = old
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return m
}

func TestLevels(t *testing.T) {
	m := capture(t, func() { Warn("warn", nil) })
	if m["level"] != string(LevelWarn) {
		t.Errorf("expected WARN level, got %v", m["level"])
	}
	m = capture(t, func() { Err("err", nil) })
	if m["level"] != string(LevelErr) {
		t.Errorf("expected ERR level, got %v", m["level"])
	}
	m = capture(t, func() { Critical("crit", nil) })
	if m["level"] != string(LevelCritical) {
		t.Errorf("expected CRITICAL level, got %v", m["level"])
	}
}
