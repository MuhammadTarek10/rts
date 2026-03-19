package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rts/inventory/internal/domain"
)

type WarehouseRepository struct {
	pool *pgxpool.Pool
}

func NewWarehouseRepository(pool *pgxpool.Pool) *WarehouseRepository {
	return &WarehouseRepository{pool: pool}
}

func (r *WarehouseRepository) GetByID(ctx context.Context, id string) (*domain.Warehouse, error) {
	var w domain.Warehouse
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, code, address_line1, city, country, is_active, is_default, created_at, updated_at
		FROM warehouses WHERE id = $1
	`, id).Scan(
		&w.ID, &w.Name, &w.Code, &w.AddressLine1, &w.City, &w.Country,
		&w.IsActive, &w.IsDefault, &w.CreatedAt, &w.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, &domain.NotFoundError{Resource: "warehouse", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("get warehouse: %w", err)
	}
	return &w, nil
}

func (r *WarehouseRepository) List(ctx context.Context) ([]domain.Warehouse, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, code, address_line1, city, country, is_active, is_default, created_at, updated_at
		FROM warehouses ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list warehouses: %w", err)
	}
	defer rows.Close()

	var warehouses []domain.Warehouse
	for rows.Next() {
		var w domain.Warehouse
		if err := rows.Scan(
			&w.ID, &w.Name, &w.Code, &w.AddressLine1, &w.City, &w.Country,
			&w.IsActive, &w.IsDefault, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan warehouse: %w", err)
		}
		warehouses = append(warehouses, w)
	}
	return warehouses, nil
}

func (r *WarehouseRepository) Create(ctx context.Context, tx pgx.Tx, input domain.CreateWarehouseInput) (*domain.Warehouse, error) {
	var w domain.Warehouse
	err := tx.QueryRow(ctx, `
		INSERT INTO warehouses (name, code, address_line1, city, country, is_default)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, code, address_line1, city, country, is_active, is_default, created_at, updated_at
	`, input.Name, input.Code, input.AddressLine1, input.City, input.Country, input.IsDefault).Scan(
		&w.ID, &w.Name, &w.Code, &w.AddressLine1, &w.City, &w.Country,
		&w.IsActive, &w.IsDefault, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create warehouse: %w", err)
	}
	return &w, nil
}

func (r *WarehouseRepository) Update(ctx context.Context, id string, input domain.UpdateWarehouseInput) (*domain.Warehouse, error) {
	// Fetch current
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	name := current.Name
	if input.Name != nil {
		name = *input.Name
	}
	addressLine1 := current.AddressLine1
	if input.AddressLine1 != nil {
		addressLine1 = input.AddressLine1
	}
	city := current.City
	if input.City != nil {
		city = input.City
	}
	country := current.Country
	if input.Country != nil {
		country = input.Country
	}

	var w domain.Warehouse
	err = r.pool.QueryRow(ctx, `
		UPDATE warehouses SET name = $1, address_line1 = $2, city = $3, country = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING id, name, code, address_line1, city, country, is_active, is_default, created_at, updated_at
	`, name, addressLine1, city, country, id).Scan(
		&w.ID, &w.Name, &w.Code, &w.AddressLine1, &w.City, &w.Country,
		&w.IsActive, &w.IsDefault, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update warehouse: %w", err)
	}
	return &w, nil
}

func (r *WarehouseRepository) Deactivate(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE warehouses SET is_active = false, updated_at = NOW() WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("deactivate warehouse: %w", err)
	}
	return nil
}

func (r *WarehouseRepository) HasStock(ctx context.Context, warehouseID string) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM stock_levels WHERE warehouse_id = $1 AND quantity_on_hand > 0
	`, warehouseID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check warehouse stock: %w", err)
	}
	return count > 0, nil
}

// ClearAllDefaults clears the default flag on all warehouses. Used before creating a new default warehouse.
func (r *WarehouseRepository) ClearAllDefaults(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `UPDATE warehouses SET is_default = false, updated_at = NOW() WHERE is_default = true`)
	if err != nil {
		return fmt.Errorf("clear default warehouses: %w", err)
	}
	return nil
}

// SetDefaultAtomic clears any existing default and sets the new one in a single atomic UPDATE.
func (r *WarehouseRepository) SetDefaultAtomic(ctx context.Context, tx pgx.Tx, id string) error {
	_, err := tx.Exec(ctx, `
		UPDATE warehouses
		SET is_default = CASE WHEN id = $1 THEN true ELSE false END,
		    updated_at = NOW()
		WHERE is_default = true OR id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("set default warehouse: %w", err)
	}
	return nil
}

func (r *WarehouseRepository) GetDefault(ctx context.Context) (*domain.Warehouse, error) {
	var w domain.Warehouse
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, code, address_line1, city, country, is_active, is_default, created_at, updated_at
		FROM warehouses WHERE is_default = true LIMIT 1
	`).Scan(
		&w.ID, &w.Name, &w.Code, &w.AddressLine1, &w.City, &w.Country,
		&w.IsActive, &w.IsDefault, &w.CreatedAt, &w.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get default warehouse: %w", err)
	}
	return &w, nil
}

func (r *WarehouseRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}
