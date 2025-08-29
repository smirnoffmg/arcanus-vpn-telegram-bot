package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestUser is the User model without DeletedAt field for testing
type TestUser struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TelegramID int64     `json:"telegram_id" gorm:"uniqueIndex;not null"`
	Username   string    `json:"username" gorm:"size:255"`
	FirstName  string    `json:"first_name" gorm:"size:255"`
	LastName   string    `json:"last_name" gorm:"size:255"`
	Status     string    `json:"status" gorm:"size:50;default:inactive"`
	QuotaLimit int64     `json:"quota_limit" gorm:"default:52428800"` // 50MB
	QuotaUsed  int64     `json:"quota_used" gorm:"default:0"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func main() {
	// Load environment variables
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Connected to database successfully")

	// Test AutoMigrate with TestUser model
	fmt.Println("Running AutoMigrate with TestUser model...")
	if err := db.AutoMigrate(&TestUser{}); err != nil {
		log.Fatalf("Failed to run AutoMigrate: %v", err)
	}

	fmt.Println("AutoMigrate completed successfully")

	// Test creating a table manually
	fmt.Println("Testing table creation...")
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	// Test a simple query
	var count int64
	if err := db.Model(&TestUser{}).Count(&count).Error; err != nil {
		log.Fatalf("Failed to count users: %v", err)
	}

	fmt.Printf("User table exists and has %d records\n", count)

	// Close connection
	if err := sqlDB.Close(); err != nil {
		log.Printf("Failed to close database connection: %v", err)
	}

	fmt.Println("Migration completed successfully")
}
