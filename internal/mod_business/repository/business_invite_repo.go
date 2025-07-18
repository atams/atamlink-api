package repository

import (
	"database/sql"
	
	"github.com/atam/atamlink/internal/mod_business/entity"
	"github.com/atam/atamlink/pkg/errors"
)

// CreateInvite membuat invite baru
func (r *businessRepository) CreateInvite(tx *sql.Tx, invite *entity.BusinessInvite) error {
	query := `
		INSERT INTO atamlink.business_invites (
			bi_b_id, bi_token, bi_role, bi_invited_by, 
			bi_is_used, bi_expires_at, bi_created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING bi_id`

	err := tx.QueryRow(
		query,
		invite.BusinessID,
		invite.Token,
		invite.Role,
		invite.InvitedBy,
		invite.IsUsed,
		invite.ExpiresAt,
		invite.CreatedAt,
	).Scan(&invite.ID)

	if err != nil {
		return errors.Wrap(err, "failed to create invite")
	}

	return nil
}

// GetInviteByToken mendapatkan invite by token
func (r *businessRepository) GetInviteByToken(token string) (*entity.BusinessInvite, error) {
	query := `
		SELECT
			bi_id, bi_b_id, bi_token, bi_role, bi_invited_by,
			bi_is_used, bi_expires_at, bi_created_at
		FROM atamlink.business_invites
		WHERE bi_token = $1`

	invite := &entity.BusinessInvite{}
	err := r.db.QueryRow(query, token).Scan(
		&invite.ID,
		&invite.BusinessID,
		&invite.Token,
		&invite.Role,
		&invite.InvitedBy,
		&invite.IsUsed,
		&invite.ExpiresAt,
		&invite.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrNotFound, "Invite tidak ditemukan", 404)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get invite")
	}

	return invite, nil
}

// UseInvite mark invite as used
func (r *businessRepository) UseInvite(tx *sql.Tx, token string) error {
	query := `
		UPDATE atamlink.business_invites
		SET bi_is_used = true
		WHERE bi_token = $1 AND bi_is_used = false`

	result, err := tx.Exec(query, token)
	if err != nil {
		return errors.Wrap(err, "failed to use invite")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrNotFound, "Invite tidak valid atau sudah digunakan", 400)
	}

	return nil
}