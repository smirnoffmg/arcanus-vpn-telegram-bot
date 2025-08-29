package logger

import (
	"fmt"
	"os"
)

// Provider represents different logging providers
type Provider string

const (
	ProviderLogrus Provider = "logrus"
	// Future providers can be added here:
	// ProviderSentry Provider = "sentry"
	// ProviderDatadog Provider = "datadog"
)

// Factory creates loggers based on configuration
type Factory struct {
	provider Provider
	config   Config
}

// NewFactory creates a new logger factory
func NewFactory(provider Provider, config Config) *Factory {
	return &Factory{
		provider: provider,
		config:   config,
	}
}

// Create creates a logger instance based on the configured provider
func (f *Factory) Create() (Logger, error) {
	switch f.provider {
	case ProviderLogrus:
		return NewLogrusLogger(f.config)
	default:
		return nil, fmt.Errorf("unsupported logger provider: %s", f.provider)
	}
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() Config {
	return Config{
		Level:        "info",
		Format:       "json",
		Output:       os.Stdout,
		AddSource:    false,
		ReportCaller: false,
	}
}

// DevelopmentConfig returns a configuration suitable for development
func DevelopmentConfig() Config {
	return Config{
		Level:        "debug",
		Format:       "text",
		Output:       os.Stdout,
		AddSource:    true,
		ReportCaller: true,
	}
}

// ProductionConfig returns a configuration suitable for production
func ProductionConfig() Config {
	return Config{
		Level:        "info",
		Format:       "json",
		Output:       os.Stdout,
		AddSource:    false,
		ReportCaller: false,
	}
}