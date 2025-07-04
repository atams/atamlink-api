package dto

import (
	"time"
)

// CreateUserRequest request untuk create user
type CreateUserRequest struct {
	Email       string `json:"email" validate:"required,email,max=255"`
	Username    string `json:"username" validate:"required,username,min=3,max=100"`
	Password    string `json:"password" validate:"required,min=8,max=100"`
	DisplayName string `json:"display_name,omitempty" validate:"max=200"`
	Phone       string `json:"phone,omitempty" validate:"phone"`
}

// UpdateUserRequest request untuk update user
type UpdateUserRequest struct {
	DisplayName string `json:"display_name,omitempty" validate:"max=200"`
	Phone       string `json:"phone,omitempty" validate:"phone"`
}

// ChangePasswordRequest request untuk change password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=100"`
}

// UserResponse response untuk user
type UserResponse struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Username    string     `json:"username"`
	IsActive    bool       `json:"is_active"`
	IsVerified  bool       `json:"is_verified"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	Profile     *ProfileResponse `json:"profile,omitempty"`
}

// ProfileResponse response untuk user profile
type ProfileResponse struct {
	ID          int64      `json:"id"`
	UserID      string     `json:"user_id"`
	DisplayName string     `json:"display_name,omitempty"`
	Phone       string     `json:"phone,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// UserListResponse response untuk list users
type UserListResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name,omitempty"`
	IsActive    bool      `json:"is_active"`
	IsVerified  bool      `json:"is_verified"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserFilter filter untuk query users
type UserFilter struct {
	Search     string    `json:"search,omitempty"`
	IsActive   *bool     `json:"is_active,omitempty"`
	IsVerified *bool     `json:"is_verified,omitempty"`
	CreatedFrom *time.Time `json:"created_from,omitempty"`
	CreatedTo   *time.Time `json:"created_to,omitempty"`
}