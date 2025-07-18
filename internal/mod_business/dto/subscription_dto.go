package dto

import "time"

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

// PlanResponse response untuk plan
type PlanResponse struct {
	ID       int64                  `json:"id"`
	Name     string                 `json:"name"`
	Price    int                    `json:"price"`
	Features map[string]interface{} `json:"features,omitempty"`
}

// ActivateSubscriptionRequest request untuk aktivasi subscription (bypass pembayaran)
type ActivateSubscriptionRequest struct {
	BusinessID int64 `json:"business_id" validate:"required,gt=0"`
	PlanID     int64 `json:"plan_id" validate:"required,gt=0"`
}