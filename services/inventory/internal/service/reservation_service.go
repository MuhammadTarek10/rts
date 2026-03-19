package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/rts/inventory/internal/cache"
	"github.com/rts/inventory/internal/domain"
	"github.com/rts/inventory/internal/publisher"
	"github.com/rts/inventory/internal/repository"
)

type ReservationService struct {
	reservationRepo *repository.ReservationRepository
	stockRepo       *repository.StockRepository
	inventoryRepo   *repository.InventoryRepository
	movementRepo    *repository.MovementRepository
	cache           *cache.RedisCache
	publisher       *publisher.EventPublisher
}

func NewReservationService(
	reservationRepo *repository.ReservationRepository,
	stockRepo *repository.StockRepository,
	inventoryRepo *repository.InventoryRepository,
	movementRepo *repository.MovementRepository,
	cache *cache.RedisCache,
	publisher *publisher.EventPublisher,
) *ReservationService {
	return &ReservationService{
		reservationRepo: reservationRepo,
		stockRepo:       stockRepo,
		inventoryRepo:   inventoryRepo,
		movementRepo:    movementRepo,
		cache:           cache,
		publisher:       publisher,
	}
}

func (s *ReservationService) Reserve(ctx context.Context, input domain.ReserveInput, performedBy string) (*domain.Reservation, error) {
	if input.Quantity <= 0 {
		return nil, &domain.ValidationError{Field: "quantity", Message: "must be positive"}
	}
	if input.OrderID == "" {
		return nil, &domain.ValidationError{Field: "order_id", Message: "is required"}
	}
	if input.SKU == "" {
		return nil, &domain.ValidationError{Field: "sku", Message: "is required"}
	}

	ttl := domain.DefaultReservationTTLMinutes
	if input.TTLMinutes != nil {
		ttl = *input.TTLMinutes
		if ttl < 1 {
			ttl = 1
		}
		if ttl > domain.MaxReservationTTLMinutes {
			ttl = domain.MaxReservationTTLMinutes
		}
	}

	item, err := s.inventoryRepo.GetBySKU(ctx, input.SKU)
	if err != nil {
		return nil, err
	}

	if !item.IsTracked {
		// Untracked items (digital goods) don't need stock reservations
		return nil, &domain.ValidationError{Field: "sku", Message: "untracked items do not require reservations"}
	}

	// Find a warehouse with sufficient available stock
	levels, err := s.stockRepo.GetByItemID(ctx, item.ID)
	if err != nil {
		return nil, err
	}

	var targetLevel *domain.StockLevel
	for i := range levels {
		if levels[i].QuantityAvailable >= input.Quantity {
			targetLevel = &levels[i]
			break
		}
	}

	if targetLevel == nil {
		totalAvailable := 0
		for _, l := range levels {
			totalAvailable += l.QuantityAvailable
		}
		return nil, &domain.InsufficientStockError{
			SKU:       input.SKU,
			Requested: input.Quantity,
			Available: totalAvailable,
		}
	}

	tx, err := s.reservationRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Lock and re-check
	sl, err := s.stockRepo.GetByItemAndWarehouseTx(ctx, tx, item.ID, targetLevel.WarehouseID)
	if err != nil {
		return nil, err
	}
	if sl == nil || sl.QuantityAvailable < input.Quantity {
		avail := 0
		if sl != nil {
			avail = sl.QuantityAvailable
		}
		return nil, &domain.InsufficientStockError{
			SKU:       input.SKU,
			Requested: input.Quantity,
			Available: avail,
		}
	}

	// Update reserved quantity
	newReserved := sl.QuantityReserved + input.Quantity
	updatedSL, err := s.stockRepo.UpdateQuantityWithVersion(ctx, tx, sl.ID, sl.QuantityOnHand, newReserved, sl.Version)
	if err != nil {
		return nil, err
	}

	// Create reservation
	expiresAt := time.Now().Add(time.Duration(ttl) * time.Minute)
	reservation := &domain.Reservation{
		InventoryItemID: item.ID,
		WarehouseID:     targetLevel.WarehouseID,
		OrderID:         input.OrderID,
		Quantity:        input.Quantity,
		Status:          domain.ReservationStatusActive,
		ExpiresAt:       expiresAt,
	}

	created, err := s.reservationRepo.Create(ctx, tx, reservation)
	if err != nil {
		return nil, err
	}

	// Record reserve movement
	refType := "reservation"
	movement := &domain.StockMovement{
		InventoryItemID: item.ID,
		WarehouseID:     targetLevel.WarehouseID,
		Type:            domain.MovementTypeReserve,
		Quantity:        -input.Quantity,
		ReferenceType:   &refType,
		ReferenceID:     &created.ID,
		PerformedBy:     &performedBy,
		Currency:        "USD",
	}
	if _, err := s.movementRepo.Create(ctx, tx, movement); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	// Post-update
	if err := s.cache.InvalidateAvailability(ctx, item.SKU); err != nil {
		slog.Warn("failed to invalidate cache", "sku", item.SKU, "error", err)
	}
	s.publisher.PublishReservationCreated(ctx, domain.ReservationEventPayload{
		ReservationID:   created.ID,
		InventoryItemID: item.ID,
		WarehouseID:     targetLevel.WarehouseID,
		OrderID:         input.OrderID,
		Quantity:        input.Quantity,
		Status:          domain.ReservationStatusActive,
	})
	s.publisher.PublishStockUpdated(ctx, domain.StockUpdatedPayload{
		InventoryItemID:   item.ID,
		WarehouseID:       updatedSL.WarehouseID,
		SKU:               item.SKU,
		QuantityOnHand:    updatedSL.QuantityOnHand,
		QuantityAvailable: updatedSL.QuantityAvailable,
	})

	return created, nil
}

