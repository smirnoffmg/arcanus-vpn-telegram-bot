package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestChain(t *testing.T) {
	t.Run("Single middleware", func(t *testing.T) {
		called := false
		middleware1 := func(next HandlerFunc) HandlerFunc {
			return func(ctx context.Context, data interface{}) error {
				called = true
				return next(ctx, data)
			}
		}
		
		handler := func(ctx context.Context, data interface{}) error {
			return nil
		}
		
		chainedHandler := Chain(handler, middleware1)
		err := chainedHandler(context.Background(), nil)
		
		assert.NoError(t, err)
		assert.True(t, called)
	})
	
	t.Run("Multiple middleware in correct order", func(t *testing.T) {
		var order []int
		
		middleware1 := func(next HandlerFunc) HandlerFunc {
			return func(ctx context.Context, data interface{}) error {
				order = append(order, 1)
				err := next(ctx, data)
				order = append(order, 4)
				return err
			}
		}
		
		middleware2 := func(next HandlerFunc) HandlerFunc {
			return func(ctx context.Context, data interface{}) error {
				order = append(order, 2)
				err := next(ctx, data)
				order = append(order, 3)
				return err
			}
		}
		
		handler := func(ctx context.Context, data interface{}) error {
			order = append(order, 100)
			return nil
		}
		
		chainedHandler := Chain(handler, middleware1, middleware2)
		err := chainedHandler(context.Background(), nil)
		
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 100, 3, 4}, order)
	})
}

func TestNewRequestDataFromUpdate(t *testing.T) {
	t.Run("Message update", func(t *testing.T) {
		update := &tgbotapi.Update{
			Message: &tgbotapi.Message{
				From: &tgbotapi.User{
					ID:       123,
					UserName: "testuser",
				},
				Chat: &tgbotapi.Chat{
					ID: 456,
				},
				Text:      "/start",
				MessageID: 789,
			},
		}
		
		data := NewRequestDataFromUpdate(update)
		
		assert.Equal(t, update, data.Update)
		assert.Equal(t, update.Message, data.Message)
		assert.Nil(t, data.Callback)
		assert.Equal(t, int64(123), data.UserID)
		assert.Equal(t, int64(456), data.ChatID)
		assert.Equal(t, "testuser", data.Username)
	})
	
	t.Run("Callback update", func(t *testing.T) {
		update := &tgbotapi.Update{
			CallbackQuery: &tgbotapi.CallbackQuery{
				From: &tgbotapi.User{
					ID:       123,
					UserName: "testuser",
				},
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{
						ID: 456,
					},
					MessageID: 789,
				},
				Data: "test_data",
			},
		}
		
		data := NewRequestDataFromUpdate(update)
		
		assert.Equal(t, update, data.Update)
		assert.Nil(t, data.Message)
		assert.Equal(t, update.CallbackQuery, data.Callback)
		assert.Equal(t, int64(123), data.UserID)
		assert.Equal(t, int64(456), data.ChatID)
		assert.Equal(t, "testuser", data.Username)
	})
}

func TestLogger(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel) // Suppress output during tests
	
	t.Run("Logs message processing", func(t *testing.T) {
		middleware := Logger(logger)
		
		handler := func(ctx context.Context, data interface{}) error {
			return nil
		}
		
		requestData := &RequestData{
			Message: &tgbotapi.Message{
				Text:      "/start",
				MessageID: 123,
			},
			UserID:   456,
			ChatID:   789,
			Username: "testuser",
		}
		
		wrappedHandler := middleware(handler)
		err := wrappedHandler(context.Background(), requestData)
		
		assert.NoError(t, err)
	})
	
	t.Run("Logs callback processing", func(t *testing.T) {
		middleware := Logger(logger)
		
		handler := func(ctx context.Context, data interface{}) error {
			return nil
		}
		
		requestData := &RequestData{
			Callback: &tgbotapi.CallbackQuery{
				Data: "test_data",
				Message: &tgbotapi.Message{
					MessageID: 123,
				},
			},
			UserID:   456,
			ChatID:   789,
			Username: "testuser",
		}
		
		wrappedHandler := middleware(handler)
		err := wrappedHandler(context.Background(), requestData)
		
		assert.NoError(t, err)
	})
	
	t.Run("Logs errors", func(t *testing.T) {
		middleware := Logger(logger)
		
		testError := errors.New("test error")
		handler := func(ctx context.Context, data interface{}) error {
			return testError
		}
		
		requestData := &RequestData{
			UserID: 456,
			ChatID: 789,
		}
		
		wrappedHandler := middleware(handler)
		err := wrappedHandler(context.Background(), requestData)
		
		assert.Error(t, err)
		assert.Equal(t, testError, err)
	})
}

