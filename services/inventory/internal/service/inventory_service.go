package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/rts/inventory/internal/domain"
	"github.com/rts/inventory/internal/publisher"
	"github.com/rts/inventory/internal/repository"
)

type InventoryService struct {
	inventoryRepo   *repository.InventoryRepository
	stockRepo       *repository.StockRepository
	warehouseRepo   *repository.WarehouseRepository
	reservationRepo *repository.ReservationRepository
	eventPublisher  *publisher.EventPublisher
}

func NewInventoryService(
	inventoryRepo *repository.InventoryRepository,
	stockRepo *repository.StockRepository,
	warehouseRepo *repository.WarehouseRepository,
	reservationRepo *repository.ReservationRepository,
	eventPublisher *publisher.EventPublisher,
) *InventoryService {
	return &InventoryService{
		inventoryRepo:   inventoryRepo,
		stockRepo:       stockRepo,
		warehouseRepo:   warehouseRepo,
		reservationRepo: reservationRepo,
		eventPublisher:  eventPublisher,
	}
}

func (s *InventoryService) GetItem(ctx context.Context, id string) (*domain.InventoryItemWithStock, error) {
	return s.inventoryRepo.GetWithStock(ctx, id)
}

func (s *InventoryService) GetItemBySKU(ctx context.Context, sku string) (*domain.InventoryItem, error) {
	return s.inventoryRepo.GetBySKU(ctx, sku)
}

func (s *InventoryService) ListItems(ctx context.Context, filter domain.InventoryItemFilter) ([]domain.InventoryItemWithStock, int, error) {
	return s.inventoryRepo.List(ctx, filter)
}

func (s *InventoryService) GetItemStockLevels(ctx context.Context, id string) ([]domain.StockLevel, error) {
	// Verify item exists
	_, err := s.inventoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.stockRepo.GetByItemID(ctx, id)
}

func (s *InventoryService) UpdateItem(ctx context.Context, id string, input domain.UpdateInventoryItemInput) (*domain.InventoryItem, error) {
	item, err := s.inventoryRepo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}

	// Update reorder settings on all stock levels if provided
	if input.ReorderPoint != nil || input.ReorderQuantity != nil {
		rp := 0
		rq := 0
		if input.ReorderPoint != nil {
			rp = *input.ReorderPoint
		}
		if input.ReorderQuantity != nil {
			rq = *input.ReorderQuantity
		}
		if err := s.stockRepo.UpdateReorderSettings(ctx, id, rp, rq); err != nil {
			return nil, err
		}
	}

	return item, nil
}

