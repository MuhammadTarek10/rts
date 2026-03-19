package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/rts/inventory/internal/domain"
	"github.com/rts/inventory/internal/middleware"
	"github.com/rts/inventory/internal/service"
)

type MovementHandler struct {
	svc *service.MovementService
}

func NewMovementHandler(svc *service.MovementService) *MovementHandler {
	return &MovementHandler{svc: svc}
}

// ListMovements godoc
// @Summary List stock movements
// @Description List stock movements with filtering. Default 30-day window, max 90 days.
// @Tags movements
// @Produce json
// @Param inventory_item_id query string false "Filter by inventory item ID"
// @Param warehouse_id query string false "Filter by warehouse ID"
// @Param type query string false "Filter by movement type"
// @Param start_date query string false "Start date (RFC3339)"
// @Param end_date query string false "End date (RFC3339)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /inventory/movements [get]
func (h *MovementHandler) ListMovements(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))

	filter := domain.MovementFilter{
		InventoryItemID: q.Get("inventory_item_id"),
		WarehouseID:     q.Get("warehouse_id"),
		Type:            q.Get("type"),
		Page:            page,
		PageSize:        pageSize,
	}

	if sd := q.Get("start_date"); sd != "" {
		t, err := time.Parse(time.RFC3339, sd)
		if err != nil {
			middleware.WriteError(w, &domain.ValidationError{Field: "start_date", Message: "invalid RFC3339 format"})
			return
		}
		filter.StartDate = &t
	}
	if ed := q.Get("end_date"); ed != "" {
		t, err := time.Parse(time.RFC3339, ed)
		if err != nil {
			middleware.WriteError(w, &domain.ValidationError{Field: "end_date", Message: "invalid RFC3339 format"})
			return
		}
		filter.EndDate = &t
	}

	movements, total, err := h.svc.ListMovements(r.Context(), filter)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	if movements == nil {
		movements = []domain.StockMovement{}
	}

	middleware.WriteJSON(w, http.StatusOK, PaginatedResponse{
		Data:     movements,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	})
}

// GetMovement godoc
// @Summary Get stock movement
// @Description Get a single stock movement by ID
// @Tags movements
// @Produce json
// @Param id path string true "Movement ID"
// @Success 200 {object} domain.StockMovement
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /inventory/movements/{id} [get]
func (h *MovementHandler) GetMovement(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	movement, err := h.svc.GetMovement(r.Context(), id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, movement)
}

// Receive godoc
// @Summary Receive stock
// @Description Receive stock into a warehouse (Admin only)
// @Tags movements
// @Accept json
// @Produce json
// @Param body body domain.ReceiveInput true "Receive data"
// @Success 201 {object} domain.StockMovement
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 422 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /inventory/movements/receive [post]
func (h *MovementHandler) Receive(w http.ResponseWriter, r *http.Request) {
	var input domain.ReceiveInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, &domain.ValidationError{Message: "invalid request body"})
		return
	}

	userID := middleware.GetUserID(r.Context())
	movement, err := h.svc.Receive(r.Context(), input, userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, movement)
}

// Ship godoc
// @Summary Ship stock
// @Description Ship stock from a warehouse (Admin only)
// @Tags movements
// @Accept json
// @Produce json
// @Param body body domain.ShipInput true "Ship data"
// @Success 201 {object} domain.StockMovement
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 409 {object} middleware.ErrorResponseWithDetail
// @Failure 422 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /inventory/movements/ship [post]
func (h *MovementHandler) Ship(w http.ResponseWriter, r *http.Request) {
	var input domain.ShipInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, &domain.ValidationError{Message: "invalid request body"})
		return
	}

	userID := middleware.GetUserID(r.Context())
	movement, err := h.svc.Ship(r.Context(), input, userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, movement)
}

// Adjust godoc
// @Summary Adjust stock
// @Description Adjust stock with a reason (Admin only). Positive = add, negative = subtract.
// @Tags movements
// @Accept json
// @Produce json
// @Param body body domain.AdjustInput true "Adjustment data"
// @Success 201 {object} domain.StockMovement
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 409 {object} middleware.ErrorResponseWithDetail
// @Failure 422 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /inventory/movements/adjust [post]
func (h *MovementHandler) Adjust(w http.ResponseWriter, r *http.Request) {
	var input domain.AdjustInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, &domain.ValidationError{Message: "invalid request body"})
		return
	}

	userID := middleware.GetUserID(r.Context())
	movement, err := h.svc.Adjust(r.Context(), input, userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, movement)
}

// Transfer godoc
// @Summary Transfer stock between warehouses
// @Description Transfer stock from one warehouse to another atomically (Admin only)
// @Tags movements
// @Accept json
// @Produce json
// @Param body body domain.TransferInput true "Transfer data"
// @Success 201 {array} domain.StockMovement
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 409 {object} middleware.ErrorResponseWithDetail
// @Failure 422 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /inventory/movements/transfer [post]
func (h *MovementHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var input domain.TransferInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, &domain.ValidationError{Message: "invalid request body"})
		return
	}

	userID := middleware.GetUserID(r.Context())
	movements, err := h.svc.Transfer(r.Context(), input, userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, movements)
}

// Return godoc
// @Summary Return stock
// @Description Process a customer return (Admin only)
// @Tags movements
// @Accept json
// @Produce json
// @Param body body domain.ReturnInput true "Return data"
// @Success 201 {object} domain.StockMovement
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 422 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /inventory/movements/return [post]
func (h *MovementHandler) Return(w http.ResponseWriter, r *http.Request) {
	var input domain.ReturnInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, &domain.ValidationError{Message: "invalid request body"})
		return
	}

	userID := middleware.GetUserID(r.Context())
	movement, err := h.svc.Return(r.Context(), input, userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, movement)
}
