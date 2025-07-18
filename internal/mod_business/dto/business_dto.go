package dto

import (
	"time"
)

// CreateBusinessRequest request untuk create business
type CreateBusinessRequest struct {
	Name    string  `json:"name" validate:"required,min=3,max=200"`
	Slug    string  `json:"slug,omitempty" validate:"omitempty,slug,min=3,max=200"`
	Type    string  `json:"type" validate:"required,business_type"`
	LogoURL *string `json:"logo_url,omitempty"`
}

// UpdateBusinessRequest request untuk update business
type UpdateBusinessRequest struct {
	Name     string  `json:"name,omitempty" validate:"omitempty,min=3,max=200"`
	Type     string  `json:"type,omitempty" validate:"omitempty,business_type"`
	IsActive *bool   `json:"is_active,omitempty"`
	LogoURL  *string `json:"logo_url,omitempty"`
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
	ActivePlan       *SubscriptionResponse  `json:"active_plan"`
}

// BusinessListResponse response untuk list business dengan user count
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

// BusinessFilter filter untuk query businesses
type BusinessFilter struct {
	Search      string     `json:"search,omitempty"`
	Type        string     `json:"type,omitempty"`
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