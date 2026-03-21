package router

// --- Base prefixes ---
const (
	BaseAPI      = "/api"
	InventoryAPI = BaseAPI + "/inventory"
)

// --- Health ---
const (
	RouteHealth = BaseAPI + "/health"
)

// --- Inventory Items ---
const (
	RouteInventoryItems     = "/items"
	RouteInventoryItemByID  = "/items/{id}"
	RouteInventoryItemBySKU = "/items/sku/{sku}"
	RouteInventoryStockByID = "/stock/{id}"
)

// --- Warehouses ---
const (
	RouteWarehouses    = "/warehouses"
	RouteWarehouseByID = "/warehouses/{id}"
)

// --- Movements ---
const (
	RouteMovements        = "/movements"
	RouteMovementByID     = "/movements/{id}"
	RouteMovementReceive  = "/movements/receive"
	RouteMovementShip     = "/movements/ship"
	RouteMovementAdjust   = "/movements/adjust"
	RouteMovementTransfer = "/movements/transfer"
	RouteMovementReturn   = "/movements/return"
)

// --- Reservations ---
const (
	RouteReservations          = "/reservations"
	RouteReservationsConfirm   = "/reservations/confirm"
	RouteReservationsRelease   = "/reservations/release"
	RouteReservationsByOrderID = "/reservations/{order_id}"
)

// --- Availability ---
const (
	RouteAvailabilityBySKU = "/availability/{sku}"
	RouteAvailabilityBulk  = "/availability/bulk"
)

// --- Methods ---
const (
	MethodGet    = "GET"
	MethodPost   = "POST"
	MethodPut    = "PUT"
	MethodPatch  = "PATCH"
	MethodDelete = "DELETE"
)
