package domain

import "context"

// Transaction represents a database transaction
type Transaction interface {
	// Commit commits the transaction
	Commit() error
	// Rollback rolls back the transaction
	Rollback() error
}

// TransactionManager manages database transactions
type TransactionManager interface {
	// WithTransaction executes a function within a database transaction
	WithTransaction(ctx context.Context, fn func(ctx context.Context, tx Transaction) error) error
	// BeginTx starts a new transaction and returns a transactional repository
	BeginTx(ctx context.Context) (UserRepository, Transaction, error)
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	Update(ctx context.Context, user *User) error
	UpdateQuota(ctx context.Context, telegramID int64, quotaUsed int64) error
}
