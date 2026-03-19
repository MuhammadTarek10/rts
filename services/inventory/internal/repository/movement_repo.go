package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rts/inventory/internal/domain"
)

type MovementRepository struct {
	pool *pgxpool.Pool
}

func NewMovementRepository(pool *pgxpool.Pool) *MovementRepository {
	return &MovementRepository{pool: pool}
}

func (r *MovementRepository) Create(ctx context.Context, tx pgx.Tx, m *domain.StockMovement) (*domain.StockMovement, error) {
	var result domain.StockMovement
	err := tx.QueryRow(ctx, `
		INSERT INTO stock_movements (inventory_item_id, warehouse_id, type, quantity, reference_type, reference_id, reason, performed_by, cost_per_unit, currency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, inventory_item_id, warehouse_id, type, quantity, reference_type, reference_id, reason, performed_by, cost_per_unit, currency, created_at
	`, m.InventoryItemID, m.WarehouseID, m.Type, m.Quantity, m.ReferenceType, m.ReferenceID,
		m.Reason, m.PerformedBy, m.CostPerUnit, m.Currency).Scan(
		&result.ID, &result.InventoryItemID, &result.WarehouseID, &result.Type, &result.Quantity,
		&result.ReferenceType, &result.ReferenceID, &result.Reason, &result.PerformedBy,
		&result.CostPerUnit, &result.Currency, &result.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create stock movement: %w", err)
	}
	return &result, nil
}

func (r *MovementRepository) GetByID(ctx context.Context, id string) (*domain.StockMovement, error) {
	var m domain.StockMovement
	err := r.pool.QueryRow(ctx, `
		SELECT id, inventory_item_id, warehouse_id, type, quantity, reference_type, reference_id, reason, performed_by, cost_per_unit, currency, created_at
		FROM stock_movements WHERE id = $1
	`, id).Scan(
		&m.ID, &m.InventoryItemID, &m.WarehouseID, &m.Type, &m.Quantity,
		&m.ReferenceType, &m.ReferenceID, &m.Reason, &m.PerformedBy,
		&m.CostPerUnit, &m.Currency, &m.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, &domain.NotFoundError{Resource: "stock_movement", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("get stock movement: %w", err)
	}
	return &m, nil
}

func (r *MovementRepository) List(ctx context.Context, filter domain.MovementFilter) ([]domain.StockMovement, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	// Default date range: 30 days
	now := time.Now()
	if filter.StartDate == nil {
		t := now.AddDate(0, 0, -30)
		filter.StartDate = &t
	}
	if filter.EndDate == nil {
		filter.EndDate = &now
	}

	// Enforce max 90-day window
	maxWindow := 90 * 24 * time.Hour
	if filter.EndDate.Sub(*filter.StartDate) > maxWindow {
		t := filter.EndDate.Add(-maxWindow)
		filter.StartDate = &t
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
	args = append(args, *filter.StartDate)
	argIdx++

	conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
	args = append(args, *filter.EndDate)
	argIdx++

	if filter.InventoryItemID != "" {
		conditions = append(conditions, fmt.Sprintf("inventory_item_id = $%d", argIdx))
		args = append(args, filter.InventoryItemID)
		argIdx++
	}
	if filter.WarehouseID != "" {
		conditions = append(conditions, fmt.Sprintf("warehouse_id = $%d", argIdx))
		args = append(args, filter.WarehouseID)
		argIdx++
	}
	if filter.Type != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, filter.Type)
		argIdx++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM stock_movements %s", whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count movements: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	query := fmt.Sprintf(`
		SELECT id, inventory_item_id, warehouse_id, type, quantity, reference_type, reference_id, reason, performed_by, cost_per_unit, currency, created_at
		FROM stock_movements %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, filter.PageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list movements: %w", err)
	}
	defer rows.Close()

	var movements []domain.StockMovement
	for rows.Next() {
		var m domain.StockMovement
		if err := rows.Scan(
			&m.ID, &m.InventoryItemID, &m.WarehouseID, &m.Type, &m.Quantity,
			&m.ReferenceType, &m.ReferenceID, &m.Reason, &m.PerformedBy,
			&m.CostPerUnit, &m.Currency, &m.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan movement: %w", err)
		}
		movements = append(movements, m)
	}
	return movements, total, nil
}