func (s *ReservationService) Confirm(ctx context.Context, input domain.ConfirmReservationInput) error {
	if input.OrderID == "" {
		return &domain.ValidationError{Field: "order_id", Message: "is required"}
	}

	tx, err := s.reservationRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	reservations, err := s.reservationRepo.GetActiveByOrderID(ctx, tx, input.OrderID)
	if err != nil {
		return err
	}

	if len(reservations) == 0 {
		// Idempotent: already confirmed or no active reservations
		return nil
	}

	for _, res := range reservations {
		if err := s.reservationRepo.Confirm(ctx, tx, res.ID); err != nil {
			return err
		}

		// Ship the confirmed quantity (decrease on_hand)
		sl, err := s.stockRepo.GetByItemAndWarehouseTx(ctx, tx, res.InventoryItemID, res.WarehouseID)
		if err != nil {
			return err
		}
		if sl != nil {
			newOnHand := sl.QuantityOnHand - res.Quantity
			if newOnHand < 0 {
				return fmt.Errorf("data integrity error: on_hand would go negative for stock level '%s' (on_hand=%d, quantity=%d)", sl.ID, sl.QuantityOnHand, res.Quantity)
			}
			newReserved := sl.QuantityReserved - res.Quantity
			if newReserved < 0 {
				return fmt.Errorf("data integrity error: reserved would go negative for stock level '%s' (reserved=%d, quantity=%d)", sl.ID, sl.QuantityReserved, res.Quantity)
			}
			updatedSL, err := s.stockRepo.UpdateQuantityWithVersion(ctx, tx, sl.ID, newOnHand, newReserved, sl.Version)
			if err != nil {
				return fmt.Errorf("version conflict during confirm for reservation '%s': %w", res.ID, err)
			}

			// Record ship movement
			refType := "sales_order"
			movement := &domain.StockMovement{
				InventoryItemID: res.InventoryItemID,
				WarehouseID:     res.WarehouseID,
				Type:            domain.MovementTypeShip,
				Quantity:        -res.Quantity,
				ReferenceType:   &refType,
				ReferenceID:     &res.OrderID,
				Currency:        "USD",
			}
			if _, err := s.movementRepo.Create(ctx, tx, movement); err != nil {
				return err
			}

			item, itemErr := s.inventoryRepo.GetByID(ctx, res.InventoryItemID)
			if itemErr != nil {
				slog.Error("failed to fetch inventory item for event publishing", "item_id", res.InventoryItemID, "error", itemErr)
			} else if item != nil {
				if cacheErr := s.cache.InvalidateAvailability(ctx, item.SKU); cacheErr != nil {
					slog.Warn("failed to invalidate cache", "sku", item.SKU, "error", cacheErr)
				}
				s.publisher.PublishStockUpdated(ctx, domain.StockUpdatedPayload{
					InventoryItemID:   item.ID,
					WarehouseID:       updatedSL.WarehouseID,
					SKU:               item.SKU,
					QuantityOnHand:    updatedSL.QuantityOnHand,
					QuantityAvailable: updatedSL.QuantityAvailable,
				})
			}
		}

		s.publisher.PublishReservationConfirmed(ctx, domain.ReservationEventPayload{
			ReservationID:   res.ID,
			InventoryItemID: res.InventoryItemID,
			WarehouseID:     res.WarehouseID,
			OrderID:         res.OrderID,
			Quantity:        res.Quantity,
			Status:          domain.ReservationStatusConfirmed,
		})
	}

	return tx.Commit(ctx)
}

