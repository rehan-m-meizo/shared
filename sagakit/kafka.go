package sagakit

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
)

func newUnifiedSaramaConfig(username, password string) *sarama.Config {
	sc := sarama.NewConfig()

	sc.Version = sarama.V2_8_0_0
	sc.Producer.Return.Successes = true

	sc.Consumer.Return.Errors = true
	sc.Consumer.Offsets.Initial = sarama.OffsetOldest

	sc.Net.SASL.Enable = true
	sc.Net.SASL.User = username
	sc.Net.SASL.Password = password
	sc.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	sc.Net.SASL.Handshake = true
	sc.Net.TLS.Enable = false

	sc.Net.DialTimeout = 10 * time.Second
	sc.Net.ReadTimeout = 10 * time.Second
	sc.Net.WriteTimeout = 10 * time.Second

	return sc
}

func NewKafkaPublisher(logger watermill.LoggerAdapter, brokers []string, username, password string) (message.Publisher, error) {
	sc := newUnifiedSaramaConfig(username, password)

	cfg := kafka.PublisherConfig{
		Brokers:               brokers,
		Marshaler:             kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: sc,
	}

	return kafka.NewPublisher(cfg, logger)
}

func NewKafkaSubscriber(logger watermill.LoggerAdapter, brokers []string, group, username, password string) (message.Subscriber, error) {
	sc := newUnifiedSaramaConfig(username, password)

	cfg := kafka.SubscriberConfig{
		Brokers:               brokers,
		ConsumerGroup:         group,
		Unmarshaler:           kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: sc,
	}

	return kafka.NewSubscriber(cfg, logger)
}
