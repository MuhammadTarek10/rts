package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rts/inventory/internal/handler"
	"github.com/rts/inventory/internal/middleware"
)

type Props struct {
	JWTSecret           string
	InventoryHandler    *handler.InventoryHandler
	WarehouseHandler    *handler.WarehouseHandler
	MovementHandler     *handler.MovementHandler
	ReservationHandler  *handler.ReservationHandler
	AvailabilityHandler *handler.AvailabilityHandler
}

// New sets up all routes, middleware, and handlers for the inventory service.
func New(props Props) http.Handler {
	r := mux.NewRouter().StrictSlash(true)

	// JWT middleware
	auth := middleware.JWTAuth(props.JWTSecret)

	publicInventory := r.PathPrefix(InventoryAPI).Subrouter()

	// --- Public routes ---
	publicInventory.HandleFunc(RouteInventoryItems, props.InventoryHandler.ListItems).Methods(MethodGet)
	publicInventory.HandleFunc(RouteInventoryItemBySKU, props.InventoryHandler.GetItemBySKU).Methods(MethodGet)
	publicInventory.HandleFunc(RouteInventoryItemByID, props.InventoryHandler.GetItem).Methods(MethodGet)
	publicInventory.HandleFunc(RouteInventoryStockByID, props.InventoryHandler.GetItemStock).Methods(MethodGet)

	publicInventory.HandleFunc(RouteMovements, props.MovementHandler.ListMovements).Methods(MethodGet)
	publicInventory.HandleFunc(RouteMovementByID, props.MovementHandler.GetMovement).Methods(MethodGet)

	publicInventory.HandleFunc(RouteWarehouses, props.WarehouseHandler.ListWarehouses).Methods(MethodGet)
	publicInventory.HandleFunc(RouteWarehouseByID, props.WarehouseHandler.GetWarehouse).Methods(MethodGet)

	publicInventory.HandleFunc(RouteAvailabilityBySKU, props.AvailabilityHandler.GetAvailability).Methods(MethodGet)
	publicInventory.HandleFunc(RouteAvailabilityBulk, props.AvailabilityHandler.BulkAvailability).Methods(MethodPost)

	authInventory := r.PathPrefix(InventoryAPI).Subrouter()
	authInventory.Use(auth)

	// Authenticated-only (not necessarily admin)
	authInventory.HandleFunc(RouteReservationsByOrderID, props.ReservationHandler.GetByOrderID).Methods(MethodGet)

	adminInventory := r.PathPrefix(InventoryAPI).Subrouter()
	adminInventory.Use(auth, middleware.RequireAdmin)

	// --- Admin-protected routes ---
	// Inventory write endpoints
	adminInventory.HandleFunc(RouteInventoryItemByID, props.InventoryHandler.UpdateItem).Methods(MethodPatch)

	// Warehouses
	adminInventory.HandleFunc(RouteWarehouses, props.WarehouseHandler.CreateWarehouse).Methods(MethodPost)
	adminInventory.HandleFunc(RouteWarehouseByID, props.WarehouseHandler.UpdateWarehouse).Methods(MethodPatch)
	adminInventory.HandleFunc(RouteWarehouseByID, props.WarehouseHandler.DeactivateWarehouse).Methods(MethodDelete)

	// Movements
	adminInventory.HandleFunc(RouteMovementReceive, props.MovementHandler.Receive).Methods(MethodPost)
	adminInventory.HandleFunc(RouteMovementShip, props.MovementHandler.Ship).Methods(MethodPost)
	adminInventory.HandleFunc(RouteMovementAdjust, props.MovementHandler.Adjust).Methods(MethodPost)
	adminInventory.HandleFunc(RouteMovementTransfer, props.MovementHandler.Transfer).Methods(MethodPost)
	adminInventory.HandleFunc(RouteMovementReturn, props.MovementHandler.Return).Methods(MethodPost)

	// Reservations
	adminInventory.HandleFunc(RouteReservations, props.ReservationHandler.Reserve).Methods(MethodPost)
	adminInventory.HandleFunc(RouteReservationsConfirm, props.ReservationHandler.Confirm).Methods(MethodPost)
	adminInventory.HandleFunc(RouteReservationsRelease, props.ReservationHandler.Release).Methods(MethodPost)

	// --- Health check ---
	r.HandleFunc(RouteHealth, func(w http.ResponseWriter, r *http.Request) {
		middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}).Methods(MethodGet)

	// --- Global middleware ---
	var h http.Handler = r
	h = middleware.Logging(h)

	return h
}
