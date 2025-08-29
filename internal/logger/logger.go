package logger

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"
)

// Logger defines the interface for structured logging
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
	
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
	WithContext(ctx context.Context) Logger
}

// Config holds logger configuration
type Config struct {
	Level      string // debug, info, warn, error, fatal, panic
	Format     string // json, text
	Output     io.Writer
	AddSource  bool
	ReportCaller bool
}

// LogrusLogger implements Logger interface using logrus
type LogrusLogger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

// NewLogrusLogger creates a new logrus-based logger
func NewLogrusLogger(config Config) (*LogrusLogger, error) {
	logger := logrus.New()
	
	// Set level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return nil, err
	}
	logger.SetLevel(level)
	
	// Set formatter
	switch config.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "@timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "caller",
			},
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	default:
		logger.SetFormatter(&logrus.JSONFormatter{})
	}
	
	// Set output
	if config.Output != nil {
		logger.SetOutput(config.Output)
	}
	
	// Set caller reporting
	logger.SetReportCaller(config.ReportCaller)
	
	return &LogrusLogger{
		logger: logger,
		entry:  logger.WithFields(logrus.Fields{}),
	}, nil
}

// Debug logs a debug message
func (l *LogrusLogger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

// Info logs an info message
func (l *LogrusLogger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

// Warn logs a warning message
func (l *LogrusLogger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

// Error logs an error message
func (l *LogrusLogger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

// Fatal logs a fatal message and exits
func (l *LogrusLogger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

// Panic logs a panic message and panics
func (l *LogrusLogger) Panic(args ...interface{}) {
	l.entry.Panic(args...)
}

// Debugf logs a formatted debug message
func (l *LogrusLogger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

// Infof logs a formatted info message
func (l *LogrusLogger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

// Warnf logs a formatted warning message
func (l *LogrusLogger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

// Errorf logs a formatted error message
func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

// Fatalf logs a formatted fatal message and exits
func (l *LogrusLogger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

// Panicf logs a formatted panic message and panics
func (l *LogrusLogger) Panicf(format string, args ...interface{}) {
	l.entry.Panicf(format, args...)
}

// WithField adds a single field to the logger
func (l *LogrusLogger) WithField(key string, value interface{}) Logger {
	return &LogrusLogger{
		logger: l.logger,
		entry:  l.entry.WithField(key, value),
	}
}

// WithFields adds multiple fields to the logger
func (l *LogrusLogger) WithFields(fields map[string]interface{}) Logger {
	logrusFields := make(logrus.Fields)
	for k, v := range fields {
		logrusFields[k] = v
	}
	return &LogrusLogger{
		logger: l.logger,
		entry:  l.entry.WithFields(logrusFields),
	}
}

// WithError adds an error field to the logger
func (l *LogrusLogger) WithError(err error) Logger {
	return &LogrusLogger{
		logger: l.logger,
		entry:  l.entry.WithError(err),
	}
}

// WithContext adds context to the logger
func (l *LogrusLogger) WithContext(ctx context.Context) Logger {
	return &LogrusLogger{
		logger: l.logger,
		entry:  l.entry.WithContext(ctx),
	}
}

// GetLogrusLogger returns the underlying logrus logger for compatibility
func (l *LogrusLogger) GetLogrusLogger() *logrus.Logger {
	return l.logger
}