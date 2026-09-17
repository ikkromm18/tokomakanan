package dto

import "time"

type CreateOrderItemRequest struct {
	ItemType string `json:"item_type" binding:"required,oneof=PRODUCT PACKAGE"`
	ItemID   uint64 `json:"item_id" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,gt=0"`
}

type CreateOrderRequest struct {
	CustomerID     *uint64                  `json:"customer_id"`
	OrderType      string                   `json:"order_type" binding:"required,oneof=DIRECT_SALE PRE_ORDER"`
	PickupDate     *string                  `json:"pickup_date"` // YYYY-MM-DD
	Notes          *string                  `json:"notes"`
	DiscountAmount float64                  `json:"discount_amount" binding:"gte=0"`
	PaymentMethod  string                   `json:"payment_method" binding:"required,oneof=CASH TRANSFER QRIS"`
	InitialPaid    float64                  `json:"initial_paid" binding:"gte=0"`
	Items          []CreateOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=DRAFT DP_PAID PAID READY COMPLETED CANCELLED"`
}

type OrderItemResponse struct {
	ID        uint64   `json:"id"`
	ItemType  string   `json:"item_type"`
	ItemID    uint64   `json:"item_id"`
	ItemName  string   `json:"item_name"`
	Quantity  int      `json:"quantity"`
	UnitPrice float64  `json:"unit_price"`
	UnitHPP   *float64 `json:"unit_hpp,omitempty"`
	Subtotal  float64  `json:"subtotal"`
}

type OrderPaymentResponse struct {
	ID            uint64    `json:"id"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	Notes         *string   `json:"notes,omitempty"`
	PaidAt        time.Time `json:"paid_at"`
}

type OrderResponse struct {
	ID             uint64                 `json:"id"`
	InvoiceNo      string                 `json:"invoice_no"`
	InvoiceToken   string                 `json:"invoice_token"`
	CustomerID     *uint64                `json:"customer_id,omitempty"`
	CustomerName   *string                `json:"customer_name,omitempty"`
	UserID         uint64                 `json:"user_id"`
	UserName       string                 `json:"user_name"`
	OrderType      string                 `json:"order_type"`
	Status         string                 `json:"status"`
	PickupDate     *string                `json:"pickup_date,omitempty"`
	Notes          *string                `json:"notes,omitempty"`
	Subtotal       float64                `json:"subtotal"`
	DiscountAmount float64                `json:"discount_amount"`
	TotalAmount    float64                `json:"total_amount"`
	TotalHPP       *float64               `json:"total_hpp,omitempty"`
	TotalPaid      float64                `json:"total_paid"`
	RemainingPaid  float64                `json:"remaining_paid"`
	PaymentMethod  string                 `json:"payment_method"`
	Items          []OrderItemResponse    `json:"items,omitempty"`
	Payments       []OrderPaymentResponse `json:"payments,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}
