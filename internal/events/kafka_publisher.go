package events

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
)

// Publisher defines the interface for event publishing
type Publisher interface {
	Publish(ctx context.Context, event *Event) error
	PublishBatch(ctx context.Context, events []*Event) error
	Close() error
}

// KafkaPublisher implements event publishing to Kafka
type KafkaPublisher struct {
	producer *kafka.Producer
	topic    string
	logger   *logrus.Logger
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers           string
	Topic             string
	SecurityProtocol  string
	SaslMechanism     string
	SaslUsername      string
	SaslPassword      string
	EnableIdempotence bool
	Acks              string
	RetryBackoffMs    int
	RequestTimeoutMs  int
}

// NewKafkaPublisher creates a new Kafka event publisher
func NewKafkaPublisher(config KafkaConfig, logger *logrus.Logger) (*KafkaPublisher, error) {
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": config.Brokers,
		"client.id":        "arcanus-vpn-bot",
		"acks":             config.Acks,
		"retries":          3,
		"retry.backoff.ms": config.RetryBackoffMs,
		"request.timeout.ms": config.RequestTimeoutMs,
		"enable.idempotence": config.EnableIdempotence,
		"compression.type":   "snappy",
		"batch.size":        16384,
		"linger.ms":         5,
	}

	// Add security configuration if provided
	if config.SecurityProtocol != "" {
		_ = kafkaConfig.SetKey("security.protocol", config.SecurityProtocol)
	}
	if config.SaslMechanism != "" {
		_ = kafkaConfig.SetKey("sasl.mechanism", config.SaslMechanism)
		_ = kafkaConfig.SetKey("sasl.username", config.SaslUsername)
		_ = kafkaConfig.SetKey("sasl.password", config.SaslPassword)
	}

	producer, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	publisher := &KafkaPublisher{
		producer: producer,
		topic:    config.Topic,
		logger:   logger,
	}

	// Start delivery report handler
	go publisher.handleDeliveryReports()

	return publisher, nil
}

// Publish publishes a single event to Kafka
func (p *KafkaPublisher) Publish(ctx context.Context, event *Event) error {
	// Serialize event to JSON
	eventData, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Create Kafka message with user-based partitioning
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(p.getPartitionKey(event)),
		Value: eventData,
		Headers: []kafka.Header{
			{Key: "event_id", Value: []byte(event.ID)},
			{Key: "event_type", Value: []byte(event.Type)},
			{Key: "version", Value: []byte(event.Version)},
			{Key: "timestamp", Value: []byte(event.Timestamp.Format(time.RFC3339))},
		},
	}

	// Add user ID to headers if available
	if event.UserID != nil {
		message.Headers = append(message.Headers, kafka.Header{
			Key:   "user_id",
			Value: []byte(strconv.FormatInt(*event.UserID, 10)),
		})
	}

	// Add correlation ID to headers if available
	if event.CorrelationID != nil {
		message.Headers = append(message.Headers, kafka.Header{
			Key:   "correlation_id",
			Value: []byte(*event.CorrelationID),
		})
	}

	// Produce message
	deliveryChan := make(chan kafka.Event)
	err = p.producer.Produce(message, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	// Wait for delivery confirmation with timeout
	select {
	case e := <-deliveryChan:
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			return fmt.Errorf("delivery failed: %v", m.TopicPartition.Error)
		}
		
		p.logger.WithFields(logrus.Fields{
			"event_id":   event.ID,
			"event_type": event.Type,
			"user_id":    event.UserID,
			"partition":  m.TopicPartition.Partition,
			"offset":     m.TopicPartition.Offset,
		}).Debug("Event published successfully")
		
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while waiting for delivery")
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout waiting for message delivery")
	}

	return nil
}

