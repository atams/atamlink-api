package entity

import (
	"database/sql"
	"time"
)

// Business entity untuk tabel businesses
type Business struct {
	ID               int64          `json:"id" db:"b_id"`
	Slug             string         `json:"slug" db:"b_slug"`
	Name             string         `json:"name" db:"b_name"`
	LogoURL          sql.NullString `json:"logo_url" db:"b_logo_url"`
	Type             string         `json:"type" db:"b_type"`
	IsActive         bool           `json:"is_active" db:"b_is_active"`
	IsSuspended      bool           `json:"is_suspended" db:"b_is_suspended"`
	SuspensionReason sql.NullString `json:"suspension_reason" db:"b_suspension_reason"`
	SuspendedBy      sql.NullInt64  `json:"suspended_by" db:"b_suspended_by"`
	SuspendedAt      *time.Time     `json:"suspended_at" db:"b_suspended_at"`
	CreatedBy        int64          `json:"created_by" db:"b_created_by"`
	CreatedAt        time.Time      `json:"created_at" db:"b_created_at"`
	UpdatedBy        sql.NullInt64  `json:"updated_by" db:"b_updated_by"`
	UpdatedAt        *time.Time     `json:"updated_at" db:"b_updated_at"`

	// Relations
	Users         []BusinessUser         `json:"users,omitempty"`
	Subscriptions []BusinessSubscription `json:"subscriptions,omitempty"`
	ActivePlan    *BusinessSubscription  `json:"active_plan,omitempty"`
}

// TableName mendapatkan nama tabel
func (Business) TableName() string {
	return "atamlink.businesses"
}

// Helper methods

// GetSuspensionReason get suspension reason dengan null handling
func (b *Business) GetSuspensionReason() string {
	if b.SuspensionReason.Valid {
		return b.SuspensionReason.String
	}
	return ""
}

// IsAccessible check apakah business bisa diakses
func (b *Business) IsAccessible() bool {
	return b.IsActive && !b.IsSuspended
}

// HasActiveSubscription check apakah punya subscription aktif
func (b *Business) HasActiveSubscription() bool {
	if b.ActivePlan == nil {
		return false
	}
	return b.ActivePlan.Status == "active" && 
	       b.ActivePlan.ExpiresAt.After(time.Now())
}