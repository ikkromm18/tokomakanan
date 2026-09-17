package dto

import "time"

type CreateProductRequest struct {
	Name      string  `json:"name" binding:"required,max=150"`
	Category  string  `json:"category" binding:"required,max=50"`
	HPP       float64 `json:"hpp" binding:"required,gte=0"`
	SellPrice float64 `json:"sell_price" binding:"required,gte=0"`
}

type UpdateProductRequest struct {
	Name      string  `json:"name" binding:"required,max=150"`
	Category  string  `json:"category" binding:"required,max=50"`
	HPP       float64 `json:"hpp" binding:"required,gte=0"`
	SellPrice float64 `json:"sell_price" binding:"required,gte=0"`
	IsActive  *bool   `json:"is_active" binding:"required"`
}

type ProductResponse struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	HPP       *float64  `json:"hpp,omitempty"`
	SellPrice float64   `json:"sell_price"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
