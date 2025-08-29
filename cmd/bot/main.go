// Package main provides the entry point for the Arcanus VPN Telegram bot.
// It initializes the bot with dependency injection using Uber FX and handles
// the application lifecycle including graceful shutdown.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/bot"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/config"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/domain"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/events"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/logger"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/repository"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/service"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewConfig creates a new configuration instance
func NewConfig() (*config.Config, error) {
	loader := config.NewEnvLoader()
	cfg, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// NewLogger creates a new logger instance with Sentry integration
func NewLogger(cfg *config.Config) (logger.Logger, error) {
	// Create logger configuration based on environment
	var loggerConfig logger.Config
	if cfg.IsDevelopment() {
		loggerConfig = logger.DevelopmentConfig()
	} else {
		loggerConfig = logger.ProductionConfig()
	}
	
	// Override with config values
	loggerConfig.Level = cfg.LogLevel
	loggerConfig.Format = cfg.LogFormat
	
	// Create Sentry configuration if DSN is provided
	if cfg.SentryDSN != "" {
		sentryConfig := logger.SentryConfig{
			DSN:              cfg.SentryDSN,
			Environment:      cfg.SentryEnvironment,
			Release:          cfg.SentryRelease,
			SampleRate:       cfg.SentrySampleRate,
			EnableTracing:    cfg.SentryEnableTracing,
			TracesSampleRate: cfg.SentryTracesSampleRate,
		}
		
		return logger.NewLoggerWithSentry(loggerConfig, sentryConfig)
	}
	
	// Create basic logger without Sentry
	factory := logger.NewFactory(logger.ProviderLogrus, loggerConfig)
	return factory.Create()
}

// NewLogrusLogger creates a logrus logger for compatibility
func NewLogrusLogger(appLogger logger.Logger) *logrus.Logger {
	// Extract the underlying logrus logger for components that need it
	if logrusLogger, ok := appLogger.(*logger.LogrusLogger); ok {
		return logrusLogger.GetLogrusLogger()
	}
	// Fallback - create a basic logrus logger
	fallbackLogger := logrus.New()
	fallbackLogger.SetLevel(logrus.InfoLevel)
	fallbackLogger.SetFormatter(&logrus.JSONFormatter{})
	return fallbackLogger
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.Config, appLogger logger.Logger) (*gorm.DB, error) {
	logrusLogger := NewLogrusLogger(appLogger)
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get the underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool based on config
	sqlDB.SetMaxOpenConns(cfg.DatabaseMaxConns)
	sqlDB.SetMaxIdleConns(cfg.DatabaseMaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DatabaseConnMaxLifetime)

	logrusLogger.WithFields(logrus.Fields{
		"max_open_conns": cfg.DatabaseMaxConns,
		"max_idle_conns": cfg.DatabaseMaxIdleConns,
		"conn_max_lifetime": cfg.DatabaseConnMaxLifetime,
	}).Info("Database connection established")
	
	return db, nil
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return repository.NewUserRepository(db)
}

// NewTransactionManager creates a new TransactionManager instance
func NewTransactionManager(db *gorm.DB) domain.TransactionManager {
	return repository.NewTransactionManager(db)
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo domain.UserRepository, txManager domain.TransactionManager, eventService *events.Service) domain.UserService {
	return service.NewUserServiceWithEvents(userRepo, txManager, eventService)
}

// NewBotHandler creates a new bot handler instance
func NewBotHandler(botAPI *tgbotapi.BotAPI, userService domain.UserService, appLogger logger.Logger, eventService *events.Service) *bot.Handler {
	logrusLogger := NewLogrusLogger(appLogger)
	return bot.NewHandlerWithEvents(botAPI, userService, logrusLogger, eventService)
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter() *bot.RateLimiter {
	return bot.NewRateLimiter()
}

// NewAuditLogger creates a new audit logger instance  
func NewAuditLogger(appLogger logger.Logger) *bot.AuditLogger {
	logrusLogger := NewLogrusLogger(appLogger)
	return bot.NewAuditLogger(logrusLogger)
}

// NewEventPublisher creates a new event publisher based on configuration
func NewEventPublisher(cfg *config.Config, appLogger logger.Logger) (events.Publisher, error) {
	logrusLogger := NewLogrusLogger(appLogger)
	if !cfg.KafkaEnabled {
		return events.NewMockPublisher(logrusLogger), nil
	}
	
	kafkaConfig := events.KafkaConfig{
		Brokers:           cfg.KafkaBrokers,
		Topic:             cfg.KafkaTopic,
		SecurityProtocol:  cfg.KafkaSecurityProtocol,
		SaslMechanism:     cfg.KafkaSaslMechanism,
		SaslUsername:      cfg.KafkaSaslUsername,
		SaslPassword:      cfg.KafkaSaslPassword,
		EnableIdempotence: cfg.KafkaEnableIdempotence,
		Acks:              cfg.KafkaAcks,
		RetryBackoffMs:    cfg.KafkaRetryBackoffMs,
		RequestTimeoutMs:  cfg.KafkaRequestTimeoutMs,
	}
	
	return events.NewKafkaPublisher(kafkaConfig, logrusLogger)
}

// NewEventService creates a new event service instance
func NewEventService(publisher events.Publisher, appLogger logger.Logger) *events.Service {
	logrusLogger := NewLogrusLogger(appLogger)
	return events.NewEventService(publisher, logrusLogger)
}

// NewBotHandlerWithMiddleware creates a new middleware-aware bot handler
func NewBotHandlerWithMiddleware(
	botAPI *tgbotapi.BotAPI, 
	userService domain.UserService, 
	appLogger logger.Logger,
	rateLimiter *bot.RateLimiter,
	auditLogger *bot.AuditLogger,
) *bot.HandlerWithMiddleware {
	logrusLogger := NewLogrusLogger(appLogger)
	return bot.NewHandlerWithMiddleware(botAPI, userService, logrusLogger, rateLimiter, auditLogger)
}

// NewTelegramBot creates a new Telegram bot instance
func NewTelegramBot(cfg *config.Config, appLogger logger.Logger) (*tgbotapi.BotAPI, error) {
	logrusLogger := NewLogrusLogger(appLogger)
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}
	
	// Enable debug mode if in development
	bot.Debug = cfg.IsDevelopment() && cfg.Debug
	
	logrusLogger.WithFields(logrus.Fields{
		"debug_mode": bot.Debug,
		"environment": cfg.Environment,
	}).Info("Telegram bot initialized")
	
	return bot, nil
}

// StartBot starts the bot application
func StartBot(
	lifecycle fx.Lifecycle, 
	botAPI *tgbotapi.BotAPI, 
	handler *bot.Handler,
	middlewareHandler *bot.HandlerWithMiddleware, 
	db *gorm.DB, 
	appLogger logger.Logger,
	cfg *config.Config,
) {
	logrusLogger := NewLogrusLogger(appLogger)
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logrusLogger.Info("Starting Arcanus VPN Telegram Bot")

			// Run database migrations
			// Temporarily disabled due to GORM issue
			// if err := db.WithContext(ctx).AutoMigrate(&domain.User{}); err != nil {
			// 	return fmt.Errorf("failed to run database migrations: %w", err)
			// }
			logrusLogger.Info("Database migrations skipped (temporarily disabled)")

			// Get bot info
			botInfo, err := botAPI.GetMe()
			if err != nil {
				return fmt.Errorf("failed to get bot info: %w", err)
			}
			logrusLogger.WithFields(logrus.Fields{
				"bot_name":        botInfo.UserName,
				"bot_id":          botInfo.ID,
				"first_name":      botInfo.FirstName,
				"can_join_groups": botInfo.CanJoinGroups,
			}).Info("Bot info retrieved")

			// Setup update configuration
			updateConfig := tgbotapi.NewUpdate(0)
			updateConfig.Timeout = 60

			// Get updates channel
			updates := botAPI.GetUpdatesChan(updateConfig)

			// Setup signal handling for graceful shutdown
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			logrusLogger.Info("Bot is ready to handle messages")
			logrusLogger.Info("Press Ctrl+C to stop the bot")

			// Start message processing loop
			go func() {
				for {
					select {
					case update := <-updates:
						// Use middleware handler in production, fallback to basic handler
						var err error
						if cfg.IsProduction() {
							err = processUpdateWithMiddleware(ctx, middlewareHandler, update)
						} else {
							err = processUpdate(ctx, handler, update)
						}
						if err != nil {
							logrusLogger.WithError(err).Error("Failed to process update")
						}
					case <-sigChan:
						logrusLogger.Info("Received shutdown signal, stopping bot...")
						botAPI.StopReceivingUpdates()
						return
					case <-ctx.Done():
						logrusLogger.Info("Context cancelled, stopping bot...")
						botAPI.StopReceivingUpdates()
						return
					}
				}
			}()

			return nil
		},
		OnStop: func(context.Context) error {
			logrusLogger.Info("Bot stopped successfully")
			return nil
		},
	})
}

