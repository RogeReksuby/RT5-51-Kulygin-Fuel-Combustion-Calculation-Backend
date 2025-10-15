package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"repback/internal/app/config"
	"strconv"
)

const servicePrefix = "repback."

type Client struct {
	client *redis.Client
}

func New(ctx context.Context, cfg config.Config) (*Client, error) {
	redisConfig := cfg.GetRedisConfig()

	redisClient := redis.NewClient(&redis.Options{
		Addr:        redisConfig.Host + ":" + strconv.Itoa(redisConfig.Port),
		Password:    redisConfig.Password,
		DB:          0,
		DialTimeout: redisConfig.DialTimeout,
		ReadTimeout: redisConfig.ReadTimeout,
	})

	// Проверяем подключение
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("cant ping redis: %w", err)
	}

	return &Client{
		client: redisClient,
	}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
