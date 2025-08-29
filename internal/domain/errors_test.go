package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserNotFoundError(t *testing.T) {
	telegramID := int64(123)
	err := UserNotFoundError{TelegramID: telegramID}

	assert.Equal(t, "user not found with telegram_id 123", err.Error())
	assert.True(t, errors.Is(err, ErrUserNotFound))
}

func TestUserAlreadyExistsError(t *testing.T) {
	telegramID := int64(123)
	err := UserAlreadyExistsError{TelegramID: telegramID}

	assert.Equal(t, "user already exists with telegram_id 123", err.Error())
	assert.True(t, errors.Is(err, ErrUserAlreadyExists))
}

func TestQuotaExceededError(t *testing.T) {
	used := int64(1500)
	limit := int64(1000)
	err := QuotaExceededError{Used: used, Limit: limit}

	assert.Equal(t, "quota usage 1500 exceeds limit 1000", err.Error())
	assert.True(t, errors.Is(err, ErrQuotaExceeded))
}

func TestValidationError(t *testing.T) {
	field := "username"
	message := "cannot be empty"
	err := ValidationError{Field: field, Message: message}

	assert.Equal(t, "validation error for field username: cannot be empty", err.Error())
	assert.True(t, errors.Is(err, ErrInvalidInput))
}

func TestErrorWrapping(t *testing.T) {
	// Test that our custom errors can be wrapped and unwrapped correctly
	originalErr := UserNotFoundError{TelegramID: 123}
	wrappedErr := errors.New("some context: " + originalErr.Error())

	// The wrapped error should contain the original error message
	assert.Contains(t, wrappedErr.Error(), "user not found with telegram_id 123")

	// Test error comparison
	assert.True(t, errors.Is(originalErr, ErrUserNotFound))
	assert.False(t, errors.Is(originalErr, ErrUserAlreadyExists))
}

func TestDatabaseError(t *testing.T) {
	originalErr := errors.New("connection timeout")
	err := DatabaseError{
		Operation: "query",
		Err:       originalErr,
	}

	assert.Equal(t, "database error during query: connection timeout", err.Error())
	assert.True(t, errors.Is(err, ErrDatabaseError))
	assert.Equal(t, originalErr, err.Unwrap())
}

func TestRateLimitError(t *testing.T) {
	err := RateLimitError{
		Resource: "api_calls",
		Limit:    100,
		Window:   "1 minute",
	}

	assert.Equal(t, "rate limit exceeded for api_calls: 100 requests per 1 minute", err.Error())
	assert.True(t, errors.Is(err, ErrRateLimitExceeded))
}

func TestAuthorizationError(t *testing.T) {
	err := AuthorizationError{
		Action: "delete_user",
		Reason: "insufficient permissions",
	}

	assert.Equal(t, "unauthorized to perform delete_user: insufficient permissions", err.Error())
	assert.True(t, errors.Is(err, ErrUnauthorized))
}

func TestForbiddenError(t *testing.T) {
	err := ForbiddenError{
		Resource: "admin_panel",
		Reason:   "user not admin",
	}

	assert.Equal(t, "access forbidden to admin_panel: user not admin", err.Error())
	assert.True(t, errors.Is(err, ErrForbidden))
}

func TestInternalError(t *testing.T) {
	originalErr := errors.New("unexpected nil pointer")
	err := InternalError{
		Component: "user_service",
		Err:       originalErr,
	}

	assert.Equal(t, "internal error in user_service: unexpected nil pointer", err.Error())
	assert.True(t, errors.Is(err, ErrInternalError))
	assert.Equal(t, originalErr, err.Unwrap())
}

func TestErrorTypeComparison(t *testing.T) {
	// Test that different error types are not equal
	userNotFound := UserNotFoundError{TelegramID: 123}
	userExists := UserAlreadyExistsError{TelegramID: 123}
	quotaExceeded := QuotaExceededError{Used: 100, Limit: 50}

	assert.False(t, errors.Is(userNotFound, ErrUserAlreadyExists))
	assert.False(t, errors.Is(userExists, ErrUserNotFound))
	assert.False(t, errors.Is(quotaExceeded, ErrUserNotFound))
	assert.False(t, errors.Is(quotaExceeded, ErrUserAlreadyExists))
}

func TestErrorWrappingChain(t *testing.T) {
	// Test error wrapping chain
	originalErr := errors.New("database connection failed")
	dbErr := DatabaseError{Operation: "create_user", Err: originalErr}
	internalErr := InternalError{Component: "user_service", Err: dbErr}

	assert.Equal(t, "internal error in user_service: database error during create_user: database connection failed", internalErr.Error())
	assert.True(t, errors.Is(internalErr, ErrInternalError))
	assert.True(t, errors.Is(internalErr, ErrDatabaseError))
	assert.Equal(t, dbErr, internalErr.Unwrap())
	assert.Equal(t, originalErr, dbErr.Unwrap())
}
