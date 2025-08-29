package events

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// Service handles event publishing with business logic
type Service struct {
	publisher Publisher
	logger    *logrus.Logger
}

// NewEventService creates a new event service
func NewEventService(publisher Publisher, logger *logrus.Logger) *Service {
	return &Service{
		publisher: publisher,
		logger:    logger,
	}
}

// PublishUserRegistered publishes a user registration event
func (s *Service) PublishUserRegistered(ctx context.Context, userID int64, username, firstName, lastName string, quotaLimit int64) error {
	event := NewUserRegisteredEvent(userID, username, firstName, lastName, quotaLimit)
	
	if err := s.publisher.Publish(ctx, event); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"event_type": event.Type,
			"user_id":    userID,
		}).Error("Failed to publish user registered event")
		return fmt.Errorf("failed to publish user registered event: %w", err)
	}
	
	s.logger.WithFields(logrus.Fields{
		"event_id":   event.ID,
		"event_type": event.Type,
		"user_id":    userID,
		"username":   username,
	}).Info("User registered event published")
	
	return nil
}

// PublishUserTrialActivated publishes a trial activation event
func (s *Service) PublishUserTrialActivated(ctx context.Context, userID int64, previousStatus, newStatus string) error {
	event := NewUserTrialActivatedEvent(userID, previousStatus, newStatus)
	
	if err := s.publisher.Publish(ctx, event); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"event_type": event.Type,
			"user_id":    userID,
		}).Error("Failed to publish user trial activated event")
		return fmt.Errorf("failed to publish user trial activated event: %w", err)
	}
	
	s.logger.WithFields(logrus.Fields{
		"event_id":        event.ID,
		"event_type":      event.Type,
		"user_id":         userID,
		"previous_status": previousStatus,
		"new_status":      newStatus,
	}).Info("User trial activated event published")
	
	return nil
}

// PublishUserQuotaUpdated publishes a quota update event
func (s *Service) PublishUserQuotaUpdated(ctx context.Context, userID int64, previousQuota, newQuota int64) error {
	event := NewUserQuotaUpdatedEvent(userID, previousQuota, newQuota)
	
	if err := s.publisher.Publish(ctx, event); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"event_type": event.Type,
			"user_id":    userID,
		}).Error("Failed to publish user quota updated event")
		return fmt.Errorf("failed to publish user quota updated event: %w", err)
	}
	
	s.logger.WithFields(logrus.Fields{
		"event_id":        event.ID,
		"event_type":      event.Type,
		"user_id":         userID,
		"previous_quota":  previousQuota,
		"new_quota":       newQuota,
		"quota_delta":     newQuota - previousQuota,
	}).Info("User quota updated event published")
	
	return nil
}

// PublishBotMessageReceived publishes a bot message received event
func (s *Service) PublishBotMessageReceived(ctx context.Context, userID int64, username string, chatID int64, messageID int, text, command string) error {
	event := NewBotMessageReceivedEvent(userID, username, chatID, messageID, text, command)
	
	if err := s.publisher.Publish(ctx, event); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"event_type": event.Type,
			"user_id":    userID,
		}).Error("Failed to publish bot message received event")
		return fmt.Errorf("failed to publish bot message received event: %w", err)
	}
	
	s.logger.WithFields(logrus.Fields{
		"event_id":   event.ID,
		"event_type": event.Type,
		"user_id":    userID,
		"chat_id":    chatID,
		"command":    command,
	}).Debug("Bot message received event published")
	
	return nil
}

// PublishBotCallbackReceived publishes a bot callback received event  
func (s *Service) PublishBotCallbackReceived(ctx context.Context, userID int64, username string, chatID int64, messageID int, callbackData string) error {
	event := NewBotCallbackReceivedEvent(userID, username, chatID, messageID, callbackData)
	
	if err := s.publisher.Publish(ctx, event); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"event_type": event.Type,
			"user_id":    userID,
		}).Error("Failed to publish bot callback received event")
		return fmt.Errorf("failed to publish bot callback received event: %w", err)
	}
	
	s.logger.WithFields(logrus.Fields{
		"event_id":      event.ID,
		"event_type":    event.Type,
		"user_id":       userID,
		"chat_id":       chatID,
		"callback_data": callbackData,
	}).Debug("Bot callback received event published")
	
	return nil
}

// PublishSystemError publishes a system error event
func (s *Service) PublishSystemError(ctx context.Context, errorType, errorMessage string, metadata map[string]string) error {
	data := map[string]interface{}{
		"error_type":    errorType,
		"error_message": errorMessage,
	}
	
	event := NewEvent(EventSystemError, nil, data)
	for k, v := range metadata {
		event.AddMetadata(k, v)
	}
	
	if err := s.publisher.Publish(ctx, event); err != nil {
		s.logger.WithError(err).WithField("event_type", event.Type).Error("Failed to publish system error event")
		return fmt.Errorf("failed to publish system error event: %w", err)
	}
	
	s.logger.WithFields(logrus.Fields{
		"event_id":      event.ID,
		"event_type":    event.Type,
		"error_type":    errorType,
		"error_message": errorMessage,
	}).Warn("System error event published")
	
	return nil
}

// PublishSystemStartup publishes a system startup event
func (s *Service) PublishSystemStartup(ctx context.Context, version string, metadata map[string]string) error {
	data := map[string]interface{}{
		"version": version,
	}
	
	event := NewEvent(EventSystemStartup, nil, data)
	for k, v := range metadata {
		event.AddMetadata(k, v)
	}
	
	if err := s.publisher.Publish(ctx, event); err != nil {
		s.logger.WithError(err).WithField("event_type", event.Type).Error("Failed to publish system startup event")
		return fmt.Errorf("failed to publish system startup event: %w", err)
	}
	
	s.logger.WithFields(logrus.Fields{
		"event_id":   event.ID,
		"event_type": event.Type,
		"version":    version,
	}).Info("System startup event published")
	
	return nil
}

// PublishSystemShutdown publishes a system shutdown event
func (s *Service) PublishSystemShutdown(ctx context.Context, reason string, metadata map[string]string) error {
	data := map[string]interface{}{
		"reason": reason,
	}
	
	event := NewEvent(EventSystemShutdown, nil, data)
	for k, v := range metadata {
		event.AddMetadata(k, v)
	}
	
	if err := s.publisher.Publish(ctx, event); err != nil {
		s.logger.WithError(err).WithField("event_type", event.Type).Error("Failed to publish system shutdown event")
		return fmt.Errorf("failed to publish system shutdown event: %w", err)
	}
	
	s.logger.WithFields(logrus.Fields{
		"event_id":   event.ID,
		"event_type": event.Type,
		"reason":     reason,
	}).Info("System shutdown event published")
	
	return nil
}

// Close closes the underlying publisher
func (s *Service) Close() error {
	return s.publisher.Close()
}