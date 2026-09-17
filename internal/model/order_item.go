package model

import "time"

const (
	ItemTypeProduct = "PRODUCT"
	ItemTypePackage = "PACKAGE"
)

type OrderItem struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderID   uint64    `gorm:"column:order_id;not null" json:"order_id"`
	ItemType  string    `gorm:"column:item_type;type:enum('PRODUCT','PACKAGE');not null" json:"item_type"`
	ItemID    uint64    `gorm:"column:item_id;not null" json:"item_id"`
	ItemName  string    `gorm:"column:item_name;type:varchar(150);not null" json:"item_name"`
	Quantity  int       `gorm:"column:quantity;not null" json:"quantity"`
	UnitPrice float64   `gorm:"column:unit_price;type:decimal(12,2);not null" json:"unit_price"`
	UnitHPP   float64   `gorm:"column:unit_hpp;type:decimal(12,2);not null" json:"unit_hpp"`
	Subtotal  float64   `gorm:"column:subtotal;type:decimal(14,2);not null" json:"subtotal"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (OrderItem) TableName() string {
	return "order_items"
}
