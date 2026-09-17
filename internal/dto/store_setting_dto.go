package dto

import "time"

type UpdateStoreSettingRequest struct {
	Name          string  `json:"name" binding:"required,min=2,max=100"`
	Address       *string `json:"address"`
	Phone         *string `json:"phone"`
	LogoURL       *string `json:"logo_url" binding:"omitempty,url"`
	ReceiptFooter *string `json:"receipt_footer"`
}

type StoreSettingResponse struct {
	ID            uint64    `json:"id"`
	Name          string    `json:"name"`
	Address       *string   `json:"address"`
	Phone         *string   `json:"phone"`
	LogoURL       *string   `json:"logo_url"`
	ReceiptFooter *string   `json:"receipt_footer"`
	UpdatedAt     time.Time `json:"updated_at"`
}