func (s *ReservationService) Release(ctx context.Context, input domain.ReleaseReservationInput) error {
	if input.OrderID == "" {
		return &domain.ValidationError{Field: "order_id", Message: "is required"}
	}

	tx, err := s.reservationRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	reservations, err := s.reservationRepo.GetActiveByOrderID(ctx, tx, input.OrderID)
	if err != nil {
		return err
	}

	if len(reservations) == 0 {
		// Idempotent
		return nil
	}

	for _, res := range reservations {
		if err := s.reservationRepo.Release(ctx, tx, res.ID); err != nil {
			return err
		}

		// Decrease reserved quantity
		sl, err := s.stockRepo.GetByItemAndWarehouseTx(ctx, tx, res.InventoryItemID, res.WarehouseID)
		if err != nil {
			return err
		}
		if sl != nil {
			newReserved := sl.QuantityReserved - res.Quantity
			if newReserved < 0 {
				return fmt.Errorf("data integrity error: reserved would go negative for stock level '%s' (reserved=%d, quantity=%d)", sl.ID, sl.QuantityReserved, res.Quantity)
			}
			updatedSL, err := s.stockRepo.UpdateQuantityWithVersion(ctx, tx, sl.ID, sl.QuantityOnHand, newReserved, sl.Version)
			if err != nil {
				return fmt.Errorf("version conflict during release for reservation '%s': %w", res.ID, err)
			}

			refType := "reservation"
			movement := &domain.StockMovement{
				InventoryItemID: res.InventoryItemID,
				WarehouseID:     res.WarehouseID,
				Type:            domain.MovementTypeRelease,
				Quantity:        res.Quantity,
				ReferenceType:   &refType,
				ReferenceID:     &res.ID,
				Currency:        "USD",
			}
			if _, err := s.movementRepo.Create(ctx, tx, movement); err != nil {
				return err
			}

			item, itemErr := s.inventoryRepo.GetByID(ctx, res.InventoryItemID)
			if itemErr != nil {
				slog.Error("failed to fetch inventory item for event publishing", "item_id", res.InventoryItemID, "error", itemErr)
			} else if item != nil {
				if cacheErr := s.cache.InvalidateAvailability(ctx, item.SKU); cacheErr != nil {
					slog.Warn("failed to invalidate cache", "sku", item.SKU, "error", cacheErr)
				}
				s.publisher.PublishStockUpdated(ctx, domain.StockUpdatedPayload{
					InventoryItemID:   item.ID,
					WarehouseID:       updatedSL.WarehouseID,
					SKU:               item.SKU,
					QuantityOnHand:    updatedSL.QuantityOnHand,
					QuantityAvailable: updatedSL.QuantityAvailable,
				})
			}
		}

		s.publisher.PublishReservationReleased(ctx, domain.ReservationEventPayload{
			ReservationID:   res.ID,
			InventoryItemID: res.InventoryItemID,
			WarehouseID:     res.WarehouseID,
			OrderID:         res.OrderID,
			Quantity:        res.Quantity,
			Status:          domain.ReservationStatusReleased,
		})
	}

	return tx.Commit(ctx)
}

