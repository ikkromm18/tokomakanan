package model

import (
	"time"
)

type AuditLog struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID     *uint64   `gorm:"column:user_id;index:idx_audit_logs_user_id" json:"user_id"`
	Action     string    `gorm:"column:action;type:varchar(30);not null" json:"action"`
	EntityType string    `gorm:"column:entity_type;type:varchar(50);not null;index:idx_audit_logs_entity,priority:1" json:"entity_type"`
	EntityID   *uint64   `gorm:"column:entity_id;index:idx_audit_logs_entity,priority:2" json:"entity_id"`
	OldValue   *string   `gorm:"column:old_value;type:json" json:"old_value"`
	NewValue   *string   `gorm:"column:new_value;type:json" json:"new_value"`
	IPAddress  *string   `gorm:"column:ip_address;type:varchar(45)" json:"ip_address"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
