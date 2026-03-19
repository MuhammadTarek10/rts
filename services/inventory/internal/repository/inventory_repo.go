package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rts/inventory/internal/domain"
)

type InventoryRepository struct {
	pool *pgxpool.Pool
}

func NewInventoryRepository(pool *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{pool: pool}
}

func (r *InventoryRepository) GetByID(ctx context.Context, id string) (*domain.InventoryItem, error) {
	var item domain.InventoryItem
	err := r.pool.QueryRow(ctx, `
		SELECT id, product_id, variant_id, sku, title, status, is_tracked, created_at, updated_at
		FROM inventory_items WHERE id = $1
	`, id).Scan(
		&item.ID, &item.ProductID, &item.VariantID, &item.SKU, &item.Title,
		&item.Status, &item.IsTracked, &item.CreatedAt, &item.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, &domain.NotFoundError{Resource: "inventory_item", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("get inventory item: %w", err)
	}
	return &item, nil
}

func (r *InventoryRepository) GetBySKU(ctx context.Context, sku string) (*domain.InventoryItem, error) {
	var item domain.InventoryItem
	err := r.pool.QueryRow(ctx, `
		SELECT id, product_id, variant_id, sku, title, status, is_tracked, created_at, updated_at
		FROM inventory_items WHERE sku = $1
	`, sku).Scan(
		&item.ID, &item.ProductID, &item.VariantID, &item.SKU, &item.Title,
		&item.Status, &item.IsTracked, &item.CreatedAt, &item.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, &domain.NotFoundError{Resource: "inventory_item", ID: sku}
	}
	if err != nil {
		return nil, fmt.Errorf("get inventory item by sku: %w", err)
	}
	return &item, nil
}

func (r *InventoryRepository) GetByProductAndVariant(ctx context.Context, productID string, variantID *string) (*domain.InventoryItem, error) {
	var item domain.InventoryItem
	var err error
	if variantID != nil {
		err = r.pool.QueryRow(ctx, `
			SELECT id, product_id, variant_id, sku, title, status, is_tracked, created_at, updated_at
			FROM inventory_items WHERE product_id = $1 AND variant_id = $2
		`, productID, *variantID).Scan(
			&item.ID, &item.ProductID, &item.VariantID, &item.SKU, &item.Title,
			&item.Status, &item.IsTracked, &item.CreatedAt, &item.UpdatedAt,
		)
	} else {
		err = r.pool.QueryRow(ctx, `
			SELECT id, product_id, variant_id, sku, title, status, is_tracked, created_at, updated_at
			FROM inventory_items WHERE product_id = $1 AND variant_id IS NULL
		`, productID).Scan(
			&item.ID, &item.ProductID, &item.VariantID, &item.SKU, &item.Title,
			&item.Status, &item.IsTracked, &item.CreatedAt, &item.UpdatedAt,
		)
	}
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get inventory item by product/variant: %w", err)
	}
	return &item, nil
}

