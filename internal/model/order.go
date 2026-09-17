package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	OrderTypeDirectSale = "DIRECT_SALE"
	OrderTypePreOrder   = "PRE_ORDER"

	OrderStatusDraft     = "DRAFT"
	OrderStatusDPPaid    = "DP_PAID"
	OrderStatusPaid      = "PAID"
	OrderStatusReady     = "READY"
	OrderStatusCompleted = "COMPLETED"
	OrderStatusCancelled = "CANCELLED"

	PaymentMethodCash     = "CASH"
	PaymentMethodTransfer = "TRANSFER"
	PaymentMethodQRIS     = "QRIS"
)

type Order struct {
	ID             uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	InvoiceNo      string         `gorm:"column:invoice_no;type:varchar(30);not null;uniqueIndex:uk_orders_invoice_no" json:"invoice_no"`
	InvoiceToken   string         `gorm:"column:invoice_token;type:char(36);not null;uniqueIndex:uk_orders_invoice_token" json:"invoice_token"`
	CustomerID     *uint64        `gorm:"column:customer_id" json:"customer_id"`
	UserID         uint64         `gorm:"column:user_id;not null" json:"user_id"`
	OrderType      string         `gorm:"column:order_type;type:enum('DIRECT_SALE','PRE_ORDER');not null" json:"order_type"`
	Status         string         `gorm:"column:status;type:enum('DRAFT','DP_PAID','PAID','READY','COMPLETED','CANCELLED');not null" json:"status"`
	PickupDate     *time.Time     `gorm:"column:pickup_date;type:date" json:"pickup_date"`
	Notes          *string        `gorm:"column:notes;type:text" json:"notes"`
	Subtotal       float64        `gorm:"column:subtotal;type:decimal(14,2);not null" json:"subtotal"`
	DiscountAmount float64        `gorm:"column:discount_amount;type:decimal(14,2);not null;default:0.00" json:"discount_amount"`
	TotalAmount    float64        `gorm:"column:total_amount;type:decimal(14,2);not null" json:"total_amount"`
	TotalHPP       float64        `gorm:"column:total_hpp;type:decimal(14,2);not null" json:"total_hpp"`
	TotalPaid      float64        `gorm:"column:total_paid;type:decimal(14,2);not null;default:0.00" json:"total_paid"`
	PaymentMethod  string         `gorm:"column:payment_method;type:enum('CASH','TRANSFER','QRIS');not null" json:"payment_method"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index:idx_orders_deleted_at" json:"-"`

	Customer *Customer      `gorm:"foreignKey:CustomerID;references:ID" json:"customer,omitempty"`
	User     *User          `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Items    []OrderItem    `gorm:"foreignKey:OrderID;references:ID" json:"items,omitempty"`
	Payments []OrderPayment `gorm:"foreignKey:OrderID;references:ID" json:"payments,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}
