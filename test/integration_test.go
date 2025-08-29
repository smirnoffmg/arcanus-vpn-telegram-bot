//go:build integration

package test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/domain"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/repository"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDatabase(t *testing.T) (*gorm.DB, func()) {
	// Get database URL from environment
	databaseURL := os.Getenv("DATABASE_URL")

	// Use SQLite for local testing if no DATABASE_URL is provided
	if databaseURL == "" {
		// Use in-memory SQLite for local testing
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err, "Failed to connect to database")

		// Auto migrate the schema
		err = db.AutoMigrate(&domain.User{})
		require.NoError(t, err, "Failed to migrate database")

		// Return cleanup function
		cleanup := func() {
			// Clean up test data
			_ = db.Exec("DELETE FROM users")
		}

		return db, cleanup
	}

	// Use PostgreSQL if DATABASE_URL is provided (for Docker testing)
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to database")

	// Auto migrate the schema
	err = db.AutoMigrate(&domain.User{})
	if err != nil {
		t.Logf("Migration error: %v", err)
		// Try to create table manually if AutoMigrate fails
		err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			telegram_id BIGINT UNIQUE NOT NULL,
			username VARCHAR(255),
			first_name VARCHAR(255),
			last_name VARCHAR(255),
			status VARCHAR(50) DEFAULT 'inactive',
			quota_limit BIGINT DEFAULT 52428800,
			quota_used BIGINT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL
		)`).Error
	}
	require.NoError(t, err, "Failed to migrate database")

	// Return cleanup function
	cleanup := func() {
		// Clean up test data
		_ = db.Exec("DELETE FROM users")
	}

	return db, cleanup
}

func TestIntegration_UserRegistration(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	// Setup dependencies
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	ctx := context.Background()

	t.Run("Register new user", func(t *testing.T) {
		// Register a new user
		user, err := userService.RegisterUser(ctx, 12345, "testuser", "Test", "User")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(12345), user.TelegramID)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "Test", user.FirstName)
		assert.Equal(t, "User", user.LastName)
		assert.Equal(t, domain.UserStatusInactive, user.Status)
		assert.Equal(t, int64(domain.DefaultQuotaLimit), user.QuotaLimit)
		assert.Equal(t, int64(0), user.QuotaUsed)
	})

	t.Run("Register existing user", func(t *testing.T) {
		// Try to register the same user again
		user, err := userService.RegisterUser(ctx, 12345, "testuser", "Test", "User")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(12345), user.TelegramID)
		// Should return existing user, not create new one
	})

	t.Run("Get user", func(t *testing.T) {
		user, err := userService.GetUser(ctx, 12345)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(12345), user.TelegramID)
	})

	t.Run("Get non-existent user", func(t *testing.T) {
		user, err := userService.GetUser(ctx, 99999)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})
}

func TestIntegration_TrialActivation(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	ctx := context.Background()

	t.Run("Activate trial for new user", func(t *testing.T) {
		// Register user first
		user, err := userService.RegisterUser(ctx, 54321, "trialuser", "Trial", "User")
		require.NoError(t, err)
		assert.Equal(t, domain.UserStatusInactive, user.Status)

		// Activate trial
		err = userService.ActivateTrial(ctx, 54321)
		require.NoError(t, err)

		// Verify trial is activated
		user, err = userService.GetUser(ctx, 54321)
		require.NoError(t, err)
		assert.Equal(t, domain.UserStatusTrial, user.Status)
		assert.True(t, user.IsActive())
	})

	t.Run("Activate trial for already active user", func(t *testing.T) {
		// Try to activate trial again
		err := userService.ActivateTrial(ctx, 54321)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUserAlreadyActive)
	})

	t.Run("Activate trial for non-existent user", func(t *testing.T) {
		err := userService.ActivateTrial(ctx, 99999)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})
}

func TestIntegration_QuotaManagement(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	ctx := context.Background()

	t.Run("Update quota for active user", func(t *testing.T) {
		// Register and activate user
		_, err := userService.RegisterUser(ctx, 11111, "quotauser", "Quota", "User")
		require.NoError(t, err)
		err = userService.ActivateTrial(ctx, 11111)
		require.NoError(t, err)

		// Update quota
		err = userService.UpdateQuota(ctx, 11111, 1024*1024) // 1MB
		require.NoError(t, err)

		// Verify quota is updated
		user, err := userService.GetUser(ctx, 11111)
		require.NoError(t, err)
		assert.Equal(t, int64(1024*1024), user.QuotaUsed)
		assert.True(t, user.HasQuotaRemaining())
	})

	t.Run("Update quota for inactive user", func(t *testing.T) {
		// Register user without activating
		_, err := userService.RegisterUser(ctx, 22222, "inactiveuser", "Inactive", "User")
		require.NoError(t, err)

		// Try to update quota
		err = userService.UpdateQuota(ctx, 22222, 1024*1024)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUserNotActive)
	})

	t.Run("Update quota exceeding limit", func(t *testing.T) {
		// Register and activate user
		_, err := userService.RegisterUser(ctx, 33333, "exceeduser", "Exceed", "User")
		require.NoError(t, err)
		err = userService.ActivateTrial(ctx, 33333)
		require.NoError(t, err)

		// Try to update quota beyond limit
		err = userService.UpdateQuota(ctx, 33333, domain.DefaultQuotaLimit+1)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrQuotaExceeded)
	})
}

func TestIntegration_DatabaseOperations(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	ctx := context.Background()

	t.Run("Create user in database", func(t *testing.T) {
		user := domain.NewUser(44444, "dbuser", "DB", "User")
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)

		// Verify user exists in database
		foundUser, err := userRepo.GetByTelegramID(ctx, 44444)
		require.NoError(t, err)
		assert.Equal(t, user.TelegramID, foundUser.TelegramID)
	})

	t.Run("Update user in database", func(t *testing.T) {
		user := domain.NewUser(55555, "updateuser", "Update", "User")
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Update user
		user.FirstName = "Updated"
		err = userRepo.Update(ctx, user)
		require.NoError(t, err)

		// Verify update
		foundUser, err := userRepo.GetByTelegramID(ctx, 55555)
		require.NoError(t, err)
		assert.Equal(t, "Updated", foundUser.FirstName)
	})

	t.Run("Update quota in database", func(t *testing.T) {
		user := domain.NewUser(66666, "quotauser", "Quota", "User")
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Update quota
		err = userRepo.UpdateQuota(ctx, 66666, 2048*1024) // 2MB
		require.NoError(t, err)

		// Verify quota update
		foundUser, err := userRepo.GetByTelegramID(ctx, 66666)
		require.NoError(t, err)
		assert.Equal(t, int64(2048*1024), foundUser.QuotaUsed)
	})
}

func TestIntegration_ConcurrentOperations(t *testing.T) {
	t.Run("Sequential user registrations", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		userRepo := repository.NewUserRepository(db)
		userService := service.NewUserService(userRepo)
		ctx := context.Background()

		const numUsers = 10

		// Register users sequentially (avoiding SQLite concurrency issues)
		for i := 0; i < numUsers; i++ {
			_, err := userService.RegisterUser(ctx, int64(70000+i),
				fmt.Sprintf("user%d", i), "Sequential", "User")
			assert.NoError(t, err)
		}

		// Verify all users were created
		for i := 0; i < numUsers; i++ {
			user, err := userService.GetUser(ctx, int64(70000+i))
			require.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("user%d", i), user.Username)
		}
	})
}

func TestIntegration_DatabaseConnection(t *testing.T) {
	t.Run("Database connectivity", func(t *testing.T) {
		databaseURL := os.Getenv("DATABASE_URL")

		if databaseURL == "" {
			// Test SQLite connection for local testing
			db, err := sql.Open("sqlite3", ":memory:")
			require.NoError(t, err)
			defer db.Close()

			// Test connection
			err = db.Ping()
			require.NoError(t, err, "Database connection failed")

			// Test query
			var result int
			err = db.QueryRow("SELECT 1").Scan(&result)
			require.NoError(t, err)
			assert.Equal(t, 1, result)
		} else {
			// Test PostgreSQL connection for Docker testing
			db, err := sql.Open("postgres", databaseURL)
			require.NoError(t, err)
			defer db.Close()

			// Test connection
			err = db.Ping()
			require.NoError(t, err, "Database connection failed")

			// Test query
			var result int
			err = db.QueryRow("SELECT 1").Scan(&result)
			require.NoError(t, err)
			assert.Equal(t, 1, result)
		}
	})
}
