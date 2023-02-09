package main

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	inner *redis.Client
}

func NewRedisClient(
	ctx context.Context,
	options *redis.Options,
) (*RedisClient, error) {
	client := redis.NewClient(options)
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return &RedisClient{
		inner: client,
	}, nil
}

func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	value, err := c.inner.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrorKeyNotFound{}
	}

	return value, err
}

func (c *RedisClient) Set(ctx context.Context, key string, value []byte) error {
	return c.inner.Set(ctx, key, value, 0).Err()
}
