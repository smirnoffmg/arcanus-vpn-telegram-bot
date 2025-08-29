package events

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEvent(t *testing.T) {
	userID := int64(12345)
	data := map[string]interface{}{
		"test": "data",
	}

	event := NewEvent(EventUserRegistered, &userID, data)

	assert.NotEmpty(t, event.ID)
	assert.Equal(t, EventUserRegistered, event.Type)
	assert.Equal(t, "1.0", event.Version)
	assert.Equal(t, userID, *event.UserID)
	assert.Equal(t, data, event.Data)
	assert.NotNil(t, event.Metadata)
	assert.WithinDuration(t, time.Now(), event.Timestamp, time.Second)
}

func TestEventSerialization(t *testing.T) {
	userID := int64(12345)
	data := map[string]interface{}{
		"telegram_id": userID,
		"username":    "testuser",
	}

	originalEvent := NewEvent(EventUserRegistered, &userID, data)
	originalEvent.SetCorrelationID("test-correlation-id")
	originalEvent.AddMetadata("source", "test")

	// Serialize to JSON
	jsonData, err := originalEvent.ToJSON()
	require.NoError(t, err)

	// Deserialize from JSON
	deserializedEvent, err := FromJSON(jsonData)
	require.NoError(t, err)

	// Verify all fields are preserved
	assert.Equal(t, originalEvent.ID, deserializedEvent.ID)
	assert.Equal(t, originalEvent.Type, deserializedEvent.Type)
	assert.Equal(t, originalEvent.Version, deserializedEvent.Version)
	assert.Equal(t, *originalEvent.UserID, *deserializedEvent.UserID)
	assert.Equal(t, *originalEvent.CorrelationID, *deserializedEvent.CorrelationID)
	
	// JSON deserialization converts numbers to float64, so we need to check differently
	assert.Equal(t, float64(userID), deserializedEvent.Data["telegram_id"])
	assert.Equal(t, "testuser", deserializedEvent.Data["username"])
	assert.Equal(t, originalEvent.Metadata, deserializedEvent.Metadata)
}

func TestEventHelperFunctions(t *testing.T) {
	userID := int64(12345)

	// Test user registered event
	userEvent := NewUserRegisteredEvent(userID, "testuser", "Test", "User", 50*1024*1024)
	assert.Equal(t, EventUserRegistered, userEvent.Type)
	assert.Equal(t, userID, *userEvent.UserID)
	assert.Equal(t, userID, userEvent.Data["telegram_id"])
	assert.Equal(t, "testuser", userEvent.Data["username"])

	// Test trial activated event
	trialEvent := NewUserTrialActivatedEvent(userID, "inactive", "active")
	assert.Equal(t, EventUserTrialActivated, trialEvent.Type)
	assert.Equal(t, userID, *trialEvent.UserID)
	assert.Equal(t, "inactive", trialEvent.Data["previous_status"])
	assert.Equal(t, "active", trialEvent.Data["new_status"])

	// Test quota updated event
	quotaEvent := NewUserQuotaUpdatedEvent(userID, 1024, 2048)
	assert.Equal(t, EventUserQuotaUpdated, quotaEvent.Type)
	assert.Equal(t, userID, *quotaEvent.UserID)
	assert.Equal(t, int64(1024), quotaEvent.Data["previous_quota"])
	assert.Equal(t, int64(2048), quotaEvent.Data["new_quota"])
	assert.Equal(t, int64(1024), quotaEvent.Data["quota_delta"])

	// Test bot message received event
	msgEvent := NewBotMessageReceivedEvent(userID, "testuser", 67890, 1, "/start", "start")
	assert.Equal(t, EventBotMessageReceived, msgEvent.Type)
	assert.Equal(t, userID, *msgEvent.UserID)
	assert.Equal(t, int64(67890), msgEvent.Data["chat_id"])
	assert.Equal(t, "start", msgEvent.Data["command"])

	// Test bot callback received event
	callbackEvent := NewBotCallbackReceivedEvent(userID, "testuser", 67890, 1, "trial")
	assert.Equal(t, EventBotCallbackReceived, callbackEvent.Type)
	assert.Equal(t, userID, *callbackEvent.UserID)
	assert.Equal(t, "trial", callbackEvent.Data["callback_data"])
}

