package kf

import (
	"fmt"
	"log"
	"strings"
)

// Config represents the application configuration
type KafkaTopic struct {
	Name string `yaml:"name"`
}

type KafkaRetry struct {
	MaxRetryCount int    `yaml:"maxRetryCount"`
	DelayMs       int    `yaml:"delayMs"`
	DLQPrefix     string `yaml:"dlqPrefix"`
}

type Config struct {
	Brokers    []string
	GroupID    string
	Topics     []string
	Username   string
	Password   string
	SASLEnable bool
	TLSEnable  bool
}

type AppConfig struct {
	Kafka Config `yaml:"kafka"`
}

// LoadConfig builds Kafka config using your central constants
func LoadConfig(brokersString string, groupID string, username string, password string, topics []string) (*Config, error) {
	brokers := strings.Split(brokersString, ",")

	// You can hardcode TLS/SASL if needed or add more config vars
	saslEnable := true
	tlsEnable := false

	cfg := &Config{
		Brokers:    brokers,
		GroupID:    groupID,
		Username:   username,
		Password:   password,
		SASLEnable: saslEnable,
		TLSEnable:  tlsEnable,
		Topics:     topics,
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid Kafka config: %w", err)
	}

	log.Printf("✅ Kafka config loaded from shared constants (GroupID: %s)", groupID)
	return cfg, nil
}

// validate checks if all required config values are present
func (c *Config) validate() error {
	if len(c.Brokers) == 0 {
		return fmt.Errorf("missing brokers")
	}
	if c.Username == "" {
		return fmt.Errorf("missing username")
	}
	if c.Password == "" {
		return fmt.Errorf("missing password")
	}
	if c.GroupID == "" {
		return fmt.Errorf("missing groupID")
	}
	return nil
}
