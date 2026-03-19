package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rts/inventory/internal/domain"
	"github.com/rts/inventory/internal/middleware"
	"github.com/rts/inventory/internal/service"
)

type AvailabilityHandler struct {
	svc *service.AvailabilityService
}

func NewAvailabilityHandler(svc *service.AvailabilityService) *AvailabilityHandler {
	return &AvailabilityHandler{svc: svc}
}

// GetAvailability godoc
// @Summary Check SKU availability
// @Description Check if a SKU is available and how many units are in stock
// @Tags Availability
// @Produce json
// @Param sku path string true "SKU"
// @Success 200 {object} service.AvailabilityResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /inventory/availability/{sku} [get]
func (h *AvailabilityHandler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	sku := r.PathValue("sku")
	resp, err := h.svc.GetAvailability(r.Context(), sku)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, resp)
}

// BulkAvailability godoc
// @Summary Bulk check SKU availability
// @Description Check availability for multiple SKUs at once (cart validation)
// @Tags Availability
// @Accept json
// @Produce json
// @Param body body BulkAvailabilityRequest true "List of SKUs"
// @Success 200 {array} service.AvailabilityResponse
// @Failure 422 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /inventory/availability/bulk [post]
func (h *AvailabilityHandler) BulkAvailability(w http.ResponseWriter, r *http.Request) {
	var input BulkAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, &domain.ValidationError{Message: "invalid request body"})
		return
	}

	if len(input.SKUs) == 0 {
		middleware.WriteError(w, &domain.ValidationError{Field: "skus", Message: "at least one SKU is required"})
		return
	}
	if len(input.SKUs) > 50 {
		middleware.WriteError(w, &domain.ValidationError{Field: "skus", Message: "maximum 50 SKUs per request"})
		return
	}

	results, err := h.svc.GetBulkAvailability(r.Context(), input.SKUs)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, results)
}

type BulkAvailabilityRequest struct {
	SKUs []string `json:"skus"`
}
