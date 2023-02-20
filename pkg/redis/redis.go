package redis

import (
	"context"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func NewRedisClent(url string) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: "",
		DB:       0,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis client - connection - redis connection failed: %w", err)
	}

	return &Redis{
		Client: client,
	}, nil
}

func (r Redis) Close() error {
	if r.Client != nil {
		if err := r.Client.Close(); err != nil {
			return fmt.Errorf("redis client - close - client close err: %w", err)
		}
		return nil
	}
	return fmt.Errorf("redis client - close - redis client unreachable %s", time.Now())
}
