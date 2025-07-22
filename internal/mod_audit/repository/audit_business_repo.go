// internal/mod_audit/repository/audit_business_repo.go
package repository

import (
	"database/sql"
	"encoding/json"

	"github.com/atam/atamlink/internal/mod_audit/entity"
	"github.com/atam/atamlink/pkg/errors"
)

// AuditRepository interface untuk business audit repository
type AuditRepository interface {
	Create(log *entity.AuditLogBusiness) error
	BatchCreate(logs []*entity.AuditLogBusiness) error
}

type auditRepository struct {
	db *sql.DB
}

// NewAuditRepository membuat instance business audit repository baru
func NewAuditRepository(db *sql.DB) AuditRepository {
	return &auditRepository{db: db}
}

// nullableJSON converts json.RawMessage to sql-safe value
func nullableJSON(data json.RawMessage) interface{} {
	if len(data) == 0 {
		return nil
	}
	
	// Validate JSON
	if !json.Valid(data) {
		return nil
	}
	
	return data
}

// Create menyimpan business audit log
func (r *auditRepository) Create(log *entity.AuditLogBusiness) error {
	contextJSON, err := json.Marshal(log.Context)
	if err != nil {
		return errors.Wrap(err, "failed to marshal context")
	}

	query := `
		INSERT INTO atamlink.audit_logs_business (
			alb_timestamp, alb_user_profile_id, alb_business_id, alb_action,
			alb_table_name, alb_record_id, alb_old_data, alb_new_data,
			alb_context, alb_reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING alb_id`

	err = r.db.QueryRow(
		query,
		log.Timestamp,
		log.UserProfileID,
		log.BusinessID,
		log.Action,
		log.Table,
		log.RecordID,
		nullableJSON(log.OldData),
		nullableJSON(log.NewData),
		contextJSON,
		log.Reason,
	).Scan(&log.ID)

	if err != nil {
		return errors.Wrap(err, "failed to create business audit log")
	}

	return nil
}

// BatchCreate menyimpan multiple business audit logs
func (r *auditRepository) BatchCreate(logs []*entity.AuditLogBusiness) error {
	if len(logs) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO atamlink.audit_logs_business (
			alb_timestamp, alb_user_profile_id, alb_business_id, alb_action,
			alb_table_name, alb_record_id, alb_old_data, alb_new_data,
			alb_context, alb_reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING alb_id`)

	if err != nil {
		return errors.Wrap(err, "failed to prepare batch insert statement")
	}
	defer stmt.Close()

	for _, log := range logs {
		contextJSON, err := json.Marshal(log.Context)
		if err != nil {
			return errors.Wrap(err, "failed to marshal context")
		}

		err = stmt.QueryRow(
			log.Timestamp,
			log.UserProfileID,
			log.BusinessID,
			log.Action,
			log.Table,
			log.RecordID,
			nullableJSON(log.OldData),
			nullableJSON(log.NewData),
			contextJSON,
			log.Reason,
		).Scan(&log.ID)

		if err != nil {
			return errors.Wrap(err, "failed to insert business audit log")
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit batch insert")
	}

	return nil
}