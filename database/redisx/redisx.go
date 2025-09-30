package redisx

import (
	"backend-go/config"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	Rdb *redis.Client
}

var Rdb *redis.Client

func InitRedis() (*Client, error) {
	addr := config.GetEnv("REDIS_ADDR", "127.0.0.1:6379")
	pass := config.GetEnv("REDIS_PASSWORD", "") // set if you enabled auth
	db := config.GetEnvInt("REDIS_DB", 0)

	Rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       db,
		// sensible defaults:
		MaxRetries:   3,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := Rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	log.Println("✅ Connected to RedisDB")

	return &Client{Rdb: Rdb}, nil
}

func (c *Client) Close() error {
	return c.Rdb.Close()
}
