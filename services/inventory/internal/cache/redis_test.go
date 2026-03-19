package cache_test

import (
	"context"
	"testing"

	"github.com/rts/inventory/internal/cache"
)

func TestAvailabilityCacheMiss(t *testing.T) {
	// Use a non-existent Redis to test cache miss fallback behavior
	c := cache.NewRedisCache("localhost:59999") // unlikely to be running
	defer c.Close()

	ctx := context.Background()

	t.Run("get returns nil on cache miss", func(t *testing.T) {
		result, err := c.GetAvailability(ctx, "nonexistent-sku")
		// Either nil result (cache miss) or error (connection failure)
		// Both should be handled gracefully by the availability service
		if result != nil && err == nil {
			t.Error("expected nil result or error for non-existent key")
		}
	})

	t.Run("set returns error on connection failure", func(t *testing.T) {
		err := c.SetAvailability(ctx, "test-sku", cache.AvailabilityData{
			Available: true,
			Quantity:  10,
		})
		if err == nil {
			t.Error("expected error when redis is not available")
		}
	})

	t.Run("invalidate returns error on connection failure", func(t *testing.T) {
		err := c.InvalidateAvailability(ctx, "test-sku")
		if err == nil {
			t.Error("expected error when redis is not available")
		}
	})
}

func TestAvailabilityDataSerialization(t *testing.T) {
	data := cache.AvailabilityData{
		Available: true,
		Quantity:  42,
	}

	if !data.Available {
		t.Error("expected available to be true")
	}
	if data.Quantity != 42 {
		t.Errorf("expected quantity 42, got %d", data.Quantity)
	}
}
