-- inventory_items: one row per product or variant
CREATE TABLE inventory_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      VARCHAR NOT NULL,
    variant_id      VARCHAR,
    sku             VARCHAR(50) NOT NULL,
    title           VARCHAR NOT NULL,
    status          VARCHAR NOT NULL DEFAULT 'active',
    is_tracked      BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_inventory_items_sku ON inventory_items(sku);
CREATE UNIQUE INDEX idx_inventory_items_product_variant
    ON inventory_items(product_id, COALESCE(variant_id, '__PRODUCT__'));

-- warehouses
CREATE TABLE warehouses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR NOT NULL,
    code            VARCHAR NOT NULL UNIQUE,
    address_line1   VARCHAR,
    city            VARCHAR,
    country         VARCHAR,
    is_active       BOOLEAN DEFAULT true,
    is_default      BOOLEAN DEFAULT false,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- stock_levels: quantity per item per warehouse
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

-- stock_movements: immutable ledger
CREATE TABLE stock_movements (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inventory_item_id   UUID NOT NULL REFERENCES inventory_items(id),
    warehouse_id        UUID NOT NULL REFERENCES warehouses(id),
    type                VARCHAR NOT NULL,
    quantity            INT NOT NULL,
    reference_type      VARCHAR,
    reference_id        VARCHAR,
    reason              VARCHAR,
    performed_by        UUID,
    cost_per_unit       DECIMAL(12,2),
    currency            VARCHAR DEFAULT 'USD',
    created_at          TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_movements_item_wh ON stock_movements(inventory_item_id, warehouse_id, created_at);
CREATE INDEX idx_movements_ref ON stock_movements(reference_type, reference_id);
CREATE INDEX idx_movements_type ON stock_movements(type, created_at);

-- reservations
CREATE TABLE reservations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inventory_item_id   UUID NOT NULL REFERENCES inventory_items(id),
    warehouse_id        UUID NOT NULL REFERENCES warehouses(id),
    order_id            VARCHAR NOT NULL,
    quantity            INT NOT NULL CHECK(quantity > 0),
    status              VARCHAR NOT NULL DEFAULT 'active',
    expires_at          TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    confirmed_at        TIMESTAMPTZ,
    released_at         TIMESTAMPTZ
);
CREATE INDEX idx_reservations_expiry ON reservations(status, expires_at);
CREATE INDEX idx_reservations_order ON reservations(order_id);
CREATE INDEX idx_reservations_item ON reservations(inventory_item_id, warehouse_id, status);
