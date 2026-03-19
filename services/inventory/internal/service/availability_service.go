package service

import (
	"context"
	"log/slog"

	"github.com/rts/inventory/internal/cache"
	"github.com/rts/inventory/internal/domain"
	"github.com/rts/inventory/internal/repository"
)

type AvailabilityResponse struct {
	SKU       string `json:"sku"`
	Available bool   `json:"available"`
	Quantity  int    `json:"quantity"`
}

type AvailabilityService struct {
	stockRepo *repository.StockRepository
	cache     *cache.RedisCache
}

func NewAvailabilityService(
	stockRepo *repository.StockRepository,
	cache *cache.RedisCache,
) *AvailabilityService {
	return &AvailabilityService{
		stockRepo: stockRepo,
		cache:     cache,
	}
}

func (s *AvailabilityService) GetAvailability(ctx context.Context, sku string) (*AvailabilityResponse, error) {
	// Check cache first
	cached, err := s.cache.GetAvailability(ctx, sku)
	if err == nil && cached != nil {
		return &AvailabilityResponse{
			SKU:       sku,
			Available: cached.Available,
			Quantity:  cached.Quantity,
		}, nil
	}

	// Fallback to database
	available, isTracked, err := s.stockRepo.GetTotalAvailableBySKU(ctx, sku)
	if err != nil {
		// If it's a not-found error, return zero availability
		var notFound *domain.NotFoundError
		if isNotFoundErr(err, &notFound) {
			return &AvailabilityResponse{
				SKU:       sku,
				Available: false,
				Quantity:  0,
			}, nil
		}
		return nil, err
	}

	resp := &AvailabilityResponse{
		SKU:       sku,
		Available: !isTracked || available > 0,
		Quantity:  available,
	}

	// Cache the result
	if cacheErr := s.cache.SetAvailability(ctx, sku, cache.AvailabilityData{
		Available: resp.Available,
		Quantity:  resp.Quantity,
	}); cacheErr != nil {
		slog.Warn("failed to cache availability", "sku", sku, "error", cacheErr)
	}

	return resp, nil
}

func (s *AvailabilityService) GetBulkAvailability(ctx context.Context, skus []string) ([]AvailabilityResponse, error) {
	var results []AvailabilityResponse
	var uncachedSKUs []string

	// Try cache for each SKU
	for _, sku := range skus {
		cached, err := s.cache.GetAvailability(ctx, sku)
		if err == nil && cached != nil {
			results = append(results, AvailabilityResponse{
				SKU:       sku,
				Available: cached.Available,
				Quantity:  cached.Quantity,
			})
		} else {
			uncachedSKUs = append(uncachedSKUs, sku)
		}
	}

	if len(uncachedSKUs) == 0 {
		return results, nil
	}

	// Fetch uncached from DB
	dbResults, err := s.stockRepo.GetBulkAvailability(ctx, uncachedSKUs)
	if err != nil {
		return nil, err
	}

	for _, sku := range uncachedSKUs {
		data, ok := dbResults[sku]
		resp := AvailabilityResponse{
			SKU:       sku,
			Available: false,
			Quantity:  0,
		}
		if ok {
			resp.Available = !data.IsTracked || data.Available > 0
			resp.Quantity = data.Available
		}

		// Cache the result
		if cacheErr := s.cache.SetAvailability(ctx, sku, cache.AvailabilityData{
			Available: resp.Available,
			Quantity:  resp.Quantity,
		}); cacheErr != nil {
			slog.Warn("failed to cache availability", "sku", sku, "error", cacheErr)
		}

		results = append(results, resp)
	}

	return results, nil
}

func isNotFoundErr(err error, target interface{}) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*domain.NotFoundError)
	return ok
}
