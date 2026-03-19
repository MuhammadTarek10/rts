package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rts/inventory/internal/cache"
	"github.com/rts/inventory/internal/domain"
	"github.com/rts/inventory/internal/publisher"
	"github.com/rts/inventory/internal/repository"
)

type MovementService struct {
	movementRepo  *repository.MovementRepository
	stockRepo     *repository.StockRepository
	inventoryRepo *repository.InventoryRepository
	cache         *cache.RedisCache
	publisher     *publisher.EventPublisher
}

func NewMovementService(
	movementRepo *repository.MovementRepository,
	stockRepo *repository.StockRepository,
	inventoryRepo *repository.InventoryRepository,
	cache *cache.RedisCache,
	publisher *publisher.EventPublisher,
) *MovementService {
	return &MovementService{
		movementRepo:  movementRepo,
		stockRepo:     stockRepo,
		inventoryRepo: inventoryRepo,
		cache:         cache,
		publisher:     publisher,
	}
}

func (s *MovementService) Receive(ctx context.Context, input domain.ReceiveInput, performedBy string) (*domain.StockMovement, error) {
	if input.Quantity <= 0 {
		return nil, &domain.ValidationError{Field: "quantity", Message: "must be positive"}
	}

	item, err := s.inventoryRepo.GetByID(ctx, input.InventoryItemID)
	if err != nil {
		return nil, err
	}

	tx, err := s.stockRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	sl, err := s.stockRepo.GetByItemAndWarehouseTx(ctx, tx, input.InventoryItemID, input.WarehouseID)
	if err != nil {
		return nil, err
	}
	if sl == nil {
		// Create stock level entry
		sl, err = s.stockRepo.UpsertWithTx(ctx, tx, input.InventoryItemID, input.WarehouseID)
		if err != nil {
			return nil, err
		}
	}

	newOnHand := sl.QuantityOnHand + input.Quantity
	updatedSL, err := s.stockRepo.UpdateQuantityWithVersion(ctx, tx, sl.ID, newOnHand, sl.QuantityReserved, sl.Version)
	if err != nil {
		return nil, err
	}

	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	movement := &domain.StockMovement{
		InventoryItemID: input.InventoryItemID,
		WarehouseID:     input.WarehouseID,
		Type:            domain.MovementTypeReceive,
		Quantity:        input.Quantity,
		ReferenceType:   input.ReferenceType,
		ReferenceID:     input.ReferenceID,
		Reason:          input.Reason,
		PerformedBy:     &performedBy,
		CostPerUnit:     input.CostPerUnit,
		Currency:        currency,
	}

	result, err := s.movementRepo.Create(ctx, tx, movement)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	s.postStockUpdate(ctx, item, updatedSL)
	return result, nil
}

func (s *MovementService) Ship(ctx context.Context, input domain.ShipInput, performedBy string) (*domain.StockMovement, error) {
	if input.Quantity <= 0 {
		return nil, &domain.ValidationError{Field: "quantity", Message: "must be positive"}
	}

	item, err := s.inventoryRepo.GetByID(ctx, input.InventoryItemID)
	if err != nil {
		return nil, err
	}

	tx, err := s.stockRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	sl, err := s.stockRepo.GetByItemAndWarehouseTx(ctx, tx, input.InventoryItemID, input.WarehouseID)
	if err != nil {
		return nil, err
	}
	if sl == nil {
		return nil, &domain.NotFoundError{Resource: "stock_level", ID: input.InventoryItemID}
	}

	if sl.QuantityOnHand < input.Quantity {
		return nil, &domain.InsufficientStockError{
			SKU:       item.SKU,
			Requested: input.Quantity,
			Available: sl.QuantityOnHand,
		}
	}

	newOnHand := sl.QuantityOnHand - input.Quantity
	// Adjust reserved if needed
	newReserved := sl.QuantityReserved
	if newReserved > newOnHand {
		newReserved = newOnHand
	}

	updatedSL, err := s.stockRepo.UpdateQuantityWithVersion(ctx, tx, sl.ID, newOnHand, newReserved, sl.Version)
	if err != nil {
		return nil, err
	}

	movement := &domain.StockMovement{
		InventoryItemID: input.InventoryItemID,
		WarehouseID:     input.WarehouseID,
		Type:            domain.MovementTypeShip,
		Quantity:        -input.Quantity,
		ReferenceType:   input.ReferenceType,
		ReferenceID:     input.ReferenceID,
		Reason:          input.Reason,
		PerformedBy:     &performedBy,
		Currency:        "USD",
	}

	result, err := s.movementRepo.Create(ctx, tx, movement)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	s.postStockUpdate(ctx, item, updatedSL)
	return result, nil
}