func TestRecovery(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel) // Suppress output during tests
	
	t.Run("Recovers from panic", func(t *testing.T) {
		middleware := Recovery(logger)
		
		handler := func(ctx context.Context, data interface{}) error {
			panic("test panic")
		}
		
		wrappedHandler := middleware(handler)
		err := wrappedHandler(context.Background(), nil)
		
		assert.Error(t, err)
		assert.Equal(t, ErrInternalError, err)
	})
	
	t.Run("Recovers from error panic", func(t *testing.T) {
		middleware := Recovery(logger)
		
		testError := errors.New("test error")
		handler := func(ctx context.Context, data interface{}) error {
			panic(testError)
		}
		
		wrappedHandler := middleware(handler)
		err := wrappedHandler(context.Background(), nil)
		
		assert.Error(t, err)
		assert.Equal(t, testError, err)
	})
	
	t.Run("Does not affect normal execution", func(t *testing.T) {
		middleware := Recovery(logger)
		
		handler := func(ctx context.Context, data interface{}) error {
			return nil
		}
		
		wrappedHandler := middleware(handler)
		err := wrappedHandler(context.Background(), nil)
		
		assert.NoError(t, err)
	})
}

func TestTimeout(t *testing.T) {
	t.Run("Handler completes within timeout", func(t *testing.T) {
		middleware := Timeout(100 * time.Millisecond)
		
		handler := func(ctx context.Context, data interface{}) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		}
		
		wrappedHandler := middleware(handler)
		err := wrappedHandler(context.Background(), nil)
		
		assert.NoError(t, err)
	})
	
	t.Run("Handler times out", func(t *testing.T) {
		middleware := Timeout(10 * time.Millisecond)
		
		handler := func(ctx context.Context, data interface{}) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		}
		
		wrappedHandler := middleware(handler)
		err := wrappedHandler(context.Background(), nil)
		
		assert.Error(t, err)
		assert.Equal(t, ErrTimeout, err)
	})
}

// MockRateLimiter for testing
type MockRateLimiter struct {
	allowResponse bool
}

func (m *MockRateLimiter) Allow(userID int64) bool {
	return m.allowResponse
}

func TestRateLimit(t *testing.T) {
	t.Run("Allows when rate limiter allows", func(t *testing.T) {
		rateLimiter := &MockRateLimiter{allowResponse: true}
		middleware := RateLimit(rateLimiter)
		
		handler := func(ctx context.Context, data interface{}) error {
			return nil
		}
		
		requestData := &RequestData{UserID: 123}
		
		wrappedHandler := middleware(handler)
		err := wrappedHandler(context.Background(), requestData)
		
		assert.NoError(t, err)
	})
	
	t.Run("Blocks when rate limiter blocks", func(t *testing.T) {
		rateLimiter := &MockRateLimiter{allowResponse: false}
		middleware := RateLimit(rateLimiter)
		
		handler := func(ctx context.Context, data interface{}) error {
			return nil
		}
		
		requestData := &RequestData{UserID: 123}
		
		wrappedHandler := middleware(handler)
		err := wrappedHandler(context.Background(), requestData)
		
		assert.Error(t, err)
		assert.Equal(t, ErrRateLimitExceeded, err)
	})
	
	t.Run("Returns error for invalid request data", func(t *testing.T) {
		rateLimiter := &MockRateLimiter{allowResponse: true}
		middleware := RateLimit(rateLimiter)
		
		handler := func(ctx context.Context, data interface{}) error {
			return nil
		}
		
		wrappedHandler := middleware(handler)
		err := wrappedHandler(context.Background(), "invalid data")
		
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidRequestData, err)
	})
}

// MockAuditLogger for testing
type MockAuditLogger struct {
	actions []string
}

func (m *MockAuditLogger) LogAction(userID int64, action string, timestamp time.Time) {
	m.actions = append(m.actions, action)
}

func TestAudit(t *testing.T) {
	t.Run("Logs message action", func(t *testing.T) {
		auditLogger := &MockAuditLogger{}
		middleware := Audit(auditLogger)
		
		handler := func(ctx context.Context, data interface{}) error {
			return nil
		}
		
		requestData := &RequestData{
			Message: &tgbotapi.Message{Text: "/start"},
			UserID:  123,
		}
		
		wrappedHandler := middleware(handler)
		err := wrappedHandler(context.Background(), requestData)
		
		assert.NoError(t, err)
		assert.Len(t, auditLogger.actions, 1)
		assert.Equal(t, "message:/start", auditLogger.actions[0])
	})
	
	t.Run("Logs callback action", func(t *testing.T) {
		auditLogger := &MockAuditLogger{}
		middleware := Audit(auditLogger)
		
		handler := func(ctx context.Context, data interface{}) error {
			return nil
		}
		
		requestData := &RequestData{
			Callback: &tgbotapi.CallbackQuery{Data: "test_data"},
			UserID:   123,
		}
		
		wrappedHandler := middleware(handler)
		err := wrappedHandler(context.Background(), requestData)
		
		assert.NoError(t, err)
		assert.Len(t, auditLogger.actions, 1)
		assert.Equal(t, "callback:test_data", auditLogger.actions[0])
	})
}