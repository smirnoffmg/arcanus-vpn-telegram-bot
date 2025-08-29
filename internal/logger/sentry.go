package logger

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// SentryConfig holds Sentry-specific configuration
type SentryConfig struct {
	DSN              string
	Environment      string
	Release          string
	SampleRate       float64
	EnableTracing    bool
	TracesSampleRate float64
}

// SentryHook implements a logrus hook for Sentry integration
// This is a placeholder implementation - in a real scenario you would use
// github.com/getsentry/sentry-go and github.com/getsentry/sentry-go/logrus
type SentryHook struct {
	config SentryConfig
	levels []logrus.Level
}

// NewSentryHook creates a new Sentry hook
func NewSentryHook(config SentryConfig) *SentryHook {
	return &SentryHook{
		config: config,
		levels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		},
	}
}

// Levels returns the levels this hook fires on
func (h *SentryHook) Levels() []logrus.Level {
	return h.levels
}

// Fire is called when a log entry is fired
func (h *SentryHook) Fire(entry *logrus.Entry) error {
	// Placeholder implementation
	// In a real implementation, you would:
	// 1. Convert logrus entry to Sentry event
	// 2. Add context, user info, tags
	// 3. Send to Sentry using sentry.CaptureEvent()
	
	// For now, just add a marker that Sentry would be called
	entry.Data["sentry_enabled"] = true
	entry.Data["sentry_dsn_configured"] = h.config.DSN != ""
	
	return nil
}

// WithSentryHook adds Sentry integration to a logrus logger
func WithSentryHook(logger *logrus.Logger, config SentryConfig) error {
	if config.DSN == "" {
		// Sentry not configured, skip
		return nil
	}
	
	hook := NewSentryHook(config)
	logger.AddHook(hook)
	
	return nil
}

// Example of how to create a logger with Sentry integration
func NewLoggerWithSentry(loggerConfig Config, sentryConfig SentryConfig) (Logger, error) {
	logrusLogger, err := NewLogrusLogger(loggerConfig)
	if err != nil {
		return nil, err
	}
	
	// Add Sentry hook if configured
	if err := WithSentryHook(logrusLogger.GetLogrusLogger(), sentryConfig); err != nil {
		return nil, err
	}
	
	return logrusLogger, nil
}

// SentryContextLogger wraps the logger with Sentry context
type SentryContextLogger struct {
	Logger
	sentryConfig SentryConfig
}

// WithSentryContext wraps a logger to add Sentry context to all log entries
func WithSentryContext(logger Logger, config SentryConfig) Logger {
	return &SentryContextLogger{
		Logger:       logger,
		sentryConfig: config,
	}
}

// WithContext adds Sentry-specific context fields
func (s *SentryContextLogger) WithContext(ctx context.Context) Logger {
	// In a real implementation, you would extract Sentry context from ctx
	enrichedLogger := s.Logger.WithContext(ctx)
	
	// Add Sentry-specific fields
	enrichedLogger = enrichedLogger.WithFields(map[string]interface{}{
		"sentry.environment": s.sentryConfig.Environment,
		"sentry.release":     s.sentryConfig.Release,
		"timestamp":          time.Now().UTC(),
	})
	
	return &SentryContextLogger{
		Logger:       enrichedLogger,
		sentryConfig: s.sentryConfig,
	}
}