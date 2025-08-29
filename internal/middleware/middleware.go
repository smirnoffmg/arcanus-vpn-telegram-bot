package middleware

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

// HandlerFunc represents a middleware-aware handler function
type HandlerFunc func(ctx context.Context, data interface{}) error

// Middleware represents a middleware function that wraps a handler
type Middleware func(next HandlerFunc) HandlerFunc

// Chain applies a series of middleware to a handler
func Chain(handler HandlerFunc, middlewares ...Middleware) HandlerFunc {
	// Apply middleware in reverse order so the first middleware in the list
	// is the outermost middleware (executes first)
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// RequestData contains the data needed by handlers
type RequestData struct {
	Update   *tgbotapi.Update
	Message  *tgbotapi.Message
	Callback *tgbotapi.CallbackQuery
	UserID   int64
	ChatID   int64
	Username string
}

// NewRequestDataFromUpdate creates RequestData from a Telegram update
func NewRequestDataFromUpdate(update *tgbotapi.Update) *RequestData {
	data := &RequestData{Update: update}
	
	if update.Message != nil {
		data.Message = update.Message
		data.UserID = update.Message.From.ID
		data.ChatID = update.Message.Chat.ID
		data.Username = update.Message.From.UserName
	}
	
	if update.CallbackQuery != nil {
		data.Callback = update.CallbackQuery
		data.UserID = update.CallbackQuery.From.ID
		data.ChatID = update.CallbackQuery.Message.Chat.ID
		data.Username = update.CallbackQuery.From.UserName
	}
	
	return data
}

// Logger creates a logging middleware
func Logger(logger *logrus.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, data interface{}) error {
			requestData, ok := data.(*RequestData)
			if !ok {
				logger.Error("Invalid request data type in logging middleware")
				return next(ctx, data)
			}
			
			start := time.Now()
			
			// Log request
			fields := logrus.Fields{
				"user_id":  requestData.UserID,
				"chat_id":  requestData.ChatID,
				"username": requestData.Username,
			}
			
			if requestData.Message != nil {
				fields["text"] = requestData.Message.Text
				fields["message_id"] = requestData.Message.MessageID
				logger.WithFields(fields).Info("Processing message")
			}
			
			if requestData.Callback != nil {
				fields["callback_data"] = requestData.Callback.Data
				fields["message_id"] = requestData.Callback.Message.MessageID
				logger.WithFields(fields).Info("Processing callback")
			}
			
			err := next(ctx, data)
			
			// Log response
			duration := time.Since(start)
			fields["duration_ms"] = duration.Milliseconds()
			
			if err != nil {
				fields["error"] = err.Error()
				logger.WithFields(fields).Error("Request failed")
			} else {
				logger.WithFields(fields).Info("Request completed")
			}
			
			return err
		}
	}
}

// Recovery creates a panic recovery middleware
func Recovery(logger *logrus.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, data interface{}) (err error) {
			defer func() {
				if r := recover(); r != nil {
					logger.WithField("panic", r).Error("Handler panicked")
					// Convert panic to error
					if panicErr, ok := r.(error); ok {
						err = panicErr
					} else {
						err = ErrInternalError
					}
				}
			}()
			
			return next(ctx, data)
		}
	}
}

// Timeout creates a timeout middleware
func Timeout(duration time.Duration) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, data interface{}) error {
			ctx, cancel := context.WithTimeout(ctx, duration)
			defer cancel()
			
			// Use a channel to communicate result
			resultChan := make(chan error, 1)
			
			go func() {
				defer func() {
					if r := recover(); r != nil {
						if err, ok := r.(error); ok {
							resultChan <- err
						} else {
							resultChan <- ErrInternalError
						}
					}
				}()
				
				resultChan <- next(ctx, data)
			}()
			
			select {
			case err := <-resultChan:
				return err
			case <-ctx.Done():
				return ErrTimeout
			}
		}
	}
}

// RateLimit creates a rate limiting middleware
func RateLimit(rateLimiter RateLimiter) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, data interface{}) error {
			requestData, ok := data.(*RequestData)
			if !ok {
				return ErrInvalidRequestData
			}
			
			if !rateLimiter.Allow(requestData.UserID) {
				return ErrRateLimitExceeded
			}
			
			return next(ctx, data)
		}
	}
}

// RateLimiter interface for rate limiting
type RateLimiter interface {
	Allow(userID int64) bool
}

// Audit creates an audit logging middleware
func Audit(auditLogger AuditLogger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, data interface{}) error {
			requestData, ok := data.(*RequestData)
			if !ok {
				return ErrInvalidRequestData
			}
			
			// Log the action before processing
			action := "unknown"
			if requestData.Message != nil {
				action = "message:" + requestData.Message.Text
			} else if requestData.Callback != nil {
				action = "callback:" + requestData.Callback.Data
			}
			
			auditLogger.LogAction(requestData.UserID, action, time.Now())
			
			return next(ctx, data)
		}
	}
}

// AuditLogger interface for audit logging
type AuditLogger interface {
	LogAction(userID int64, action string, timestamp time.Time)
}