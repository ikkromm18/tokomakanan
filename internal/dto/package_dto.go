package dto

import "time"

type PackageItemRequest struct {
	ProductID uint64 `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,gt=0"`
}

type CreatePackageRequest struct {
	Name      string               `json:"name" binding:"required,max=150"`
	SellPrice float64              `json:"sell_price" binding:"required,gte=0"`
	Items     []PackageItemRequest `json:"items" binding:"required,min=1,dive"`
}

type UpdatePackageRequest struct {
	Name      string               `json:"name" binding:"required,max=150"`
	SellPrice float64              `json:"sell_price" binding:"required,gte=0"`
	IsActive  *bool                `json:"is_active" binding:"required"`
	Items     []PackageItemRequest `json:"items" binding:"required,min=1,dive"`
}

type PackageItemResponse struct {
	ID          uint64   `json:"id"`
	ProductID   uint64   `json:"product_id"`
	ProductName string   `json:"product_name"`
	Quantity    int      `json:"quantity"`
	UnitHPP     *float64 `json:"unit_hpp,omitempty"`
	UnitSell    float64  `json:"unit_sell"`
}

type PackageResponse struct {
	ID        uint64                `json:"id"`
	Name      string                `json:"name"`
	TotalHPP  *float64              `json:"total_hpp,omitempty"`
	SellPrice float64               `json:"sell_price"`
	IsActive  bool                  `json:"is_active"`
	Items     []PackageItemResponse `json:"items,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}
