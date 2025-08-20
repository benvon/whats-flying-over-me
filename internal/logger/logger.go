package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Level string

const (
	LevelWarn     Level = "WARN"
	LevelErr      Level = "ERR"
	LevelCritical Level = "CRITICAL"
)

// logEntry represents a single log entry.
type logEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     Level                  `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

func log(level Level, msg string, fields map[string]interface{}) {
	entry := logEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   msg,
		Fields:    fields,
	}
	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(entry); err != nil {
		if _, err2 := fmt.Fprintf(os.Stdout, "{\"timestamp\":%q,\"level\":%q,\"message\":%q,\"error\":%q}\n", entry.Timestamp, level, msg, err); err2 != nil {
			_ = err2
		}
	}
}

func Warn(msg string, fields map[string]interface{}) {
	log(LevelWarn, msg, fields)
}

func Err(msg string, fields map[string]interface{}) {
	log(LevelErr, msg, fields)
}

func Critical(msg string, fields map[string]interface{}) {
	log(LevelCritical, msg, fields)
}