func (s *MovementService) Adjust(ctx context.Context, input domain.AdjustInput, performedBy string) (*domain.StockMovement, error) {
	if input.Quantity == 0 {
		return nil, &domain.ValidationError{Field: "quantity", Message: "must be non-zero"}
	}

	item, err := s.inventoryRepo.GetByID(ctx, input.InventoryItemID)
	if err != nil {
		return nil, err
	}

	tx, err := s.stockRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	sl, err := s.stockRepo.GetByItemAndWarehouseTx(ctx, tx, input.InventoryItemID, input.WarehouseID)
	if err != nil {
		return nil, err
	}
	if sl == nil {
		return nil, &domain.NotFoundError{Resource: "stock_level", ID: input.InventoryItemID}
	}

	newOnHand := sl.QuantityOnHand + input.Quantity
	if newOnHand < 0 {
		return nil, &domain.InsufficientStockError{
			SKU:       item.SKU,
			Requested: -input.Quantity,
			Available: sl.QuantityOnHand,
		}
	}

	newReserved := sl.QuantityReserved
	if newReserved > newOnHand {
		newReserved = newOnHand
	}

	updatedSL, err := s.stockRepo.UpdateQuantityWithVersion(ctx, tx, sl.ID, newOnHand, newReserved, sl.Version)
	if err != nil {
		return nil, err
	}

	refType := "adjustment"
	movement := &domain.StockMovement{
		InventoryItemID: input.InventoryItemID,
		WarehouseID:     input.WarehouseID,
		Type:            domain.MovementTypeAdjust,
		Quantity:        input.Quantity,
		ReferenceType:   &refType,
		Reason:          input.Reason,
		PerformedBy:     &performedBy,
		Currency:        "USD",
	}

	result, err := s.movementRepo.Create(ctx, tx, movement)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	s.postStockUpdate(ctx, item, updatedSL)
	return result, nil
}

func (s *MovementService) Transfer(ctx context.Context, input domain.TransferInput, performedBy string) ([]domain.StockMovement, error) {
	if input.Quantity <= 0 {
		return nil, &domain.ValidationError{Field: "quantity", Message: "must be positive"}
	}
	if input.FromWarehouseID == input.ToWarehouseID {
		return nil, &domain.ValidationError{Field: "to_warehouse_id", Message: "must differ from source warehouse"}
	}

	item, err := s.inventoryRepo.GetByID(ctx, input.InventoryItemID)
	if err != nil {
		return nil, err
	}

	tx, err := s.stockRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Lock source
	fromSL, err := s.stockRepo.GetByItemAndWarehouseTx(ctx, tx, input.InventoryItemID, input.FromWarehouseID)
	if err != nil {
		return nil, err
	}
	if fromSL == nil {
		return nil, &domain.NotFoundError{Resource: "stock_level", ID: input.FromWarehouseID}
	}
	if fromSL.QuantityAvailable < input.Quantity {
		return nil, &domain.InsufficientStockError{
			SKU:       item.SKU,
			Requested: input.Quantity,
			Available: fromSL.QuantityAvailable,
		}
	}

	// Lock or create destination
	toSL, err := s.stockRepo.GetByItemAndWarehouseTx(ctx, tx, input.InventoryItemID, input.ToWarehouseID)
	if err != nil {
		return nil, err
	}
	if toSL == nil {
		toSL, err = s.stockRepo.UpsertWithTx(ctx, tx, input.InventoryItemID, input.ToWarehouseID)
		if err != nil {
			return nil, err
		}
	}

	// Update source
	newFromOnHand := fromSL.QuantityOnHand - input.Quantity
	updatedFromSL, err := s.stockRepo.UpdateQuantityWithVersion(ctx, tx, fromSL.ID, newFromOnHand, fromSL.QuantityReserved, fromSL.Version)
	if err != nil {
		return nil, err
	}

	// Update destination
	newToOnHand := toSL.QuantityOnHand + input.Quantity
	updatedToSL, err := s.stockRepo.UpdateQuantityWithVersion(ctx, tx, toSL.ID, newToOnHand, toSL.QuantityReserved, toSL.Version)
	if err != nil {
		return nil, err
	}

	refType := "transfer"
	outMovement := &domain.StockMovement{
		InventoryItemID: input.InventoryItemID,
		WarehouseID:     input.FromWarehouseID,
		Type:            domain.MovementTypeTransferOut,
		Quantity:        -input.Quantity,
		ReferenceType:   &refType,
		Reason:          input.Reason,
		PerformedBy:     &performedBy,
		Currency:        "USD",
	}
	inMovement := &domain.StockMovement{
		InventoryItemID: input.InventoryItemID,
		WarehouseID:     input.ToWarehouseID,
		Type:            domain.MovementTypeTransferIn,
		Quantity:        input.Quantity,
		ReferenceType:   &refType,
		Reason:          input.Reason,
		PerformedBy:     &performedBy,
		Currency:        "USD",
	}

	outResult, err := s.movementRepo.Create(ctx, tx, outMovement)
	if err != nil {
		return nil, err
	}
	inResult, err := s.movementRepo.Create(ctx, tx, inMovement)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	s.postStockUpdate(ctx, item, updatedFromSL)
	s.postStockUpdate(ctx, item, updatedToSL)
	return []domain.StockMovement{*outResult, *inResult}, nil
}

