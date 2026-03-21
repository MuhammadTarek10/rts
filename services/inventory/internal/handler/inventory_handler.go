package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/rts/inventory/internal/domain"
	"github.com/rts/inventory/internal/middleware"
	"github.com/rts/inventory/internal/service"
)

type InventoryHandler struct {
	svc *service.InventoryService
}

func NewInventoryHandler(svc *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{svc: svc}
}

// ListItems godoc
// @Summary List inventory items
// @Description List all inventory items with optional filtering and pagination
// @Tags Inventory
// @Accept json
// @Produce json
// @Param status query string false "Filter by status (active, archived, draft)"
// @Param sku query string false "Filter by SKU (partial match)"
// @Param product_id query string false "Filter by product ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /inventory/items [get]
func (h *InventoryHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))

	filter := domain.InventoryItemFilter{
		Status:    q.Get("status"),
		SKU:       q.Get("sku"),
		ProductID: q.Get("product_id"),
		Page:      page,
		PageSize:  pageSize,
	}

	items, total, err := h.svc.ListItems(r.Context(), filter)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	if items == nil {
		items = []domain.InventoryItemWithStock{}
	}

	middleware.WriteJSON(w, http.StatusOK, PaginatedResponse{
		Data:     items,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	})
}

// GetItem godoc
// @Summary Get inventory item
// @Description Get a single inventory item with aggregated stock levels
// @Tags Inventory
// @Produce json
// @Param id path string true "Inventory item ID"
// @Success 200 {object} domain.InventoryItemWithStock
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /inventory/items/{id} [get]
func (h *InventoryHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	item, err := h.svc.GetItem(r.Context(), id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, item)
}

// GetItemBySKU godoc
// @Summary Get inventory item by SKU
// @Description Lookup an inventory item by its SKU
// @Tags Inventory
// @Produce json
// @Param sku path string true "SKU"
// @Success 200 {object} domain.InventoryItem
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /inventory/items/sku/{sku} [get]
func (h *InventoryHandler) GetItemBySKU(w http.ResponseWriter, r *http.Request) {
	sku := r.PathValue("sku")
	item, err := h.svc.GetItemBySKU(r.Context(), sku)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, item)
}

// GetItemStock godoc
// @Summary Get stock levels for an item
// @Description Get stock levels per warehouse for a specific inventory item
// @Tags Inventory
// @Produce json
// @Param id path string true "Inventory item ID"
// @Success 200 {array} domain.StockLevel
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /inventory/stock/{id} [get]
func (h *InventoryHandler) GetItemStock(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	levels, err := h.svc.GetItemStockLevels(r.Context(), id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if levels == nil {
		levels = []domain.StockLevel{}
	}
	middleware.WriteJSON(w, http.StatusOK, levels)
}

// UpdateItem godoc
// @Summary Update inventory item settings
// @Description Update an inventory item's tracking and reorder settings
// @Tags Inventory
// @Accept json
// @Produce json
// @Param id path string true "Inventory item ID"
// @Param body body domain.UpdateInventoryItemInput true "Update fields"
// @Success 200 {object} domain.InventoryItem
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 422 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /inventory/items/{id} [patch]
func (h *InventoryHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var input domain.UpdateInventoryItemInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, &domain.ValidationError{Message: "invalid request body"})
		return
	}

	item, err := h.svc.UpdateItem(r.Context(), id, input)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, item)
}

// PaginatedResponse wraps list responses with pagination metadata.
type PaginatedResponse struct {
	Data     any `json:"data"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}
