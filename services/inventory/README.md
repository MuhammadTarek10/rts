# Inventory Service

Real-time inventory management service for the RTS platform. Tracks stock quantities across multiple warehouses, manages reservations to prevent overselling, maintains an immutable movement ledger, and synchronizes with the catalog service via RabbitMQ events.

## Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                        Inventory Service (Go)                        │
│                                                                      │
│  ┌──────────┐   ┌──────────┐   ┌─────────────┐   ┌──────────────┐  │
│  │ Handlers │──▶│ Services │──▶│ Repositories│──▶│  PostgreSQL  │  │
│  │ (HTTP)   │   │ (Logic)  │   │   (SQL/pgx) │   │ inventory_db │  │
│  └──────────┘   └────┬─────┘   └─────────────┘   └──────────────┘  │
│                      │                                               │
│            ┌─────────┼──────────┐                                    │
│            ▼         ▼          ▼                                    │
│     ┌──────────┐ ┌────────┐ ┌────────────┐                         │
│     │  Redis   │ │RabbitMQ│ │  RabbitMQ  │                         │
│     │  Cache   │ │Publisher│ │  Consumer  │                         │
│     │ (30s TTL)│ │(events)│ │(catalog)   │                         │
│     └──────────┘ └────────┘ └────────────┘                         │
└──────────────────────────────────────────────────────────────────────┘
```

| Component | Technology                                          | Purpose                                              |
| --------- | --------------------------------------------------- | ---------------------------------------------------- |
| Language  | Go (stdlib `net/http`, Go 1.22+ routing)            | No framework, minimal dependencies                   |
| Database  | PostgreSQL (`inventory_db` on shared instance)      | Source of truth for all inventory data               |
| Cache     | Redis (30s TTL)                                     | Availability query cache, DB fallback if unavailable |
| Messaging | RabbitMQ                                            | Consume catalog events, publish inventory events     |
| Auth      | JWT (shared secret with auth service)               | Role-based: `admin` for mutations                    |
| Docs      | Swagger (swaggo/swag + http-swagger)                | Auto-generated from code annotations                 |
| Port      | 8080 (internal), 3004 (external via docker-compose) |                                                      |

## Project Structure

```
services/inventory/
├── cmd/server/
│   └── main.go                    # Entry point: server, migrations, sweeper
├── internal/
│   ├── config/config.go           # Environment-based configuration
│   ├── domain/                    # Plain structs, constants, error types
│   │   ├── inventory_item.go
│   │   ├── warehouse.go
│   │   ├── stock_level.go
│   │   ├── stock_movement.go
│   │   ├── reservation.go
│   │   ├── events.go
│   │   └── errors.go
│   ├── handler/                   # HTTP handlers (one file per resource)
│   │   ├── inventory_handler.go
│   │   ├── warehouse_handler.go
│   │   ├── movement_handler.go
│   │   ├── reservation_handler.go
│   │   └── availability_handler.go
│   ├── middleware/
│   │   ├── auth.go                # JWT validation + RequireAdmin
│   │   ├── logging.go             # Structured request logging
│   │   ├── error.go               # Domain error → HTTP JSON response
│   │   └── swagger.go             # Basic auth for Swagger UI
│   ├── repository/                # Raw SQL with pgx (no ORM)
│   │   ├── inventory_repo.go
│   │   ├── warehouse_repo.go
│   │   ├── stock_repo.go
│   │   ├── movement_repo.go
│   │   └── reservation_repo.go
│   ├── service/                   # Business logic orchestration
│   │   ├── inventory_service.go
│   │   ├── movement_service.go
│   │   ├── reservation_service.go
│   │   └── availability_service.go
│   ├── consumer/
│   │   └── catalog_consumer.go    # RabbitMQ consumer for catalog events
│   ├── publisher/
│   │   └── event_publisher.go     # Publishes inventory events to RabbitMQ
│   ├── cache/
│   │   └── redis.go               # Redis availability cache
│   └── router/
│       └── router.go              # Route registration + auth middleware
├── migrations/
│   ├── 001_initial.up.sql
│   └── 001_initial.down.sql
├── docs/                          # Generated by swag (do not edit manually)
├── .env.development
├── docker-compose.yml             # Local dev PostgreSQL (port 5442)
├── Dockerfile
├── Makefile
└── go.mod / go.sum
```

## Database Schema

Migrations run automatically on service startup. Five tables form the data model:

### 1. `inventory_items` — Product/Variant Registry

Mirrors the catalog service's products and variants. One row per SKU. Created automatically when the catalog publishes a `product.created` event.

```sql
CREATE TABLE inventory_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      VARCHAR NOT NULL,
    variant_id      VARCHAR,                -- NULL = product-level (no variants)
    sku             VARCHAR(50) NOT NULL,
    title           VARCHAR NOT NULL,       -- denormalized from catalog, kept in sync
    status          VARCHAR NOT NULL DEFAULT 'active',  -- active | archived | draft
    is_tracked      BOOLEAN DEFAULT true,   -- false = unlimited quantity (digital goods)
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Handles NULL variant_id correctly for uniqueness
CREATE UNIQUE INDEX idx_inventory_items_product_variant
    ON inventory_items(product_id, COALESCE(variant_id, '__PRODUCT__'));
