package router

import (
	"net/http"

	"github.com/rts/inventory/internal/handler"
	"github.com/rts/inventory/internal/middleware"
)

func New(
	jwtSecret string,
	swaggerUser string,
	swaggerPass string,
	inventoryHandler *handler.InventoryHandler,
	warehouseHandler *handler.WarehouseHandler,
	movementHandler *handler.MovementHandler,
	reservationHandler *handler.ReservationHandler,
	availabilityHandler *handler.AvailabilityHandler,
) http.Handler {
	mux := http.NewServeMux()

	auth := middleware.JWTAuth(jwtSecret)
	adminAuth := func(h http.HandlerFunc) http.Handler {
		return auth(middleware.RequireAdmin(http.HandlerFunc(h)))
	}

	// Health check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Inventory items (read endpoints are public, write requires admin)
	mux.HandleFunc("GET /api/inventory/items", inventoryHandler.ListItems)
	mux.HandleFunc("GET /api/inventory/items/sku/{sku}", inventoryHandler.GetItemBySKU)
	mux.HandleFunc("GET /api/inventory/stock/{id}", inventoryHandler.GetItemStock)
	mux.HandleFunc("GET /api/inventory/items/{id}", inventoryHandler.GetItem)
	mux.Handle("PUT /api/inventory/items/{id}", adminAuth(inventoryHandler.UpdateItem))

	// Warehouses
	mux.HandleFunc("GET /api/inventory/warehouses", warehouseHandler.ListWarehouses)
	mux.HandleFunc("GET /api/inventory/warehouses/{id}", warehouseHandler.GetWarehouse)
	mux.Handle("POST /api/inventory/warehouses", adminAuth(warehouseHandler.CreateWarehouse))
	mux.Handle("PUT /api/inventory/warehouses/{id}", adminAuth(warehouseHandler.UpdateWarehouse))
	mux.Handle("DELETE /api/inventory/warehouses/{id}", adminAuth(warehouseHandler.DeactivateWarehouse))

	// Movements
	mux.HandleFunc("GET /api/inventory/movements", movementHandler.ListMovements)
	mux.HandleFunc("GET /api/inventory/movements/{id}", movementHandler.GetMovement)
	mux.Handle("POST /api/inventory/movements/receive", adminAuth(movementHandler.Receive))
	mux.Handle("POST /api/inventory/movements/ship", adminAuth(movementHandler.Ship))
	mux.Handle("POST /api/inventory/movements/adjust", adminAuth(movementHandler.Adjust))
	mux.Handle("POST /api/inventory/movements/transfer", adminAuth(movementHandler.Transfer))
	mux.Handle("POST /api/inventory/movements/return", adminAuth(movementHandler.Return))

	// Reservations (all require auth)
	mux.Handle("POST /api/inventory/reservations", adminAuth(reservationHandler.Reserve))
	mux.Handle("POST /api/inventory/reservations/confirm", adminAuth(reservationHandler.Confirm))
	mux.Handle("POST /api/inventory/reservations/release", adminAuth(reservationHandler.Release))
	mux.Handle("GET /api/inventory/reservations/{order_id}", auth(http.HandlerFunc(reservationHandler.GetByOrderID)))

	// Availability (public)
	mux.HandleFunc("GET /api/inventory/availability/{sku}", availabilityHandler.GetAvailability)
	mux.HandleFunc("POST /api/inventory/availability/bulk", availabilityHandler.BulkAvailability)

	// Apply global middleware
	var h http.Handler = mux
	h = middleware.Logging(h)

	return h
}
