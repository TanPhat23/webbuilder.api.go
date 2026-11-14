package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Config holds the Redis configuration parameters
type Config struct {
	URL      string
	Host     string
	Port     string
	Password string
	DB       int
}

// NewRedisClient initializes a new Redis client.
// It prioritizes the URL (e.g. REDIS_URL) if provided, otherwise it uses the individual host/port fields.
func NewRedisClient(cfg Config) (*redis.Client, error) {
	var options *redis.Options

	if cfg.URL != "" {
		var err error
		options, err = redis.ParseURL(cfg.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse redis url: %w", err)
		}
	} else {
		options = &redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
			Password: cfg.Password,
			DB:       cfg.DB,
		}
	}

	rdb := redis.NewClient(options)

	// Test connection
	ctx := context.Background()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return rdb, nil
}