```

**Use cases:**

- Lookup item by SKU for availability checks and reservations
- Filter by status (active/archived) for listing endpoints
- `is_tracked = false` skips stock validation (digital goods, services)
- The `COALESCE` index ensures one inventory item per product when no variants exist, and one per variant when they do

### 2. `warehouses` — Physical Storage Locations

Represents fulfillment centers or storage locations. Exactly one warehouse must be marked as default at all times.

```sql
CREATE TABLE warehouses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR NOT NULL,
    code            VARCHAR NOT NULL UNIQUE,   -- e.g. "WH-CAIRO-01"
    address_line1   VARCHAR,
    city            VARCHAR,
    country         VARCHAR,
    is_active       BOOLEAN DEFAULT true,      -- soft-delete
    is_default      BOOLEAN DEFAULT false,     -- exactly one default enforced in app
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
```

**Use cases:**

- Default warehouse receives initial stock (qty 0) when catalog creates a product
- Warehouse deactivation is rejected if it holds stock or is the default
- Default enforcement uses an atomic UPDATE to prevent race conditions

### 3. `stock_levels` — Quantity per Item per Warehouse

The core quantity tracker. `quantity_available` is a PostgreSQL generated column: `on_hand - reserved`. Optimistic locking via `version` prevents concurrent update conflicts.

```sql
CREATE TABLE stock_levels (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inventory_item_id   UUID NOT NULL REFERENCES inventory_items(id),
    warehouse_id        UUID NOT NULL REFERENCES warehouses(id),
    quantity_on_hand    INT NOT NULL DEFAULT 0 CHECK(quantity_on_hand >= 0),
    quantity_reserved   INT NOT NULL DEFAULT 0 CHECK(quantity_reserved >= 0),
    quantity_available  INT GENERATED ALWAYS AS (quantity_on_hand - quantity_reserved) STORED,
    reorder_point       INT DEFAULT 0,
    reorder_quantity    INT DEFAULT 0,
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    version             INT DEFAULT 1,
    UNIQUE(inventory_item_id, warehouse_id),
    CHECK(quantity_reserved <= quantity_on_hand)
);
```

**Use cases:**

- `quantity_available` drives the public availability API — "can the customer buy this?"
- `reorder_point` triggers `inventory.stock.low` events when available drops below it
- `version` field enables optimistic concurrency: UPDATE only succeeds if version matches what was read, preventing lost updates from concurrent requests
- CHECK constraints guarantee data integrity at the database level (reserved never exceeds on_hand, quantities never go negative)

### 4. `stock_movements` — Immutable Audit Ledger

Every stock change creates an append-only ledger entry. Stock is never directly edited — all changes flow through movements.

```sql
CREATE TABLE stock_movements (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inventory_item_id   UUID NOT NULL REFERENCES inventory_items(id),
    warehouse_id        UUID NOT NULL REFERENCES warehouses(id),
    type                VARCHAR NOT NULL,      -- receive | ship | adjust | transfer_in |
                                               -- transfer_out | reserve | release | return
    quantity            INT NOT NULL,           -- positive = inbound, negative = outbound
    reference_type      VARCHAR,               -- purchase_order | sales_order | reservation | etc.
    reference_id        VARCHAR,               -- PO number, order ID, reservation ID
    reason              VARCHAR,               -- free text: "damaged", "customer return"
    performed_by        UUID,                  -- user ID from JWT claims
    cost_per_unit       DECIMAL(12,2),
    currency            VARCHAR DEFAULT 'USD',
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_movements_item_wh ON stock_movements(inventory_item_id, warehouse_id, created_at);
CREATE INDEX idx_movements_ref ON stock_movements(reference_type, reference_id);
CREATE INDEX idx_movements_type ON stock_movements(type, created_at);
```

**Use cases:**

- Full audit trail: who changed what, when, and why
- `reference_type` + `reference_id` links movements to external entities (purchase orders, sales orders)
- Transfer creates two linked movements: `transfer_out` (source) + `transfer_in` (destination) in a single transaction
- List endpoint defaults to 30-day window (max 90) to prevent full table scans

### 5. `reservations` — TTL-Based Stock Holds

Holds stock for pending orders. Prevents overselling by temporarily reducing available quantity. Unreleased reservations auto-expire via a background sweeper.

```sql
CREATE TABLE reservations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inventory_item_id   UUID NOT NULL REFERENCES inventory_items(id),
    warehouse_id        UUID NOT NULL REFERENCES warehouses(id),
    order_id            VARCHAR NOT NULL,
    quantity            INT NOT NULL CHECK(quantity > 0),
    status              VARCHAR NOT NULL DEFAULT 'active',  -- active | confirmed | released | expired
    expires_at          TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    confirmed_at        TIMESTAMPTZ,
    released_at         TIMESTAMPTZ
);

