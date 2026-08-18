package database

import (
	"agropal/config"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedis(cfg *config.Config) *redis.Client {
	ctx := context.Background()

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Db,
		Protocol: cfg.Redis.Protocol,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("could not connect to Redis: %s", err))
		return nil
	}

	return client
}
