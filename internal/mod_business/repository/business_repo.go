package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/mod_business/entity"
	"github.com/atam/atamlink/pkg/database"
	"github.com/atam/atamlink/pkg/errors"
)

// BusinessRepository interface untuk business repository
type BusinessRepository interface {
	Create(tx *sql.Tx, business *entity.Business) error
	GetByID(id int64) (*entity.Business, error)
	GetBySlug(slug string) (*entity.Business, error)
	List(filter ListFilter) ([]*entity.Business, int64, error)
	Update(tx *sql.Tx, business *entity.Business) error
	Delete(tx *sql.Tx, id int64) error

	// Business User methods
	AddUser(tx *sql.Tx, businessUser *entity.BusinessUser) error
	GetUsersByBusinessID(businessID int64) ([]*entity.BusinessUser, error)
	GetUserByBusinessAndProfile(businessID, profileID int64) (*entity.BusinessUser, error)
	UpdateUserRole(tx *sql.Tx, businessID, profileID int64, role string) error
	RemoveUser(tx *sql.Tx, businessID, profileID int64) error

	// Business Invite methods
	CreateInvite(tx *sql.Tx, invite *entity.BusinessInvite) error
	GetInviteByToken(token string) (*entity.BusinessInvite, error)
	UseInvite(tx *sql.Tx, token string) error

	// Business Subscription methods
	GetActiveSubscription(businessID int64) (*entity.BusinessSubscription, error)
	CreateSubscription(tx *sql.Tx, subscription *entity.BusinessSubscription) error
	UpdateSubscription(tx *sql.Tx, subscription *entity.BusinessSubscription) error

	// Helper methods
	IsSlugExists(slug string) (bool, error)
	CountUserBusinesses(profileID int64) (int, error)
}

type businessRepository struct {
	db *sql.DB
}

// NewBusinessRepository membuat instance business repository baru
func NewBusinessRepository(db *sql.DB) BusinessRepository {
	return &businessRepository{db: db}
}

// ListFilter filter untuk list businesses
type ListFilter struct {
	Search      string
	Type        string
	IsActive    *bool
	IsSuspended *bool
	ProfileID   int64
	Limit       int
	Offset      int
	OrderBy     string
}

// Create membuat business baru
func (r *businessRepository) Create(tx *sql.Tx, business *entity.Business) error {
	query := `
		INSERT INTO atamlink.businesses (
			b_slug, b_name, b_logo_url, b_type, b_is_active, b_is_suspended,
			b_created_by, b_created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING b_id`

	err := tx.QueryRow(
		query,
		business.Slug,
		business.Name,
		business.LogoURL,
		business.Type,
		business.IsActive,
		business.IsSuspended,
		business.CreatedBy,
		business.CreatedAt,
	).Scan(&business.ID)

	if err != nil {
		return errors.Wrap(err, "failed to create business")
	}

	return nil
}

// GetByID mendapatkan business berdasarkan ID
func (r *businessRepository) GetByID(id int64) (*entity.Business, error) {
	query := `
		SELECT 
			b_id, b_slug, b_name, b_logo_url, b_type, b_is_active, b_is_suspended,
			b_suspension_reason, b_suspended_by, b_suspended_at,
			b_created_by, b_created_at, b_updated_by, b_updated_at
		FROM atamlink.businesses
		WHERE b_id = $1`

	business := &entity.Business{}
	err := r.db.QueryRow(query, id).Scan(
		&business.ID,
		&business.Slug,
		&business.Name,
		&business.LogoURL,
		&business.Type,
		&business.IsActive,
		&business.IsSuspended,
		&business.SuspensionReason,
		&business.SuspendedBy,
		&business.SuspendedAt,
		&business.CreatedBy,
		&business.CreatedAt,
		&business.UpdatedBy,
		&business.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrBusinessNotFound, constant.ErrMsgBusinessNotFound, 404)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get business")
	}

	return business, nil
}

// GetBySlug mendapatkan business berdasarkan slug
func (r *businessRepository) GetBySlug(slug string) (*entity.Business, error) {
	query := `
		SELECT 
			b_id, b_slug, b_name, b_logo_url, b_type, b_is_active, b_is_suspended,
			b_suspension_reason, b_suspended_by, b_suspended_at,
			b_created_by, b_created_at, b_updated_by, b_updated_at
		FROM atamlink.businesses
		WHERE b_slug = $1`

	business := &entity.Business{}
	err := r.db.QueryRow(query, slug).Scan(
		&business.ID,
		&business.Slug,
		&business.Name,
		&business.LogoURL,
		&business.Type,
		&business.IsActive,
		&business.IsSuspended,
		&business.SuspensionReason,
		&business.SuspendedBy,
		&business.SuspendedAt,
		&business.CreatedBy,
		&business.CreatedAt,
		&business.UpdatedBy,
		&business.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrBusinessNotFound, constant.ErrMsgBusinessNotFound, 404)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get business by slug")
	}

	return business, nil
}

