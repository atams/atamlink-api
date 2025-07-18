package dto

import "time"

// BusinessUserResponse response untuk business user
type BusinessUserResponse struct {
	ID        int64            `json:"id"`
	ProfileID int64            `json:"profile_id"`
	Role      string           `json:"role"`
	IsOwner   bool             `json:"is_owner"`
	IsActive  bool             `json:"is_active"`
	JoinedAt  time.Time        `json:"joined_at"`
	Profile   *ProfileResponse `json:"profile,omitempty"`
}

// ProfileResponse profile user
type ProfileResponse struct {
	ID          int64  `json:"id"`
	UserID      string `json:"user_id"`
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

// AddUserRequest request untuk add user to business
type AddUserRequest struct {
	ProfileID int64  `json:"profile_id" validate:"required,gt=0"`
	Role      string `json:"role" validate:"required,business_role"`
}

// UpdateUserRoleRequest request untuk update user role
type UpdateUserRoleRequest struct {
	Role string `json:"role" validate:"required,business_role"`
}