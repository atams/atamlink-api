package entity

import (
	"time"
)

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

// TableName mendapatkan nama tabel
func (BusinessSubscription) TableName() string {
	return "atamlink.business_subscriptions"
}

// IsActive check apakah subscription aktif
func (bs *BusinessSubscription) IsActive() bool {
	return bs.Status == "active" && 
	       bs.ExpiresAt.After(time.Now())
}

// GetDurationDays mendapatkan durasi dalam hari
func (mp *MasterPlan) GetDurationDays() int {
	switch mp.Duration {
	case "monthly":
		return 30
	case "quarterly":
		return 90
	case "yearly":
		return 365
	default:
		return 30
	}
}