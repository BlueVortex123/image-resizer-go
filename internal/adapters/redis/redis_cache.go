package redisadapter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCacheInstance(addr, pass string, db int, ttl time.Duration) (*RedisCache, error) {
	fmt.Printf("Checking Redis Client Connection...\n")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       db,
	})

	// Ping to check connection
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		fmt.Printf("Redis Client Connection Failed!\n")
		fmt.Printf("Error: %v\n", err)
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	fmt.Printf("Redis Client Connected!\n")
	return &RedisCache{client: redisClient, ttl: ttl}, nil
}

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, string, bool, error) {
	bytes, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, "", false, nil
	}
	if err != nil {
		return nil, "", false, err
	}

	// get meta
	metaKey := key + ":meta"
	contentType, err := r.client.Get(ctx, metaKey).Result()
	if err == redis.Nil {
		contentType = "application/octet-stream"
	} else if err != nil {
		return nil, "", false, err
	}

	return bytes, contentType, true, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value []byte, contentType string) error {
	metaKey := key + ":meta"
	pipe := r.client.Pipeline() // use pipeline to set both key and meta
	pipe.Set(ctx, key, value, r.ttl)
	pipe.Set(ctx, metaKey, contentType, r.ttl)
	_, err := pipe.Exec(ctx) // execute pipeline
	return err
}
