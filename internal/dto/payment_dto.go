package dto

import "time"

type CreatePaymentRequest struct {
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	PaymentMethod string  `json:"payment_method" binding:"required,oneof=CASH TRANSFER QRIS"`
	Notes         *string `json:"notes"`
}

type PaymentResponse struct {
	ID            uint64    `json:"id"`
	OrderID       uint64    `json:"order_id"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	Notes         *string   `json:"notes,omitempty"`
	PaidAt        time.Time `json:"paid_at"`
	CreatedAt     time.Time `json:"created_at"`
}
