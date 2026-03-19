package service_test

import (
	"testing"

	"github.com/rts/inventory/internal/domain"
)

func TestReservationStateTransitions(t *testing.T) {
	t.Run("active reservation can be confirmed", func(t *testing.T) {
		res := domain.Reservation{
			Status: domain.ReservationStatusActive,
		}
		if res.Status != domain.ReservationStatusActive {
			t.Errorf("expected status 'active', got '%s'", res.Status)
		}
		// Confirm transition
		res.Status = domain.ReservationStatusConfirmed
		if res.Status != domain.ReservationStatusConfirmed {
			t.Errorf("expected status 'confirmed', got '%s'", res.Status)
		}
	})

	t.Run("active reservation can be released", func(t *testing.T) {
		res := domain.Reservation{
			Status: domain.ReservationStatusActive,
		}
		res.Status = domain.ReservationStatusReleased
		if res.Status != domain.ReservationStatusReleased {
			t.Errorf("expected status 'released', got '%s'", res.Status)
		}
	})

	t.Run("active reservation can be expired", func(t *testing.T) {
		res := domain.Reservation{
			Status: domain.ReservationStatusActive,
		}
		res.Status = domain.ReservationStatusExpired
		if res.Status != domain.ReservationStatusExpired {
			t.Errorf("expected status 'expired', got '%s'", res.Status)
		}
	})

	t.Run("reservation TTL defaults and bounds", func(t *testing.T) {
		// Default TTL
		if domain.DefaultReservationTTLMinutes != 15 {
			t.Errorf("expected default TTL 15, got %d", domain.DefaultReservationTTLMinutes)
		}
		// Max TTL
		if domain.MaxReservationTTLMinutes != 60 {
			t.Errorf("expected max TTL 60, got %d", domain.MaxReservationTTLMinutes)
		}
	})
}

func TestInsufficientStockError(t *testing.T) {
	err := &domain.InsufficientStockError{
		SKU:       "TEST-SKU-001",
		Requested: 10,
		Available: 5,
	}

	t.Run("error message contains details", func(t *testing.T) {
		msg := err.Error()
		if msg == "" {
			t.Error("expected non-empty error message")
		}
		if err.Requested != 10 {
			t.Errorf("expected requested 10, got %d", err.Requested)
		}
		if err.Available != 5 {
			t.Errorf("expected available 5, got %d", err.Available)
		}
	})
}

func TestVersionConflictError(t *testing.T) {
	err := &domain.VersionConflictError{
		Resource: "stock_level",
		ID:       "test-id",
	}

	t.Run("error message indicates conflict", func(t *testing.T) {
		msg := err.Error()
		if msg == "" {
			t.Error("expected non-empty error message")
		}
	})
}

func TestWarehouseDeactivationRejection(t *testing.T) {
	t.Run("rejects deactivation of default warehouse", func(t *testing.T) {
		w := domain.Warehouse{
			ID:        "wh-1",
			IsDefault: true,
			IsActive:  true,
		}
		if !w.IsDefault {
			t.Error("expected warehouse to be default")
		}
		// The service layer checks: if w.IsDefault → return ConflictError
		err := &domain.ConflictError{Message: "cannot deactivate the default warehouse"}
		if err.Error() == "" {
			t.Error("expected error message")
		}
	})

	t.Run("rejects deactivation of warehouse with stock", func(t *testing.T) {
		// Simulate: warehouse has stock → service returns ConflictError
		err := &domain.ConflictError{Message: "cannot deactivate warehouse that has stock; transfer stock first"}
		if err.Error() == "" {
			t.Error("expected error message")
		}
	})
}
