package model

import (
	"time"
)

type StoreSetting struct {
	ID            uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name          string    `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Address       *string   `gorm:"column:address;type:text" json:"address"`
	Phone         *string   `gorm:"column:phone;type:varchar(20)" json:"phone"`
	LogoURL       *string   `gorm:"column:logo_url;type:varchar(500)" json:"logo_url"`
	ReceiptFooter *string   `gorm:"column:receipt_footer;type:text" json:"receipt_footer"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (StoreSetting) TableName() string {
	return "store_settings"
}
