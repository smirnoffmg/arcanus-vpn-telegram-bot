package bot

import (
	"context"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupTestHandler creates a handler with mocked dependencies for testing
func setupTestHandler() (*MockBotAPI, *MockUserService, *Handler) {
	mockBotAPI := new(MockBotAPI)
	mockService := new(MockUserService)
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel) // Reduce test noise
	handler := NewHandler(mockBotAPI, mockService, logger)
	return mockBotAPI, mockService, handler
}

// MockBotAPI is a mock implementation of the Telegram Bot API
type MockBotAPI struct {
	mock.Mock
}

func (m *MockBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	args := m.Called(c)
	return args.Get(0).(tgbotapi.Message), args.Error(1)
}

func (m *MockBotAPI) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	args := m.Called(c)
	return args.Get(0).(*tgbotapi.APIResponse), args.Error(1)
}

func (m *MockBotAPI) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	args := m.Called(config)
	return args.Get(0).(tgbotapi.UpdatesChannel)
}

func (m *MockBotAPI) StopReceivingUpdates() {
	m.Called()
}

func (m *MockBotAPI) GetMe() (tgbotapi.User, error) {
	args := m.Called()
	return args.Get(0).(tgbotapi.User), args.Error(1)
}

// MockUserService is a mock implementation of domain.UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) RegisterUser(ctx context.Context, telegramID int64, username, firstName, lastName string) (*domain.User, error) {
	args := m.Called(ctx, telegramID, username, firstName, lastName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, telegramID int64) (*domain.User, error) {
	args := m.Called(ctx, telegramID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) ActivateTrial(ctx context.Context, telegramID int64) error {
	args := m.Called(ctx, telegramID)
	return args.Error(0)
}

func (m *MockUserService) UpdateQuota(ctx context.Context, telegramID int64, quotaUsed int64) error {
	args := m.Called(ctx, telegramID, quotaUsed)
	return args.Error(0)
}

func TestHandler_HandleUpdate_StartCommand(t *testing.T) {
	mockBotAPI, mockService, handler := setupTestHandler()

	// Create a test message
	message := &tgbotapi.Message{
		Text: "/start",
		From: &tgbotapi.User{
			ID:        123,
			UserName:  "testuser",
			FirstName: "Test",
			LastName:  "User",
		},
		Chat: &tgbotapi.Chat{
			ID: 456,
		},
	}

	update := tgbotapi.Update{Message: message}

	// Mock the service response
	expectedUser := domain.NewUser(123, "testuser", "Test", "User")
	mockService.On("RegisterUser", mock.Anything, int64(123), "testuser", "Test", "User").
		Return(expectedUser, nil)
		
	// Mock the bot API response
	mockBotAPI.On("Send", mock.AnythingOfType("tgbotapi.MessageConfig")).
		Return(tgbotapi.Message{}, nil)

	err := handler.HandleUpdate(context.Background(), update)

	assert.NoError(t, err)
	mockService.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestHandler_HandleUpdate_AccountCommand(t *testing.T) {
	mockBotAPI, mockService, handler := setupTestHandler()

	// Create a test message
	message := &tgbotapi.Message{
		Text: "/account",
		From: &tgbotapi.User{
			ID:        123,
			UserName:  "testuser",
			FirstName: "Test",
			LastName:  "User",
		},
		Chat: &tgbotapi.Chat{
			ID: 456,
		},
	}

	update := tgbotapi.Update{Message: message}

	// Mock the service response
	expectedUser := domain.NewUser(123, "testuser", "Test", "User")
	mockService.On("GetUser", mock.Anything, int64(123)).
		Return(expectedUser, nil)
		
	// Mock the bot API response
	mockBotAPI.On("Send", mock.AnythingOfType("tgbotapi.MessageConfig")).
		Return(tgbotapi.Message{}, nil)

	err := handler.HandleUpdate(context.Background(), update)

	assert.NoError(t, err)
	mockService.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestHandler_HandleUpdate_HelpCommand(t *testing.T) {
	mockBotAPI, mockService, handler := setupTestHandler()

	// Create a test message
	message := &tgbotapi.Message{
		Text: "/help",
		From: &tgbotapi.User{
			ID:        123,
			UserName:  "testuser",
			FirstName: "Test",
			LastName:  "User",
		},
		Chat: &tgbotapi.Chat{
			ID: 456,
		},
	}

	update := tgbotapi.Update{Message: message}
	
	// Mock the bot API response
	mockBotAPI.On("Send", mock.AnythingOfType("tgbotapi.MessageConfig")).
		Return(tgbotapi.Message{}, nil)

	err := handler.HandleUpdate(context.Background(), update)

	assert.NoError(t, err)
	mockBotAPI.AssertExpectations(t)
	// No service expectations needed for help command
	_ = mockService
}

func TestHandler_HandleUpdate_UnknownCommand(t *testing.T) {
	mockBotAPI, mockService, handler := setupTestHandler()

	// Create a test message
	message := &tgbotapi.Message{
		Text: "/unknown",
		From: &tgbotapi.User{
			ID:        123,
			UserName:  "testuser",
			FirstName: "Test",
			LastName:  "User",
		},
		Chat: &tgbotapi.Chat{
			ID: 456,
		},
	}

	update := tgbotapi.Update{Message: message}
	
	// Mock the bot API response for unknown command
	mockBotAPI.On("Send", mock.AnythingOfType("tgbotapi.MessageConfig")).
		Return(tgbotapi.Message{}, nil)

	err := handler.HandleUpdate(context.Background(), update)

	assert.NoError(t, err)
	mockBotAPI.AssertExpectations(t)
	// No service expectations needed for unknown command
	_ = mockService
}

func TestHandler_HandleUpdate_NilMessage(t *testing.T) {
	mockBotAPI, mockService, handler := setupTestHandler()

	update := tgbotapi.Update{Message: nil}

	err := handler.HandleUpdate(context.Background(), update)

	assert.NoError(t, err)
	// Suppress unused variable warnings
	_, _ = mockBotAPI, mockService
}

func TestHandler_HandleCallback_Trial(t *testing.T) {
	mockBotAPI, mockService, handler := setupTestHandler()

	// Create a test callback
	callback := &tgbotapi.CallbackQuery{
		ID: "test_callback_id",
		From: &tgbotapi.User{
			ID:        123,
			UserName:  "testuser",
			FirstName: "Test",
			LastName:  "User",
		},
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID: 456,
			},
			MessageID: 789,
		},
		Data: "trial",
	}

	// Mock the service responses
	mockService.On("ActivateTrial", mock.Anything, int64(123)).Return(nil)

	expectedUser := domain.NewUser(123, "testuser", "Test", "User")
	expectedUser.Status = domain.UserStatusTrial
	mockService.On("GetUser", mock.Anything, int64(123)).Return(expectedUser, nil)
	
	// Mock the bot API responses
	mockBotAPI.On("Request", mock.AnythingOfType("tgbotapi.CallbackConfig")).
		Return(&tgbotapi.APIResponse{Ok: true}, nil).Maybe()
	mockBotAPI.On("Send", mock.AnythingOfType("tgbotapi.EditMessageTextConfig")).
		Return(tgbotapi.Message{}, nil)

	err := handler.HandleCallback(context.Background(), callback)

	assert.NoError(t, err)
	mockService.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestHandler_HandleCallback_Account(t *testing.T) {
	mockBotAPI, mockService, handler := setupTestHandler()

	// Create a test callback
	callback := &tgbotapi.CallbackQuery{
		ID: "test_callback_id",
		From: &tgbotapi.User{
			ID:        123,
			UserName:  "testuser",
			FirstName: "Test",
			LastName:  "User",
		},
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID: 456,
			},
			MessageID: 789,
		},
		Data: "account",
	}

	// Mock the service response
	expectedUser := domain.NewUser(123, "testuser", "Test", "User")
	mockService.On("GetUser", mock.Anything, int64(123)).Return(expectedUser, nil)
	
	// Mock the bot API responses
	mockBotAPI.On("Request", mock.AnythingOfType("tgbotapi.CallbackConfig")).
		Return(&tgbotapi.APIResponse{Ok: true}, nil).Maybe()
	mockBotAPI.On("Send", mock.AnythingOfType("tgbotapi.EditMessageTextConfig")).
		Return(tgbotapi.Message{}, nil)

	err := handler.HandleCallback(context.Background(), callback)

	assert.NoError(t, err)
	mockService.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestHandler_HandleCallback_Help(t *testing.T) {
	mockBotAPI, mockService, handler := setupTestHandler()

	// Create a test callback
	callback := &tgbotapi.CallbackQuery{
		ID: "test_callback_id",
		From: &tgbotapi.User{
			ID:        123,
			UserName:  "testuser",
			FirstName: "Test",
			LastName:  "User",
		},
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID: 456,
			},
			MessageID: 789,
		},
		Data: "help",
	}
	
	// Mock the bot API responses for help callback
	mockBotAPI.On("Request", mock.AnythingOfType("tgbotapi.CallbackConfig")).
		Return(&tgbotapi.APIResponse{Ok: true}, nil).Maybe()
	mockBotAPI.On("Send", mock.AnythingOfType("tgbotapi.EditMessageTextConfig")).
		Return(tgbotapi.Message{}, nil)

	err := handler.HandleCallback(context.Background(), callback)

	assert.NoError(t, err)
	mockBotAPI.AssertExpectations(t)
	// No service expectations needed for help callback
	_ = mockService
}

func TestHandler_HandleCallback_Unknown(t *testing.T) {
	mockBotAPI, mockService, handler := setupTestHandler()

	// Create a test callback
	callback := &tgbotapi.CallbackQuery{
		ID: "test_callback_id",
		From: &tgbotapi.User{
			ID:        123,
			UserName:  "testuser",
			FirstName: "Test",
			LastName:  "User",
		},
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID: 456,
			},
			MessageID: 789,
		},
		Data: "unknown",
	}
	
	// Mock the bot API response for unknown callback
	mockBotAPI.On("Request", mock.AnythingOfType("tgbotapi.CallbackConfig")).
		Return(&tgbotapi.APIResponse{Ok: true}, nil)

	err := handler.HandleCallback(context.Background(), callback)

	assert.NoError(t, err)
	mockBotAPI.AssertExpectations(t)
	// No service expectations needed for unknown callback
	_ = mockService
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"Zero bytes", 0, "0 B"},
		{"Small bytes", 1023, "1023 B"},
		{"One KB", 1024, "1.0 KB"},
		{"One MB", 1024 * 1024, "1.0 MB"},
		{"One GB", 1024 * 1024 * 1024, "1.0 GB"},
		{"Mixed bytes", 1536, "1.5 KB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}
