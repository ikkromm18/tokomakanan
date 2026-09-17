package model

import (
	"time"

	"gorm.io/gorm"
)

type ProductPackage struct {
	ID        uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"column:name;type:varchar(150);not null" json:"name"`
	TotalHPP  float64        `gorm:"column:total_hpp;type:decimal(12,2);not null" json:"total_hpp"`
	SellPrice float64        `gorm:"column:sell_price;type:decimal(12,2);not null" json:"sell_price"`
	IsActive  bool           `gorm:"column:is_active;type:tinyint(1);not null;default:1" json:"is_active"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index:idx_product_packages_deleted_at" json:"-"`

	Items []PackageItem `gorm:"foreignKey:PackageID;references:ID" json:"items,omitempty"`
}

func (ProductPackage) TableName() string {
	return "product_packages"
}
