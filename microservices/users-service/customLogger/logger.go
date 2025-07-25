package customLogger

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"sync"
)

// Logger structure encapsulates the logger instance and configuration
type Logger struct {
	instance *logrus.Logger
}

var (
	loggerInstance *Logger
	once           sync.Once
)

// GetLogger returns the singleton logger instance
func GetLogger() *Logger {
	once.Do(func() {
		logInstance := logrus.New()

		// Attempt to open the log file with restricted permissions
		logFile, err := os.OpenFile("/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			// Fallback to stderr if file cannot be opened
			logInstance.Warn("Failed to open log file, falling back to stderr:", err)
			logInstance.SetOutput(os.Stderr)
		} else {
			// Configure log rotation with Lumberjack
			logInstance.SetOutput(&lumberjack.Logger{
				Filename:   "app.log",
				MaxSize:    10,   // Max size in MB
				MaxBackups: 3,    // Max backup files
				MaxAge:     28,   // Max days to keep old logs
				Compress:   true, // Compress old log files
			})
			defer logFile.Close() // Ensure the file is closed properly
		}

		// Set JSON formatter for structured logging
		logInstance.SetFormatter(&logrus.JSONFormatter{})

		// Set log level
		logInstance.SetLevel(logrus.InfoLevel)

		loggerInstance = &Logger{
			instance: logInstance,
		}
	})
	return loggerInstance
}

// addDefaultFields adds automated fields: timestamp, source, and event_id
func (l *Logger) addDefaultFields(fields logrus.Fields) logrus.Fields {
	if fields == nil {
		fields = logrus.Fields{}
	}
	fields["timestamp"] = time.Now().Format(time.RFC3339) // ISO 8601 format
	fields["source"] = "user-server"
	fields["event_id"] = uuid.NewString() // Generate a new UUID for event_id
	return fields
}

// Info logs an informational message
func (l *Logger) Info(fields logrus.Fields, message string) {
	l.instance.WithFields(l.addDefaultFields(fields)).Info(message)
}

// Error logs an error message
func (l *Logger) Error(fields logrus.Fields, message string) {
	l.instance.WithFields(l.addDefaultFields(fields)).Error(message)
}

// Warn logs a warning message
func (l *Logger) Warn(fields logrus.Fields, message string) {
	l.instance.WithFields(l.addDefaultFields(fields)).Warn(message)
}

// Debug logs a debug message
func (l *Logger) Debug(fields logrus.Fields, message string) {
	l.instance.WithFields(l.addDefaultFields(fields)).Debug(message)
}
