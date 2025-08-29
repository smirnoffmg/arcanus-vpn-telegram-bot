package bot

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// AuditLog represents a security audit event
type AuditLog struct {
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip,omitempty"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Details   string    `json:"details,omitempty"`
	SessionID string    `json:"session_id,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
}

// AuditLogger handles security audit logging
type AuditLogger struct {
	logger *logrus.Logger
}

// NewAuditLogger creates a new audit logger instance
func NewAuditLogger(logger *logrus.Logger) *AuditLogger {
	return &AuditLogger{logger: logger}
}

// LogEvent logs an audit event
func (al *AuditLogger) LogEvent(userID int64, username, action string, success bool, err error, details map[string]interface{}) {
	audit := AuditLog{
		UserID:    userID,
		Username:  username,
		Action:    action,
		Timestamp: time.Now(),
		Success:   success,
	}

	if err != nil {
		audit.Error = err.Error()
	}

	if details != nil {
		if detailsJSON, err := json.Marshal(details); err == nil {
			audit.Details = string(detailsJSON)
		}
	}

	// Log with structured fields for easy filtering
	fields := logrus.Fields{
		"audit":     true,
		"user_id":   audit.UserID,
		"username":  audit.Username,
		"action":    audit.Action,
		"success":   audit.Success,
		"timestamp": audit.Timestamp,
	}

	if audit.Error != "" {
		fields["error"] = audit.Error
	}

	if audit.Details != "" {
		fields["details"] = audit.Details
	}

	level := logrus.InfoLevel
	if !success {
		level = logrus.WarnLevel
	}

	al.logger.WithFields(fields).Log(level, "Audit event")
}

// LogUserRegistration logs user registration events
func (al *AuditLogger) LogUserRegistration(userID int64, username, firstName, lastName string, success bool, err error) {
	details := map[string]interface{}{
		"first_name": firstName,
		"last_name":  lastName,
		"event_type": "registration",
	}

	al.LogEvent(userID, username, "user_registration", success, err, details)
}

// LogCommand logs command execution events
func (al *AuditLogger) LogCommand(userID int64, username, command string, success bool, err error) {
	details := map[string]interface{}{
		"command":    command,
		"event_type": "command",
	}

	al.LogEvent(userID, username, "command_execution", success, err, details)
}

// LogCallback logs callback query events
func (al *AuditLogger) LogCallback(userID int64, username, callbackData string, success bool, err error) {
	details := map[string]interface{}{
		"callback_data": callbackData,
		"event_type":    "callback",
	}

	al.LogEvent(userID, username, "callback_query", success, err, details)
}

// LogQuotaUsage logs quota usage events
func (al *AuditLogger) LogQuotaUsage(userID int64, username string, quotaUsed int64, success bool, err error) {
	details := map[string]interface{}{
		"quota_used": quotaUsed,
		"event_type": "quota_usage",
	}

	al.LogEvent(userID, username, "quota_usage", success, err, details)
}

// LogSecurityEvent logs security-related events
func (al *AuditLogger) LogSecurityEvent(userID int64, username, eventType string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["event_type"] = "security"

	al.LogEvent(userID, username, fmt.Sprintf("security_%s", eventType), true, nil, details)
}
