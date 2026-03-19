package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rts/inventory/internal/domain"
)

type ReservationRepository struct {
	pool *pgxpool.Pool
}

func NewReservationRepository(pool *pgxpool.Pool) *ReservationRepository {
	return &ReservationRepository{pool: pool}
}

func (r *ReservationRepository) Create(ctx context.Context, tx pgx.Tx, res *domain.Reservation) (*domain.Reservation, error) {
	var result domain.Reservation
	err := tx.QueryRow(ctx, `
		INSERT INTO reservations (inventory_item_id, warehouse_id, order_id, quantity, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, inventory_item_id, warehouse_id, order_id, quantity, status, expires_at, created_at, confirmed_at, released_at
	`, res.InventoryItemID, res.WarehouseID, res.OrderID, res.Quantity, res.Status, res.ExpiresAt).Scan(
		&result.ID, &result.InventoryItemID, &result.WarehouseID, &result.OrderID, &result.Quantity,
		&result.Status, &result.ExpiresAt, &result.CreatedAt, &result.ConfirmedAt, &result.ReleasedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create reservation: %w", err)
	}
	return &result, nil
}

func (r *ReservationRepository) GetByOrderID(ctx context.Context, orderID string) ([]domain.Reservation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, inventory_item_id, warehouse_id, order_id, quantity, status, expires_at, created_at, confirmed_at, released_at
		FROM reservations WHERE order_id = $1
		ORDER BY created_at ASC
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("get reservations by order: %w", err)
	}
	defer rows.Close()

	var reservations []domain.Reservation
	for rows.Next() {
		var res domain.Reservation
		if err := rows.Scan(
			&res.ID, &res.InventoryItemID, &res.WarehouseID, &res.OrderID, &res.Quantity,
			&res.Status, &res.ExpiresAt, &res.CreatedAt, &res.ConfirmedAt, &res.ReleasedAt,
		); err != nil {
			return nil, fmt.Errorf("scan reservation: %w", err)
		}
		reservations = append(reservations, res)
	}
	return reservations, nil
}

func (r *ReservationRepository) GetActiveByOrderID(ctx context.Context, tx pgx.Tx, orderID string) ([]domain.Reservation, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, inventory_item_id, warehouse_id, order_id, quantity, status, expires_at, created_at, confirmed_at, released_at
		FROM reservations WHERE order_id = $1 AND status = 'active' FOR UPDATE
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("get active reservations: %w", err)
	}
	defer rows.Close()

	var reservations []domain.Reservation
	for rows.Next() {
		var res domain.Reservation
		if err := rows.Scan(
			&res.ID, &res.InventoryItemID, &res.WarehouseID, &res.OrderID, &res.Quantity,
			&res.Status, &res.ExpiresAt, &res.CreatedAt, &res.ConfirmedAt, &res.ReleasedAt,
		); err != nil {
			return nil, fmt.Errorf("scan reservation: %w", err)
		}
		reservations = append(reservations, res)
	}
	return reservations, nil
}

func (r *ReservationRepository) Confirm(ctx context.Context, tx pgx.Tx, id string) error {
	now := time.Now()
	tag, err := tx.Exec(ctx, `
		UPDATE reservations SET status = 'confirmed', confirmed_at = $1 WHERE id = $2 AND status = 'active'
	`, now, id)
	if err != nil {
		return fmt.Errorf("confirm reservation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return &domain.ConflictError{Message: fmt.Sprintf("reservation '%s' is not in active state", id)}
	}
	return nil
}

func (r *ReservationRepository) Release(ctx context.Context, tx pgx.Tx, id string) error {
	now := time.Now()
	tag, err := tx.Exec(ctx, `
		UPDATE reservations SET status = 'released', released_at = $1 WHERE id = $2 AND status = 'active'
	`, now, id)
	if err != nil {
		return fmt.Errorf("release reservation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return &domain.ConflictError{Message: fmt.Sprintf("reservation '%s' is not in active state", id)}
	}
	return nil
}

func (r *ReservationRepository) GetExpired(ctx context.Context, limit int) ([]domain.Reservation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, inventory_item_id, warehouse_id, order_id, quantity, status, expires_at, created_at, confirmed_at, released_at
		FROM reservations WHERE status = 'active' AND expires_at <= NOW()
		ORDER BY expires_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("get expired reservations: %w", err)
	}
	defer rows.Close()

	var reservations []domain.Reservation
	for rows.Next() {
		var res domain.Reservation
		if err := rows.Scan(
			&res.ID, &res.InventoryItemID, &res.WarehouseID, &res.OrderID, &res.Quantity,
			&res.Status, &res.ExpiresAt, &res.CreatedAt, &res.ConfirmedAt, &res.ReleasedAt,
		); err != nil {
			return nil, fmt.Errorf("scan expired reservation: %w", err)
		}
		reservations = append(reservations, res)
	}
	return reservations, nil
}

func (r *ReservationRepository) Expire(ctx context.Context, tx pgx.Tx, id string) error {
	now := time.Now()
	tag, err := tx.Exec(ctx, `
		UPDATE reservations SET status = 'expired', released_at = $1 WHERE id = $2 AND status = 'active'
	`, now, id)
	if err != nil {
		return fmt.Errorf("expire reservation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return &domain.ConflictError{Message: fmt.Sprintf("reservation '%s' is not in active state", id)}
	}
	return nil
}

func (r *ReservationRepository) GetActiveByItemAndWarehouse(ctx context.Context, itemID, warehouseID string) ([]domain.Reservation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, inventory_item_id, warehouse_id, order_id, quantity, status, expires_at, created_at, confirmed_at, released_at
		FROM reservations WHERE inventory_item_id = $1 AND warehouse_id = $2 AND status = 'active'
	`, itemID, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("get active reservations by item/warehouse: %w", err)
	}
	defer rows.Close()

	var reservations []domain.Reservation
	for rows.Next() {
		var res domain.Reservation
		if err := rows.Scan(
			&res.ID, &res.InventoryItemID, &res.WarehouseID, &res.OrderID, &res.Quantity,
			&res.Status, &res.ExpiresAt, &res.CreatedAt, &res.ConfirmedAt, &res.ReleasedAt,
		); err != nil {
			return nil, fmt.Errorf("scan reservation: %w", err)
		}
		reservations = append(reservations, res)
	}
	return reservations, nil
}

func (r *ReservationRepository) ReleaseByProductID(ctx context.Context, tx pgx.Tx, productID string) error {
	now := time.Now()
	_, err := tx.Exec(ctx, `
		UPDATE reservations SET status = 'released', released_at = $1
		WHERE inventory_item_id IN (SELECT id FROM inventory_items WHERE product_id = $2)
		AND status = 'active'
	`, now, productID)
	if err != nil {
		return fmt.Errorf("release reservations by product: %w", err)
	}
	return nil
}

func (r *ReservationRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}
