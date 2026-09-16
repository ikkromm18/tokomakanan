package dto

import "time"

type AuditLogResponse struct {
	ID         uint64    `json:"id"`
	UserID     *uint64   `json:"user_id"`
	Action     string    `json:"action"`
	EntityType string    `json:"entity_type"`
	EntityID   *uint64   `json:"entity_id"`
	OldValue   any       `json:"old_value"`
	NewValue   any       `json:"new_value"`
	IPAddress  *string   `json:"ip_address"`
	CreatedAt  time.Time `json:"created_at"`
}
