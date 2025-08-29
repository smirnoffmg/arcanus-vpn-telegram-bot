package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogrusLogger(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "valid json config",
			config: Config{
				Level:  "info",
				Format: "json",
			},
			expectError: false,
		},
		{
			name: "valid text config",
			config: Config{
				Level:  "debug",
				Format: "text",
			},
			expectError: false,
		},
		{
			name: "invalid level",
			config: Config{
				Level:  "invalid",
				Format: "json",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogrusLogger(tt.config)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			}
		})
	}
}

func TestLogrusLoggerMethods(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:  "debug",
		Format: "json",
		Output: &buf,
	}
	
	logger, err := NewLogrusLogger(config)
	require.NoError(t, err)
	
	// Test different log methods
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")
	
	// Test formatted methods
	logger.Debugf("debug %s", "formatted")
	logger.Infof("info %d", 42)
	logger.Warnf("warn %v", true)
	logger.Errorf("error %s", "formatted")
	
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	// Should have 8 log lines
	assert.Len(t, lines, 8)
	
	// Verify JSON format
	for _, line := range lines {
		var logEntry map[string]interface{}
		err := json.Unmarshal([]byte(line), &logEntry)
		assert.NoError(t, err, "Each line should be valid JSON")
		
		// Check required fields
		assert.Contains(t, logEntry, "@timestamp")
		assert.Contains(t, logEntry, "level")
		assert.Contains(t, logEntry, "message")
	}
}

func TestLogrusLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:  "info",
		Format: "json",
		Output: &buf,
	}
	
	logger, err := NewLogrusLogger(config)
	require.NoError(t, err)
	
	// Test WithField
	fieldLogger := logger.WithField("user_id", 12345)
	fieldLogger.Info("test message")
	
	// Test WithFields
	fieldsLogger := logger.WithFields(map[string]interface{}{
		"user_id": 67890,
		"action":  "login",
	})
	fieldsLogger.Info("another test message")
	
	// Test WithError
	testErr := errors.New("test error")
	errorLogger := logger.WithError(testErr)
	errorLogger.Error("error occurred")
	
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Len(t, lines, 3)
	
	// Verify first log entry has user_id field
	var firstEntry map[string]interface{}
	err = json.Unmarshal([]byte(lines[0]), &firstEntry)
	require.NoError(t, err)
	assert.Equal(t, float64(12345), firstEntry["user_id"])
	
	// Verify second log entry has multiple fields
	var secondEntry map[string]interface{}
	err = json.Unmarshal([]byte(lines[1]), &secondEntry)
	require.NoError(t, err)
	assert.Equal(t, float64(67890), secondEntry["user_id"])
	assert.Equal(t, "login", secondEntry["action"])
	
	// Verify third log entry has error field
	var thirdEntry map[string]interface{}
	err = json.Unmarshal([]byte(lines[2]), &thirdEntry)
	require.NoError(t, err)
	assert.Equal(t, "test error", thirdEntry["error"])
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

func TestLogrusLoggerWithContext(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:  "info",
		Format: "json",
		Output: &buf,
	}
	
	logger, err := NewLogrusLogger(config)
	require.NoError(t, err)
	
	ctx := context.WithValue(context.Background(), contextKey("request_id"), "test-123")
	contextLogger := logger.WithContext(ctx)
	contextLogger.Info("test with context")
	
	output := buf.String()
	assert.Contains(t, output, "test with context")
}

func TestLoggerFactory(t *testing.T) {
	tests := []struct {
		name        string
		provider    Provider
		config      Config
		expectError bool
	}{
		{
			name:     "logrus provider",
			provider: ProviderLogrus,
			config: Config{
				Level:  "info",
				Format: "json",
			},
			expectError: false,
		},
		{
			name:     "unsupported provider",
			provider: Provider("unsupported"),
			config:   DefaultConfig(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewFactory(tt.provider, tt.config)
			logger, err := factory.Create()
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			}
		})
	}
}

func TestDefaultConfigs(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultConfig()
		assert.Equal(t, "info", config.Level)
		assert.Equal(t, "json", config.Format)
		assert.NotNil(t, config.Output)
	})
	
	t.Run("DevelopmentConfig", func(t *testing.T) {
		config := DevelopmentConfig()
		assert.Equal(t, "debug", config.Level)
		assert.Equal(t, "text", config.Format)
		assert.True(t, config.AddSource)
		assert.True(t, config.ReportCaller)
	})
	
	t.Run("ProductionConfig", func(t *testing.T) {
		config := ProductionConfig()
		assert.Equal(t, "info", config.Level)
		assert.Equal(t, "json", config.Format)
		assert.False(t, config.AddSource)
		assert.False(t, config.ReportCaller)
	})
}

func TestSentryHook(t *testing.T) {
	config := SentryConfig{
		DSN:         "https://test@sentry.io/123456",
		Environment: "test",
		Release:     "1.0.0",
	}
	
	hook := NewSentryHook(config)
	assert.NotNil(t, hook)
	
	levels := hook.Levels()
	assert.Contains(t, levels, logrus.ErrorLevel)
	assert.Contains(t, levels, logrus.WarnLevel)
}

func TestSentryContextLogger(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:  "info",
		Format: "json",
		Output: &buf,
	}
	
	baseLogger, err := NewLogrusLogger(config)
	require.NoError(t, err)
	
	sentryConfig := SentryConfig{
		Environment: "test",
		Release:     "1.0.0",
	}
	
	sentryLogger := WithSentryContext(baseLogger, sentryConfig)
	ctx := context.Background()
	contextLogger := sentryLogger.WithContext(ctx)
	contextLogger.Info("test message")
	
	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "sentry.environment")
	assert.Contains(t, output, "sentry.release")
}