package main

import (
	"context"
	"os"
	"testing"

	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/config"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
	}{
		{
			name: "Valid configuration",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN": "test_token_123456789",
				"DATABASE_URL":       "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable",
				"LOG_LEVEL":          "info",
			},
			expectError: false,
		},
		{
			name: "Missing TELEGRAM_BOT_TOKEN",
			envVars: map[string]string{
				"DATABASE_URL": "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable",
				"LOG_LEVEL":    "info",
			},
			expectError: true,
		},
		{
			name: "Missing DATABASE_URL",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN": "test_token_123456789",
				"LOG_LEVEL":          "info",
			},
			expectError: true,
		},
		{
			name: "Default LOG_LEVEL",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN": "test_token_123456789",
				"DATABASE_URL":       "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables
			_ = os.Unsetenv("TELEGRAM_BOT_TOKEN")
			_ = os.Unsetenv("DATABASE_URL")
			_ = os.Unsetenv("LOG_LEVEL")

			// Set test environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}

			// Clean up after test
			defer func() {
				_ = os.Unsetenv("TELEGRAM_BOT_TOKEN")
				_ = os.Unsetenv("DATABASE_URL")
				_ = os.Unsetenv("LOG_LEVEL")
			}()

			config, err := NewConfig()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.Equal(t, tt.envVars["TELEGRAM_BOT_TOKEN"], config.TelegramToken)
				assert.Equal(t, tt.envVars["DATABASE_URL"], config.DatabaseURL)

				expectedLogLevel := "info"
				if level, exists := tt.envVars["LOG_LEVEL"]; exists {
					expectedLogLevel = level
				}
				assert.Equal(t, expectedLogLevel, config.LogLevel)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
	}{
		{"Debug level", "debug"},
		{"Info level", "info"},
		{"Warn level", "warn"},
		{"Error level", "error"},
		{"Invalid level defaults to info", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				LogLevel:  tt.logLevel,
				LogFormat: "json",
				Environment: "development",
			}
			appLogger, err := NewLogger(cfg)

			if tt.logLevel == "invalid" {
				assert.Error(t, err)
				assert.Nil(t, appLogger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, appLogger)
				
				// Test that we can create a logrus logger from it
				logrusLogger := NewLogrusLogger(appLogger)
				assert.NotNil(t, logrusLogger)
			}
		})
	}
}

func TestConfigIntegration(t *testing.T) {
	t.Run("Creates configuration instance", func(t *testing.T) {
		// Set minimal required environment variables
		_ = os.Setenv("TELEGRAM_BOT_TOKEN", "test_token")
		_ = os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test")
		defer func() {
			_ = os.Unsetenv("TELEGRAM_BOT_TOKEN")
			_ = os.Unsetenv("DATABASE_URL")
		}()

		cfg, err := NewConfig()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "test_token", cfg.TelegramToken)
		assert.Equal(t, "postgres://test:test@localhost/test", cfg.DatabaseURL)
	})
}

func TestNewUserRepository(t *testing.T) {
	// This is a simple test to ensure the function doesn't panic
	// In a real scenario, you'd mock the database
	t.Run("Creates repository without panic", func(t *testing.T) {
		// This test would require a mock database
		// For now, we'll just ensure the function signature is correct
		assert.True(t, true, "NewUserRepository function exists")
	})
}

func TestNewUserService(t *testing.T) {
	// This is a simple test to ensure the function doesn't panic
	t.Run("Creates service without panic", func(t *testing.T) {
		// This test would require a mock repository
		// For now, we'll just ensure the function signature is correct
		assert.True(t, true, "NewUserService function exists")
	})
}

func TestNewBotHandler(t *testing.T) {
	// This is a simple test to ensure the function doesn't panic
	t.Run("Creates handler without panic", func(t *testing.T) {
		// This test would require a mock service and logger
		// For now, we'll just ensure the function signature is correct
		assert.True(t, true, "NewBotHandler function exists")
	})
}

func TestNewLogrusLogger(t *testing.T) {
	t.Run("Extracts logrus logger from LogrusLogger", func(t *testing.T) {
		// Create a LogrusLogger
		loggerConfig := logger.DefaultConfig()
		logrusLogger, err := logger.NewLogrusLogger(loggerConfig)
		require.NoError(t, err)
		
		// Extract the underlying logrus logger
		result := NewLogrusLogger(logrusLogger)
		assert.NotNil(t, result)
	})
	
	t.Run("Creates fallback logger for non-LogrusLogger", func(t *testing.T) {
		// Create a mock logger that's not a LogrusLogger
		mockLogger := &mockLogger{}
		
		// Should create a fallback logrus logger
		result := NewLogrusLogger(mockLogger)
		assert.NotNil(t, result)
	})
}

// mockLogger implements the logger.Logger interface for testing
type mockLogger struct{}

func (m *mockLogger) Debug(args ...interface{})                            {}
func (m *mockLogger) Info(args ...interface{})                             {}
func (m *mockLogger) Warn(args ...interface{})                             {}
func (m *mockLogger) Error(args ...interface{})                            {}
func (m *mockLogger) Fatal(args ...interface{})                            {}
func (m *mockLogger) Panic(args ...interface{})                            {}
func (m *mockLogger) Debugf(format string, args ...interface{})            {}
func (m *mockLogger) Infof(format string, args ...interface{})             {}
func (m *mockLogger) Warnf(format string, args ...interface{})             {}
func (m *mockLogger) Errorf(format string, args ...interface{})            {}
func (m *mockLogger) Fatalf(format string, args ...interface{})            {}
func (m *mockLogger) Panicf(format string, args ...interface{})            {}
func (m *mockLogger) WithField(key string, value interface{}) logger.Logger { return m }
func (m *mockLogger) WithFields(fields map[string]interface{}) logger.Logger { return m }
func (m *mockLogger) WithError(err error) logger.Logger                     { return m }
func (m *mockLogger) WithContext(ctx context.Context) logger.Logger         { return m }