func TestMockPublisher(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests
	
	publisher := NewMockPublisher(logger)
	
	// Test single event publishing
	event := NewUserRegisteredEvent(12345, "testuser", "Test", "User", 1024)
	ctx := context.Background()
	
	err := publisher.Publish(ctx, event)
	require.NoError(t, err)
	
	publishedEvents := publisher.GetPublishedEvents()
	assert.Len(t, publishedEvents, 1)
	assert.Equal(t, event.ID, publishedEvents[0].ID)
	assert.Equal(t, event.Type, publishedEvents[0].Type)
	
	// Test batch publishing
	events := []*Event{
		NewUserTrialActivatedEvent(12345, "inactive", "active"),
		NewUserQuotaUpdatedEvent(12345, 0, 512),
	}
	
	err = publisher.PublishBatch(ctx, events)
	require.NoError(t, err)
	
	publishedEvents = publisher.GetPublishedEvents()
	assert.Len(t, publishedEvents, 3) // 1 from before + 2 from batch
	
	// Test error mode
	publisher.SetShouldError(true)
	err = publisher.Publish(ctx, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock error")
	
	err = publisher.PublishBatch(ctx, events)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock batch error")
	
	// Test clear events
	publisher.SetShouldError(false)
	publisher.ClearEvents()
	publishedEvents = publisher.GetPublishedEvents()
	assert.Len(t, publishedEvents, 0)
}

func TestEventService(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests
	
	publisher := NewMockPublisher(logger)
	service := NewEventService(publisher, logger)
	ctx := context.Background()
	
	// Test user registered event
	err := service.PublishUserRegistered(ctx, 12345, "testuser", "Test", "User", 1024)
	require.NoError(t, err)
	
	// Test trial activated event
	err = service.PublishUserTrialActivated(ctx, 12345, "inactive", "active")
	require.NoError(t, err)
	
	// Test quota updated event
	err = service.PublishUserQuotaUpdated(ctx, 12345, 512, 1024)
	require.NoError(t, err)
	
	// Test bot message received event
	err = service.PublishBotMessageReceived(ctx, 12345, "testuser", 67890, 1, "/start", "start")
	require.NoError(t, err)
	
	// Test bot callback received event
	err = service.PublishBotCallbackReceived(ctx, 12345, "testuser", 67890, 1, "trial")
	require.NoError(t, err)
	
	// Test system error event
	metadata := map[string]string{"component": "test"}
	err = service.PublishSystemError(ctx, "test_error", "Test error message", metadata)
	require.NoError(t, err)
	
	// Test system startup event
	err = service.PublishSystemStartup(ctx, "1.0.0", metadata)
	require.NoError(t, err)
	
	// Test system shutdown event
	err = service.PublishSystemShutdown(ctx, "graceful shutdown", metadata)
	require.NoError(t, err)
	
	// Verify all events were published
	publishedEvents := publisher.GetPublishedEvents()
	assert.Len(t, publishedEvents, 8)
	
	// Verify event types
	eventTypes := make(map[EventType]int)
	for _, event := range publishedEvents {
		eventTypes[event.Type]++
	}
	
	assert.Equal(t, 1, eventTypes[EventUserRegistered])
	assert.Equal(t, 1, eventTypes[EventUserTrialActivated])
	assert.Equal(t, 1, eventTypes[EventUserQuotaUpdated])
	assert.Equal(t, 1, eventTypes[EventBotMessageReceived])
	assert.Equal(t, 1, eventTypes[EventBotCallbackReceived])
	assert.Equal(t, 1, eventTypes[EventSystemError])
	assert.Equal(t, 1, eventTypes[EventSystemStartup])
	assert.Equal(t, 1, eventTypes[EventSystemShutdown])
}

func TestEventServiceErrorHandling(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests
	
	publisher := NewMockPublisher(logger)
	publisher.SetShouldError(true) // Make publisher return errors
	service := NewEventService(publisher, logger)
	ctx := context.Background()
	
	// All service methods should return errors when publisher fails
	err := service.PublishUserRegistered(ctx, 12345, "testuser", "Test", "User", 1024)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to publish user registered event")
	
	err = service.PublishUserTrialActivated(ctx, 12345, "inactive", "active")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to publish user trial activated event")
	
	err = service.PublishUserQuotaUpdated(ctx, 12345, 512, 1024)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to publish user quota updated event")
}