package testhelpers

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresContainer wraps a test PostgreSQL connection.
// Set TEST_DATABASE_URL env var to point to a running PostgreSQL instance.
// Example: TEST_DATABASE_URL=postgresql://test:test@localhost:5432/inventory_test
type PostgresContainer struct {
	URI  string
	Pool *pgxpool.Pool
}

// StartPostgres connects to a PostgreSQL instance for integration tests.
// Requires TEST_DATABASE_URL environment variable.
func StartPostgres(t *testing.T, ctx context.Context) *PostgresContainer {
	t.Helper()

	uri := os.Getenv("TEST_DATABASE_URL")
	if uri == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping integration test")
	}

	pool, err := pgxpool.New(ctx, uri)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	if err := runTestMigrations(pool, ctx); err != nil {
		pool.Close()
		t.Fatalf("failed to run test migrations: %v", err)
	}

	return &PostgresContainer{
		URI:  uri,
		Pool: pool,
	}
}

func runTestMigrations(pool *pgxpool.Pool, ctx context.Context) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS inventory_items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			product_id VARCHAR NOT NULL,
			variant_id VARCHAR,
			sku VARCHAR(50) NOT NULL,
			title VARCHAR NOT NULL,
			status VARCHAR NOT NULL DEFAULT 'active',
			is_tracked BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_inventory_items_product_variant
			ON inventory_items(product_id, COALESCE(variant_id, '__PRODUCT__'))`,
		`CREATE TABLE IF NOT EXISTS warehouses (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR NOT NULL,
			code VARCHAR NOT NULL UNIQUE,
			address_line1 VARCHAR,
			city VARCHAR,
			country VARCHAR,
			is_active BOOLEAN DEFAULT true,
			is_default BOOLEAN DEFAULT false,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS stock_levels (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			inventory_item_id UUID NOT NULL REFERENCES inventory_items(id),
			warehouse_id UUID NOT NULL REFERENCES warehouses(id),
			quantity_on_hand INT NOT NULL DEFAULT 0 CHECK(quantity_on_hand >= 0),
			quantity_reserved INT NOT NULL DEFAULT 0 CHECK(quantity_reserved >= 0),
			quantity_available INT GENERATED ALWAYS AS (quantity_on_hand - quantity_reserved) STORED,
			reorder_point INT DEFAULT 0,
			reorder_quantity INT DEFAULT 0,
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			version INT DEFAULT 1,
			UNIQUE(inventory_item_id, warehouse_id),
			CHECK(quantity_reserved <= quantity_on_hand)
		)`,
		`CREATE TABLE IF NOT EXISTS stock_movements (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			inventory_item_id UUID NOT NULL REFERENCES inventory_items(id),
			warehouse_id UUID NOT NULL REFERENCES warehouses(id),
			type VARCHAR NOT NULL,
			quantity INT NOT NULL,
			reference_type VARCHAR,
			reference_id VARCHAR,
			reason VARCHAR,
			performed_by UUID,
			cost_per_unit DECIMAL(12,2),
			currency VARCHAR DEFAULT 'USD',
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS reservations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			inventory_item_id UUID NOT NULL REFERENCES inventory_items(id),
			warehouse_id UUID NOT NULL REFERENCES warehouses(id),
			order_id VARCHAR NOT NULL,
			quantity INT NOT NULL CHECK(quantity > 0),
			status VARCHAR NOT NULL DEFAULT 'active',
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			confirmed_at TIMESTAMPTZ,
			released_at TIMESTAMPTZ
		)`,
	}

	for _, m := range migrations {
		if _, err := pool.Exec(ctx, m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	return nil
}

// Cleanup closes the pool. With env-var-based connections there's no container to terminate.
func (pc *PostgresContainer) Cleanup(t *testing.T, _ context.Context) {
	t.Helper()
	pc.Pool.Close()
}
