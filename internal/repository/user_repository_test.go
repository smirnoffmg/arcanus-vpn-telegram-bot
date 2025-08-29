package repository

import (
	"context"
	"testing"

	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate the schema
	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	return db, func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}
}

func TestUserRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	user := domain.NewUser(123, "testuser", "Test", "User")

	err := repo.Create(context.Background(), user)

	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
}

func TestUserRepository_Create_DuplicateTelegramID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	user1 := domain.NewUser(123, "testuser1", "Test", "User")
	user2 := domain.NewUser(123, "testuser2", "Test", "User")

	err := repo.Create(context.Background(), user1)
	assert.NoError(t, err)

	err = repo.Create(context.Background(), user2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create user")
}

func TestUserRepository_GetByTelegramID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	expectedUser := domain.NewUser(123, "testuser", "Test", "User")

	// Create user first
	err := repo.Create(context.Background(), expectedUser)
	require.NoError(t, err)

	// Retrieve user
	user, err := repo.GetByTelegramID(context.Background(), 123)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser.TelegramID, user.TelegramID)
	assert.Equal(t, expectedUser.Username, user.Username)
	assert.Equal(t, expectedUser.FirstName, user.FirstName)
	assert.Equal(t, expectedUser.LastName, user.LastName)
}

func TestUserRepository_GetByTelegramID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	user, err := repo.GetByTelegramID(context.Background(), 123)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUserRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	user := domain.NewUser(123, "testuser", "Test", "User")

	// Create user first
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Update user
	user.Username = "updateduser"
	user.Status = domain.UserStatusTrial

	err = repo.Update(context.Background(), user)

	assert.NoError(t, err)

	// Verify update
	updatedUser, err := repo.GetByTelegramID(context.Background(), 123)
	assert.NoError(t, err)
	assert.Equal(t, "updateduser", updatedUser.Username)
	assert.Equal(t, domain.UserStatusTrial, updatedUser.Status)
}

func TestUserRepository_UpdateQuota(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	user := domain.NewUser(123, "testuser", "Test", "User")

	// Create user first
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Update quota
	newQuotaUsed := int64(1024)
	err = repo.UpdateQuota(context.Background(), 123, newQuotaUsed)

	assert.NoError(t, err)

	// Verify quota update
	updatedUser, err := repo.GetByTelegramID(context.Background(), 123)
	assert.NoError(t, err)
	assert.Equal(t, newQuotaUsed, updatedUser.QuotaUsed)
}

func TestUserRepository_UpdateQuota_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	err := repo.UpdateQuota(context.Background(), 999, 1024)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}