// List mendapatkan list businesses dengan filter
func (r *businessRepository) List(filter ListFilter) ([]*entity.Business, int64, error) {
	// Build query dengan filter
	qb := database.NewQueryBuilder()
	qb.Select(
		"b_id", "b_slug", "b_name", "b_logo_url", "b_type", "b_is_active", "b_is_suspended",
		"b_suspension_reason", "b_suspended_by", "b_suspended_at",
		"b_created_by", "b_created_at", "b_updated_by", "b_updated_at",
	).From("atamlink.businesses")

	// Apply filters
	if filter.Search != "" {
		qb.Where("(LOWER(b_name) LIKE LOWER($1) OR LOWER(b_slug) LIKE LOWER($1))", "%"+filter.Search+"%")
	}

	if filter.Type != "" {
		qb.Where("b_type = ?", filter.Type)
	}

	if filter.IsActive != nil {
		qb.Where("b_is_active = ?", *filter.IsActive)
	}

	if filter.IsSuspended != nil {
		qb.Where("b_is_suspended = ?", *filter.IsSuspended)
	}

	if filter.ProfileID > 0 {
		qb.InnerJoin("atamlink.business_users", "bu_b_id = b_id")
		qb.Where("bu_up_id = ? AND bu_is_active = true", filter.ProfileID)
	}

	// Count total
	countQuery, countArgs := qb.BuildCount()
	var total int64
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to count businesses")
	}

	// Get data
	qb.OrderBy(filter.OrderBy)
	qb.Limit(filter.Limit)
	qb.Offset(filter.Offset)

	query, args := qb.Build()
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to query businesses")
	}
	defer rows.Close()

	businesses := make([]*entity.Business, 0)
	for rows.Next() {
		business := &entity.Business{}
		err := rows.Scan(
			&business.ID,
			&business.Slug,
			&business.Name,
			&business.LogoURL,
			&business.Type,
			&business.IsActive,
			&business.IsSuspended,
			&business.SuspensionReason,
			&business.SuspendedBy,
			&business.SuspendedAt,
			&business.CreatedBy,
			&business.CreatedAt,
			&business.UpdatedBy,
			&business.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errors.Wrap(err, "failed to scan business row")
		}
		businesses = append(businesses, business)
	}

	return businesses, total, nil
}

// Update update business
func (r *businessRepository) Update(tx *sql.Tx, business *entity.Business) error {
	query := `
		UPDATE atamlink.businesses SET
			b_name = $2,
			b_logo_url = $3,
			b_type = $4,
			b_is_active = $5,
			b_is_suspended = $6,
			b_suspension_reason = $7,
			b_suspended_by = $8,
			b_suspended_at = $9,
			b_updated_by = $10,
			b_updated_at = $11
		WHERE b_id = $1`

	result, err := tx.Exec(
		query,
		business.ID,
		business.Name,
		business.LogoURL,
		business.Type,
		business.IsActive,
		business.IsSuspended,
		business.SuspensionReason,
		business.SuspendedBy,
		business.SuspendedAt,
		business.UpdatedBy,
		time.Now(),
	)

	if err != nil {
		return errors.Wrap(err, "failed to update business")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrBusinessNotFound, constant.ErrMsgBusinessNotFound, 404)
	}

	return nil
}

// Delete soft delete business
func (r *businessRepository) Delete(tx *sql.Tx, id int64) error {
	query := `
		UPDATE atamlink.businesses 
		SET b_is_active = false, b_updated_at = $2
		WHERE b_id = $1`

	result, err := tx.Exec(query, id, time.Now())
	if err != nil {
		return errors.Wrap(err, "failed to delete business")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrBusinessNotFound, constant.ErrMsgBusinessNotFound, 404)
	}

	return nil
}

// AddUser menambahkan user ke business
func (r *businessRepository) AddUser(tx *sql.Tx, businessUser *entity.BusinessUser) error {
	query := `
		INSERT INTO atamlink.business_users (
			bu_b_id, bu_up_id, bu_role, bu_is_owner, bu_is_active, bu_created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING bu_id`

	err := tx.QueryRow(
		query,
		businessUser.BusinessID,
		businessUser.ProfileID,
		businessUser.Role,
		businessUser.IsOwner,
		businessUser.IsActive,
		businessUser.CreatedAt,
	).Scan(&businessUser.ID)

	if err != nil {
		return errors.Wrap(err, "failed to add user to business")
	}

	return nil
}

