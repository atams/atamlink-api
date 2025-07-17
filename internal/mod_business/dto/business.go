package dto

import (
	// Remove "mime/multipart" import as it's no longer part of DTO
	"time"
)

// CreateBusinessRequest request untuk create business
type CreateBusinessRequest struct {
	Name string `json:"name" validate:"required,min=3,max=200"`
	Slug string `json:"slug,omitempty" validate:"omitempty,slug,min=3,max=100"`
	Type string `json:"type" validate:"required,oneof=retail service manufacturing technology hospitality healthcare education other"`
	LogoURL *string `json:"logo_url,omitempty"` // Change from LogoFile to LogoURL
}

// UpdateBusinessRequest request untuk update business
type UpdateBusinessRequest struct {
	Name     string `json:"name,omitempty" validate:"omitempty,min=3,max=200"`
	Type     string `json:"type,omitempty" validate:"omitempty,oneof=retail service manufacturing technology hospitality healthcare education other"`
	IsActive *bool  `json:"is_active,omitempty"`
	LogoURL *string `json:"logo_url,omitempty"` // Change from LogoFile to LogoURL
}

// BusinessResponse response untuk business
type BusinessResponse struct {
	ID               int64                  `json:"id"`
	Slug             string                 `json:"slug"`
	Name             string                 `json:"name"`
	LogoURL          *string                `json:"logo_url,omitempty"`
	Type             string                 `json:"type"`
	IsActive         bool                   `json:"is_active"`
	IsSuspended      bool                   `json:"is_suspended"`
	SuspensionReason string                 `json:"suspension_reason,omitempty"`
	CreatedBy        int64                  `json:"created_by"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        *time.Time             `json:"updated_at,omitempty"`
	Users            []BusinessUserResponse `json:"users,omitempty"`
	ActivePlan       *SubscriptionResponse  `json:"active_plan,omitempty"`
}

// BusinessListResponse response untuk list businesses
type BusinessListResponse struct {
	ID          int64      `json:"id"`
	Slug        string     `json:"slug"`
	Name        string     `json:"name"`
	LogoURL     *string    `json:"logo_url,omitempty"`
	Type        string     `json:"type"`
	IsActive    bool       `json:"is_active"`
	IsSuspended bool       `json:"is_suspended"`
	UserCount   int        `json:"user_count"`
	UserRole    *string    `json:"user_role,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// BusinessUserResponse response untuk business user
type BusinessUserResponse struct {
	ID          int64            `json:"id"`
	ProfileID   int64            `json:"profile_id"`
	Role        string           `json:"role"`
	IsOwner     bool             `json:"is_owner"`
	IsActive    bool             `json:"is_active"`
	JoinedAt    time.Time        `json:"joined_at"`
	Profile     *ProfileResponse `json:"profile,omitempty"`
}

// ProfileResponse simple profile response
type ProfileResponse struct {
	ID          int64  `json:"id"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email,omitempty"`
}

// SubscriptionResponse response untuk subscription
type SubscriptionResponse struct {
	ID        int64         `json:"id"`
	PlanID    int64         `json:"plan_id"`
	PlanName  string        `json:"plan_name"`
	Status    string        `json:"status"`
	StartsAt  time.Time     `json:"starts_at"`
	ExpiresAt time.Time     `json:"expires_at"`
	CreatedAt time.Time     `json:"created_at"`
	Plan      *PlanResponse `json:"plan,omitempty"`
}

// PlanResponse simple plan response
type PlanResponse struct {
	ID       int64                  `json:"id"`
	Name     string                 `json:"name"`
	Price    int                    `json:"price"`
	Features map[string]interface{} `json:"features"`
}

// AddUserRequest request untuk add user to business
type AddUserRequest struct {
	ProfileID int64  `json:"profile_id" validate:"required,gt=0"`
	Role      string `json:"role" validate:"required,oneof=owner admin editor viewer"`
}

// UpdateUserRoleRequest request untuk update user role
type UpdateUserRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=owner admin editor viewer"`
}

// CreateInviteRequest request untuk create invite
type CreateInviteRequest struct {
	Role      string    `json:"role" validate:"required,oneof=admin editor viewer"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// InviteResponse response untuk invite
type InviteResponse struct {
	ID        int64     `json:"id"`
	Token     string    `json:"token"`
	Role      string    `json:"role"`
	InvitedBy int64     `json:"invited_by"`
	IsUsed    bool      `json:"is_used"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	InviteURL string    `json:"invite_url,omitempty"`
}

// AcceptInviteRequest request untuk accept invite
type AcceptInviteRequest struct {
	Token     string `json:"token" validate:"required"`
	ProfileID int64  `json:"profile_id" validate:"required,gt=0"`
}

// BusinessFilter filter untuk query businesses
type BusinessFilter struct {
	Search      string     `json:"search,omitempty"`
	Type        string     `json:"type,omitempty"`
	// IsActive    *bool      `json:"is_active,omitempty"`
	IsSuspended *bool      `json:"is_suspended,omitempty"`
	UserID      string     `json:"user_id,omitempty"`
	ProfileID   int64      `json:"profile_id,omitempty"`
	CreatedFrom *time.Time `json:"created_from,omitempty"`
	CreatedTo   *time.Time `json:"created_to,omitempty"`
}

// BusinessStats statistik business
type BusinessStats struct {
	TotalUsers    int `json:"total_users"`
	TotalCatalogs int `json:"total_catalogs"`
	TotalProducts int `json:"total_products"`
	TotalViews    int `json:"total_views"`
}