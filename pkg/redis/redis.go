package redis

import (
	"context"

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
		return nil, err
	}

	return &Redis{
		Client: client,
	}, nil
}

func (r Redis) Close() {
	if r.Client != nil {
		r.Client.Close()
	}
}
