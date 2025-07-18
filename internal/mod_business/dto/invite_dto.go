package dto

import "time"

// CreateInviteRequest request untuk create invite
type CreateInviteRequest struct {
	Role      string    `json:"role" validate:"required,business_role"`
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