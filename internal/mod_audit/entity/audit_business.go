package entity

import (
	"encoding/json"
	"time"
)

// AuditLogBusiness entity untuk tabel audit_logs_business
type AuditLogBusiness struct {
	ID            int64                  `json:"id" db:"alb_id"`
	Timestamp     time.Time              `json:"timestamp" db:"alb_timestamp"`
	UserProfileID *int64                 `json:"user_profile_id" db:"alb_user_profile_id"`
	BusinessID    *int64                 `json:"business_id" db:"alb_business_id"`
	Action        string                 `json:"action" db:"alb_action"`
	Table         string                 `json:"table_name" db:"alb_table_name"`
	RecordID      string                 `json:"record_id" db:"alb_record_id"`
	OldData       json.RawMessage        `json:"old_data" db:"alb_old_data"`
	NewData       json.RawMessage        `json:"new_data" db:"alb_new_data"`
	Context       map[string]interface{} `json:"context" db:"alb_context"`
	Reason        string                 `json:"reason" db:"alb_reason"`
}

// TableName mendapatkan nama tabel
func (AuditLogBusiness) TableName() string {
	return "atamlink.audit_logs_business"
}