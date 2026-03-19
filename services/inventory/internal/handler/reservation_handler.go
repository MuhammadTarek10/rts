package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rts/inventory/internal/domain"
	"github.com/rts/inventory/internal/middleware"
	"github.com/rts/inventory/internal/service"
)

type ReservationHandler struct {
	svc *service.ReservationService
}

func NewReservationHandler(svc *service.ReservationService) *ReservationHandler {
	return &ReservationHandler{svc: svc}
}

// Reserve godoc
// @Summary Reserve stock
// @Description Reserve stock for an order. Auto-selects warehouse with sufficient stock.
// @Tags Reservations
// @Accept json
// @Produce json
// @Param body body domain.ReserveInput true "Reserve data"
// @Success 201 {object} domain.Reservation
// @Failure 409 {object} middleware.ErrorResponseWithDetail
// @Failure 422 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /inventory/reservations [post]
func (h *ReservationHandler) Reserve(w http.ResponseWriter, r *http.Request) {
	var input domain.ReserveInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, &domain.ValidationError{Message: "invalid request body"})
		return
	}

	userID := middleware.GetUserID(r.Context())
	reservation, err := h.svc.Reserve(r.Context(), input, userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, reservation)
}

// Confirm godoc
// @Summary Confirm reservation
// @Description Confirm a reservation after payment. Idempotent.
// @Tags Reservations
// @Accept json
// @Produce json
// @Param body body domain.ConfirmReservationInput true "Confirm data"
// @Success 200 {object} map[string]string
// @Failure 422 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /inventory/reservations/confirm [post]
func (h *ReservationHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	var input domain.ConfirmReservationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, &domain.ValidationError{Message: "invalid request body"})
		return
	}

	if err := h.svc.Confirm(r.Context(), input); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "confirmed"})
}

// Release godoc
// @Summary Release reservation
// @Description Release a reservation (order cancelled). Idempotent.
// @Tags Reservations
// @Accept json
// @Produce json
// @Param body body domain.ReleaseReservationInput true "Release data"
// @Success 200 {object} map[string]string
// @Failure 422 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /inventory/reservations/release [post]
func (h *ReservationHandler) Release(w http.ResponseWriter, r *http.Request) {
	var input domain.ReleaseReservationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, &domain.ValidationError{Message: "invalid request body"})
		return
	}

	if err := h.svc.Release(r.Context(), input); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "released"})
}

// GetByOrderID godoc
// @Summary Get reservations by order
// @Description Get all reservations for a specific order
// @Tags Reservations
// @Produce json
// @Param order_id path string true "Order ID"
// @Success 200 {array} domain.Reservation
// @Failure 500 {object} middleware.ErrorResponse
// @Security BearerAuth
// @Router /inventory/reservations/{order_id} [get]
func (h *ReservationHandler) GetByOrderID(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("order_id")
	reservations, err := h.svc.GetByOrderID(r.Context(), orderID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if reservations == nil {
		reservations = []domain.Reservation{}
	}
	middleware.WriteJSON(w, http.StatusOK, reservations)
}
