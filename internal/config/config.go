package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Telegram Bot configuration
	TelegramToken string
	
	// Database configuration
	DatabaseURL        string
	DatabaseMaxConns   int
	DatabaseMaxIdleConns int
	DatabaseConnMaxLifetime time.Duration
	
	// Kafka configuration
	KafkaBrokers           string
	KafkaTopic             string
	KafkaSecurityProtocol  string
	KafkaSaslMechanism     string
	KafkaSaslUsername      string
	KafkaSaslPassword      string
	KafkaEnableIdempotence bool
	KafkaAcks              string
	KafkaRetryBackoffMs    int
	KafkaRequestTimeoutMs  int
	KafkaEnabled           bool
	
	// Logging configuration
	LogLevel string
	LogFormat string // json or text
	
	// Sentry configuration
	SentryDSN              string
	SentryEnvironment      string
	SentryRelease          string
	SentrySampleRate       float64
	SentryEnableTracing    bool
	SentryTracesSampleRate float64
	
	// Server configuration
	Port    int
	Timeout time.Duration
	
	// Application settings
	Environment string // development, staging, production
	Debug       bool
}

// Validator interface for configuration validation
type Validator interface {
	Validate() error
}

// Loader interface for configuration loading
type Loader interface {
	Load() (*Config, error)
}

// EnvLoader loads configuration from environment variables
type EnvLoader struct{}

// NewEnvLoader creates a new environment loader
func NewEnvLoader() *EnvLoader {
	return &EnvLoader{}
}

// Load loads configuration from environment variables
func (e *EnvLoader) Load() (*Config, error) {
	// Try to load .env file (ignore errors as it's optional)
	_ = godotenv.Load()
	
	config := &Config{
		// Required fields
		TelegramToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		
		// Optional fields with defaults
		LogLevel:        getEnvOrDefault("LOG_LEVEL", "info"),
		LogFormat:       getEnvOrDefault("LOG_FORMAT", "json"),
		Environment:     getEnvOrDefault("ENVIRONMENT", "development"),
		Port:           getEnvAsIntOrDefault("PORT", 8080),
		Debug:          getEnvAsBoolOrDefault("DEBUG", false),
		Timeout:        getEnvAsDurationOrDefault("TIMEOUT", 30*time.Second),
		
		// Sentry configuration
		SentryDSN:              getEnvOrDefault("SENTRY_DSN", ""),
		SentryEnvironment:      getEnvOrDefault("SENTRY_ENVIRONMENT", ""),
		SentryRelease:          getEnvOrDefault("SENTRY_RELEASE", ""),
		SentrySampleRate:       getEnvAsFloatOrDefault("SENTRY_SAMPLE_RATE", 1.0),
		SentryEnableTracing:    getEnvAsBoolOrDefault("SENTRY_ENABLE_TRACING", false),
		SentryTracesSampleRate: getEnvAsFloatOrDefault("SENTRY_TRACES_SAMPLE_RATE", 0.1),
		
		// Database connection settings
		DatabaseMaxConns:        getEnvAsIntOrDefault("DB_MAX_CONNS", 25),
		DatabaseMaxIdleConns:    getEnvAsIntOrDefault("DB_MAX_IDLE_CONNS", 5),
		DatabaseConnMaxLifetime: getEnvAsDurationOrDefault("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		
		// Kafka configuration
		KafkaBrokers:           getEnvOrDefault("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopic:             getEnvOrDefault("KAFKA_TOPIC", "arcanus-events"),
		KafkaSecurityProtocol:  getEnvOrDefault("KAFKA_SECURITY_PROTOCOL", ""),
		KafkaSaslMechanism:     getEnvOrDefault("KAFKA_SASL_MECHANISM", ""),
		KafkaSaslUsername:      getEnvOrDefault("KAFKA_SASL_USERNAME", ""),
		KafkaSaslPassword:      getEnvOrDefault("KAFKA_SASL_PASSWORD", ""),
		KafkaEnableIdempotence: getEnvAsBoolOrDefault("KAFKA_ENABLE_IDEMPOTENCE", true),
		KafkaAcks:              getEnvOrDefault("KAFKA_ACKS", "all"),
		KafkaRetryBackoffMs:    getEnvAsIntOrDefault("KAFKA_RETRY_BACKOFF_MS", 100),
		KafkaRequestTimeoutMs:  getEnvAsIntOrDefault("KAFKA_REQUEST_TIMEOUT_MS", 30000),
		KafkaEnabled:           getEnvAsBoolOrDefault("KAFKA_ENABLED", true),
	}
	
	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.TelegramToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	
	// Validate log level
	validLogLevels := map[string]bool{
		"panic": true, "fatal": true, "error": true,
		"warn": true, "info": true, "debug": true, "trace": true,
	}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}
	
	// Validate log format
	if c.LogFormat != "json" && c.LogFormat != "text" {
		return fmt.Errorf("invalid log format: %s, must be 'json' or 'text'", c.LogFormat)
	}
	
	// Validate environment
	validEnvs := map[string]bool{
		"development": true, "staging": true, "production": true,
	}
	if !validEnvs[c.Environment] {
		return fmt.Errorf("invalid environment: %s", c.Environment)
	}
	
	// Validate port range
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d, must be between 1-65535", c.Port)
	}
	
	// Validate database connection settings
	if c.DatabaseMaxConns < 1 {
		return fmt.Errorf("database max connections must be at least 1")
	}
	if c.DatabaseMaxIdleConns < 0 || c.DatabaseMaxIdleConns > c.DatabaseMaxConns {
		return fmt.Errorf("database max idle connections must be between 0 and max connections")
	}
	
	// Validate Kafka settings if enabled
	if c.KafkaEnabled {
		if c.KafkaBrokers == "" {
			return fmt.Errorf("KAFKA_BROKERS is required when Kafka is enabled")
		}
		if c.KafkaTopic == "" {
			return fmt.Errorf("KAFKA_TOPIC is required when Kafka is enabled")
		}
		if c.KafkaAcks != "all" && c.KafkaAcks != "1" && c.KafkaAcks != "0" {
			return fmt.Errorf("invalid KAFKA_ACKS value: %s, must be 'all', '1', or '0'", c.KafkaAcks)
		}
	}
	
	return nil
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// GetServerAddr returns the server address string
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf(":%d", c.Port)
}

// Helper functions for environment variable parsing
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvAsDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsFloatOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}