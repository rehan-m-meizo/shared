package kf

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

// HandlerFunc defines the signature for message handlers
type HandlerFunc func(message *sarama.ConsumerMessage) error

// ConsumerManager manages the Kafka consumer lifecycle
type ConsumerManager struct {
	config        *Config
	consumerGroup sarama.ConsumerGroup
	handler       HandlerFunc
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewConsumerManager creates a new consumer manager
func NewConsumerManager(cfg *Config, handler HandlerFunc) (*ConsumerManager, error) {
	saramaConfig := newSaramaConfig(cfg)

	consumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ConsumerManager{
		config:        cfg,
		consumerGroup: consumerGroup,
		handler:       handler,
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// Start begins consuming messages
func (cm *ConsumerManager) Start() error {
	if len(cm.config.Topics) == 0 {
		return fmt.Errorf("no topics provided in configuration")
	}

	handler := &ConsumerGroupHandler{
		handler: cm.handler,
	}

	go func() {
		for {
			select {
			case <-cm.ctx.Done():
				return
			default:
				if err := cm.consumerGroup.Consume(cm.ctx, cm.config.Topics, handler); err != nil {
					log.Printf("❌ Kafka consume error: %v", err)
				}
			}
		}
	}()

	log.Printf("✅ Consumer started for topics: %v, group: %s, brokers: %v", 
		cm.config.Topics, cm.config.GroupID, cm.config.Brokers)
	log.Printf("📋 Consumer config: offset=OffsetOldest (consumes ALL messages, including those published before consumer starts)")
	return nil
}

// Stop gracefully shuts down the consumer
func (cm *ConsumerManager) Stop() error {
	log.Println("🛑 Stopping consumer...")

	cm.cancel()

	if err := cm.consumerGroup.Close(); err != nil {
		return fmt.Errorf("failed to close consumer group: %w", err)
	}

	log.Println("✅ Consumer stopped")
	return nil
}

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler
type ConsumerGroupHandler struct {
	handler HandlerFunc
}

func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	log.Println("✅ Consumer group session setup")
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("🛑 Consumer group session cleanup")
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Printf("📥 Starting to consume from partition %d, initial offset: %d", claim.Partition(), claim.InitialOffset())
	
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				log.Println("⚠️ Received nil message, continuing...")
				continue
			}

			log.Printf("📨 Consuming message: topic=[%s], partition=%d, offset=%d, key=%s, value_len=%d", 
				message.Topic, message.Partition, message.Offset, string(message.Key), len(message.Value))

			if err := h.handler(message); err != nil {
				log.Printf("❌ Handler error on topic [%s], partition %d, offset %d: %v", 
					message.Topic, message.Partition, message.Offset, err)
			} else {
				log.Printf("✅ Successfully processed message: topic=[%s], partition=%d, offset=%d", 
					message.Topic, message.Partition, message.Offset)
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			log.Printf("🛑 Consumer claim context done for partition %d", claim.Partition())
			return nil
		}
	}
}

// newSaramaConfig creates a new Sarama configuration
func newSaramaConfig(cfg *Config) *sarama.Config {
	config := sarama.NewConfig()

	config.Producer.Return.Successes = true
	// Use OffsetOldest for development to consume all messages, including those published before consumer starts
	// Change to OffsetNewest for production to only consume new messages
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Version = sarama.V2_8_0_0
	config.Net.SASL.Enable = true
	config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	config.Net.SASL.User = cfg.Username
	config.Net.SASL.Password = cfg.Password
	config.Net.TLS.Enable = false

	return config
}
