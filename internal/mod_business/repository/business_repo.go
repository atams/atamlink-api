package repository

import (
	"database/sql"
	"time"

	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/mod_business/entity"
	"github.com/atam/atamlink/pkg/database"
	"github.com/atam/atamlink/pkg/errors"
)

// BusinessRepository interface untuk business repository
type BusinessRepository interface {
	// Core business methods
	Create(tx *sql.Tx, business *entity.Business) error
	GetByID(id int64) (*entity.Business, error)
	GetBySlug(slug string) (*entity.Business, error)
	List(filter ListFilter) ([]*entity.Business, int64, error)
	Update(tx *sql.Tx, business *entity.Business) error
	Delete(tx *sql.Tx, id int64) error
	IsSlugExists(slug string) (bool, error)
	GetBusinessesWithUserCount(filter ListFilter) ([]*BusinessWithUserCount, int64, error)

	// Business User methods
	AddUser(tx *sql.Tx, businessUser *entity.BusinessUser) error
	GetUsersByBusinessID(businessID int64) ([]*entity.BusinessUser, error)
	GetUserByBusinessAndProfile(businessID, profileID int64) (*entity.BusinessUser, error)
	UpdateUserRole(tx *sql.Tx, businessID, profileID int64, role string) error
	RemoveUser(tx *sql.Tx, businessID, profileID int64) error
	CountUserBusinesses(profileID int64) (int, error)
	GetUserCountByBusinessID(businessID int64) (int, error)
	GetUserRoleInBusiness(businessID, profileID int64) (string, error)

	// Business Invite methods
	CreateInvite(tx *sql.Tx, invite *entity.BusinessInvite) error
	GetInviteByToken(token string) (*entity.BusinessInvite, error)
	UseInvite(tx *sql.Tx, token string) error

	// Business Subscription methods
	GetActiveSubscription(businessID int64) (*entity.BusinessSubscription, error)
	CreateSubscription(tx *sql.Tx, subscription *entity.BusinessSubscription) error
	UpdateSubscription(tx *sql.Tx, subscription *entity.BusinessSubscription) error
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
	IsSuspended *bool
	ProfileID   int64
	Limit       int
	Offset      int
	OrderBy     string
}

// BusinessWithUserCount struct untuk business dengan user count dan role
type BusinessWithUserCount struct {
	entity.Business
	UserCount int     `json:"user_count"`
	UserRole  *string `json:"user_role,omitempty"`
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
			b_id, b_slug, b_name, b_logo_url, b_type, b_is_active,
			b_is_suspended, b_suspension_reason, b_suspended_by, b_suspended_at,
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
		return nil, errors.New(errors.ErrNotFound, constant.ErrMsgBusinessNotFound, 404)
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
			b_id, b_slug, b_name, b_logo_url, b_type, b_is_active,
			b_is_suspended, b_suspension_reason, b_suspended_by, b_suspended_at,
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
		return nil, errors.New(errors.ErrNotFound, constant.ErrMsgBusinessNotFound, 404)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get business by slug")
	}

	return business, nil
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
		business.UpdatedAt,
	)

	if err != nil {
		return errors.Wrap(err, "failed to update business")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrNotFound, constant.ErrMsgBusinessNotFound, 404)
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
		return errors.New(errors.ErrNotFound, constant.ErrMsgBusinessNotFound, 404)
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

// GetBusinessesWithUserCount mendapatkan businesses dengan user count dan role
func (r *businessRepository) GetBusinessesWithUserCount(filter ListFilter) ([]*BusinessWithUserCount, int64, error) {
	qb := database.NewQueryBuilder()
	qb.Select(`b.b_id, b.b_slug, b.b_name, b.b_logo_url, b.b_type, 
		b.b_is_active, b.b_is_suspended, b.b_suspension_reason, 
		b.b_suspended_by, b.b_suspended_at,
		b.b_created_by, b.b_created_at, b.b_updated_by, b.b_updated_at,
		COUNT(DISTINCT bu_all.bu_id) FILTER (WHERE bu_all.bu_is_active = true) as user_count`)
	
	// Add user role jika ada profileID
	if filter.ProfileID > 0 {
		qb.Select("bu.bu_role as user_role")
	}

	qb.From("atamlink.businesses b")
	qb.LeftJoin("atamlink.business_users bu_all", "bu_all.bu_b_id = b.b_id")
	
	// Join untuk mendapatkan role user jika ada profileID
	if filter.ProfileID > 0 {
		qb.LeftJoin("atamlink.business_users bu", "bu.bu_b_id = b.b_id AND bu.bu_is_active = true")
		qb.Where("bu.bu_up_id = ?", filter.ProfileID)
	}

	// Base filter - only active businesses
	qb.Where("b.b_is_active = true")
	
	// If profile ID is provided, only show businesses where user is member
	if filter.ProfileID > 0 {
		qb.Where("EXISTS (SELECT 1 FROM atamlink.business_users WHERE bu_b_id = b.b_id AND bu_up_id = ? AND bu_is_active = true)", filter.ProfileID)
	}

	// Apply filters
	if filter.Search != "" {
		qb.Where("(LOWER(b.b_name) LIKE LOWER(?) OR LOWER(b.b_slug) LIKE LOWER(?))", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	if filter.Type != "" {
		qb.Where("b.b_type = ?", filter.Type)
	}

	if filter.IsSuspended != nil {
		qb.Where("b.b_is_suspended = ?", *filter.IsSuspended)
	}

	// Group by all selected columns
	qb.GroupBy(`b.b_id, b.b_slug, b.b_name, b.b_logo_url, b.b_type, 
		b.b_is_active, b.b_is_suspended, b.b_suspension_reason, 
		b.b_suspended_by, b.b_suspended_at,
		b.b_created_by, b.b_created_at, b.b_updated_by, b.b_updated_at`)
	
	if filter.ProfileID > 0 {
		qb.GroupBy("bu.bu_role")
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

	businesses := make([]*BusinessWithUserCount, 0)
	for rows.Next() {
		business := &BusinessWithUserCount{}
		scanArgs := []interface{}{
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
			&business.UserCount,
		}

		// Tambahkan role jika ada profileID
		if filter.ProfileID > 0 {
			scanArgs = append(scanArgs, &business.UserRole)
		}

		err := rows.Scan(scanArgs...)
		if err != nil {
			return nil, 0, errors.Wrap(err, "failed to scan business row")
		}
		businesses = append(businesses, business)
	}

	return businesses, total, nil
}

// List mendapatkan list businesses (deprecated - use GetBusinessesWithUserCount)
func (r *businessRepository) List(filter ListFilter) ([]*entity.Business, int64, error) {
	// Implementation sama dengan GetBusinessesWithUserCount tapi tanpa user count
	// Untuk backward compatibility
	return nil, 0, nil
}