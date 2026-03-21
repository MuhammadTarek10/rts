package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rts/inventory/internal/domain"
	"github.com/rts/inventory/internal/middleware"
	"github.com/rts/inventory/internal/service"
)

type WarehouseHandler struct {
	svc *service.InventoryService
}

func NewWarehouseHandler(svc *service.InventoryService) *WarehouseHandler {
	return &WarehouseHandler{svc: svc}
}

// ListWarehouses godoc
// @Summary List warehouses
// @Description Get all warehouses
// @Tags Warehouses
// @Produce json
// @Success 200 {array} domain.Warehouse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /inventory/warehouses [get]
func (h *WarehouseHandler) ListWarehouses(w http.ResponseWriter, r *http.Request) {
	warehouses, err := h.svc.ListWarehouses(r.Context())
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if warehouses == nil {
		warehouses = []domain.Warehouse{}
	}
	middleware.WriteJSON(w, http.StatusOK, warehouses)
}

// GetWarehouse godoc
// @Summary Get warehouse
// @Description Get a single warehouse by ID
// @Tags Warehouses
// @Produce json
// @Param id path string true "Warehouse ID"
// @Success 200 {object} domain.Warehouse
// @Failure 404 {object} middleware.ErrorResponse
// @Router /inventory/warehouses/{id} [get]
func (h *WarehouseHandler) GetWarehouse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	warehouse, err := h.svc.GetWarehouse(r.Context(), id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, warehouse)
}

// CreateWarehouse godoc
// @Summary Create warehouse
// @Description Create a new warehouse (Admin only)
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param body body domain.CreateWarehouseInput true "Warehouse data"
// @Success 201 {object} domain.Warehouse
// @Failure 409 {object} middleware.ErrorResponse
// @Failure 422 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /inventory/warehouses [post]
func (h *WarehouseHandler) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateWarehouseInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, &domain.ValidationError{Message: "invalid request body"})
		return
	}

	if input.Name == "" {
		middleware.WriteError(w, &domain.ValidationError{Field: "name", Message: "is required"})
		return
	}
	if input.Code == "" {
		middleware.WriteError(w, &domain.ValidationError{Field: "code", Message: "is required"})
		return
	}

	warehouse, err := h.svc.CreateWarehouse(r.Context(), input)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, warehouse)
}

// UpdateWarehouse godoc
// @Summary Update warehouse
// @Description Update an existing warehouse (Admin only)
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param id path string true "Warehouse ID"
// @Param body body domain.UpdateWarehouseInput true "Update fields"
// @Success 200 {object} domain.Warehouse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 422 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /inventory/warehouses/{id} [patch]
func (h *WarehouseHandler) UpdateWarehouse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var input domain.UpdateWarehouseInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, &domain.ValidationError{Message: "invalid request body"})
		return
	}

	warehouse, err := h.svc.UpdateWarehouse(r.Context(), id, input)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, warehouse)
}

// DeactivateWarehouse godoc
// @Summary Deactivate warehouse
// @Description Soft-deactivate a warehouse (Admin only). Rejects if has stock or is default.
// @Tags Warehouses
// @Param id path string true "Warehouse ID"
// @Success 204
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 409 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /inventory/warehouses/{id} [delete]
func (h *WarehouseHandler) DeactivateWarehouse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.DeactivateWarehouse(r.Context(), id); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
