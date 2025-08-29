package repository

import (
	"context"
	"testing"

	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/domain"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTransactionTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	
	// Auto-migrate the User model
	err = db.AutoMigrate(&domain.User{})
	assert.NoError(t, err)
	
	return db
}

func TestTransactionManager_WithTransaction(t *testing.T) {
	db := setupTransactionTestDB(t)
	tm := NewTransactionManager(db)
	
	t.Run("Successful transaction commits", func(t *testing.T) {
		user := domain.NewUser(12345, "testuser", "Test", "User")
		
		err := tm.WithTransaction(context.Background(), func(ctx context.Context, tx domain.Transaction) error {
			// Within WithTransaction, we need to use a transaction-aware repository
			// We'll simulate this by getting the tx object and using it directly
			txImpl := tx.(*Transaction)
			txRepo := NewUserRepositoryWithTx(txImpl.tx)
			return txRepo.Create(ctx, user)
		})
		
		assert.NoError(t, err)
		
		// Verify user was created
		repo := NewUserRepository(db)
		createdUser, err := repo.GetByTelegramID(context.Background(), 12345)
		assert.NoError(t, err)
		assert.Equal(t, user.TelegramID, createdUser.TelegramID)
	})
	
	t.Run("Failed transaction rolls back", func(t *testing.T) {
		user := domain.NewUser(67890, "testuser2", "Test", "User")
		
		err := tm.WithTransaction(context.Background(), func(ctx context.Context, tx domain.Transaction) error {
			// Within WithTransaction, we need to use a transaction-aware repository
			txImpl := tx.(*Transaction)
			txRepo := NewUserRepositoryWithTx(txImpl.tx)
			if err := txRepo.Create(ctx, user); err != nil {
				return err
			}
			// Force an error to trigger rollback
			return domain.ErrInternalError
		})
		
		assert.Error(t, err)
		assert.Equal(t, domain.ErrInternalError, err)
		
		// Verify user was NOT created (rolled back)
		repo := NewUserRepository(db)
		_, err = repo.GetByTelegramID(context.Background(), 67890)
		assert.Error(t, err)
		assert.IsType(t, domain.UserNotFoundError{}, err)
	})
}

func TestTransactionManager_BeginTx(t *testing.T) {
	db := setupTransactionTestDB(t)
	tm := NewTransactionManager(db)
	
	t.Run("Begin and commit transaction", func(t *testing.T) {
		ctx := context.Background()
		
		txRepo, tx, err := tm.BeginTx(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, txRepo)
		assert.NotNil(t, tx)
		
		// Create user within transaction
		user := domain.NewUser(11111, "txuser", "Transaction", "User")
		err = txRepo.Create(ctx, user)
		assert.NoError(t, err)
		
		// Commit transaction
		err = tx.Commit()
		assert.NoError(t, err)
		
		// Verify user was created
		repo := NewUserRepository(db)
		createdUser, err := repo.GetByTelegramID(ctx, 11111)
		assert.NoError(t, err)
		assert.Equal(t, user.TelegramID, createdUser.TelegramID)
	})
	
	t.Run("Begin and rollback transaction", func(t *testing.T) {
		ctx := context.Background()
		
		txRepo, tx, err := tm.BeginTx(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, txRepo)
		assert.NotNil(t, tx)
		
		// Create user within transaction
		user := domain.NewUser(22222, "rollbackuser", "Rollback", "User")
		err = txRepo.Create(ctx, user)
		assert.NoError(t, err)
		
		// Rollback transaction
		err = tx.Rollback()
		assert.NoError(t, err)
		
		// Verify user was NOT created (rolled back)
		repo := NewUserRepository(db)
		_, err = repo.GetByTelegramID(ctx, 22222)
		assert.Error(t, err)
		assert.IsType(t, domain.UserNotFoundError{}, err)
	})
}

func TestTransaction_CommitAndRollback(t *testing.T) {
	db := setupTransactionTestDB(t)
	
	t.Run("Transaction commit", func(t *testing.T) {
		tx := db.Begin()
		transaction := &Transaction{tx: tx}
		
		// Create user within transaction
		user := domain.NewUser(33333, "commituser", "Commit", "User")
		result := tx.Create(user)
		assert.NoError(t, result.Error)
		
		// Commit
		err := transaction.Commit()
		assert.NoError(t, err)
		
		// Verify user exists
		var foundUser domain.User
		result = db.Where("telegram_id = ?", 33333).First(&foundUser)
		assert.NoError(t, result.Error)
	})
	
	t.Run("Transaction rollback", func(t *testing.T) {
		tx := db.Begin()
		transaction := &Transaction{tx: tx}
		
		// Create user within transaction
		user := domain.NewUser(44444, "rollbackuser2", "Rollback", "User")
		result := tx.Create(user)
		assert.NoError(t, result.Error)
		
		// Rollback
		err := transaction.Rollback()
		assert.NoError(t, err)
		
		// Verify user does not exist
		var foundUser domain.User
		result = db.Where("telegram_id = ?", 44444).First(&foundUser)
		assert.Error(t, result.Error)
		assert.Equal(t, gorm.ErrRecordNotFound, result.Error)
	})
}