// HandleProductCreated processes a catalog product.created event.
func (s *InventoryService) HandleProductCreated(ctx context.Context, payload domain.CatalogProductCreatedPayload) error {
	// Get default warehouse
	defaultWH, err := s.warehouseRepo.GetDefault(ctx)
	if err != nil {
		return err
	}

	tx, err := s.inventoryRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if len(payload.Variants) == 0 {
		// Product-level inventory item
		item := &domain.InventoryItem{
			ProductID: payload.ProductID,
			SKU:       payload.SKU,
			Title:     payload.Title,
			Status:    "active",
			IsTracked: true,
		}
		created, err := s.inventoryRepo.Upsert(ctx, tx, item)
		if err != nil {
			return err
		}

		if defaultWH != nil {
			if _, err := s.stockRepo.UpsertWithTx(ctx, tx, created.ID, defaultWH.ID); err != nil {
				return err
			}
		}
	} else {
		// One inventory item per variant
		for _, v := range payload.Variants {
			variantID := v.VariantID
			item := &domain.InventoryItem{
				ProductID: payload.ProductID,
				VariantID: &variantID,
				SKU:       v.SKU,
				Title:     payload.Title,
				Status:    "active",
				IsTracked: true,
			}
			created, err := s.inventoryRepo.Upsert(ctx, tx, item)
			if err != nil {
				return err
			}

			if defaultWH != nil {
				if _, err := s.stockRepo.UpsertWithTx(ctx, tx, created.ID, defaultWH.ID); err != nil {
					return err
				}
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	slog.Info("handled product.created", "product_id", payload.ProductID)
	return nil
}

// HandleProductUpdated processes a catalog product.updated event.
func (s *InventoryService) HandleProductUpdated(ctx context.Context, payload domain.CatalogProductUpdatedPayload) error {
	tx, err := s.inventoryRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	defaultWH, err := s.warehouseRepo.GetDefault(ctx)
	if err != nil {
		return err
	}

	if len(payload.Variants) == 0 {
		// Update or create product-level item
		item := &domain.InventoryItem{
			ProductID: payload.ProductID,
			SKU:       payload.SKU,
			Title:     payload.Title,
			Status:    "active",
			IsTracked: true,
		}
		created, err := s.inventoryRepo.Upsert(ctx, tx, item)
		if err != nil {
			return err
		}
		if defaultWH != nil {
			if _, err := s.stockRepo.UpsertWithTx(ctx, tx, created.ID, defaultWH.ID); err != nil {
				return err
			}
		}
	} else {
		// Upsert each variant
		for _, v := range payload.Variants {
			variantID := v.VariantID
			item := &domain.InventoryItem{
				ProductID: payload.ProductID,
				VariantID: &variantID,
				SKU:       v.SKU,
				Title:     payload.Title,
				Status:    "active",
				IsTracked: true,
			}
			created, err := s.inventoryRepo.Upsert(ctx, tx, item)
			if err != nil {
				return err
			}
			if defaultWH != nil {
				if _, err := s.stockRepo.UpsertWithTx(ctx, tx, created.ID, defaultWH.ID); err != nil {
					return err
				}
			}
		}

		// Archive inventory items for variants that no longer exist
		existingItems, err := s.inventoryRepo.GetByProductID(ctx, payload.ProductID)
		if err != nil {
			return err
		}
		variantSet := make(map[string]bool)
		for _, v := range payload.Variants {
			variantSet[v.VariantID] = true
		}
		for _, item := range existingItems {
			if item.VariantID != nil && !variantSet[*item.VariantID] {
				archiveItem := &domain.InventoryItem{
					ProductID: item.ProductID,
					VariantID: item.VariantID,
					SKU:       item.SKU,
					Title:     item.Title,
					Status:    "archived",
					IsTracked: item.IsTracked,
				}
				if _, err := s.inventoryRepo.Upsert(ctx, tx, archiveItem); err != nil {
					return err
				}
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	slog.Info("handled product.updated", "product_id", payload.ProductID)
	return nil
}

// HandleProductStatusChanged processes a catalog product.status_changed event.
func (s *InventoryService) HandleProductStatusChanged(ctx context.Context, payload domain.CatalogProductStatusChangedPayload) error {
	statusMap := map[string]string{
		"Draft":    "draft",
		"Active":   "active",
		"Archived": "archived",
	}

	newStatus, ok := statusMap[payload.NewStatus]
	if !ok {
		newStatus = payload.NewStatus
	}

	if err := s.inventoryRepo.UpdateStatus(ctx, payload.ProductID, newStatus); err != nil {
		return err
	}

	// If archived, release active reservations
	if newStatus == "archived" {
		tx, err := s.inventoryRepo.BeginTx(ctx)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback(ctx) }()

		if err := s.releaseReservationsForProduct(ctx, tx, payload.ProductID); err != nil {
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}

	slog.Info("handled product.status_changed", "product_id", payload.ProductID, "new_status", newStatus)
	return nil
}

// HandleProductDeleted processes a catalog product.deleted event.
func (s *InventoryService) HandleProductDeleted(ctx context.Context, payload domain.CatalogProductDeletedPayload) error {
	if err := s.inventoryRepo.UpdateStatus(ctx, payload.ProductID, "archived"); err != nil {
		return err
	}
	slog.Info("handled product.deleted", "product_id", payload.ProductID)
	return nil
}

func (s *InventoryService) releaseReservationsForProduct(ctx context.Context, tx pgx.Tx, productID string) error {
	items, err := s.inventoryRepo.GetByProductID(ctx, productID)
	if err != nil {
		return err
	}

	for _, item := range items {
		levels, err := s.stockRepo.GetByItemID(ctx, item.ID)
		if err != nil {
			return err
		}
		for _, level := range levels {
			reservations, err := s.reservationRepo.GetActiveByItemAndWarehouse(ctx, item.ID, level.WarehouseID)
			if err != nil {
				return err
			}
			for _, res := range reservations {
				if err := s.reservationRepo.Release(ctx, tx, res.ID); err != nil {
					return err
				}
				// Decrease reserved quantity
				sl, err := s.stockRepo.GetByItemAndWarehouseTx(ctx, tx, item.ID, level.WarehouseID)
				if err != nil || sl == nil {
					continue
				}
				newReserved := sl.QuantityReserved - res.Quantity
				if newReserved < 0 {
					slog.Error("data integrity error: reserved would go negative during product release", "stock_level_id", sl.ID, "reserved", sl.QuantityReserved, "quantity", res.Quantity)
					continue
				}
				if _, err := s.stockRepo.UpdateQuantityWithVersion(ctx, tx, sl.ID, sl.QuantityOnHand, newReserved, sl.Version); err != nil {
					return fmt.Errorf("failed to update stock after release for stock level '%s': %w", sl.ID, err)
				}
			}
		}
	}
	return nil
}

// CreateWarehouse creates a new warehouse with default enforcement.
func (s *InventoryService) CreateWarehouse(ctx context.Context, input domain.CreateWarehouseInput) (*domain.Warehouse, error) {
	tx, err := s.warehouseRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if input.IsDefault {
		if err := s.warehouseRepo.ClearAllDefaults(ctx, tx); err != nil {
			return nil, err
		}
	}

	w, err := s.warehouseRepo.Create(ctx, tx, input)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return w, nil
}

// UpdateWarehouse updates a warehouse. If setting as default, clears existing default in a transaction.
func (s *InventoryService) UpdateWarehouse(ctx context.Context, id string, input domain.UpdateWarehouseInput) (*domain.Warehouse, error) {
	if input.IsDefault != nil && *input.IsDefault {
		tx, err := s.warehouseRepo.BeginTx(ctx)
		if err != nil {
			return nil, err
		}
		defer func() { _ = tx.Rollback(ctx) }()

		if err := s.warehouseRepo.SetDefaultAtomic(ctx, tx, id); err != nil {
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
	}

	return s.warehouseRepo.Update(ctx, id, input)
}

// DeactivateWarehouse rejects if warehouse has stock or is default.
func (s *InventoryService) DeactivateWarehouse(ctx context.Context, id string) error {
	w, err := s.warehouseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if w.IsDefault {
		return &domain.ConflictError{Message: "cannot deactivate the default warehouse"}
	}

	hasStock, err := s.warehouseRepo.HasStock(ctx, id)
	if err != nil {
		return err
	}
	if hasStock {
		return &domain.ConflictError{Message: "cannot deactivate warehouse that has stock; transfer stock first"}
	}

	return s.warehouseRepo.Deactivate(ctx, id)
}

func (s *InventoryService) ListWarehouses(ctx context.Context) ([]domain.Warehouse, error) {
	return s.warehouseRepo.List(ctx)
}

func (s *InventoryService) GetWarehouse(ctx context.Context, id string) (*domain.Warehouse, error) {
	return s.warehouseRepo.GetByID(ctx, id)
}
