package domain

import "time"

const (
	ReservationStatusActive    = "active"
	ReservationStatusConfirmed = "confirmed"
	ReservationStatusReleased  = "released"
	ReservationStatusExpired   = "expired"
)

type Reservation struct {
	ID              string     `json:"id"`
	InventoryItemID string     `json:"inventory_item_id"`
	WarehouseID     string     `json:"warehouse_id"`
	OrderID         string     `json:"order_id"`
	Quantity        int        `json:"quantity"`
	Status          string     `json:"status"`
	ExpiresAt       time.Time  `json:"expires_at"`
	CreatedAt       time.Time  `json:"created_at"`
	ConfirmedAt     *time.Time `json:"confirmed_at,omitempty"`
	ReleasedAt      *time.Time `json:"released_at,omitempty"`
}

type ReserveInput struct {
	SKU        string `json:"sku"`
	Quantity   int    `json:"quantity"`
	OrderID    string `json:"order_id"`
	TTLMinutes *int   `json:"ttl_minutes,omitempty"`
}

type ConfirmReservationInput struct {
	OrderID string `json:"order_id"`
}

type ReleaseReservationInput struct {
	OrderID string `json:"order_id"`
}

const (
	DefaultReservationTTLMinutes = 15
	MaxReservationTTLMinutes     = 60
)
