package domain

import (
	"encoding/json"
	"time"
)

// Inbound catalog events (camelCase JSON from catalog service)

type CatalogEventEnvelope struct {
	EventType  string          `json:"eventType"`
	OccurredOn time.Time       `json:"occurredOn"`
	Payload    json.RawMessage `json:"payload"`
}

type CatalogProductCreatedPayload struct {
	ProductID string               `json:"productId"`
	SKU       string               `json:"sku"`
	Title     string               `json:"title"`
	BrandID   *string              `json:"brandId"`
	Price     float64              `json:"price"`
	Currency  string               `json:"currency"`
	Variants  []CatalogVariantData `json:"variants"`
}

type CatalogProductUpdatedPayload struct {
	ProductID     string               `json:"productId"`
	SKU           string               `json:"sku"`
	Title         string               `json:"title"`
	ChangedFields []string             `json:"changedFields"`
	Variants      []CatalogVariantData `json:"variants"`
}

type CatalogProductStatusChangedPayload struct {
	ProductID string `json:"productId"`
	OldStatus string `json:"oldStatus"`
	NewStatus string `json:"newStatus"`
}

type CatalogProductDeletedPayload struct {
	ProductID string `json:"productId"`
	SKU       string `json:"sku"`
}

type CatalogVariantData struct {
	VariantID  string            `json:"variantId"`
	SKU        string            `json:"sku"`
	Attributes map[string]string `json:"attributes"`
}

// Outbound inventory events (snake_case JSON)

type InventoryEvent struct {
	EventType  string      `json:"event_type"`
	OccurredOn time.Time   `json:"occurred_on"`
	Payload    interface{} `json:"payload"`
}

type StockUpdatedPayload struct {
	InventoryItemID   string `json:"inventory_item_id"`
	WarehouseID       string `json:"warehouse_id"`
	SKU               string `json:"sku"`
	QuantityOnHand    int    `json:"quantity_on_hand"`
	QuantityAvailable int    `json:"quantity_available"`
}

type StockLowPayload struct {
	InventoryItemID string `json:"inventory_item_id"`
	WarehouseID     string `json:"warehouse_id"`
	SKU             string `json:"sku"`
	Available       int    `json:"available"`
	ReorderPoint    int    `json:"reorder_point"`
}

type StockOutPayload struct {
	InventoryItemID string `json:"inventory_item_id"`
	WarehouseID     string `json:"warehouse_id"`
	SKU             string `json:"sku"`
}

type ReservationEventPayload struct {
	ReservationID   string `json:"reservation_id"`
	InventoryItemID string `json:"inventory_item_id"`
	WarehouseID     string `json:"warehouse_id"`
	OrderID         string `json:"order_id"`
	Quantity        int    `json:"quantity"`
	Status          string `json:"status"`
}
