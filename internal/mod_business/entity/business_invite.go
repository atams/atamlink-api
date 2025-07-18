package entity

import (
	"time"
)

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

// TableName mendapatkan nama tabel
func (BusinessInvite) TableName() string {
	return "atamlink.business_invites"
}

// IsExpired check apakah invite sudah expired
func (bi *BusinessInvite) IsExpired() bool {
	return bi.ExpiresAt.Before(time.Now())
}

// IsValid check apakah invite masih valid
func (bi *BusinessInvite) IsValid() bool {
	return !bi.IsUsed && !bi.IsExpired()
}