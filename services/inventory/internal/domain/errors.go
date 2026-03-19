package domain

import "fmt"

// NotFoundError is returned when a resource is not found.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s '%s' not found", e.Resource, e.ID)
}

// ConflictError is returned when a resource conflicts with existing state.
type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return e.Message
}

// InsufficientStockError is returned when there is not enough stock.
type InsufficientStockError struct {
	SKU       string
	Requested int
	Available int
}

func (e *InsufficientStockError) Error() string {
	return fmt.Sprintf("insufficient stock for SKU '%s': requested %d, available %d", e.SKU, e.Requested, e.Available)
}

// ValidationError is returned for invalid input.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// InternalError is returned for unexpected server errors.
type InternalError struct {
	Message string
	Err     error
}

func (e *InternalError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("internal error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("internal error: %s", e.Message)
}

func (e *InternalError) Unwrap() error {
	return e.Err
}

// VersionConflictError is returned when optimistic locking fails.
type VersionConflictError struct {
	Resource string
	ID       string
}

func (e *VersionConflictError) Error() string {
	return fmt.Sprintf("version conflict on %s '%s': record was modified by another request", e.Resource, e.ID)
}