func (r *InventoryRepository) List(ctx context.Context, filter domain.InventoryItemFilter) ([]domain.InventoryItemWithStock, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("i.status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.SKU != "" {
		conditions = append(conditions, fmt.Sprintf("i.sku ILIKE $%d", argIdx))
		args = append(args, "%"+filter.SKU+"%")
		argIdx++
	}
	if filter.ProductID != "" {
		conditions = append(conditions, fmt.Sprintf("i.product_id = $%d", argIdx))
		args = append(args, filter.ProductID)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM inventory_items i %s", whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count inventory items: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	query := fmt.Sprintf(`
		SELECT i.id, i.product_id, i.variant_id, i.sku, i.title, i.status, i.is_tracked, i.created_at, i.updated_at,
			COALESCE(SUM(sl.quantity_on_hand), 0) AS total_on_hand,
			COALESCE(SUM(sl.quantity_reserved), 0) AS total_reserved,
			COALESCE(SUM(sl.quantity_available), 0) AS total_available
		FROM inventory_items i
		LEFT JOIN stock_levels sl ON sl.inventory_item_id = i.id
		%s
		GROUP BY i.id
		ORDER BY i.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, filter.PageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list inventory items: %w", err)
	}
	defer rows.Close()

	var items []domain.InventoryItemWithStock
	for rows.Next() {
		var item domain.InventoryItemWithStock
		if err := rows.Scan(
			&item.ID, &item.ProductID, &item.VariantID, &item.SKU, &item.Title,
			&item.Status, &item.IsTracked, &item.CreatedAt, &item.UpdatedAt,
			&item.TotalOnHand, &item.TotalReserved, &item.TotalAvailable,
		); err != nil {
			return nil, 0, fmt.Errorf("scan inventory item: %w", err)
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *InventoryRepository) GetWithStock(ctx context.Context, id string) (*domain.InventoryItemWithStock, error) {
	var item domain.InventoryItemWithStock
	err := r.pool.QueryRow(ctx, `
		SELECT i.id, i.product_id, i.variant_id, i.sku, i.title, i.status, i.is_tracked, i.created_at, i.updated_at,
			COALESCE(SUM(sl.quantity_on_hand), 0),
			COALESCE(SUM(sl.quantity_reserved), 0),
			COALESCE(SUM(sl.quantity_available), 0)
		FROM inventory_items i
		LEFT JOIN stock_levels sl ON sl.inventory_item_id = i.id
		WHERE i.id = $1
		GROUP BY i.id
	`, id).Scan(
		&item.ID, &item.ProductID, &item.VariantID, &item.SKU, &item.Title,
		&item.Status, &item.IsTracked, &item.CreatedAt, &item.UpdatedAt,
		&item.TotalOnHand, &item.TotalReserved, &item.TotalAvailable,
	)
	if err == pgx.ErrNoRows {
		return nil, &domain.NotFoundError{Resource: "inventory_item", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("get inventory item with stock: %w", err)
	}
	return &item, nil
}

func (r *InventoryRepository) Upsert(ctx context.Context, tx pgx.Tx, item *domain.InventoryItem) (*domain.InventoryItem, error) {
	var result domain.InventoryItem
	err := tx.QueryRow(ctx, `
		INSERT INTO inventory_items (product_id, variant_id, sku, title, status, is_tracked)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (product_id, COALESCE(variant_id, '__PRODUCT__'))
		DO UPDATE SET sku = EXCLUDED.sku, title = EXCLUDED.title, status = EXCLUDED.status, updated_at = NOW()
		RETURNING id, product_id, variant_id, sku, title, status, is_tracked, created_at, updated_at
	`, item.ProductID, item.VariantID, item.SKU, item.Title, item.Status, item.IsTracked).Scan(
		&result.ID, &result.ProductID, &result.VariantID, &result.SKU, &result.Title,
		&result.Status, &result.IsTracked, &result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert inventory item: %w", err)
	}
	return &result, nil
}

func (r *InventoryRepository) UpdateStatus(ctx context.Context, productID, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE inventory_items SET status = $1, updated_at = NOW() WHERE product_id = $2
	`, status, productID)
	if err != nil {
		return fmt.Errorf("update inventory item status: %w", err)
	}
	return nil
}

func (r *InventoryRepository) Update(ctx context.Context, id string, input domain.UpdateInventoryItemInput) (*domain.InventoryItem, error) {
	var sets []string
	var args []interface{}
	argIdx := 1

	if input.IsTracked != nil {
		sets = append(sets, fmt.Sprintf("is_tracked = $%d", argIdx))
		args = append(args, *input.IsTracked)
		argIdx++
	}

	if len(sets) == 0 {
		return r.GetByID(ctx, id)
	}

	sets = append(sets, "updated_at = NOW()")
	query := fmt.Sprintf("UPDATE inventory_items SET %s WHERE id = $%d RETURNING id, product_id, variant_id, sku, title, status, is_tracked, created_at, updated_at",
		strings.Join(sets, ", "), argIdx)
	args = append(args, id)

	var item domain.InventoryItem
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&item.ID, &item.ProductID, &item.VariantID, &item.SKU, &item.Title,
		&item.Status, &item.IsTracked, &item.CreatedAt, &item.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, &domain.NotFoundError{Resource: "inventory_item", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("update inventory item: %w", err)
	}
	return &item, nil
}

func (r *InventoryRepository) GetByProductID(ctx context.Context, productID string) ([]domain.InventoryItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, product_id, variant_id, sku, title, status, is_tracked, created_at, updated_at
		FROM inventory_items WHERE product_id = $1
	`, productID)
	if err != nil {
		return nil, fmt.Errorf("get inventory items by product: %w", err)
	}
	defer rows.Close()

	var items []domain.InventoryItem
	for rows.Next() {
		var item domain.InventoryItem
		if err := rows.Scan(
			&item.ID, &item.ProductID, &item.VariantID, &item.SKU, &item.Title,
			&item.Status, &item.IsTracked, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan inventory item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *InventoryRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}
