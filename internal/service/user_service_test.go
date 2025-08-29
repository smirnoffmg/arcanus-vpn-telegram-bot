package service

import (
	"context"
	"testing"

	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of domain.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	args := m.Called(ctx, telegramID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateQuota(ctx context.Context, telegramID int64, quotaUsed int64) error {
	args := m.Called(ctx, telegramID, quotaUsed)
	return args.Error(0)
}

func TestUserService_RegisterUser_NewUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	telegramID := int64(123)
	username := "testuser"
	firstName := "Test"
	lastName := "User"

	// Mock GetByTelegramID to return "not found"
	mockRepo.On("GetByTelegramID", mock.Anything, telegramID).
		Return((*domain.User)(nil), domain.UserNotFoundError{TelegramID: telegramID})

	// Mock Create to succeed
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
		Return(nil)

	user, err := service.RegisterUser(context.Background(), telegramID, username, firstName, lastName)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, telegramID, user.TelegramID)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, firstName, user.FirstName)
	assert.Equal(t, lastName, user.LastName)
	assert.Equal(t, domain.UserStatusInactive, user.Status)

	mockRepo.AssertExpectations(t)
}

func TestUserService_RegisterUser_ExistingUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	telegramID := int64(123)
	username := "testuser"
	firstName := "Test"
	lastName := "User"

	existingUser := domain.NewUser(telegramID, username, firstName, lastName)
	existingUser.Status = domain.UserStatusTrial

	// Mock GetByTelegramID to return existing user
	mockRepo.On("GetByTelegramID", mock.Anything, telegramID).
		Return(existingUser, nil)

	user, err := service.RegisterUser(context.Background(), telegramID, username, firstName, lastName)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, existingUser.TelegramID, user.TelegramID)
	assert.Equal(t, existingUser.Status, user.Status)

	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	telegramID := int64(123)
	expectedUser := domain.NewUser(telegramID, "testuser", "Test", "User")

	mockRepo.On("GetByTelegramID", mock.Anything, telegramID).
		Return(expectedUser, nil)

	user, err := service.GetUser(context.Background(), telegramID)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)

	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUser_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	telegramID := int64(123)

	mockRepo.On("GetByTelegramID", mock.Anything, telegramID).
		Return((*domain.User)(nil), domain.UserNotFoundError{TelegramID: telegramID})

	user, err := service.GetUser(context.Background(), telegramID)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to get user")

	mockRepo.AssertExpectations(t)
}

func TestUserService_ActivateTrial(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	telegramID := int64(123)
	user := domain.NewUser(telegramID, "testuser", "Test", "User")

	mockRepo.On("GetByTelegramID", mock.Anything, telegramID).
		Return(user, nil)

	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).
		Return(nil)

	err := service.ActivateTrial(context.Background(), telegramID)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestUserService_ActivateTrial_AlreadyActive(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	telegramID := int64(123)
	user := domain.NewUser(telegramID, "testuser", "Test", "User")
	user.Status = domain.UserStatusTrial

	mockRepo.On("GetByTelegramID", mock.Anything, telegramID).
		Return(user, nil)

	err := service.ActivateTrial(context.Background(), telegramID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserAlreadyActive)

	mockRepo.AssertExpectations(t)
}

