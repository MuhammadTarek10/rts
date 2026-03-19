package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	availabilityKeyPrefix = "inventory:availability:"
	availabilityTTL       = 30 * time.Second
)

type AvailabilityData struct {
	Available bool `json:"available"`
	Quantity  int  `json:"quantity"`
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr string) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisCache{client: client}
}

func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

func (c *RedisCache) GetAvailability(ctx context.Context, sku string) (*AvailabilityData, error) {
	key := availabilityKeyPrefix + sku
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var data AvailabilityData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("unmarshal availability: %w", err)
	}
	return &data, nil
}

func (c *RedisCache) SetAvailability(ctx context.Context, sku string, data AvailabilityData) error {
	key := availabilityKeyPrefix + sku
	val, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal availability: %w", err)
	}
	return c.client.Set(ctx, key, val, availabilityTTL).Err()
}

func (c *RedisCache) InvalidateAvailability(ctx context.Context, sku string) error {
	key := availabilityKeyPrefix + sku
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		slog.Warn("failed to invalidate availability cache", "sku", sku, "error", err)
		return err
	}
	return nil
}
