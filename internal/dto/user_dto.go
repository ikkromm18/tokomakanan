package dto

import "time"

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required,oneof=superadmin owner admin"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name" binding:"omitempty,min=2,max=100"`
	Email    *string `json:"email" binding:"omitempty,email"`
	Password *string `json:"password" binding:"omitempty,min=8"`
	Role     *string `json:"role" binding:"omitempty,oneof=superadmin owner admin"`
	IsActive *bool   `json:"is_active"`
}

type UserResponse struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
