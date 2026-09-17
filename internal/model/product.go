package model

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID        uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"column:name;type:varchar(150);not null" json:"name"`
	Category  string         `gorm:"column:category;type:varchar(50);not null" json:"category"`
	HPP       float64        `gorm:"column:hpp;type:decimal(12,2);not null" json:"hpp"`
	SellPrice float64        `gorm:"column:sell_price;type:decimal(12,2);not null" json:"sell_price"`
	IsActive  bool           `gorm:"column:is_active;type:tinyint(1);not null;default:1" json:"is_active"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index:idx_products_deleted_at" json:"-"`
}

func (Product) TableName() string {
	return "products"
}
