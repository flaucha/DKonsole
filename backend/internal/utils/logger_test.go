package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLoggerFunctions(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	Logger.SetOutput(&buf)
	Logger.SetFormatter(&logrus.JSONFormatter{
		DisableTimestamp: true, // easier to test
	})

	// Reset after test
	defer func() {
		Logger.SetOutput(os.Stdout)
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}()

	t.Run("LogInfo", func(t *testing.T) {
		buf.Reset()
		LogInfo("test info", map[string]interface{}{"key": "value"})

		var entry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("failed to parse log: %v", err)
		}

		if entry["level"] != "info" {
			t.Errorf("expected level info, got %v", entry["level"])
		}
		if entry["msg"] != "test info" {
			t.Errorf("expected msg 'test info', got %v", entry["msg"])
		}
		if entry["key"] != "value" {
			t.Errorf("expected key 'value', got %v", entry["key"])
		}
		if entry["type"] != "info" {
			t.Errorf("expected type 'info', got %v", entry["type"])
		}
	})

	t.Run("LogWarn", func(t *testing.T) {
		buf.Reset()
		LogWarn("test warn", map[string]interface{}{"key": "warn_val"})

		var entry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("failed to parse log: %v", err)
		}

		if entry["level"] != "warning" {
			t.Errorf("expected level warning, got %v", entry["level"])
		}
		if entry["msg"] != "test warn" {
			t.Errorf("expected msg 'test warn', got %v", entry["msg"])
		}
		if entry["key"] != "warn_val" {
			t.Errorf("expected key 'warn_val', got %v", entry["key"])
		}
	})

	t.Run("LogError", func(t *testing.T) {
		buf.Reset()
		err := errors.New("test error")
		LogError(err, "test msg", map[string]interface{}{"ctx": "err_ctx"})

		var entry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("failed to parse log: %v", err)
		}

		if entry["level"] != "error" {
			t.Errorf("expected level error, got %v", entry["level"])
		}
		if entry["msg"] != "test msg" {
			t.Errorf("expected msg 'test msg', got %v", entry["msg"])
		}
		if entry["error"] != "test error" {
			t.Errorf("expected error 'test error', got %v", entry["error"])
		}
		if entry["ctx"] != "err_ctx" {
			t.Errorf("expected ctx 'err_ctx', got %v", entry["ctx"])
		}
	})

	t.Run("LogAuditEntry", func(t *testing.T) {
		buf.Reset()
		entry := AuditLogEntry{
			User:    "user1",
			Action:  "login",
			Success: true,
			IP:      "127.0.0.1",
			Details: map[string]interface{}{"method": "POST"},
		}
		LogAuditEntry(entry)

		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("failed to parse log: %v", err)
		}

		if logEntry["type"] != "audit" {
			t.Errorf("expected type audit, got %v", logEntry["type"])
		}
		if logEntry["user"] != "user1" {
			t.Errorf("expected user user1, got %v", logEntry["user"])
		}
		if logEntry["success"] != true {
			t.Errorf("expected success true, got %v", logEntry["success"])
		}
		if logEntry["method"] != "POST" {
			t.Errorf("expected flattened detail method=POST, got %v", logEntry["method"])
		}
	})
}
