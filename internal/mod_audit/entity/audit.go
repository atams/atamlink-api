package entity

import (
	"encoding/json"
	"time"
)

// AuditLog entity untuk tabel audit_logs
type AuditLog struct {
	ID            int64                  `json:"id" db:"al_id"`
	Timestamp     time.Time              `json:"timestamp" db:"al_timestamp"`
	UserProfileID *int64                 `json:"user_profile_id" db:"al_user_profile_id"`
	BusinessID    *int64                 `json:"business_id" db:"al_business_id"`
	Action        string                 `json:"action" db:"al_action"`
	Table		  string                 `json:"table_name" db:"al_table_name"`
	RecordID      string                 `json:"record_id" db:"al_record_id"`
	OldData       json.RawMessage        `json:"old_data" db:"al_old_data"`
	NewData       json.RawMessage        `json:"new_data" db:"al_new_data"`
	Context       map[string]interface{} `json:"context" db:"al_context"`
	Reason        string                 `json:"reason" db:"al_reason"`
}

// TableName mendapatkan nama tabel
func (AuditLog) TableName() string {
	return "atamlink.audit_logs"
}