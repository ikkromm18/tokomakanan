package model

type PackageItem struct {
	ID        uint64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PackageID uint64  `gorm:"column:package_id;not null" json:"package_id"`
	ProductID uint64  `gorm:"column:product_id;not null" json:"product_id"`
	Quantity  int     `gorm:"column:quantity;not null" json:"quantity"`
	Product   Product `gorm:"foreignKey:ProductID;references:ID" json:"product,omitempty"`
}

func (PackageItem) TableName() string {
	return "package_items"
}
