package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	RoleSuperadmin = "superadmin"
	RoleOwner      = "owner"
	RoleAdmin      = "admin"
)

type User struct {
	ID           uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name         string         `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Email        string         `gorm:"column:email;type:varchar(150);not null;uniqueIndex:uk_users_email" json:"email"`
	PasswordHash string         `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	Role         string         `gorm:"column:role;type:enum('superadmin','owner','admin');not null" json:"role"`
	IsActive     bool           `gorm:"column:is_active;type:tinyint(1);not null;default:1" json:"is_active"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index:idx_users_deleted_at" json:"-"`
}

func (User) TableName() string {
	return "users"
}