func TestUserService_ActivateTrial_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	telegramID := int64(123)

	mockRepo.On("GetByTelegramID", mock.Anything, telegramID).
		Return((*domain.User)(nil), domain.UserNotFoundError{TelegramID: telegramID})

	err := service.ActivateTrial(context.Background(), telegramID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user for trial activation")

	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateQuota(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	telegramID := int64(123)
	user := domain.NewUser(telegramID, "testuser", "Test", "User")
	user.Status = domain.UserStatusTrial
	user.QuotaLimit = 1000

	mockRepo.On("GetByTelegramID", mock.Anything, telegramID).
		Return(user, nil)

	mockRepo.On("UpdateQuota", mock.Anything, telegramID, int64(500)).
		Return(nil)

	err := service.UpdateQuota(context.Background(), telegramID, 500)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateQuota_UserNotActive(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	telegramID := int64(123)
	user := domain.NewUser(telegramID, "testuser", "Test", "User")
	user.Status = domain.UserStatusInactive

	mockRepo.On("GetByTelegramID", mock.Anything, telegramID).
		Return(user, nil)

	err := service.UpdateQuota(context.Background(), telegramID, 500)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserNotActive)

	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateQuota_ExceedsLimit(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	telegramID := int64(123)
	user := domain.NewUser(telegramID, "testuser", "Test", "User")
	user.Status = domain.UserStatusTrial
	user.QuotaLimit = 1000

	mockRepo.On("GetByTelegramID", mock.Anything, telegramID).
		Return(user, nil)

	err := service.UpdateQuota(context.Background(), telegramID, 1500)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrQuotaExceeded)

	mockRepo.AssertExpectations(t)
}

func TestUserService_RegisterUser_InvalidInput(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	tests := []struct {
		name       string
		telegramID int64
		username   string
		firstName  string
		lastName   string
		wantErr    bool
	}{
		{
			name:       "Invalid telegram ID",
			telegramID: 0,
			username:   "testuser",
			firstName:  "Test",
			lastName:   "User",
			wantErr:    true,
		},
		{
			name:       "Empty first name",
			telegramID: 123,
			username:   "testuser",
			firstName:  "",
			lastName:   "User",
			wantErr:    true,
		},
		{
			name:       "Empty last name",
			telegramID: 123,
			username:   "testuser",
			firstName:  "Test",
			lastName:   "",
			wantErr:    false, // Last name can be empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Only set up mock expectations if we expect the service to call the repository
			if !tt.wantErr || (tt.telegramID > 0 && tt.firstName != "") {
				mockRepo.On("GetByTelegramID", mock.Anything, tt.telegramID).
					Return((*domain.User)(nil), domain.UserNotFoundError{TelegramID: tt.telegramID})
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
					Return(nil)
			}

			user, err := service.RegisterUser(context.Background(), tt.telegramID, tt.username, tt.firstName, tt.lastName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUser_InvalidInput(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	user, err := service.GetUser(context.Background(), 0)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestUserService_ActivateTrial_InvalidInput(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	err := service.ActivateTrial(context.Background(), 0)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestUserService_ActivateTrial_UpdateError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	telegramID := int64(123)
	user := domain.NewUser(telegramID, "testuser", "Test", "User")

	mockRepo.On("GetByTelegramID", mock.Anything, telegramID).
		Return(user, nil)

	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).
		Return(domain.ErrUserNotFound)

	err := service.ActivateTrial(context.Background(), telegramID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update user for trial activation")

	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateQuota_InvalidInput(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	tests := []struct {
		name       string
		telegramID int64
		quotaUsed  int64
		wantErr    bool
	}{
		{
			name:       "Invalid telegram ID",
			telegramID: 0,
			quotaUsed:  500,
			wantErr:    true,
		},
		{
			name:       "Negative quota used",
			telegramID: 123,
			quotaUsed:  -100,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateQuota(context.Background(), tt.telegramID, tt.quotaUsed)

			assert.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrInvalidInput)
		})
	}
}

func TestUserService_UpdateQuota_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	telegramID := int64(123)

	mockRepo.On("GetByTelegramID", mock.Anything, telegramID).
		Return((*domain.User)(nil), domain.UserNotFoundError{TelegramID: telegramID})

	err := service.UpdateQuota(context.Background(), telegramID, 500)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user for quota update")

	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateQuota_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	telegramID := int64(123)
	user := domain.NewUser(telegramID, "testuser", "Test", "User")
	user.Status = domain.UserStatusTrial
	user.QuotaLimit = 1000

	mockRepo.On("GetByTelegramID", mock.Anything, telegramID).
		Return(user, nil)

	mockRepo.On("UpdateQuota", mock.Anything, telegramID, int64(500)).
		Return(domain.ErrUserNotFound)

	err := service.UpdateQuota(context.Background(), telegramID, 500)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update quota")

	mockRepo.AssertExpectations(t)
}
