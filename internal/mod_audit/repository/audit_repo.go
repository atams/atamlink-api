package repository

import (
	"database/sql"
	"encoding/json"

	"github.com/atam/atamlink/internal/mod_audit/entity"
	"github.com/atam/atamlink/pkg/errors"
)

// AuditRepository interface untuk audit repository
type AuditRepository interface {
	Create(log *entity.AuditLog) error
	BatchCreate(logs []*entity.AuditLog) error
}

type auditRepository struct {
	db *sql.DB
}

// NewAuditRepository membuat instance audit repository baru
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

// Create menyimpan audit log
func (r *auditRepository) Create(log *entity.AuditLog) error {
	contextJSON, err := json.Marshal(log.Context)
	if err != nil {
		return errors.Wrap(err, "failed to marshal context")
	}

	query := `
		INSERT INTO atamlink.audit_logs (
			al_timestamp, al_user_profile_id, al_business_id, al_action,
			al_table_name, al_record_id, al_old_data, al_new_data,
			al_context, al_reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING al_id`

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
		return errors.Wrap(err, "failed to create audit log")
	}

	return nil
}

// BatchCreate menyimpan multiple audit logs
func (r *auditRepository) BatchCreate(logs []*entity.AuditLog) error {
	if len(logs) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO atamlink.audit_logs (
			al_timestamp, al_user_profile_id, al_business_id, al_action,
			al_table_name, al_record_id, al_old_data, al_new_data,
			al_context, al_reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	for _, log := range logs {
		contextJSON, err := json.Marshal(log.Context)
		if err != nil {
			return errors.Wrap(err, "failed to marshal context")
		}

		_, err = stmt.Exec(
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
		)
		if err != nil {
			return errors.Wrap(err, "failed to execute statement")
		}
	}

	return tx.Commit()
}