package middleware

import "errors"

var (
	// ErrTimeout is returned when a handler times out
	ErrTimeout = errors.New("handler timeout")
	
	// ErrRateLimitExceeded is returned when rate limit is exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	
	// ErrInvalidRequestData is returned when request data is invalid
	ErrInvalidRequestData = errors.New("invalid request data")
	
	// ErrInternalError is returned for internal errors
	ErrInternalError = errors.New("internal error")
)