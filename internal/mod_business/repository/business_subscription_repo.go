package repository

import (
	"database/sql"
	"encoding/json"
	
	"github.com/atam/atamlink/internal/mod_business/entity"
	"github.com/atam/atamlink/pkg/errors"
)

// GetActiveSubscription mendapatkan active subscription dengan detail plan
func (r *businessRepository) GetActiveSubscription(businessID int64) (*entity.BusinessSubscription, error) {
	query := `
		SELECT
			bs.bs_id, bs.bs_b_id, bs.bs_mp_id, bs.bs_status,
			bs.bs_starts_at, bs.bs_expires_at, bs.bs_created_at, bs.bs_updated_at,
			mp.mp_id, mp.mp_name, mp.mp_price, mp.mp_duration, mp.mp_features, mp.mp_is_active
		FROM atamlink.business_subscriptions bs
		INNER JOIN atamlink.master_plans mp ON mp.mp_id = bs.bs_mp_id
		WHERE bs.bs_b_id = $1
			AND bs.bs_status = 'active'
			AND bs.bs_expires_at > NOW()
		ORDER BY bs.bs_created_at DESC
		LIMIT 1`

	sub := &entity.BusinessSubscription{
		Plan: &entity.MasterPlan{},
	}

	var featuresJSON []byte
	err := r.db.QueryRow(query, businessID).Scan(
		&sub.ID,
		&sub.BusinessID,
		&sub.PlanID,
		&sub.Status,
		&sub.StartsAt,
		&sub.ExpiresAt,
		&sub.CreatedAt,
		&sub.UpdatedAt,
		&sub.Plan.ID,
		&sub.Plan.Name,
		&sub.Plan.Price,
		&sub.Plan.Duration,
		&featuresJSON,
		&sub.Plan.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Return nil jika tidak ada subscription aktif
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get active subscription")
	}

	// Parse features JSON
	if len(featuresJSON) > 0 {
		if err := json.Unmarshal(featuresJSON, &sub.Plan.Features); err != nil {
			return nil, errors.Wrap(err, "failed to parse plan features")
		}
	}

	return sub, nil
}

// CreateSubscription membuat subscription baru
func (r *businessRepository) CreateSubscription(tx *sql.Tx, subscription *entity.BusinessSubscription) error {
	query := `
		INSERT INTO atamlink.business_subscriptions (
			bs_b_id, bs_mp_id, bs_status, bs_starts_at, 
			bs_expires_at, bs_created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING bs_id`

	err := tx.QueryRow(
		query,
		subscription.BusinessID,
		subscription.PlanID,
		subscription.Status,
		subscription.StartsAt,
		subscription.ExpiresAt,
		subscription.CreatedAt,
	).Scan(&subscription.ID)

	if err != nil {
		return errors.Wrap(err, "failed to create subscription")
	}

	return nil
}

// UpdateSubscription update subscription
func (r *businessRepository) UpdateSubscription(tx *sql.Tx, subscription *entity.BusinessSubscription) error {
	query := `
		UPDATE atamlink.business_subscriptions SET
			bs_status = $2,
			bs_expires_at = $3,
			bs_updated_at = NOW()
		WHERE bs_id = $1`

	result, err := tx.Exec(
		query,
		subscription.ID,
		subscription.Status,
		subscription.ExpiresAt,
	)

	if err != nil {
		return errors.Wrap(err, "failed to update subscription")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrNotFound, "Subscription tidak ditemukan", 404)
	}

	return nil
}