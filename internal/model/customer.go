package model

import (
	"time"

	"gorm.io/gorm"
)

type Customer struct {
	ID        uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Phone     string         `gorm:"column:phone;type:varchar(20);not null;uniqueIndex:uk_customers_phone" json:"phone"`
	Address   *string        `gorm:"column:address;type:text" json:"address"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index:idx_customers_deleted_at" json:"-"`
}

func (Customer) TableName() string {
	return "customers"
}
