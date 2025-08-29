// Package service implements business logic for the Arcanus VPN bot.
// It provides user management, trial activation, and quota tracking services
// following clean architecture principles with proper error handling and event publishing.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/domain"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/events"
)

// UserService implements domain.UserService
type UserService struct {
	userRepo     domain.UserRepository
	txManager    domain.TransactionManager
	eventService *events.Service
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo domain.UserRepository) domain.UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// NewUserServiceWithEvents creates a new UserService instance with event publishing
func NewUserServiceWithEvents(userRepo domain.UserRepository, txManager domain.TransactionManager, eventService *events.Service) domain.UserService {
	return &UserService{
		userRepo:     userRepo,
		txManager:    txManager,
		eventService: eventService,
	}
}

// NewUserServiceWithTx creates a new UserService instance with transaction support
func NewUserServiceWithTx(userRepo domain.UserRepository, txManager domain.TransactionManager) domain.UserService {
	return &UserService{
		userRepo:  userRepo,
		txManager: txManager,
	}
}

// RegisterUser registers a new user or returns existing user
func (s *UserService) RegisterUser(ctx context.Context, telegramID int64, username, firstName, lastName string) (*domain.User, error) {
	// Validate input
	if telegramID <= 0 {
		return nil, domain.ErrInvalidInput
	}
	if firstName == "" {
		return nil, domain.ErrInvalidInput
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetByTelegramID(ctx, telegramID)
	if err == nil {
		// User exists, return it
		return existingUser, nil
	}

	// Check if it's a "not found" error
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Create new user
	user := domain.NewUser(telegramID, username, firstName, lastName)

	// Validate the created user
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("invalid user data: %w", err)
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Publish user registration event
	if s.eventService != nil {
		if err := s.eventService.PublishUserRegistered(ctx, user.TelegramID, user.Username, user.FirstName, user.LastName, user.QuotaLimit); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to publish user registered event: %v\n", err)
		}
	}

	return user, nil
}

// GetUser retrieves a user by their Telegram ID
func (s *UserService) GetUser(ctx context.Context, telegramID int64) (*domain.User, error) {
	// Validate input
	if telegramID <= 0 {
		return nil, domain.ErrInvalidInput
	}

	user, err := s.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// ActivateTrial activates the trial for a user
func (s *UserService) ActivateTrial(ctx context.Context, telegramID int64) error {
	// Validate input
	if telegramID <= 0 {
		return domain.ErrInvalidInput
	}

	user, err := s.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("failed to get user for trial activation: %w", err)
	}

	// Check if user can activate trial
	if !user.CanActivateTrial() {
		return domain.ErrUserAlreadyActive
	}

	// Store previous status for event
	previousStatus := user.Status
	
	// Activate trial
	user.ActivateTrial()

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user for trial activation: %w", err)
	}

	// Publish trial activation event
	if s.eventService != nil {
		if err := s.eventService.PublishUserTrialActivated(ctx, user.TelegramID, previousStatus, user.Status); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to publish user trial activated event: %v\n", err)
		}
	}

	return nil
}

// UpdateQuota updates the quota usage for a user
func (s *UserService) UpdateQuota(ctx context.Context, telegramID int64, quotaUsed int64) error {
	// Validate input
	if telegramID <= 0 {
		return domain.ErrInvalidInput
	}
	if quotaUsed < 0 {
		return domain.ErrInvalidInput
	}

	user, err := s.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("failed to get user for quota update: %w", err)
	}

	// Check if user is active
	if !user.IsActive() {
		return domain.ErrUserNotActive
	}

	// Check if user has sufficient quota
	if !user.CanUseQuota(quotaUsed) {
		return domain.QuotaExceededError{Used: quotaUsed, Limit: user.QuotaLimit}
	}

	// Store previous quota for event
	previousQuota := user.QuotaUsed

	err = s.userRepo.UpdateQuota(ctx, telegramID, quotaUsed)
	if err != nil {
		return fmt.Errorf("failed to update quota: %w", err)
	}

	// Publish quota update event
	if s.eventService != nil {
		if err := s.eventService.PublishUserQuotaUpdated(ctx, user.TelegramID, previousQuota, quotaUsed); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to publish user quota updated event: %v\n", err)
		}
	}

	return nil
}
