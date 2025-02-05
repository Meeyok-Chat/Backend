package configs

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient() (*RedisClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr:     GetEnv("REDIS_URI") + ":" + GetEnv("REDIS_PORT"),
		Password: GetEnv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("ping redis failed: %w", err)
	}
	fmt.Println("ping redis success")
	return &RedisClient{Client: client}, nil
}
