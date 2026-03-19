package middleware

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/rts/inventory/internal/domain"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponseWithDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Available int    `json:"available,omitempty"`
}

// WriteError maps domain errors to HTTP status codes and writes a JSON response.
func WriteError(w http.ResponseWriter, err error) {
	var notFound *domain.NotFoundError
	var conflict *domain.ConflictError
	var insufficientStock *domain.InsufficientStockError
	var validation *domain.ValidationError
	var versionConflict *domain.VersionConflictError
	var internal *domain.InternalError

	switch {
	case errors.As(err, &notFound):
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: notFound.Error(),
		})
	case errors.As(err, &conflict):
		writeJSON(w, http.StatusConflict, ErrorResponse{
			Code:    "CONFLICT",
			Message: conflict.Message,
		})
	case errors.As(err, &insufficientStock):
		writeJSON(w, http.StatusConflict, ErrorResponseWithDetail{
			Code:      "INSUFFICIENT_STOCK",
			Message:   insufficientStock.Error(),
			Available: insufficientStock.Available,
		})
	case errors.As(err, &validation):
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: validation.Error(),
		})
	case errors.As(err, &versionConflict):
		writeJSON(w, http.StatusConflict, ErrorResponse{
			Code:    "VERSION_CONFLICT",
			Message: versionConflict.Error(),
		})
	case errors.As(err, &internal):
		slog.Error("internal error", "error", internal.Error())
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "An internal error occurred.",
		})
	default:
		slog.Error("unhandled error", "error", err.Error())
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "An internal error occurred.",
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write response", "error", err)
	}
}

// WriteJSON is an exported helper for handlers to write JSON responses.
func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	writeJSON(w, status, v)
}
