package repository

import (
	"database/sql"

	"github.com/atam/atamlink/internal/mod_user/entity"
	"github.com/atam/atamlink/pkg/errors"
)

// UserRepository interface untuk user repository
type UserRepository interface {
	GetProfileByID(profileID int64) (*entity.UserProfile, error)
	GetProfileByUserID(userID string) (*entity.UserProfile, error)
	GetUserByID(userID string) (*entity.User, error)
}

type userRepository struct {
	db *sql.DB
}

// NewUserRepository membuat instance user repository baru
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// GetProfileByID mendapatkan user profile by ID
func (r *userRepository) GetProfileByID(profileID int64) (*entity.UserProfile, error) {
	query := `
		SELECT 
			up.up_id, up.up_u_id, up.up_phone, up.up_display_name,
			up.up_created_at, up.up_updated_at,
			u.u_id, u.u_email, u.u_username, u.u_is_active, u.u_is_verified
		FROM atamlink.user_profiles up
		INNER JOIN atamlink.users u ON u.u_id = up.up_u_id
		WHERE up.up_id = $1`

	profile := &entity.UserProfile{
		User: &entity.User{},
	}

	err := r.db.QueryRow(query, profileID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.Phone,
		&profile.DisplayName,
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&profile.User.ID,
		&profile.User.Email,
		&profile.User.Username,
		&profile.User.IsActive,
		&profile.User.IsVerified,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrNotFound, "Profile tidak ditemukan", 404)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get profile")
	}

	return profile, nil
}

// GetProfileByUserID mendapatkan user profile by user ID
func (r *userRepository) GetProfileByUserID(userID string) (*entity.UserProfile, error) {
	query := `
		SELECT 
			up_id, up_u_id, up_phone, up_display_name,
			up_created_at, up_updated_at
		FROM atamlink.user_profiles
		WHERE up_u_id = $1`

	profile := &entity.UserProfile{}
	err := r.db.QueryRow(query, userID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.Phone,
		&profile.DisplayName,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrNotFound, "Profile tidak ditemukan", 404)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get profile by user ID")
	}

	return profile, nil
}

// GetUserByID mendapatkan user by ID
func (r *userRepository) GetUserByID(userID string) (*entity.User, error) {
	query := `
		SELECT 
			u_id, u_email, u_username, u_is_active, u_is_verified,
			u_is_locked, u_email_verified_at, u_last_login_at,
			u_failed_login_attempts, u_locked_until, u_metadata,
			u_ip_address, u_user_agent, created_at, updated_at
		FROM atamlink.users
		WHERE u_id = $1`

	user := &entity.User{}
	err := r.db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.IsActive,
		&user.IsVerified,
		&user.IsLocked,
		&user.EmailVerifiedAt,
		&user.LastLoginAt,
		&user.FailedLoginAttempts,
		&user.LockedUntil,
		&user.Metadata,
		&user.IPAddress,
		&user.UserAgent,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrNotFound, "User tidak ditemukan", 404)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	return user, nil
}