// processUpdate processes a single Telegram update
func processUpdate(ctx context.Context, handler *bot.Handler, update tgbotapi.Update) error {
	// Handle callback queries
	if update.CallbackQuery != nil {
		return handler.HandleCallback(ctx, update.CallbackQuery)
	}

	// Handle messages
	if update.Message != nil {
		return handler.HandleUpdate(ctx, update)
	}

	return nil
}

// processUpdateWithMiddleware processes a single Telegram update using middleware
func processUpdateWithMiddleware(ctx context.Context, handler *bot.HandlerWithMiddleware, update tgbotapi.Update) error {
	// Handle callback queries
	if update.CallbackQuery != nil {
		return handler.HandleCallback(ctx, update.CallbackQuery)
	}

	// Handle messages  
	if update.Message != nil {
		return handler.HandleUpdate(ctx, update)
	}

	return nil
}


func main() {
	app := fx.New(
		fx.Provide(
			NewConfig,
			NewLogger,
			NewDatabase,
			NewUserRepository,
			NewTransactionManager,
			NewEventPublisher,
			NewEventService,
			NewUserService,
			NewRateLimiter,
			NewAuditLogger,
			NewBotHandler,
			NewBotHandlerWithMiddleware,
			NewTelegramBot,
		),
		fx.Invoke(StartBot),
	)

	if err := app.Start(context.Background()); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	<-app.Done()
}