CREATE INDEX idx_reservations_expiry ON reservations(status, expires_at);
CREATE INDEX idx_reservations_order ON reservations(order_id);
CREATE INDEX idx_reservations_item ON reservations(inventory_item_id, warehouse_id, status);
```

**Reservation state machine:**

```
              ┌──────────┐
    Reserve   │  active   │  TTL: 15min default, 60min max
              └────┬──┬───┘
       Payment     │  │  Timeout (sweeper) or Cancel
       confirmed   │  │
              ┌────▼┐ └───▶┌──────────┐
              │confirmed│   │ released  │  stock returned to available
              └────────┘    └──────────┘
                            ┌──────────┐
                            │ expired   │  auto-released by sweeper
                            └──────────┘
```

**Use cases:**

- Order service calls `POST /reservations` when customer starts checkout — stock is held
- `POST /reservations/confirm` on payment success — ships the stock (decreases on_hand)
- `POST /reservations/release` on order cancel — returns reserved qty to available
- Background sweeper runs every 30s, processes expired reservations in batches of 100
- Confirm and release are idempotent: calling twice returns success without side effects

## API Endpoints

### Health Check

| Method | Path          | Auth   | Description                  |
| ------ | ------------- | ------ | ---------------------------- |
| GET    | `/api/health` | Public | Returns `{ "status": "ok" }` |

### Inventory Items

| Method | Path                             | Auth   | Description                                                            |
| ------ | -------------------------------- | ------ | ---------------------------------------------------------------------- |
| GET    | `/api/inventory/items`           | Public | List items. Params: `status`, `sku`, `product_id`, `page`, `page_size` |
| GET    | `/api/inventory/items/{id}`      | Public | Get item with aggregated stock totals across all warehouses            |
| GET    | `/api/inventory/items/sku/{sku}` | Public | Lookup item by SKU                                                     |
| GET    | `/api/inventory/stock/{id}`      | Public | Stock levels per warehouse for an item                                 |
| PUT    | `/api/inventory/items/{id}`      | Admin  | Update `is_tracked`, `reorder_point`, `reorder_quantity`               |

### Warehouses

| Method | Path                             | Auth   | Description                                                                              |
| ------ | -------------------------------- | ------ | ---------------------------------------------------------------------------------------- |
| GET    | `/api/inventory/warehouses`      | Public | List all warehouses                                                                      |
| GET    | `/api/inventory/warehouses/{id}` | Public | Get warehouse by ID                                                                      |
| POST   | `/api/inventory/warehouses`      | Admin  | Create warehouse. Body: `name`, `code`, `address_line1`, `city`, `country`, `is_default` |
| PUT    | `/api/inventory/warehouses/{id}` | Admin  | Update warehouse fields                                                                  |
| DELETE | `/api/inventory/warehouses/{id}` | Admin  | Soft-deactivate. Rejected if has stock or is default                                     |

### Stock Movements

| Method | Path                                | Auth   | Description                                                                                                                                                                |
| ------ | ----------------------------------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| GET    | `/api/inventory/movements`          | Public | List movements. Params: `inventory_item_id`, `warehouse_id`, `type`, `start_date` (RFC3339), `end_date` (RFC3339), `page`, `page_size`. Default 30-day window, max 90 days |
| GET    | `/api/inventory/movements/{id}`     | Public | Get single movement detail                                                                                                                                                 |
| POST   | `/api/inventory/movements/receive`  | Admin  | Receive stock into a warehouse. Auto-creates stock_level if needed                                                                                                         |
| POST   | `/api/inventory/movements/ship`     | Admin  | Ship stock. Returns 409 if insufficient available quantity                                                                                                                 |
| POST   | `/api/inventory/movements/adjust`   | Admin  | Positive or negative adjustment with reason                                                                                                                                |
| POST   | `/api/inventory/movements/transfer` | Admin  | Atomic transfer between two warehouses. Returns 2 movements                                                                                                                |
| POST   | `/api/inventory/movements/return`   | Admin  | Customer return. Increases on_hand                                                                                                                                         |

### Reservations

| Method | Path                                     | Auth  | Description                                                                                                                     |
| ------ | ---------------------------------------- | ----- | ------------------------------------------------------------------------------------------------------------------------------- |
| POST   | `/api/inventory/reservations`            | Admin | Reserve stock by SKU. Auto-selects warehouse. Body: `sku`, `quantity`, `order_id`, `ttl_minutes` (optional, default 15, max 60) |
| POST   | `/api/inventory/reservations/confirm`    | Admin | Confirm reservation (ships stock). Idempotent                                                                                   |
| POST   | `/api/inventory/reservations/release`    | Admin | Release reservation (returns stock). Idempotent                                                                                 |
| GET    | `/api/inventory/reservations/{order_id}` | Auth  | Get all reservations for an order                                                                                               |

### Availability

| Method | Path                                | Auth   | Description                                                                 |
| ------ | ----------------------------------- | ------ | --------------------------------------------------------------------------- |
| GET    | `/api/inventory/availability/{sku}` | Public | Returns `{ sku, available, quantity }`. Redis-cached (30s TTL), DB fallback |
| POST   | `/api/inventory/availability/bulk`  | Public | Check multiple SKUs. Body: `{ "skus": [...] }`. Max 50 per request          |

## Event-Driven Architecture

### Consumed Events (from Catalog Service)

The inventory service binds to the `catalog.exchange` via its own queue `inventory.catalog-events`. All handlers are idempotent (UPSERT logic).

| Catalog Event                    | Inventory Action                                                                                                                      |
| -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| `catalog.product.created`        | Create `inventory_item`(s) — one per variant, or one for product if no variants. Create `stock_level` at default warehouse with qty 0 |
| `catalog.product.updated`        | Update denormalized `title` and `sku`. Create/archive items for added/removed variants                                                |
| `catalog.product.status_changed` | Update item status. If archived → release all active reservations                                                                     |
| `catalog.product.deleted`        | Archive all inventory items for the product                                                                                           |

Dead-letter queue (`inventory.catalog-events.dlq`) catches malformed messages. Processing errors are NACK'd with requeue for retry.

### Published Events (to Inventory Exchange)

Published to `inventory.exchange` → `inventory.events` queue. Fire-and-forget (logs warning if RabbitMQ unavailable, does not block the API response).

| Event                             | Trigger                                                          |
| --------------------------------- | ---------------------------------------------------------------- |
| `inventory.stock.updated`         | Any stock_level quantity change                                  |
| `inventory.stock.low`             | `quantity_available` drops to or below `reorder_point` (but > 0) |
| `inventory.stock.out`             | `quantity_available` reaches 0                                   |
| `inventory.reservation.created`   | New reservation created                                          |
| `inventory.reservation.confirmed` | Reservation confirmed (payment received)                         |
| `inventory.reservation.released`  | Reservation released (cancelled) or expired (sweeper)            |

## Error Handling

Domain error types map to HTTP status codes:

| Error Type               | HTTP Status | Code                 | When                                                  |
| ------------------------ | ----------- | -------------------- | ----------------------------------------------------- |
| `NotFoundError`          | 404         | `NOT_FOUND`          | Resource doesn't exist                                |
| `ValidationError`        | 422         | `VALIDATION_ERROR`   | Invalid input (includes field name)                   |
| `ConflictError`          | 409         | `CONFLICT`           | State conflict (e.g., warehouse has stock)            |
| `InsufficientStockError` | 409         | `INSUFFICIENT_STOCK` | Not enough available qty (includes `available` count) |
| `VersionConflictError`   | 409         | `VERSION_CONFLICT`   | Optimistic locking failure (concurrent update)        |
| `InternalError`          | 500         | `INTERNAL_ERROR`     | Unexpected errors (details not leaked to client)      |

## Running the Service

### Standalone (Local Development)

```bash
# 1. Start the local PostgreSQL database (port 5442)
cd services/inventory
docker compose up -d

