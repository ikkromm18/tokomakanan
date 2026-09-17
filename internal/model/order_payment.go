package model

import "time"

type OrderPayment struct {
	ID            uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderID       uint64    `gorm:"column:order_id;not null" json:"order_id"`
	Amount        float64   `gorm:"column:amount;type:decimal(14,2);not null" json:"amount"`
	PaymentMethod string    `gorm:"column:payment_method;type:enum('CASH','TRANSFER','QRIS');not null" json:"payment_method"`
	Notes         *string   `gorm:"column:notes;type:varchar(255)" json:"notes"`
	PaidAt        time.Time `gorm:"column:paid_at;autoCreateTime" json:"paid_at"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (OrderPayment) TableName() string {
	return "order_payments"
}
