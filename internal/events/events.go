package events

import (
	"encoding/json"
	"time"
)

// EventType represents the type of event that occurred
type EventType string

const (
	// User Events
	EventUserRegistered    EventType = "user.registered"
	EventUserTrialActivated EventType = "user.trial_activated"
	EventUserQuotaUpdated   EventType = "user.quota_updated"
	EventUserStatusChanged  EventType = "user.status_changed"
	
	// Bot Events
	EventBotMessageReceived EventType = "bot.message_received"
	EventBotCallbackReceived EventType = "bot.callback_received"
	EventBotCommandExecuted  EventType = "bot.command_executed"
	
	// System Events
	EventSystemError        EventType = "system.error"
	EventSystemStartup      EventType = "system.startup"
	EventSystemShutdown     EventType = "system.shutdown"
)

// Event represents a domain event in the system
type Event struct {
	// Event metadata
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	
	// Event context
	UserID       *int64             `json:"user_id,omitempty"`
	SessionID    *string            `json:"session_id,omitempty"`
	CorrelationID *string           `json:"correlation_id,omitempty"`
	
	// Event payload
	Data     map[string]interface{} `json:"data"`
	Metadata map[string]string      `json:"metadata,omitempty"`
}

// NewEvent creates a new event with required fields
func NewEvent(eventType EventType, userID *int64, data map[string]interface{}) *Event {
	return &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Version:   "1.0",
		Timestamp: time.Now().UTC(),
		UserID:    userID,
		Data:      data,
		Metadata:  make(map[string]string),
	}
}

// ToJSON serializes the event to JSON
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON deserializes an event from JSON
func FromJSON(data []byte) (*Event, error) {
	var event Event
	err := json.Unmarshal(data, &event)
	return &event, err
}

// SetCorrelationID sets the correlation ID for request tracing
func (e *Event) SetCorrelationID(correlationID string) {
	e.CorrelationID = &correlationID
}

// SetSessionID sets the session ID for user session tracking
func (e *Event) SetSessionID(sessionID string) {
	e.SessionID = &sessionID
}

// AddMetadata adds metadata to the event
func (e *Event) AddMetadata(key, value string) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
}

// GetPartitionKey returns the partition key for Kafka partitioning
// Uses UserID for user-specific events, falls back to event type
func (e *Event) GetPartitionKey() string {
	if e.UserID != nil {
		return string(rune(*e.UserID)) // Convert user ID to string for partitioning
	}
	return string(e.Type) // System events partitioned by type
}

// generateEventID generates a unique event ID
// In production, you might want to use UUID or snowflake IDs
func generateEventID() string {
	return time.Now().Format("20060102150405.000000")
}

// UserRegisteredEventData represents data for user registration event
type UserRegisteredEventData struct {
	TelegramID int64  `json:"telegram_id"`
	Username   string `json:"username"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	QuotaLimit int64  `json:"quota_limit"`
}

// UserTrialActivatedEventData represents data for trial activation event
type UserTrialActivatedEventData struct {
	TelegramID     int64     `json:"telegram_id"`
	PreviousStatus string    `json:"previous_status"`
	NewStatus      string    `json:"new_status"`
	ActivatedAt    time.Time `json:"activated_at"`
}

// UserQuotaUpdatedEventData represents data for quota update event
type UserQuotaUpdatedEventData struct {
	TelegramID      int64 `json:"telegram_id"`
	PreviousQuota   int64 `json:"previous_quota"`
	NewQuota        int64 `json:"new_quota"`
	QuotaDelta      int64 `json:"quota_delta"`
}

// BotMessageReceivedEventData represents data for bot message event
type BotMessageReceivedEventData struct {
	TelegramID int64  `json:"telegram_id"`
	Username   string `json:"username"`
	ChatID     int64  `json:"chat_id"`
	MessageID  int    `json:"message_id"`
	Text       string `json:"text"`
	Command    string `json:"command,omitempty"`
}

// BotCallbackReceivedEventData represents data for bot callback event
type BotCallbackReceivedEventData struct {
	TelegramID   int64  `json:"telegram_id"`
	Username     string `json:"username"`
	ChatID       int64  `json:"chat_id"`
	MessageID    int    `json:"message_id"`
	CallbackData string `json:"callback_data"`
}

// Helper functions to create specific events

// NewUserRegisteredEvent creates a user registration event
func NewUserRegisteredEvent(userID int64, username, firstName, lastName string, quotaLimit int64) *Event {
	data := map[string]interface{}{
		"telegram_id": userID,
		"username":    username,
		"first_name":  firstName,
		"last_name":   lastName,
		"quota_limit": quotaLimit,
	}
	return NewEvent(EventUserRegistered, &userID, data)
}

// NewUserTrialActivatedEvent creates a trial activation event
func NewUserTrialActivatedEvent(userID int64, previousStatus, newStatus string) *Event {
	data := map[string]interface{}{
		"telegram_id":      userID,
		"previous_status":  previousStatus,
		"new_status":       newStatus,
		"activated_at":     time.Now().UTC(),
	}
	return NewEvent(EventUserTrialActivated, &userID, data)
}

// NewUserQuotaUpdatedEvent creates a quota update event
func NewUserQuotaUpdatedEvent(userID int64, previousQuota, newQuota int64) *Event {
	data := map[string]interface{}{
		"telegram_id":     userID,
		"previous_quota":  previousQuota,
		"new_quota":       newQuota,
		"quota_delta":     newQuota - previousQuota,
	}
	return NewEvent(EventUserQuotaUpdated, &userID, data)
}

// NewBotMessageReceivedEvent creates a bot message received event
func NewBotMessageReceivedEvent(userID int64, username string, chatID int64, messageID int, text, command string) *Event {
	data := map[string]interface{}{
		"telegram_id": userID,
		"username":    username,
		"chat_id":     chatID,
		"message_id":  messageID,
		"text":        text,
		"command":     command,
	}
	return NewEvent(EventBotMessageReceived, &userID, data)
}

// NewBotCallbackReceivedEvent creates a bot callback received event
func NewBotCallbackReceivedEvent(userID int64, username string, chatID int64, messageID int, callbackData string) *Event {
	data := map[string]interface{}{
		"telegram_id":   userID,
		"username":      username,
		"chat_id":       chatID,
		"message_id":    messageID,
		"callback_data": callbackData,
	}
	return NewEvent(EventBotCallbackReceived, &userID, data)
}