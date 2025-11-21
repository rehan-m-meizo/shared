package redisclient

import (
	"context"
	"fmt"
	"log"
	"shared/constants"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	once        sync.Once
)

func InitRedisClient() error {
	once.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     constants.RedisHost + ":" + constants.RedisPort,
			Username: constants.RedisUsername,
			Password: constants.RedisPassword,
			DB:       0,
		})
	})
	return nil
}

func GetRedisClient() (*redis.Client, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}
	return redisClient, nil
}

func CloseRedisClient() error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return redisClient.Close()
}

func InitBloomFilter(name string, capacity int64, errorRate float64) error {
	client, err := GetRedisClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Check if bloom filter key already exists
	exists, err := client.Exists(ctx, name).Result()
	if err != nil {
		return fmt.Errorf("failed to check bloom existence: %w", err)
	}

	if exists > 0 {
		log.Printf("✅ Bloom filter '%s' already exists, skipping creation", name)
		return nil
	}

	// Try to create it
	_, err = client.BFReserve(ctx, name, errorRate, capacity).Result()
	if err != nil {
		// Ignore the “already exists” errors gracefully
		if strings.Contains(err.Error(), "item exists") || strings.Contains(err.Error(), "BUSYKEY") {
			log.Printf("✅ Bloom filter '%s' already exists (from another instance)", name)
			return nil
		}
		return fmt.Errorf("failed to create bloom filter: %w", err)
	}

	log.Printf("✅ Bloom filter '%s' initialized", name)
	return nil
}