// GetUsersByBusinessID mendapatkan users by business ID
func (r *businessRepository) GetUsersByBusinessID(businessID int64) ([]*entity.BusinessUser, error) {
	query := `
		SELECT 
			bu.bu_id, bu.bu_b_id, bu.bu_up_id, bu.bu_role, 
			bu.bu_is_owner, bu.bu_is_active, bu.bu_created_at,
			up.up_phone, up.up_display_name
		FROM atamlink.business_users bu
		INNER JOIN atamlink.user_profiles up ON up.up_id = bu.bu_up_id
		WHERE bu.bu_b_id = $1 AND bu.bu_is_active = true
		ORDER BY bu.bu_created_at ASC`

	rows, err := r.db.Query(query, businessID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get business users")
	}
	defer rows.Close()

	users := make([]*entity.BusinessUser, 0)
	for rows.Next() {
		user := &entity.BusinessUser{
			Profile: &entity.UserProfile{},
		}
		err := rows.Scan(
			&user.ID,
			&user.BusinessID,
			&user.ProfileID,
			&user.Role,
			&user.IsOwner,
			&user.IsActive,
			&user.CreatedAt,
			&user.Profile.Phone,
			&user.Profile.DisplayName,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan business user")
		}
		user.Profile.ID = user.ProfileID
		users = append(users, user)
	}

	return users, nil
}

// GetUserByBusinessAndProfile mendapatkan user by business dan profile ID
func (r *businessRepository) GetUserByBusinessAndProfile(businessID, profileID int64) (*entity.BusinessUser, error) {
	query := `
		SELECT bu_id, bu_b_id, bu_up_id, bu_role, bu_is_owner, bu_is_active, bu_created_at
		FROM atamlink.business_users
		WHERE bu_b_id = $1 AND bu_up_id = $2`

	user := &entity.BusinessUser{}
	err := r.db.QueryRow(query, businessID, profileID).Scan(
		&user.ID,
		&user.BusinessID,
		&user.ProfileID,
		&user.Role,
		&user.IsOwner,
		&user.IsActive,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get business user")
	}

	return user, nil
}

// UpdateUserRole update user role
func (r *businessRepository) UpdateUserRole(tx *sql.Tx, businessID, profileID int64, role string) error {
	query := `
		UPDATE atamlink.business_users 
		SET bu_role = $3
		WHERE bu_b_id = $1 AND bu_up_id = $2`

	result, err := tx.Exec(query, businessID, profileID, role)
	if err != nil {
		return errors.Wrap(err, "failed to update user role")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrNotFound, "User tidak ditemukan dalam bisnis", 404)
	}

	return nil
}

// RemoveUser remove user dari business
func (r *businessRepository) RemoveUser(tx *sql.Tx, businessID, profileID int64) error {
	query := `
		UPDATE atamlink.business_users 
		SET bu_is_active = false
		WHERE bu_b_id = $1 AND bu_up_id = $2`

	result, err := tx.Exec(query, businessID, profileID)
	if err != nil {
		return errors.Wrap(err, "failed to remove user")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrNotFound, "User tidak ditemukan dalam bisnis", 404)
	}

	return nil
}

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

// GetActiveSubscription mendapatkan active subscription
func (r *businessRepository) GetActiveSubscription(businessID int64) (*entity.BusinessSubscription, error) {
	query := `
		SELECT 
			bs.bs_id, bs.bs_b_id, bs.bs_mp_id, bs.bs_status,
			bs.bs_starts_at, bs.bs_expires_at, bs.bs_created_at, bs.bs_updated_at,
			mp.mp_id, mp.mp_name, mp.mp_price, mp.mp_features
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
		&featuresJSON,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get active subscription")
	}

	// Parse JSON features
	if err := json.Unmarshal(featuresJSON, &sub.Plan.Features); err != nil {
		return nil, errors.Wrap(err, "failed to parse plan features")
	}

	return sub, nil
}

// CreateSubscription create subscription
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
			bs_updated_at = $4
		WHERE bs_id = $1`

	result, err := tx.Exec(
		query,
		subscription.ID,
		subscription.Status,
		subscription.ExpiresAt,
		time.Now(),
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

// IsSlugExists check apakah slug sudah ada
func (r *businessRepository) IsSlugExists(slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM atamlink.businesses WHERE b_slug = $1)`

	var exists bool
	err := r.db.QueryRow(query, slug).Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "failed to check slug exists")
	}

	return exists, nil
}

// CountUserBusinesses count business yang dimiliki user
func (r *businessRepository) CountUserBusinesses(profileID int64) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM atamlink.business_users 
		WHERE bu_up_id = $1 AND bu_is_active = true`

	var count int
	err := r.db.QueryRow(query, profileID).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "failed to count user businesses")
	}

	return count, nil
}