// PublishBatch publishes multiple events in batch
func (p *KafkaPublisher) PublishBatch(ctx context.Context, events []*Event) error {
	if len(events) == 0 {
		return nil
	}

	deliveryChans := make([]chan kafka.Event, len(events))
	
	// Produce all messages
	for i, event := range events {
		eventData, err := event.ToJSON()
		if err != nil {
			return fmt.Errorf("failed to serialize event %s: %w", event.ID, err)
		}

		message := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &p.topic,
				Partition: kafka.PartitionAny,
			},
			Key:   []byte(p.getPartitionKey(event)),
			Value: eventData,
			Headers: []kafka.Header{
				{Key: "event_id", Value: []byte(event.ID)},
				{Key: "event_type", Value: []byte(event.Type)},
				{Key: "version", Value: []byte(event.Version)},
			},
		}

		deliveryChan := make(chan kafka.Event)
		deliveryChans[i] = deliveryChan
		
		err = p.producer.Produce(message, deliveryChan)
		if err != nil {
			return fmt.Errorf("failed to produce message for event %s: %w", event.ID, err)
		}
	}

	// Wait for all deliveries
	for i, deliveryChan := range deliveryChans {
		select {
		case e := <-deliveryChan:
			m := e.(*kafka.Message)
			if m.TopicPartition.Error != nil {
				return fmt.Errorf("delivery failed for event %s: %v", events[i].ID, m.TopicPartition.Error)
			}
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for batch delivery")
		case <-time.After(30 * time.Second):
			return fmt.Errorf("timeout waiting for batch delivery")
		}
	}

	p.logger.WithField("batch_size", len(events)).Debug("Event batch published successfully")
	return nil
}

// Close closes the Kafka producer
func (p *KafkaPublisher) Close() error {
	// Flush any pending messages
	remaining := p.producer.Flush(15 * 1000) // 15 seconds
	if remaining > 0 {
		p.logger.WithField("remaining_messages", remaining).Warn("Some messages were not flushed")
	}
	
	p.producer.Close()
	return nil
}

// getPartitionKey returns the partition key for an event
// This ensures events for the same user go to the same partition for ordering
func (p *KafkaPublisher) getPartitionKey(event *Event) string {
	if event.UserID != nil {
		return fmt.Sprintf("user_%d", *event.UserID)
	}
	return string(event.Type)
}

// handleDeliveryReports handles delivery reports in the background
func (p *KafkaPublisher) handleDeliveryReports() {
	for e := range p.producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				p.logger.WithError(ev.TopicPartition.Error).Error("Message delivery failed")
			}
		case kafka.Error:
			p.logger.WithError(ev).Error("Kafka error")
		default:
			// Ignore other event types
		}
	}
}

// MockPublisher implements Publisher interface for testing
type MockPublisher struct {
	events      []*Event
	shouldError bool
	logger      *logrus.Logger
}

// NewMockPublisher creates a new mock publisher for testing
func NewMockPublisher(logger *logrus.Logger) *MockPublisher {
	return &MockPublisher{
		events: make([]*Event, 0),
		logger: logger,
	}
}

// Publish stores the event in memory for testing
func (m *MockPublisher) Publish(ctx context.Context, event *Event) error {
	if m.shouldError {
		return fmt.Errorf("mock error")
	}
	
	m.events = append(m.events, event)
	m.logger.WithFields(logrus.Fields{
		"event_id":   event.ID,
		"event_type": event.Type,
		"user_id":    event.UserID,
	}).Debug("Mock event published")
	
	return nil
}

// PublishBatch stores multiple events in memory for testing
func (m *MockPublisher) PublishBatch(ctx context.Context, events []*Event) error {
	if m.shouldError {
		return fmt.Errorf("mock batch error")
	}
	
	m.events = append(m.events, events...)
	m.logger.WithField("batch_size", len(events)).Debug("Mock event batch published")
	
	return nil
}

// Close is a no-op for mock publisher
func (m *MockPublisher) Close() error {
	return nil
}

// GetPublishedEvents returns all published events (for testing)
func (m *MockPublisher) GetPublishedEvents() []*Event {
	return m.events
}

// SetShouldError makes the mock publisher return errors (for testing)
func (m *MockPublisher) SetShouldError(shouldError bool) {
	m.shouldError = shouldError
}

// ClearEvents clears all stored events (for testing)
func (m *MockPublisher) ClearEvents() {
	m.events = make([]*Event, 0)
}