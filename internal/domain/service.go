package domain

import "context"

// UserService defines the interface for user business logic
type UserService interface {
	RegisterUser(ctx context.Context, telegramID int64, username, firstName, lastName string) (*User, error)
	GetUser(ctx context.Context, telegramID int64) (*User, error)
	ActivateTrial(ctx context.Context, telegramID int64) error
	UpdateQuota(ctx context.Context, telegramID int64, quotaUsed int64) error
}