func (s *ReservationService) GetByOrderID(ctx context.Context, orderID string) ([]domain.Reservation, error) {
	return s.reservationRepo.GetByOrderID(ctx, orderID)
}

// ExpireBatch processes expired reservations in batches. Called by the sweeper goroutine.
func (s *ReservationService) ExpireBatch(ctx context.Context, batchSize int) (int, error) {
	expired, err := s.reservationRepo.GetExpired(ctx, batchSize)
	if err != nil {
		return 0, err
	}

	if len(expired) == 0 {
		return 0, nil
	}

	processed := 0
	for _, res := range expired {
		tx, err := s.reservationRepo.BeginTx(ctx)
		if err != nil {
			slog.Error("failed to begin tx for expiry", "error", err)
			continue
		}

		if err := s.reservationRepo.Expire(ctx, tx, res.ID); err != nil {
			_ = tx.Rollback(ctx)
			slog.Error("failed to expire reservation", "id", res.ID, "error", err)
			continue
		}

		// Release reserved stock
		sl, err := s.stockRepo.GetByItemAndWarehouseTx(ctx, tx, res.InventoryItemID, res.WarehouseID)
		if err != nil || sl == nil {
			_ = tx.Rollback(ctx)
			slog.Error("failed to get stock level for expiry", "error", err)
			continue
		}

		newReserved := sl.QuantityReserved - res.Quantity
		if newReserved < 0 {
			_ = tx.Rollback(ctx)
			slog.Error("data integrity error: reserved would go negative during expiry", "stock_level_id", sl.ID, "reserved", sl.QuantityReserved, "quantity", res.Quantity)
			continue
		}

		updatedSL, err := s.stockRepo.UpdateQuantityWithVersion(ctx, tx, sl.ID, sl.QuantityOnHand, newReserved, sl.Version)
		if err != nil {
			_ = tx.Rollback(ctx)
			slog.Error("version conflict during expiry", "reservation_id", res.ID, "error", err)
			continue
		}

		refType := "reservation"
		movement := &domain.StockMovement{
			InventoryItemID: res.InventoryItemID,
			WarehouseID:     res.WarehouseID,
			Type:            domain.MovementTypeRelease,
			Quantity:        res.Quantity,
			ReferenceType:   &refType,
			ReferenceID:     &res.ID,
			Currency:        "USD",
		}
		if _, err := s.movementRepo.Create(ctx, tx, movement); err != nil {
			_ = tx.Rollback(ctx)
			slog.Error("failed to create release movement", "error", err)
			continue
		}

		if err := tx.Commit(ctx); err != nil {
			slog.Error("failed to commit expiry tx", "error", err)
			continue
		}

		processed++

		item, itemErr := s.inventoryRepo.GetByID(ctx, res.InventoryItemID)
		if itemErr != nil {
			slog.Error("failed to fetch inventory item for event publishing", "item_id", res.InventoryItemID, "error", itemErr)
		} else if item != nil {
			if cacheErr := s.cache.InvalidateAvailability(ctx, item.SKU); cacheErr != nil {
				slog.Warn("failed to invalidate cache after expiry", "sku", item.SKU, "error", cacheErr)
			}
			s.publisher.PublishStockUpdated(ctx, domain.StockUpdatedPayload{
				InventoryItemID:   item.ID,
				WarehouseID:       updatedSL.WarehouseID,
				SKU:               item.SKU,
				QuantityOnHand:    updatedSL.QuantityOnHand,
				QuantityAvailable: updatedSL.QuantityAvailable,
			})
		}

		s.publisher.PublishReservationReleased(ctx, domain.ReservationEventPayload{
			ReservationID:   res.ID,
			InventoryItemID: res.InventoryItemID,
			WarehouseID:     res.WarehouseID,
			OrderID:         res.OrderID,
			Quantity:        res.Quantity,
			Status:          domain.ReservationStatusExpired,
		})
	}

	return processed, nil
}