# 2. Copy env file
cp .env.development .env

# 3. Run the service (auto-runs migrations on startup)
make run

# The service is now available at http://localhost:8080
# Swagger UI at http://localhost:8080/swagger/index.html (admin/admin)
```

If you also need RabbitMQ and Redis locally (for event publishing and caching):

```bash
# From repo root — start shared infrastructure
docker compose up -d rabbitmq redis
```

### With Full Stack (Docker Compose)

```bash
# From repo root
docker compose up -d --build inventory

# Service is available at:
#   Direct:  http://localhost:3004
#   Nginx:   http://localhost/api/inventory/items
#   Swagger: http://localhost/inventory-swagger/index.html
#   Docs redirect: http://localhost/api/inventory/docs
```

### Migrations Only

```bash
# Run migrations and exit (useful for CI/CD)
make migrate
# or
go run ./cmd/server -migrate-only
```

### Environment Variables

| Variable                 | Description                                      | Default                    |
| ------------------------ | ------------------------------------------------ | -------------------------- |
| `PORT`                   | HTTP server port                                 | `8080`                     |
| `DATABASE_URL`           | PostgreSQL connection string                     | **required**               |
| `RABBITMQ_URI`           | RabbitMQ AMQP URI                                | **required**               |
| `REDIS_URL`              | Redis address (host:port)                        | `localhost:6379`           |
| `JWT_ACCESS_SECRET`      | Shared JWT signing secret (same as auth service) | **required**               |
| `RABBITMQ_EXCHANGE_NAME` | Exchange for publishing inventory events         | `inventory.exchange`       |
| `RABBITMQ_QUEUE_NAME`    | Queue for inventory events                       | `inventory.events`         |
| `CATALOG_QUEUE_NAME`     | Queue for consuming catalog events               | `inventory.catalog-events` |
| `SWAGGER_USERNAME`       | Swagger UI basic auth username                   | `admin`                    |
| `SWAGGER_PASSWORD`       | Swagger UI basic auth password                   | `admin`                    |
| `MIGRATIONS_PATH`        | Path to migration files                          | `migrations`               |

## Updating Swagger Documentation

Swagger docs are auto-generated from Go code annotations (comments above handler functions). After changing any handler annotations or adding new endpoints:

```bash
# 1. Install swag CLI (one-time)
go install github.com/swaggo/swag/cmd/swag@latest

