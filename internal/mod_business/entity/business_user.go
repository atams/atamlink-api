package entity

import (
	"database/sql"
	"time"
)

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

// UserProfile untuk relasi (dari mod_user)
type UserProfile struct {
	ID          int64          `json:"id" db:"up_id"`
	UserID      string         `json:"user_id" db:"up_u_id"`
	Phone       sql.NullString `json:"phone" db:"up_phone"`
	DisplayName sql.NullString `json:"display_name" db:"up_display_name"`
}

// TableName mendapatkan nama tabel
func (BusinessUser) TableName() string {
	return "atamlink.business_users"
}

// GetDisplayName helper untuk mendapatkan display name
func (up *UserProfile) GetDisplayName() string {
	if up.DisplayName.Valid {
		return up.DisplayName.String
	}
	return ""
}