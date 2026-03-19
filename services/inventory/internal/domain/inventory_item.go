package domain

import "time"

type InventoryItem struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	VariantID *string   `json:"variant_id,omitempty"`
	SKU       string    `json:"sku"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	IsTracked bool      `json:"is_tracked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type InventoryItemWithStock struct {
	InventoryItem
	TotalOnHand    int `json:"total_on_hand"`
	TotalReserved  int `json:"total_reserved"`
	TotalAvailable int `json:"total_available"`
}

type UpdateInventoryItemInput struct {
	IsTracked       *bool `json:"is_tracked,omitempty"`
	ReorderPoint    *int  `json:"reorder_point,omitempty"`
	ReorderQuantity *int  `json:"reorder_quantity,omitempty"`
}

type InventoryItemFilter struct {
	Status    string
	SKU       string
	ProductID string
	Page      int
	PageSize  int
}
