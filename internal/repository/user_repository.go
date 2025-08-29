package repository

import (
	"context"
	"fmt"

	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/domain"
	"gorm.io/gorm"
)

// UserRepository implements domain.UserRepository for PostgreSQL using GORM
type UserRepository struct {
	db *gorm.DB
}

// Transaction wraps a GORM database transaction
type Transaction struct {
	tx *gorm.DB
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	return t.tx.Commit().Error
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	return t.tx.Rollback().Error
}

// TransactionManager implements domain.TransactionManager
type TransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *gorm.DB) domain.TransactionManager {
	return &TransactionManager{db: db}
}

// WithTransaction executes a function within a database transaction
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context, tx domain.Transaction) error) error {
	return tm.db.Transaction(func(tx *gorm.DB) error {
		transaction := &Transaction{tx: tx}
		return fn(ctx, transaction)
	})
}

// BeginTx starts a new transaction and returns a transactional repository
func (tm *TransactionManager) BeginTx(ctx context.Context) (domain.UserRepository, domain.Transaction, error) {
	tx := tm.db.Begin()
	if tx.Error != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	
	txRepo := &UserRepository{db: tx}
	transaction := &Transaction{tx: tx}
	
	return txRepo, transaction, nil
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &UserRepository{db: db}
}

// NewUserRepositoryWithTx creates a new UserRepository instance with a transaction
func NewUserRepositoryWithTx(tx *gorm.DB) domain.UserRepository {
	return &UserRepository{db: tx}
}

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		return fmt.Errorf("failed to create user: %w", result.Error)
	}
	return nil
}

// GetByTelegramID retrieves a user by their Telegram ID
func (r *UserRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	var user domain.User
	result := r.db.WithContext(ctx).Where("telegram_id = ?", telegramID).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domain.UserNotFoundError{TelegramID: telegramID}
		}
		return nil, fmt.Errorf("failed to get user: %w", result.Error)
	}
	return &user, nil
}

// Update updates an existing user in the database
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.UserNotFoundError{TelegramID: user.TelegramID}
	}
	return nil
}

// UpdateQuota updates only the quota_used field for a user
func (r *UserRepository) UpdateQuota(ctx context.Context, telegramID int64, quotaUsed int64) error {
	result := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("telegram_id = ?", telegramID).
		Updates(map[string]interface{}{
			"quota_used": quotaUsed,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update quota: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.UserNotFoundError{TelegramID: telegramID}
	}
	return nil
}
