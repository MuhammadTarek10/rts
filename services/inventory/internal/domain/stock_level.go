package domain

import "time"

type StockLevel struct {
	ID                string    `json:"id"`
	InventoryItemID   string    `json:"inventory_item_id"`
	WarehouseID       string    `json:"warehouse_id"`
	QuantityOnHand    int       `json:"quantity_on_hand"`
	QuantityReserved  int       `json:"quantity_reserved"`
	QuantityAvailable int       `json:"quantity_available"`
	ReorderPoint      int       `json:"reorder_point"`
	ReorderQuantity   int       `json:"reorder_quantity"`
	UpdatedAt         time.Time `json:"updated_at"`
	Version           int       `json:"version"`
}
