package domain

import "time"

const (
	MovementTypeReceive     = "receive"
	MovementTypeShip        = "ship"
	MovementTypeAdjust      = "adjust"
	MovementTypeTransferIn  = "transfer_in"
	MovementTypeTransferOut = "transfer_out"
	MovementTypeReserve     = "reserve"
	MovementTypeRelease     = "release"
	MovementTypeReturn      = "return"
)

type StockMovement struct {
	ID              string    `json:"id"`
	InventoryItemID string    `json:"inventory_item_id"`
	WarehouseID     string    `json:"warehouse_id"`
	Type            string    `json:"type"`
	Quantity        int       `json:"quantity"`
	ReferenceType   *string   `json:"reference_type,omitempty"`
	ReferenceID     *string   `json:"reference_id,omitempty"`
	Reason          *string   `json:"reason,omitempty"`
	PerformedBy     *string   `json:"performed_by,omitempty"`
	CostPerUnit     *float64  `json:"cost_per_unit,omitempty"`
	Currency        string    `json:"currency"`
	CreatedAt       time.Time `json:"created_at"`
}

type ReceiveInput struct {
	InventoryItemID string   `json:"inventory_item_id"`
	WarehouseID     string   `json:"warehouse_id"`
	Quantity        int      `json:"quantity"`
	ReferenceType   *string  `json:"reference_type,omitempty"`
	ReferenceID     *string  `json:"reference_id,omitempty"`
	Reason          *string  `json:"reason,omitempty"`
	CostPerUnit     *float64 `json:"cost_per_unit,omitempty"`
	Currency        string   `json:"currency"`
}

type ShipInput struct {
	InventoryItemID string  `json:"inventory_item_id"`
	WarehouseID     string  `json:"warehouse_id"`
	Quantity        int     `json:"quantity"`
	ReferenceType   *string `json:"reference_type,omitempty"`
	ReferenceID     *string `json:"reference_id,omitempty"`
	Reason          *string `json:"reason,omitempty"`
}

type AdjustInput struct {
	InventoryItemID string  `json:"inventory_item_id"`
	WarehouseID     string  `json:"warehouse_id"`
	Quantity        int     `json:"quantity"`
	Reason          *string `json:"reason,omitempty"`
}

type TransferInput struct {
	InventoryItemID string  `json:"inventory_item_id"`
	FromWarehouseID string  `json:"from_warehouse_id"`
	ToWarehouseID   string  `json:"to_warehouse_id"`
	Quantity        int     `json:"quantity"`
	Reason          *string `json:"reason,omitempty"`
}

type ReturnInput struct {
	InventoryItemID string  `json:"inventory_item_id"`
	WarehouseID     string  `json:"warehouse_id"`
	Quantity        int     `json:"quantity"`
	ReferenceType   *string `json:"reference_type,omitempty"`
	ReferenceID     *string `json:"reference_id,omitempty"`
	Reason          *string `json:"reason,omitempty"`
}

type MovementFilter struct {
	InventoryItemID string
	WarehouseID     string
	Type            string
	StartDate       *time.Time
	EndDate         *time.Time
	Page            int
	PageSize        int
}