func (s *MovementService) Return(ctx context.Context, input domain.ReturnInput, performedBy string) (*domain.StockMovement, error) {
	if input.Quantity <= 0 {
		return nil, &domain.ValidationError{Field: "quantity", Message: "must be positive"}
	}

	item, err := s.inventoryRepo.GetByID(ctx, input.InventoryItemID)
	if err != nil {
		return nil, err
	}

	tx, err := s.stockRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	sl, err := s.stockRepo.GetByItemAndWarehouseTx(ctx, tx, input.InventoryItemID, input.WarehouseID)
	if err != nil {
		return nil, err
	}
	if sl == nil {
		sl, err = s.stockRepo.UpsertWithTx(ctx, tx, input.InventoryItemID, input.WarehouseID)
		if err != nil {
			return nil, err
		}
	}

	newOnHand := sl.QuantityOnHand + input.Quantity
	updatedSL, err := s.stockRepo.UpdateQuantityWithVersion(ctx, tx, sl.ID, newOnHand, sl.QuantityReserved, sl.Version)
	if err != nil {
		return nil, err
	}

	movement := &domain.StockMovement{
		InventoryItemID: input.InventoryItemID,
		WarehouseID:     input.WarehouseID,
		Type:            domain.MovementTypeReturn,
		Quantity:        input.Quantity,
		ReferenceType:   input.ReferenceType,
		ReferenceID:     input.ReferenceID,
		Reason:          input.Reason,
		PerformedBy:     &performedBy,
		Currency:        "USD",
	}

	result, err := s.movementRepo.Create(ctx, tx, movement)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	s.postStockUpdate(ctx, item, updatedSL)
	return result, nil
}

func (s *MovementService) GetMovement(ctx context.Context, id string) (*domain.StockMovement, error) {
	return s.movementRepo.GetByID(ctx, id)
}

func (s *MovementService) ListMovements(ctx context.Context, filter domain.MovementFilter) ([]domain.StockMovement, int, error) {
	return s.movementRepo.List(ctx, filter)
}

func (s *MovementService) postStockUpdate(ctx context.Context, item *domain.InventoryItem, sl *domain.StockLevel) {
	// Invalidate cache
	if err := s.cache.InvalidateAvailability(ctx, item.SKU); err != nil {
		slog.Warn("failed to invalidate cache", "sku", item.SKU, "error", err)
	}

	// Publish stock updated event
	s.publisher.PublishStockUpdated(ctx, domain.StockUpdatedPayload{
		InventoryItemID:   item.ID,
		WarehouseID:       sl.WarehouseID,
		SKU:               item.SKU,
		QuantityOnHand:    sl.QuantityOnHand,
		QuantityAvailable: sl.QuantityAvailable,
	})

	// Check low stock
	if sl.ReorderPoint > 0 && sl.QuantityAvailable <= sl.ReorderPoint {
		if sl.QuantityAvailable == 0 {
			s.publisher.PublishStockOut(ctx, domain.StockOutPayload{
				InventoryItemID: item.ID,
				WarehouseID:     sl.WarehouseID,
				SKU:             item.SKU,
			})
		} else {
			s.publisher.PublishStockLow(ctx, domain.StockLowPayload{
				InventoryItemID: item.ID,
				WarehouseID:     sl.WarehouseID,
				SKU:             item.SKU,
				Available:       sl.QuantityAvailable,
				ReorderPoint:    sl.ReorderPoint,
			})
		}
	}
}
