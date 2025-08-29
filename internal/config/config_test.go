package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEnvLoader_Load(t *testing.T) {
	// Save original environment variables
	originalTelegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	originalDatabaseURL := os.Getenv("DATABASE_URL")
	originalLogLevel := os.Getenv("LOG_LEVEL")
	
	// Cleanup function to restore original values
	cleanup := func() {
		_ = os.Setenv("TELEGRAM_BOT_TOKEN", originalTelegramToken)
		_ = os.Setenv("DATABASE_URL", originalDatabaseURL)
		_ = os.Setenv("LOG_LEVEL", originalLogLevel)
		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("DEBUG")
		_ = os.Unsetenv("LOG_FORMAT")
		_ = os.Unsetenv("ENVIRONMENT")
	}
	defer cleanup()

	t.Run("Load with all environment variables", func(t *testing.T) {
		// Set test environment variables
		_ = os.Setenv("TELEGRAM_BOT_TOKEN", "test_token")
		_ = os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test")
		_ = os.Setenv("LOG_LEVEL", "debug")
		_ = os.Setenv("PORT", "9000")
		_ = os.Setenv("DEBUG", "true")
		_ = os.Setenv("LOG_FORMAT", "text")
		_ = os.Setenv("ENVIRONMENT", "staging")

		loader := NewEnvLoader()
		config, err := loader.Load()

		assert.NoError(t, err)
		assert.Equal(t, "test_token", config.TelegramToken)
		assert.Equal(t, "postgres://test:test@localhost/test", config.DatabaseURL)
		assert.Equal(t, "debug", config.LogLevel)
		assert.Equal(t, 9000, config.Port)
		assert.True(t, config.Debug)
		assert.Equal(t, "text", config.LogFormat)
		assert.Equal(t, "staging", config.Environment)
	})

	t.Run("Load with defaults", func(t *testing.T) {
		// Clear all optional environment variables
		_ = os.Unsetenv("LOG_LEVEL")
		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("DEBUG")
		_ = os.Unsetenv("LOG_FORMAT")
		_ = os.Unsetenv("ENVIRONMENT")

		loader := NewEnvLoader()
		config, err := loader.Load()

		assert.NoError(t, err)
		assert.Equal(t, "info", config.LogLevel)
		assert.Equal(t, 8080, config.Port)
		assert.False(t, config.Debug)
		assert.Equal(t, "json", config.LogFormat)
		assert.Equal(t, "development", config.Environment)
		assert.Equal(t, 25, config.DatabaseMaxConns)
		assert.Equal(t, 5, config.DatabaseMaxIdleConns)
		assert.Equal(t, 5*time.Minute, config.DatabaseConnMaxLifetime)
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("Valid configuration", func(t *testing.T) {
		config := &Config{
			TelegramToken:           "test_token",
			DatabaseURL:             "postgres://test:test@localhost/test",
			LogLevel:               "info",
			LogFormat:              "json",
			Environment:            "development",
			Port:                   8080,
			DatabaseMaxConns:       25,
			DatabaseMaxIdleConns:   5,
		}

		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("Missing telegram token", func(t *testing.T) {
		config := &Config{
			DatabaseURL: "postgres://test:test@localhost/test",
			LogLevel:   "info",
			LogFormat:  "json",
			Environment: "development",
			Port:       8080,
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TELEGRAM_BOT_TOKEN is required")
	})

	t.Run("Missing database URL", func(t *testing.T) {
		config := &Config{
			TelegramToken: "test_token",
			LogLevel:     "info",
			LogFormat:    "json",
			Environment:  "development",
			Port:         8080,
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DATABASE_URL is required")
	})

	t.Run("Invalid log level", func(t *testing.T) {
		config := &Config{
			TelegramToken: "test_token",
			DatabaseURL:   "postgres://test:test@localhost/test",
			LogLevel:     "invalid",
			LogFormat:    "json",
			Environment:  "development",
			Port:         8080,
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid log level")
	})

	t.Run("Invalid log format", func(t *testing.T) {
		config := &Config{
			TelegramToken: "test_token",
			DatabaseURL:   "postgres://test:test@localhost/test",
			LogLevel:     "info",
			LogFormat:    "invalid",
			Environment:  "development",
			Port:         8080,
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid log format")
	})

	t.Run("Invalid environment", func(t *testing.T) {
		config := &Config{
			TelegramToken: "test_token",
			DatabaseURL:   "postgres://test:test@localhost/test",
			LogLevel:     "info",
			LogFormat:    "json",
			Environment:  "invalid",
			Port:         8080,
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid environment")
	})

	t.Run("Invalid port", func(t *testing.T) {
		config := &Config{
			TelegramToken: "test_token",
			DatabaseURL:   "postgres://test:test@localhost/test",
			LogLevel:     "info",
			LogFormat:    "json",
			Environment:  "development",
			Port:         70000,
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid port")
	})

	t.Run("Invalid database connections", func(t *testing.T) {
		config := &Config{
			TelegramToken:        "test_token",
			DatabaseURL:          "postgres://test:test@localhost/test",
			LogLevel:            "info",
			LogFormat:           "json",
			Environment:         "development",
			Port:                8080,
			DatabaseMaxConns:    0,
			DatabaseMaxIdleConns: 5,
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database max connections must be at least 1")
	})
}

func TestConfig_HelperMethods(t *testing.T) {
	t.Run("IsProduction", func(t *testing.T) {
		config := &Config{Environment: "production"}
		assert.True(t, config.IsProduction())
		
		config.Environment = "development"
		assert.False(t, config.IsProduction())
	})

	t.Run("IsDevelopment", func(t *testing.T) {
		config := &Config{Environment: "development"}
		assert.True(t, config.IsDevelopment())
		
		config.Environment = "production"
		assert.False(t, config.IsDevelopment())
	})

	t.Run("GetServerAddr", func(t *testing.T) {
		config := &Config{Port: 9000}
		assert.Equal(t, ":9000", config.GetServerAddr())
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("getEnvOrDefault", func(t *testing.T) {
		// Set a test environment variable
		_ = os.Setenv("TEST_VAR", "test_value")
		defer func() { _ = os.Unsetenv("TEST_VAR") }()
		
		result := getEnvOrDefault("TEST_VAR", "default")
		assert.Equal(t, "test_value", result)
		
		result = getEnvOrDefault("NON_EXISTENT_VAR", "default")
		assert.Equal(t, "default", result)
	})

	t.Run("getEnvAsIntOrDefault", func(t *testing.T) {
		_ = os.Setenv("TEST_INT", "42")
		defer func() { _ = os.Unsetenv("TEST_INT") }()
		
		result := getEnvAsIntOrDefault("TEST_INT", 10)
		assert.Equal(t, 42, result)
		
		result = getEnvAsIntOrDefault("NON_EXISTENT_INT", 10)
		assert.Equal(t, 10, result)
		
		// Test invalid integer
		_ = os.Setenv("INVALID_INT", "not_a_number")
		defer func() { _ = os.Unsetenv("INVALID_INT") }()
		
		result = getEnvAsIntOrDefault("INVALID_INT", 10)
		assert.Equal(t, 10, result)
	})

	t.Run("getEnvAsBoolOrDefault", func(t *testing.T) {
		_ = os.Setenv("TEST_BOOL", "true")
		defer func() { _ = os.Unsetenv("TEST_BOOL") }()
		
		result := getEnvAsBoolOrDefault("TEST_BOOL", false)
		assert.True(t, result)
		
		result = getEnvAsBoolOrDefault("NON_EXISTENT_BOOL", false)
		assert.False(t, result)
	})

	t.Run("getEnvAsDurationOrDefault", func(t *testing.T) {
		_ = os.Setenv("TEST_DURATION", "10s")
		defer func() { _ = os.Unsetenv("TEST_DURATION") }()
		
		result := getEnvAsDurationOrDefault("TEST_DURATION", 5*time.Second)
		assert.Equal(t, 10*time.Second, result)
		
		result = getEnvAsDurationOrDefault("NON_EXISTENT_DURATION", 5*time.Second)
		assert.Equal(t, 5*time.Second, result)
	})
}