package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rts/inventory/internal/domain"
)

type StockRepository struct {
	pool *pgxpool.Pool
}

func NewStockRepository(pool *pgxpool.Pool) *StockRepository {
	return &StockRepository{pool: pool}
}

func (r *StockRepository) GetByItemAndWarehouse(ctx context.Context, itemID, warehouseID string) (*domain.StockLevel, error) {
	var sl domain.StockLevel
	err := r.pool.QueryRow(ctx, `
		SELECT id, inventory_item_id, warehouse_id, quantity_on_hand, quantity_reserved, quantity_available,
			reorder_point, reorder_quantity, updated_at, version
		FROM stock_levels WHERE inventory_item_id = $1 AND warehouse_id = $2
	`, itemID, warehouseID).Scan(
		&sl.ID, &sl.InventoryItemID, &sl.WarehouseID, &sl.QuantityOnHand, &sl.QuantityReserved,
		&sl.QuantityAvailable, &sl.ReorderPoint, &sl.ReorderQuantity, &sl.UpdatedAt, &sl.Version,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get stock level: %w", err)
	}
	return &sl, nil
}

func (r *StockRepository) GetByItemID(ctx context.Context, itemID string) ([]domain.StockLevel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, inventory_item_id, warehouse_id, quantity_on_hand, quantity_reserved, quantity_available,
			reorder_point, reorder_quantity, updated_at, version
		FROM stock_levels WHERE inventory_item_id = $1
	`, itemID)
	if err != nil {
		return nil, fmt.Errorf("get stock levels by item: %w", err)
	}
	defer rows.Close()

	var levels []domain.StockLevel
	for rows.Next() {
		var sl domain.StockLevel
		if err := rows.Scan(
			&sl.ID, &sl.InventoryItemID, &sl.WarehouseID, &sl.QuantityOnHand, &sl.QuantityReserved,
			&sl.QuantityAvailable, &sl.ReorderPoint, &sl.ReorderQuantity, &sl.UpdatedAt, &sl.Version,
		); err != nil {
			return nil, fmt.Errorf("scan stock level: %w", err)
		}
		levels = append(levels, sl)
	}
	return levels, nil
}

func (r *StockRepository) UpsertWithTx(ctx context.Context, tx pgx.Tx, itemID, warehouseID string) (*domain.StockLevel, error) {
	var sl domain.StockLevel
	err := tx.QueryRow(ctx, `
		INSERT INTO stock_levels (inventory_item_id, warehouse_id)
		VALUES ($1, $2)
		ON CONFLICT (inventory_item_id, warehouse_id) DO NOTHING
		RETURNING id, inventory_item_id, warehouse_id, quantity_on_hand, quantity_reserved, quantity_available,
			reorder_point, reorder_quantity, updated_at, version
	`, itemID, warehouseID).Scan(
		&sl.ID, &sl.InventoryItemID, &sl.WarehouseID, &sl.QuantityOnHand, &sl.QuantityReserved,
		&sl.QuantityAvailable, &sl.ReorderPoint, &sl.ReorderQuantity, &sl.UpdatedAt, &sl.Version,
	)
	if err == pgx.ErrNoRows {
		// Already exists, fetch it
		return r.GetByItemAndWarehouseTx(ctx, tx, itemID, warehouseID)
	}
	if err != nil {
		return nil, fmt.Errorf("upsert stock level: %w", err)
	}
	return &sl, nil
}

func (r *StockRepository) GetByItemAndWarehouseTx(ctx context.Context, tx pgx.Tx, itemID, warehouseID string) (*domain.StockLevel, error) {
	var sl domain.StockLevel
	err := tx.QueryRow(ctx, `
		SELECT id, inventory_item_id, warehouse_id, quantity_on_hand, quantity_reserved, quantity_available,
			reorder_point, reorder_quantity, updated_at, version
		FROM stock_levels WHERE inventory_item_id = $1 AND warehouse_id = $2 FOR UPDATE
	`, itemID, warehouseID).Scan(
		&sl.ID, &sl.InventoryItemID, &sl.WarehouseID, &sl.QuantityOnHand, &sl.QuantityReserved,
		&sl.QuantityAvailable, &sl.ReorderPoint, &sl.ReorderQuantity, &sl.UpdatedAt, &sl.Version,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get stock level for update: %w", err)
	}
	return &sl, nil
}

func (r *StockRepository) UpdateQuantityWithVersion(ctx context.Context, tx pgx.Tx, id string, onHand, reserved, version int) (*domain.StockLevel, error) {
	var sl domain.StockLevel
	err := tx.QueryRow(ctx, `
		UPDATE stock_levels
		SET quantity_on_hand = $1, quantity_reserved = $2, version = version + 1, updated_at = NOW()
		WHERE id = $3 AND version = $4
		RETURNING id, inventory_item_id, warehouse_id, quantity_on_hand, quantity_reserved, quantity_available,
			reorder_point, reorder_quantity, updated_at, version
	`, onHand, reserved, id, version).Scan(
		&sl.ID, &sl.InventoryItemID, &sl.WarehouseID, &sl.QuantityOnHand, &sl.QuantityReserved,
		&sl.QuantityAvailable, &sl.ReorderPoint, &sl.ReorderQuantity, &sl.UpdatedAt, &sl.Version,
	)
	if err == pgx.ErrNoRows {
		return nil, &domain.VersionConflictError{Resource: "stock_level", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("update stock level: %w", err)
	}
	return &sl, nil
}

func (r *StockRepository) UpdateReorderSettings(ctx context.Context, itemID string, reorderPoint, reorderQuantity int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE stock_levels SET reorder_point = $1, reorder_quantity = $2, updated_at = NOW()
		WHERE inventory_item_id = $3
	`, reorderPoint, reorderQuantity, itemID)
	if err != nil {
		return fmt.Errorf("update reorder settings: %w", err)
	}
	return nil
}

func (r *StockRepository) GetTotalAvailableBySKU(ctx context.Context, sku string) (int, bool, error) {
	var available int
	var isTracked bool
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(sl.quantity_available), 0), i.is_tracked
		FROM inventory_items i
		LEFT JOIN stock_levels sl ON sl.inventory_item_id = i.id
		WHERE i.sku = $1 AND i.status = 'active'
		GROUP BY i.is_tracked
	`, sku).Scan(&available, &isTracked)
	if err == pgx.ErrNoRows {
		return 0, true, &domain.NotFoundError{Resource: "inventory_item", ID: sku}
	}
	if err != nil {
		return 0, true, fmt.Errorf("get total available: %w", err)
	}
	return available, isTracked, nil
}

func (r *StockRepository) GetBulkAvailability(ctx context.Context, skus []string) (map[string]struct {
	Available int
	IsTracked bool
}, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT i.sku, COALESCE(SUM(sl.quantity_available), 0), i.is_tracked
		FROM inventory_items i
		LEFT JOIN stock_levels sl ON sl.inventory_item_id = i.id
		WHERE i.sku = ANY($1) AND i.status = 'active'
		GROUP BY i.sku, i.is_tracked
	`, skus)
	if err != nil {
		return nil, fmt.Errorf("get bulk availability: %w", err)
	}
	defer rows.Close()

	result := make(map[string]struct {
		Available int
		IsTracked bool
	})
	for rows.Next() {
		var sku string
		var available int
		var isTracked bool
		if err := rows.Scan(&sku, &available, &isTracked); err != nil {
			return nil, fmt.Errorf("scan bulk availability: %w", err)
		}
		result[sku] = struct {
			Available int
			IsTracked bool
		}{Available: available, IsTracked: isTracked}
	}
	return result, nil
}

func (r *StockRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}
