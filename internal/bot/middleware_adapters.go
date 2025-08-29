package bot

import (
	"time"

	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/middleware"
)

// RateLimiterAdapter adapts the bot's RateLimiter to the middleware interface
type RateLimiterAdapter struct {
	rateLimiter *RateLimiter
}

// NewRateLimiterAdapter creates a new rate limiter adapter
func NewRateLimiterAdapter(rateLimiter *RateLimiter) middleware.RateLimiter {
	return &RateLimiterAdapter{rateLimiter: rateLimiter}
}

// Allow checks if a user is allowed to make a request
func (a *RateLimiterAdapter) Allow(userID int64) bool {
	return a.rateLimiter.Allow(userID)
}

// AuditLoggerAdapter adapts the bot's AuditLogger to the middleware interface
type AuditLoggerAdapter struct {
	auditLogger *AuditLogger
}

// NewAuditLoggerAdapter creates a new audit logger adapter
func NewAuditLoggerAdapter(auditLogger *AuditLogger) middleware.AuditLogger {
	return &AuditLoggerAdapter{auditLogger: auditLogger}
}

// LogAction logs an action performed by a user
func (a *AuditLoggerAdapter) LogAction(userID int64, action string, timestamp time.Time) {
	details := map[string]interface{}{
		"action": action,
		"timestamp": timestamp,
	}
	a.auditLogger.LogSecurityEvent(userID, "", "user_action", details)
}