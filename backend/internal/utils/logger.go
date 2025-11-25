package utils

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	// Logger is the global structured logger instance
	Logger *logrus.Logger
)

func init() {
	Logger = logrus.New()

	// Set log level from environment variable
	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch logLevel {
	case "debug":
		Logger.SetLevel(logrus.DebugLevel)
	case "info":
		Logger.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		Logger.SetLevel(logrus.WarnLevel)
	case "error":
		Logger.SetLevel(logrus.ErrorLevel)
	default:
		Logger.SetLevel(logrus.InfoLevel) // Default to info
	}

	// Set JSON formatter for structured logging
	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})

	// Set output to stdout (can be redirected in container)
	Logger.SetOutput(os.Stdout)
}

// AuditLogEntry represents a structured audit log entry
type AuditLogEntry struct {
	User      string                 `json:"user"`
	IP        string                 `json:"ip"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource,omitempty"`
	Namespace string                 `json:"namespace,omitempty"`
	Method    string                 `json:"method,omitempty"`
	Path      string                 `json:"path,omitempty"`
	Status    int                    `json:"status,omitempty"`
	Duration  string                 `json:"duration,omitempty"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// auditLogInternal writes a structured audit log entry (internal function)
func auditLogInternal(entry AuditLogEntry) {
	fields := logrus.Fields{
		"type":      "audit",
		"user":      entry.User,
		"ip":        entry.IP,
		"action":    entry.Action,
		"success":   entry.Success,
	}

	if entry.Resource != "" {
		fields["resource"] = entry.Resource
	}
	if entry.Namespace != "" {
		fields["namespace"] = entry.Namespace
	}
	if entry.Method != "" {
		fields["method"] = entry.Method
	}
	if entry.Path != "" {
		fields["path"] = entry.Path
	}
	if entry.Status > 0 {
		fields["status"] = entry.Status
	}
	if entry.Duration != "" {
		fields["duration"] = entry.Duration
	}
	if entry.Error != "" {
		fields["error"] = entry.Error
	}
	if len(entry.Details) > 0 {
		for k, v := range entry.Details {
			fields[k] = v
		}
	}

	Logger.WithFields(fields).Info("audit log")
}

// LogError logs an error with context
func LogError(err error, message string, fields map[string]interface{}) {
	logFields := logrus.Fields{
		"type":    "error",
		"message": message,
	}
	if err != nil {
		logFields["error"] = err.Error()
	}
	if fields != nil {
		for k, v := range fields {
			logFields[k] = v
		}
	}
	Logger.WithFields(logFields).Error(message)
}

// LogInfo logs an info message with context
func LogInfo(message string, fields map[string]interface{}) {
	logFields := logrus.Fields{
		"type": "info",
	}
	if fields != nil {
		for k, v := range fields {
			logFields[k] = v
		}
	}
	Logger.WithFields(logFields).Info(message)
}

// LogDebug logs a debug message with context
func LogDebug(message string, fields map[string]interface{}) {
	logFields := logrus.Fields{
		"type": "debug",
	}
	if fields != nil {
		for k, v := range fields {
			logFields[k] = v
		}
	}
	Logger.WithFields(logFields).Debug(message)
}

// LogWarn logs a warning message with context
func LogWarn(message string, fields map[string]interface{}) {
	logFields := logrus.Fields{
		"type": "warn",
	}
	if fields != nil {
		for k, v := range fields {
			logFields[k] = v
		}
	}
	Logger.WithFields(logFields).Warn(message)
}

