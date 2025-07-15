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

// BusinessUser entity untuk tabel business_users
type BusinessUser struct {
	ID        int64     `json:"id" db:"bu_id"`
	BusinessID int64    `json:"business_id" db:"bu_b_id"`
	ProfileID  int64    `json:"profile_id" db:"bu_up_id"`
	Role      string    `json:"role" db:"bu_role"`
	IsOwner   bool      `json:"is_owner" db:"bu_is_owner"`
	IsActive  bool      `json:"is_active" db:"bu_is_active"`
	CreatedAt time.Time `json:"created_at" db:"bu_created_at"`

	// Relations
	Business *Business    `json:"business,omitempty"`
	Profile  *UserProfile `json:"profile,omitempty"`
}

// BusinessInvite entity untuk tabel business_invites
type BusinessInvite struct {
	ID        int64     `json:"id" db:"bi_id"`
	BusinessID int64    `json:"business_id" db:"bi_b_id"`
	Token     string    `json:"token" db:"bi_token"`
	Role      string    `json:"role" db:"bi_role"`
	InvitedBy int64     `json:"invited_by" db:"bi_invited_by"`
	IsUsed    bool      `json:"is_used" db:"bi_is_used"`
	ExpiresAt time.Time `json:"expires_at" db:"bi_expires_at"`
	CreatedAt time.Time `json:"created_at" db:"bi_created_at"`

	// Relations
	Business   *Business    `json:"business,omitempty"`
	InvitedUser *UserProfile `json:"invited_user,omitempty"`
}

// BusinessSubscription entity untuk tabel business_subscriptions
type BusinessSubscription struct {
	ID        int64      `json:"id" db:"bs_id"`
	BusinessID int64     `json:"business_id" db:"bs_b_id"`
	PlanID    int64      `json:"plan_id" db:"bs_mp_id"`
	Status    string     `json:"status" db:"bs_status"`
	StartsAt  time.Time  `json:"starts_at" db:"bs_starts_at"`
	ExpiresAt time.Time  `json:"expires_at" db:"bs_expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"bs_created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"bs_updated_at"`

	// Relations
	Business *Business   `json:"business,omitempty"`
	Plan     *MasterPlan `json:"plan,omitempty"`
}

// MasterPlan entity untuk tabel master_plans
type MasterPlan struct {
	ID        int64                  `json:"id" db:"mp_id"`
	Name      string                 `json:"name" db:"mp_name"`
	Price     int                    `json:"price" db:"mp_price"`
	Duration  string                 `json:"duration" db:"mp_duration"`
	Features  map[string]interface{} `json:"features" db:"mp_features"`
	IsActive  bool                   `json:"is_active" db:"mp_is_active"`
	CreatedAt time.Time              `json:"created_at" db:"mp_created_at"`
}

// UserProfile untuk relasi (dari mod_user)
type UserProfile struct {
	ID          int64          `json:"id" db:"up_id"`
	UserID      string         `json:"user_id" db:"up_u_id"`
	Phone       sql.NullString `json:"phone" db:"up_phone"`
	DisplayName sql.NullString `json:"display_name" db:"up_display_name"`
}

func (up *UserProfile) GetDisplayName() string {
	if up.DisplayName.Valid {
		return up.DisplayName.String
	}
	return ""
}

// TableName mendapatkan nama tabel
func (Business) TableName() string {
	return "atamlink.businesses"
}

func (BusinessUser) TableName() string {
	return "atamlink.business_users"
}

func (BusinessInvite) TableName() string {
	return "atamlink.business_invites"
}

func (BusinessSubscription) TableName() string {
	return "atamlink.business_subscriptions"
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

// IsExpired check apakah invite sudah expired
func (bi *BusinessInvite) IsExpired() bool {
	return bi.ExpiresAt.Before(time.Now())
}

// IsValid check apakah invite masih valid
func (bi *BusinessInvite) IsValid() bool {
	return !bi.IsUsed && !bi.IsExpired()
}

// IsActive check apakah subscription aktif
func (bs *BusinessSubscription) IsActive() bool {
	return bs.Status == "active" && 
	       bs.ExpiresAt.After(time.Now())
}