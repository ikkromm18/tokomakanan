package dto

import "time"

type CreateCustomerRequest struct {
	Name    string  `json:"name" binding:"required,max=100"`
	Phone   string  `json:"phone" binding:"required,max=20"`
	Address *string `json:"address"`
}

type UpdateCustomerRequest struct {
	Name    string  `json:"name" binding:"required,max=100"`
	Phone   string  `json:"phone" binding:"required,max=20"`
	Address *string `json:"address"`
}

type CustomerResponse struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Address   *string   `json:"address,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
