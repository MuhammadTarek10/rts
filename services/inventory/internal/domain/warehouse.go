package domain

import "time"

type Warehouse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Code         string    `json:"code"`
	AddressLine1 *string   `json:"address_line1,omitempty"`
	City         *string   `json:"city,omitempty"`
	Country      *string   `json:"country,omitempty"`
	IsActive     bool      `json:"is_active"`
	IsDefault    bool      `json:"is_default"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateWarehouseInput struct {
	Name         string  `json:"name"`
	Code         string  `json:"code"`
	AddressLine1 *string `json:"address_line1,omitempty"`
	City         *string `json:"city,omitempty"`
	Country      *string `json:"country,omitempty"`
	IsDefault    bool    `json:"is_default"`
}

type UpdateWarehouseInput struct {
	Name         *string `json:"name,omitempty"`
	AddressLine1 *string `json:"address_line1,omitempty"`
	City         *string `json:"city,omitempty"`
	Country      *string `json:"country,omitempty"`
	IsDefault    *bool   `json:"is_default,omitempty"`
}
