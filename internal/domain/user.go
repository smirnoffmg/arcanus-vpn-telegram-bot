package domain

import (
	"fmt"
	"time"
)

// User represents a VPN bot user
type User struct {
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

// UserStatus constants
const (
	UserStatusInactive = "inactive"
	UserStatusTrial    = "trial"
	UserStatusActive   = "active"
)

// DefaultQuotaLimit is 50MB in bytes
const DefaultQuotaLimit = 52428800

// NewUser creates a new user with default values
func NewUser(telegramID int64, username, firstName, lastName string) *User {
	now := time.Now()
	return &User{
		TelegramID: telegramID,
		Username:   username,
		FirstName:  firstName,
		LastName:   lastName,
		Status:     UserStatusInactive,
		QuotaLimit: DefaultQuotaLimit,
		QuotaUsed:  0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// ActivateTrial activates the user's trial
func (u *User) ActivateTrial() {
	u.Status = UserStatusTrial
	u.UpdatedAt = time.Now()
}

// GetQuotaRemaining returns the remaining quota in bytes
func (u *User) GetQuotaRemaining() int64 {
	return u.QuotaLimit - u.QuotaUsed
}

// HasQuotaRemaining checks if user has quota remaining
func (u *User) HasQuotaRemaining() bool {
	return u.GetQuotaRemaining() > 0
}

// IsActive checks if user is active or on trial
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive || u.Status == UserStatusTrial
}

// Validate validates user data
func (u *User) Validate() error {
	if u.TelegramID <= 0 {
		return ValidationError{Field: "telegram_id", Message: "must be positive"}
	}
	if u.FirstName == "" {
		return ValidationError{Field: "first_name", Message: "cannot be empty"}
	}
	if u.Status != UserStatusInactive && u.Status != UserStatusTrial && u.Status != UserStatusActive {
		return ValidationError{Field: "status", Message: fmt.Sprintf("invalid status: %s", u.Status)}
	}
	if u.QuotaLimit < 0 {
		return ValidationError{Field: "quota_limit", Message: "cannot be negative"}
	}
	if u.QuotaUsed < 0 {
		return ValidationError{Field: "quota_used", Message: "cannot be negative"}
	}
	return nil
}

// CanActivateTrial checks if the user can activate trial
func (u *User) CanActivateTrial() bool {
	return u.Status == UserStatusInactive
}

// CanUseQuota checks if the user can use the specified amount of quota
func (u *User) CanUseQuota(amount int64) bool {
	if !u.IsActive() {
		return false
	}
	if amount <= 0 {
		return false
	}
	return u.GetQuotaRemaining() >= amount
}

// AddQuotaUsage adds quota usage if possible
func (u *User) AddQuotaUsage(amount int64) error {
	if !u.CanUseQuota(amount) {
		return QuotaExceededError{Used: u.QuotaUsed + amount, Limit: u.QuotaLimit}
	}
	u.QuotaUsed += amount
	u.UpdatedAt = time.Now()
	return nil
}

// GetQuotaUsagePercentage returns the percentage of quota used
func (u *User) GetQuotaUsagePercentage() float64 {
	if u.QuotaLimit == 0 {
		return 0.0
	}
	return float64(u.QuotaUsed) / float64(u.QuotaLimit) * 100.0
}