# 2. Regenerate docs
make swagger
# or
swag init -g cmd/server/main.go -o docs

# This regenerates three files in docs/:
#   docs.go       — Go source (registered via blank import in main.go)
#   swagger.json  — OpenAPI spec
#   swagger.yaml  — OpenAPI spec (YAML)
```

Swagger annotations live in the handler files (`internal/handler/*.go`) and the main file (`cmd/server/main.go` for global config). Example:

```go
// @Summary Reserve stock
// @Tags reservations
// @Accept json
// @Param body body domain.ReserveInput true "Reservation input"
// @Success 201 {object} domain.Reservation
// @Failure 409 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /inventory/reservations [post]
func (h *ReservationHandler) Reserve(w http.ResponseWriter, r *http.Request) {
```

The generated `docs/` directory is committed to the repo so the Docker build includes it. Always regenerate and commit after changing annotations.

## Running Tests

```bash
# Run all tests
make test
# or
go test ./... -v -count=1

# Run unit tests only (skip integration tests)
go test ./... -short -v

# Run tests for a specific package
go test ./internal/service/... -v
go test ./internal/middleware/... -v

# Run with race detector
go test ./... -race -v
```

### Integration Tests

Integration tests require a running PostgreSQL instance. Set the `TEST_DATABASE_URL` environment variable:

```bash
# Using the local dev database
export TEST_DATABASE_URL="postgresql://postgres:postgres@localhost:5442/inventory_db?sslmode=disable"
go test ./... -v -count=1

# Tests that need the database are skipped if TEST_DATABASE_URL is not set
```

## Makefile Targets

| Target         | Command                                   | Description               |
| -------------- | ----------------------------------------- | ------------------------- |
| `make run`     | `go run ./cmd/server`                     | Start the service locally |
| `make test`    | `go test ./... -v -count=1`               | Run all tests             |
| `make migrate` | `go run ./cmd/server -migrate-only`       | Run migrations and exit   |
| `make swagger` | `swag init -g cmd/server/main.go -o docs` | Regenerate Swagger docs   |
| `make fmt`     | `gofumpt -w .`                            | Format Go code            |
| `make lint`    | `golangci-lint run ./...`                 | Run linter                |

## Concurrency & Data Integrity

The service uses several strategies to ensure correctness under concurrent access:

- **Optimistic locking:** `stock_levels.version` is checked on every update. If another request modified the row since it was read, the update fails with a `VersionConflictError` and the transaction is rolled back
- **`SELECT ... FOR UPDATE`:** Reservation creation locks the target stock level row within the transaction to prevent double-reserving
- **Atomic warehouse default:** Setting a new default warehouse uses a single UPDATE with `CASE WHEN` to clear the old default and set the new one atomically
- **Database CHECK constraints:** `quantity_on_hand >= 0`, `quantity_reserved >= 0`, and `quantity_reserved <= quantity_on_hand` are enforced at the database level as a safety net
- **Immutable ledger:** Stock movements are append-only — no UPDATE or DELETE on `stock_movements`

## Phase 2 — Planned Features

### Low Stock Alerts & Thresholds

- `alert_rules` table: per-item or global thresholds with notification channels
- Alert deduplication with 1-hour cooldown (fire once per threshold crossing)
- API: CRUD for alert rules + list items currently below their reorder point

### Batch/Lot Tracking

- `batches` table: batch number, quantity, cost per unit, expiry date, supplier
- FEFO (First Expired, First Out) shipping logic
- Expiring batch report endpoint
- Batch quantity reconciliation with `stock_levels.quantity_on_hand`

### Inventory Counts / Audits

- `inventory_counts` + `count_lines` tables
- Workflow: draft → in_progress → completed
- Snapshot expected quantity at count start
- Auto-generate adjustment movements for variances on completion

### Reporting Endpoints

```
GET /api/inventory/reports/stock-summary        — current stock by warehouse
GET /api/inventory/reports/movement-history      — aggregated movement history
GET /api/inventory/reports/low-stock             — items below reorder point
GET /api/inventory/reports/out-of-stock          — items with zero available
```

### Grafana Dashboard

- Stock levels heatmap, reservation fill rate, movement velocity
- Cache hit ratio, event processing lag
- Prometheus-compatible `/metrics` endpoint

## TODOs

### Outbox Pattern for Event Publishing (P2)

Currently fire-and-forget — if RabbitMQ is down when a stock change happens, the event is lost. Implement transactional outbox: write events to an `outbox` table in the same transaction as the stock change, then a background worker publishes and marks them as sent. Guarantees at-least-once delivery. Should be done before the orders service ships, since orders will depend on inventory events for stock validation.

### Service-to-Service Authentication (P2)

Currently uses admin JWT tokens for machine-to-machine calls. When the orders service is planned, implement proper service API keys or service-scoped JWTs to distinguish human admin actions from automated service calls in the audit trail.

### Movement Table Partitioning (P3)

When `stock_movements` exceeds ~1M rows, implement PostgreSQL range partitioning by month (`created_at`). Improves query performance for the movement list endpoint and enables efficient archival of old partitions.

## Phase 3 — Future

- Stock valuation (FIFO / LIFO / Weighted Average)
- Supplier / Purchase Order integration
- Multi-currency cost tracking
- Demand forecasting hooks (for recommendation service)
- Real-time WebSocket stock updates for admin UI
