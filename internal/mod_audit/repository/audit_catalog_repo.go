// internal/mod_audit/repository/audit_catalog_repo.go
package repository

import (
	"database/sql"
	"encoding/json"

	"github.com/atam/atamlink/internal/mod_audit/entity"
	"github.com/atam/atamlink/pkg/errors"
)

// AuditCatalogRepository interface untuk catalog audit repository
type AuditCatalogRepository interface {
	Create(log *entity.AuditLogCatalog) error
	BatchCreate(logs []*entity.AuditLogCatalog) error
}

type auditCatalogRepository struct {
	db *sql.DB
}

// NewAuditCatalogRepository membuat instance catalog audit repository baru
func NewAuditCatalogRepository(db *sql.DB) AuditCatalogRepository {
	return &auditCatalogRepository{db: db}
}

// nullableJSONCatalog converts json.RawMessage to sql-safe value
func nullableJSONCatalog(data json.RawMessage) interface{} {
	if len(data) == 0 {
		return nil
	}
	
	// Validate JSON
	if !json.Valid(data) {
		return nil
	}
	
	return data
}

// Create menyimpan catalog audit log
func (r *auditCatalogRepository) Create(log *entity.AuditLogCatalog) error {
	contextJSON, err := json.Marshal(log.Context)
	if err != nil {
		return errors.Wrap(err, "failed to marshal context")
	}

	query := `
		INSERT INTO atamlink.audit_logs_catalog (
			alc_timestamp, alc_user_profile_id, alc_catalog_id, alc_action,
			alc_table_name, alc_record_id, alc_old_data, alc_new_data,
			alc_context, alc_reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING alc_id`

	err = r.db.QueryRow(
		query,
		log.Timestamp,
		log.UserProfileID,
		log.CatalogID,
		log.Action,
		log.Table,
		log.RecordID,
		nullableJSONCatalog(log.OldData),
		nullableJSONCatalog(log.NewData),
		contextJSON,
		log.Reason,
	).Scan(&log.ID)

	if err != nil {
		return errors.Wrap(err, "failed to create catalog audit log")
	}

	return nil
}

// BatchCreate menyimpan multiple catalog audit logs
func (r *auditCatalogRepository) BatchCreate(logs []*entity.AuditLogCatalog) error {
	if len(logs) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO atamlink.audit_logs_catalog (
			alc_timestamp, alc_user_profile_id, alc_catalog_id, alc_action,
			alc_table_name, alc_record_id, alc_old_data, alc_new_data,
			alc_context, alc_reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING alc_id`)

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
			log.CatalogID,
			log.Action,
			log.Table,
			log.RecordID,
			nullableJSONCatalog(log.OldData),
			nullableJSONCatalog(log.NewData),
			contextJSON,
			log.Reason,
		).Scan(&log.ID)

		if err != nil {
			return errors.Wrap(err, "failed to insert catalog audit log")
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit batch insert")
	}

	return nil
}