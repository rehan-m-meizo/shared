package kf

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"shared/constants"
	"shared/utils"
	"sync"
	"time"
)

type KafkaPublisher struct {
	producer sarama.SyncProducer
	redis    *redis.Client
}

var (
	singletonPub *KafkaPublisher
	pubInitOnce  sync.Once
	pubInitErr   error
	ctx          = context.Background()
)

const (
	redisTTL = 24 * time.Hour
)

// InitKafkaPublisher initializes the singleton publisher manually
func InitKafkaPublisher(redisClient *redis.Client) error {
	pubInitOnce.Do(func() {
		cfg := &Config{
			Brokers:  utils.SplitAndTrim(constants.KafkaBrokers),
			Username: constants.KafkaUser,
			Password: constants.KafkaPass,
			GroupID:  "", // not needed for publisher
		}

		saramaCfg := newSaramaConfig(cfg)
		producer, err := sarama.NewSyncProducer(cfg.Brokers, saramaCfg)
		if err != nil {
			pubInitErr = fmt.Errorf("failed to create Kafka producer: %w", err)
			return
		}

		singletonPub = &KafkaPublisher{
			producer: producer,
			redis:    redisClient,
		}
	})
	return pubInitErr
}

// GetKafkaPublisher returns the singleton instance (after Init)
func GetKafkaPublisher() (*KafkaPublisher, error) {
	if singletonPub == nil {
		return nil, fmt.Errorf("KafkaPublisher not initialized. Call InitKafkaPublisher() first")
	}
	return singletonPub, nil
}

// Publish sends a message to Kafka after caching it in Redis.
// If Kafka fails, Redis retains the message as fallback for consumer use.
func (kp *KafkaPublisher) Publish(topic string, key string, data any, retries int) error {
	value, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("❌ failed to marshal data: %w", err)
	}

	redisKey := fmt.Sprintf("kafka:%s:%s", topic, key)

	// Step 1: Cache to Redis
	if err := kp.redis.Set(ctx, redisKey, value, redisTTL).Err(); err != nil {
		return fmt.Errorf("❌ failed to cache to Redis: %w", err)
	}

	// Step 2: Try Kafka publish
	for attempt := 0; attempt <= retries; attempt++ {
		_, _, err = kp.producer.SendMessage(&sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(key),
			Value: sarama.ByteEncoder(value),
		})
		if err == nil {
			// Optionally delete or shorten TTL after Kafka success
			_ = kp.redis.Expire(ctx, redisKey, 1*time.Hour).Err()
			return nil
		}

		// Optional: backoff logic could go here
	}

	// Step 3: Fallback to DLQ if Kafka fails
	dlxTopic := fmt.Sprintf("dlq.%s", topic)
	_, _, dlqErr := kp.producer.SendMessage(&sarama.ProducerMessage{
		Topic: dlxTopic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	})
	if dlqErr != nil {
		return fmt.Errorf("❌ failed to publish to DLQ topic %s: %v (original error: %v)", dlxTopic, dlqErr, err)
	}

	return fmt.Errorf("⚠️ max retries reached. Sent to DLQ topic %s: %w", dlxTopic, err)
}

// CloseKafkaPublisher safely closes the producer
func CloseKafkaPublisher() error {
	if singletonPub != nil && singletonPub.producer != nil {
		err := singletonPub.producer.Close()
		if err != nil {
			return fmt.Errorf("error closing Kafka producer: %w", err)
		}
		singletonPub = nil
	}
	return nil
}

// PublishEvent publishes an event to Kafka with a standardized event structure.
// This is a centralized function that can be used across all services.
//
// Parameters:
//   - ctx: Context for the operation
//   - eventType: The type of event (also used as the Kafka topic name)
//   - aggregateID: The aggregate ID (used as the Kafka message key)
//   - payload: The event payload data
//
// The event structure published to Kafka:
//   {
//     "event_type": "...",
//     "aggregate_id": "...",
//     "payload": { ... },
//     "timestamp": 1234567890
//   }
func PublishEvent(ctx context.Context, eventType string, aggregateID string, payload map[string]interface{}) error {
	log.Printf("═══════════════════════════════════════════════════════════════")
	log.Printf("📤 [EVENT PUBLISH] Publishing event to Kafka")

	// Validate required parameters
	if eventType == "" {
		log.Printf("❌ [EVENT PUBLISH] Validation failed: event_type is required")
		return fmt.Errorf("event_type is required")
	}
	if aggregateID == "" {
		log.Printf("❌ [EVENT PUBLISH] Validation failed: aggregate_id is required")
		return fmt.Errorf("aggregate_id is required")
	}
	if payload == nil {
		log.Printf("❌ [EVENT PUBLISH] Validation failed: payload is required")
		return fmt.Errorf("payload is required")
	}

	// Get payload keys for logging
	payloadKeys := getPayloadKeys(payload)
	log.Printf("   event_type=%s", eventType)
	log.Printf("   aggregate_id=%s", aggregateID)
	log.Printf("   payload_keys=%v", payloadKeys)
	log.Printf("   payload_size=%d keys", len(payloadKeys))

	// Get Kafka publisher
	kafkaPublisher, err := GetKafkaPublisher()
	if err != nil {
		log.Printf("❌ [EVENT PUBLISH] Failed to get Kafka publisher: %v", err)
		log.Printf("═══════════════════════════════════════════════════════════════")
		return fmt.Errorf("failed to get Kafka publisher: %w", err)
	}

	// Create event message with required structure
	event := map[string]interface{}{
		"event_type":   eventType,
		"aggregate_id": aggregateID,
		"payload":      payload,
		"timestamp":    time.Now().Unix(),
	}

	// Publish to Kafka topic (event type is the topic name)
	topic := eventType
	key := aggregateID

	eventJSON, _ := json.Marshal(event)
	log.Printf("   topic=%s", topic)
	log.Printf("   key=%s", key)
	log.Printf("   event_payload_length=%d bytes", len(eventJSON))

	// Publish with retry (3 retries built into Publish method)
	if err := kafkaPublisher.Publish(topic, key, event, 3); err != nil {
		log.Printf("❌ [EVENT PUBLISH] Failed to publish to Kafka: %v", err)
		log.Printf("═══════════════════════════════════════════════════════════════")
		return fmt.Errorf("failed to publish to Kafka: %w", err)
	}

	log.Printf("✅ [EVENT PUBLISH] Event published successfully")
	log.Printf("   topic=%s", topic)
	log.Printf("   key=%s", key)
	log.Printf("═══════════════════════════════════════════════════════════════")
	return nil
}

// getPayloadKeys is a helper function to extract keys from a map for debugging
func getPayloadKeys(payload map[string]interface{}) []string {
	if payload == nil {
		return []string{}
	}
	keys := make([]string, 0, len(payload))
	for k := range payload {
		keys = append(keys, k)
	}
	return keys
}
