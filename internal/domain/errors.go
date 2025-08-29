package domain

import (
	"errors"
	"fmt"
)

// Domain errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotActive     = errors.New("user is not active")
	ErrUserAlreadyActive = errors.New("user is already active")
	ErrQuotaExceeded     = errors.New("quota usage exceeds limit")
	ErrInvalidInput      = errors.New("invalid input")

	ErrDatabaseError     = errors.New("database error")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrInternalError     = errors.New("internal error")
)

// UserNotFoundError represents when a user is not found
type UserNotFoundError struct {
	TelegramID int64
}

func (e UserNotFoundError) Error() string {
	return fmt.Sprintf("user not found with telegram_id %d", e.TelegramID)
}

func (e UserNotFoundError) Is(target error) bool {
	return target == ErrUserNotFound
}

// UserAlreadyExistsError represents when a user already exists
type UserAlreadyExistsError struct {
	TelegramID int64
}

func (e UserAlreadyExistsError) Error() string {
	return fmt.Sprintf("user already exists with telegram_id %d", e.TelegramID)
}

func (e UserAlreadyExistsError) Is(target error) bool {
	return target == ErrUserAlreadyExists
}

// QuotaExceededError represents when quota usage exceeds limit
type QuotaExceededError struct {
	Used  int64
	Limit int64
}

func (e QuotaExceededError) Error() string {
	return fmt.Sprintf("quota usage %d exceeds limit %d", e.Used, e.Limit)
}

func (e QuotaExceededError) Is(target error) bool {
	return target == ErrQuotaExceeded
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field %s: %s", e.Field, e.Message)
}

func (e ValidationError) Is(target error) bool {
	return target == ErrInvalidInput
}

// DatabaseError represents database-related errors
type DatabaseError struct {
	Operation string
	Err       error
}

func (e DatabaseError) Error() string {
	return fmt.Sprintf("database error during %s: %v", e.Operation, e.Err)
}

func (e DatabaseError) Is(target error) bool {
	return target == ErrDatabaseError
}

func (e DatabaseError) Unwrap() error {
	return e.Err
}

// RateLimitError represents rate limiting errors
type RateLimitError struct {
	Resource string
	Limit    int
	Window   string
}

func (e RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded for %s: %d requests per %s", e.Resource, e.Limit, e.Window)
}

func (e RateLimitError) Is(target error) bool {
	return target == ErrRateLimitExceeded
}

// AuthorizationError represents authorization errors
type AuthorizationError struct {
	Action string
	Reason string
}

func (e AuthorizationError) Error() string {
	return fmt.Sprintf("unauthorized to perform %s: %s", e.Action, e.Reason)
}

func (e AuthorizationError) Is(target error) bool {
	return target == ErrUnauthorized
}

// ForbiddenError represents forbidden access errors
type ForbiddenError struct {
	Resource string
	Reason   string
}

func (e ForbiddenError) Error() string {
	return fmt.Sprintf("access forbidden to %s: %s", e.Resource, e.Reason)
}

func (e ForbiddenError) Is(target error) bool {
	return target == ErrForbidden
}

// InternalError represents internal server errors
type InternalError struct {
	Component string
	Err       error
}

func (e InternalError) Error() string {
	return fmt.Sprintf("internal error in %s: %v", e.Component, e.Err)
}

func (e InternalError) Is(target error) bool {
	return target == ErrInternalError
}

func (e InternalError) Unwrap() error {
	return e.Err
}
