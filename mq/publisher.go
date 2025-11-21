package mq

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn      *amqp091.Connection
	channel   *amqp091.Channel
	exchange  string
	exchType  string
	connected bool
	mu        sync.RWMutex
	rabbitURL string
}

var (
	instance *Publisher
	once     sync.Once
)

// InitPublisher initializes the singleton publisher
func InitPublisher(rabbitURL, exchange, exchangeType string) error {
	var err error
	once.Do(func() {
		p := &Publisher{
			exchange:  exchange,
			exchType:  exchangeType,
			rabbitURL: rabbitURL,
		}
		if e := p.connect(); e != nil {
			err = e
			return
		}
		instance = p
		log.Println("✅ RabbitMQ Publisher initialized:", exchange)
	})
	return err
}

// GetPublisher returns the singleton instance
func GetPublisher() (*Publisher, error) {
	if instance == nil {
		return nil, errors.New("publisher not initialized")
	}
	return instance, nil
}

// connect initializes connection and channel
func (p *Publisher) connect() error {
	conn, err := amqp091.Dial(p.rabbitURL)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return err
	}
	if err := ch.ExchangeDeclare(
		p.exchange, p.exchType,
		true, false, false, false, nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.conn = conn
	p.channel = ch
	p.connected = true

	return nil
}

// reconnect safely closes and reopens the connection
func (p *Publisher) reconnect() error {
	p.Close()
	return p.connect()
}

// Publish sends a message to the exchange with the routing key
func (p *Publisher) Publish(ctx context.Context, routingKey string, data any) error {
	p.mu.RLock()
	connected := p.connected
	ch := p.channel
	p.mu.RUnlock()

	if !connected || ch == nil {
		if err := p.reconnect(); err != nil {
			return errors.New("failed to reconnect: " + err.Error())
		}
		// retry with fresh connection
		return p.Publish(ctx, routingKey, data)
	}

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	pub := amqp091.Publishing{
		ContentType: "application/json",
		Body:        body,
		Timestamp:   time.Now(),
	}

	return ch.PublishWithContext(ctx, p.exchange, routingKey, false, false, pub)
}

// Close cleans up the publisher connection
func (p *Publisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.connected = false

	if p.channel != nil {
		_ = p.channel.Close()
		p.channel = nil
	}
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
	}
	return nil
}
