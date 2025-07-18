package repository

import (
	"database/sql"

	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/mod_business/entity"
	"github.com/atam/atamlink/pkg/errors"
)

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
			up.up_id, up.up_u_id, up.up_phone, up.up_display_name
		FROM atamlink.business_users bu
		LEFT JOIN atamlink.user_profiles up ON up.up_id = bu.bu_up_id
		WHERE bu.bu_b_id = $1 AND bu.bu_is_active = true
		ORDER BY bu.bu_is_owner DESC, bu.bu_created_at ASC`

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
			&user.Profile.ID,
			&user.Profile.UserID,
			&user.Profile.Phone,
			&user.Profile.DisplayName,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan business user")
		}

		users = append(users, user)
	}

	return users, nil
}

// GetUserByBusinessAndProfile mendapatkan user by business dan profile ID
func (r *businessRepository) GetUserByBusinessAndProfile(businessID, profileID int64) (*entity.BusinessUser, error) {
	query := `
		SELECT
			bu_id, bu_b_id, bu_up_id, bu_role, bu_is_owner, bu_is_active, bu_created_at
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

// UpdateUserRole update role user dalam business
func (r *businessRepository) UpdateUserRole(tx *sql.Tx, businessID, profileID int64, role string) error {
	query := `
		UPDATE atamlink.business_users
		SET bu_role = $3, bu_is_owner = $4
		WHERE bu_b_id = $1 AND bu_up_id = $2 AND bu_is_active = true`

	isOwner := role == constant.RoleOwner

	result, err := tx.Exec(query, businessID, profileID, role, isOwner)
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

// RemoveUser hapus user dari business
func (r *businessRepository) RemoveUser(tx *sql.Tx, businessID, profileID int64) error {
	query := `
		UPDATE atamlink.business_users
		SET bu_is_active = false
		WHERE bu_b_id = $1 AND bu_up_id = $2 AND bu_is_active = true`

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

// CountUserBusinesses menghitung jumlah business yang dimiliki user
func (r *businessRepository) CountUserBusinesses(profileID int64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM atamlink.business_users bu
		INNER JOIN atamlink.businesses b ON b.b_id = bu.bu_b_id
		WHERE bu.bu_up_id = $1 AND bu.bu_is_active = true AND b.b_is_active = true`

	var count int
	err := r.db.QueryRow(query, profileID).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "failed to count user businesses")
	}

	return count, nil
}

// GetUserCountByBusinessID mendapatkan jumlah user dalam business
func (r *businessRepository) GetUserCountByBusinessID(businessID int64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM atamlink.business_users
		WHERE bu_b_id = $1 AND bu_is_active = true`

	var count int
	err := r.db.QueryRow(query, businessID).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "failed to count business users")
	}

	return count, nil
}

// GetUserRoleInBusiness mendapatkan role user dalam business
func (r *businessRepository) GetUserRoleInBusiness(businessID, profileID int64) (string, error) {
	query := `
		SELECT bu_role
		FROM atamlink.business_users
		WHERE bu_b_id = $1 AND bu_up_id = $2 AND bu_is_active = true`

	var role string
	err := r.db.QueryRow(query, businessID, profileID).Scan(&role)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", errors.Wrap(err, "failed to get user role")
	}

	return role, nil
}