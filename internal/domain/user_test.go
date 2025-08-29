package domain

import (
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	telegramID := int64(123456789)
	username := "testuser"
	firstName := "Test"
	lastName := "User"

	user := NewUser(telegramID, username, firstName, lastName)

	if user.TelegramID != telegramID {
		t.Errorf("Expected TelegramID %d, got %d", telegramID, user.TelegramID)
	}
	if user.Username != username {
		t.Errorf("Expected Username %s, got %s", username, user.Username)
	}
	if user.FirstName != firstName {
		t.Errorf("Expected FirstName %s, got %s", firstName, user.FirstName)
	}
	if user.LastName != lastName {
		t.Errorf("Expected LastName %s, got %s", lastName, user.LastName)
	}
	if user.Status != UserStatusInactive {
		t.Errorf("Expected Status %s, got %s", UserStatusInactive, user.Status)
	}
	if user.QuotaLimit != DefaultQuotaLimit {
		t.Errorf("Expected QuotaLimit %d, got %d", DefaultQuotaLimit, user.QuotaLimit)
	}
	if user.QuotaUsed != 0 {
		t.Errorf("Expected QuotaUsed 0, got %d", user.QuotaUsed)
	}
	if user.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if user.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestUser_ActivateTrial(t *testing.T) {
	user := NewUser(123, "test", "Test", "User")
	originalUpdatedAt := user.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	user.ActivateTrial()

	if user.Status != UserStatusTrial {
		t.Errorf("Expected Status %s, got %s", UserStatusTrial, user.Status)
	}
	if !user.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestUser_GetQuotaRemaining(t *testing.T) {
	tests := []struct {
		name       string
		quotaLimit int64
		quotaUsed  int64
		expected   int64
	}{
		{
			name:       "Full quota remaining",
			quotaLimit: 100,
			quotaUsed:  0,
			expected:   100,
		},
		{
			name:       "Half quota remaining",
			quotaLimit: 100,
			quotaUsed:  50,
			expected:   50,
		},
		{
			name:       "No quota remaining",
			quotaLimit: 100,
			quotaUsed:  100,
			expected:   0,
		},
		{
			name:       "Over quota",
			quotaLimit: 100,
			quotaUsed:  150,
			expected:   -50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				QuotaLimit: tt.quotaLimit,
				QuotaUsed:  tt.quotaUsed,
			}

			result := user.GetQuotaRemaining()
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestUser_HasQuotaRemaining(t *testing.T) {
	tests := []struct {
		name       string
		quotaLimit int64
		quotaUsed  int64
		expected   bool
	}{
		{
			name:       "Has quota remaining",
			quotaLimit: 100,
			quotaUsed:  50,
			expected:   true,
		},
		{
			name:       "No quota remaining",
			quotaLimit: 100,
			quotaUsed:  100,
			expected:   false,
		},
		{
			name:       "Over quota",
			quotaLimit: 100,
			quotaUsed:  150,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				QuotaLimit: tt.quotaLimit,
				QuotaUsed:  tt.quotaUsed,
			}

			result := user.HasQuotaRemaining()
			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "Inactive user",
			status:   UserStatusInactive,
			expected: false,
		},
		{
			name:     "Trial user",
			status:   UserStatusTrial,
			expected: true,
		},
		{
			name:     "Active user",
			status:   UserStatusActive,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Status: tt.status}

			result := user.IsActive()
			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

// New tests for validation methods
func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		wantErr bool
	}{
		{
			name: "Valid user",
			user: &User{
				TelegramID: 123456789,
				Username:   "testuser",
				FirstName:  "Test",
				LastName:   "User",
				Status:     UserStatusInactive,
				QuotaLimit: 100,
				QuotaUsed:  0,
			},
			wantErr: false,
		},
		{
			name: "Invalid telegram ID",
			user: &User{
				TelegramID: 0,
				Username:   "testuser",
				FirstName:  "Test",
				LastName:   "User",
			},
			wantErr: true,
		},
		{
			name: "Empty first name",
			user: &User{
				TelegramID: 123456789,
				Username:   "testuser",
				FirstName:  "",
				LastName:   "User",
			},
			wantErr: true,
		},
		{
			name: "Invalid status",
			user: &User{
				TelegramID: 123456789,
				Username:   "testuser",
				FirstName:  "Test",
				LastName:   "User",
				Status:     "invalid_status",
			},
			wantErr: true,
		},
		{
			name: "Negative quota limit",
			user: &User{
				TelegramID: 123456789,
				Username:   "testuser",
				FirstName:  "Test",
				LastName:   "User",
				QuotaLimit: -100,
			},
			wantErr: true,
		},
		{
			name: "Negative quota used",
			user: &User{
				TelegramID: 123456789,
				Username:   "testuser",
				FirstName:  "Test",
				LastName:   "User",
				QuotaUsed:  -50,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_CanActivateTrial(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		expected bool
	}{
		{
			name: "Can activate trial - inactive user",
			user: &User{
				Status: UserStatusInactive,
			},
			expected: true,
		},
		{
			name: "Cannot activate trial - already trial",
			user: &User{
				Status: UserStatusTrial,
			},
			expected: false,
		},
		{
			name: "Cannot activate trial - already active",
			user: &User{
				Status: UserStatusActive,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.CanActivateTrial()
			if result != tt.expected {
				t.Errorf("User.CanActivateTrial() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestUser_CanUseQuota(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		amount   int64
		expected bool
	}{
		{
			name: "Can use quota - active user with sufficient quota",
			user: &User{
				Status:     UserStatusActive,
				QuotaLimit: 100,
				QuotaUsed:  50,
			},
			amount:   25,
			expected: true,
		},
		{
			name: "Cannot use quota - inactive user",
			user: &User{
				Status:     UserStatusInactive,
				QuotaLimit: 100,
				QuotaUsed:  0,
			},
			amount:   25,
			expected: false,
		},
		{
			name: "Cannot use quota - insufficient quota",
			user: &User{
				Status:     UserStatusActive,
				QuotaLimit: 100,
				QuotaUsed:  80,
			},
			amount:   25,
			expected: false,
		},
		{
			name: "Cannot use quota - negative amount",
			user: &User{
				Status:     UserStatusActive,
				QuotaLimit: 100,
				QuotaUsed:  0,
			},
			amount:   -10,
			expected: false,
		},
		{
			name: "Can use quota - exact remaining quota",
			user: &User{
				Status:     UserStatusActive,
				QuotaLimit: 100,
				QuotaUsed:  80,
			},
			amount:   20,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.CanUseQuota(tt.amount)
			if result != tt.expected {
				t.Errorf("User.CanUseQuota(%d) = %v, expected %v", tt.amount, result, tt.expected)
			}
		})
	}
}

func TestUser_AddQuotaUsage(t *testing.T) {
	tests := []struct {
		name         string
		user         *User
		amount       int64
		expectError  bool
		expectedUsed int64
	}{
		{
			name: "Successfully add quota usage",
			user: &User{
				Status:     UserStatusActive,
				QuotaLimit: 100,
				QuotaUsed:  50,
			},
			amount:       25,
			expectError:  false,
			expectedUsed: 75,
		},
		{
			name: "Error - inactive user",
			user: &User{
				Status:     UserStatusInactive,
				QuotaLimit: 100,
				QuotaUsed:  0,
			},
			amount:       25,
			expectError:  true,
			expectedUsed: 0,
		},
		{
			name: "Error - insufficient quota",
			user: &User{
				Status:     UserStatusActive,
				QuotaLimit: 100,
				QuotaUsed:  80,
			},
			amount:       25,
			expectError:  true,
			expectedUsed: 80,
		},
		{
			name: "Error - negative amount",
			user: &User{
				Status:     UserStatusActive,
				QuotaLimit: 100,
				QuotaUsed:  50,
			},
			amount:       -10,
			expectError:  true,
			expectedUsed: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalUsed := tt.user.QuotaUsed
			err := tt.user.AddQuotaUsage(tt.amount)

			if (err != nil) != tt.expectError {
				t.Errorf("User.AddQuotaUsage(%d) error = %v, expectError %v", tt.amount, err, tt.expectError)
			}

			if !tt.expectError && tt.user.QuotaUsed != tt.expectedUsed {
				t.Errorf("User.AddQuotaUsage(%d) = %d, expected %d", tt.amount, tt.user.QuotaUsed, tt.expectedUsed)
			}

			if tt.expectError && tt.user.QuotaUsed != originalUsed {
				t.Errorf("User.AddQuotaUsage(%d) changed QuotaUsed to %d, expected unchanged %d", tt.amount, tt.user.QuotaUsed, originalUsed)
			}
		})
	}
}

func TestUser_GetQuotaUsagePercentage(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		expected float64
	}{
		{
			name: "0% usage",
			user: &User{
				QuotaLimit: 100,
				QuotaUsed:  0,
			},
			expected: 0.0,
		},
		{
			name: "50% usage",
			user: &User{
				QuotaLimit: 100,
				QuotaUsed:  50,
			},
			expected: 50.0,
		},
		{
			name: "100% usage",
			user: &User{
				QuotaLimit: 100,
				QuotaUsed:  100,
			},
			expected: 100.0,
		},
		{
			name: "Over 100% usage",
			user: &User{
				QuotaLimit: 100,
				QuotaUsed:  150,
			},
			expected: 150.0,
		},
		{
			name: "Zero quota limit",
			user: &User{
				QuotaLimit: 0,
				QuotaUsed:  50,
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.GetQuotaUsagePercentage()
			if result != tt.expected {
				t.Errorf("User.GetQuotaUsagePercentage() = %f, expected %f", result, tt.expected)
			}
		})
	}
}
