package entity

import (
	"encoding/json"
	"time"
)

// AuditLogCatalog entity untuk tabel audit_logs_catalog
type AuditLogCatalog struct {
	ID            int64                  `json:"id" db:"alc_id"`
	Timestamp     time.Time              `json:"timestamp" db:"alc_timestamp"`
	UserProfileID *int64                 `json:"user_profile_id" db:"alc_user_profile_id"`
	CatalogID     *int64                 `json:"catalog_id" db:"alc_catalog_id"`
	Action        string                 `json:"action" db:"alc_action"`
	Table         string                 `json:"table_name" db:"alc_table_name"`
	RecordID      string                 `json:"record_id" db:"alc_record_id"`
	OldData       json.RawMessage        `json:"old_data" db:"alc_old_data"`
	NewData       json.RawMessage        `json:"new_data" db:"alc_new_data"`
	Context       map[string]interface{} `json:"context" db:"alc_context"`
	Reason        string                 `json:"reason" db:"alc_reason"`
}

// TableName mendapatkan nama tabel
func (AuditLogCatalog) TableName() string {
	return "atamlink.audit_logs_catalog"
}