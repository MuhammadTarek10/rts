package middleware_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/rts/inventory/internal/domain"
	"github.com/rts/inventory/internal/middleware"
)

func TestWriteError(t *testing.T) {
	t.Run("NotFoundError returns 404", func(t *testing.T) {
		rec := httptest.NewRecorder()
		middleware.WriteError(rec, &domain.NotFoundError{Resource: "item", ID: "123"})
		if rec.Code != 404 {
			t.Errorf("expected 404, got %d", rec.Code)
		}
		var resp middleware.ErrorResponse
		json.NewDecoder(rec.Body).Decode(&resp)
		if resp.Code != "NOT_FOUND" {
			t.Errorf("expected NOT_FOUND, got %s", resp.Code)
		}
	})

	t.Run("ConflictError returns 409", func(t *testing.T) {
		rec := httptest.NewRecorder()
		middleware.WriteError(rec, &domain.ConflictError{Message: "duplicate"})
		if rec.Code != 409 {
			t.Errorf("expected 409, got %d", rec.Code)
		}
	})

	t.Run("InsufficientStockError returns 409 with available quantity", func(t *testing.T) {
		rec := httptest.NewRecorder()
		middleware.WriteError(rec, &domain.InsufficientStockError{SKU: "SKU-1", Requested: 10, Available: 5})
		if rec.Code != 409 {
			t.Errorf("expected 409, got %d", rec.Code)
		}
		var resp middleware.ErrorResponseWithDetail
		json.NewDecoder(rec.Body).Decode(&resp)
		if resp.Code != "INSUFFICIENT_STOCK" {
			t.Errorf("expected INSUFFICIENT_STOCK, got %s", resp.Code)
		}
		if resp.Available != 5 {
			t.Errorf("expected available 5, got %d", resp.Available)
		}
	})

	t.Run("ValidationError returns 422", func(t *testing.T) {
		rec := httptest.NewRecorder()
		middleware.WriteError(rec, &domain.ValidationError{Field: "qty", Message: "must be positive"})
		if rec.Code != 422 {
			t.Errorf("expected 422, got %d", rec.Code)
		}
	})

	t.Run("VersionConflictError returns 409", func(t *testing.T) {
		rec := httptest.NewRecorder()
		middleware.WriteError(rec, &domain.VersionConflictError{Resource: "stock_level", ID: "123"})
		if rec.Code != 409 {
			t.Errorf("expected 409, got %d", rec.Code)
		}
	})

	t.Run("InternalError returns 500", func(t *testing.T) {
		rec := httptest.NewRecorder()
		middleware.WriteError(rec, &domain.InternalError{Message: "db down"})
		if rec.Code != 500 {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})
}
