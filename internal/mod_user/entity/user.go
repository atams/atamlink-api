package entity

import (
	"database/sql"
	"time"
)

// User entity untuk tabel users
type User struct {
	ID                  string         `json:"id" db:"u_id"`
	Email               string         `json:"email" db:"u_email"`
	Username            string         `json:"username" db:"u_username"`
	PasswordHash        string         `json:"-" db:"u_password_hash"`
	IsActive            bool           `json:"is_active" db:"u_is_active"`
	IsVerified          bool           `json:"is_verified" db:"u_is_verified"`
	IsLocked            bool           `json:"is_locked" db:"u_is_locked"`
	EmailVerifiedAt     *time.Time     `json:"email_verified_at" db:"u_email_verified_at"`
	LastLoginAt         *time.Time     `json:"last_login_at" db:"u_last_login_at"`
	FailedLoginAttempts int            `json:"failed_login_attempts" db:"u_failed_login_attempts"`
	LockedUntil         *time.Time     `json:"locked_until" db:"u_locked_until"`
	Metadata            sql.NullString `json:"metadata" db:"u_metadata"`
	IPAddress           sql.NullString `json:"ip_address" db:"u_ip_address"`
	UserAgent           sql.NullString `json:"user_agent" db:"u_user_agent"`
	CreatedAt           time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt           *time.Time     `json:"updated_at" db:"updated_at"`
}

// UserProfile entity untuk tabel user_profiles
type UserProfile struct {
	ID          int64          `json:"id" db:"up_id"`
	UserID      string         `json:"user_id" db:"up_u_id"`
	Phone       sql.NullString `json:"phone" db:"up_phone"`
	DisplayName sql.NullString `json:"display_name" db:"up_display_name"`
	CreatedAt   time.Time      `json:"created_at" db:"up_created_at"`
	UpdatedAt   *time.Time     `json:"updated_at" db:"up_updated_at"`
	
	// Relations
	User *User `json:"user,omitempty"`
}



// TableName mendapatkan nama tabel
func (User) TableName() string {
	return "atamlink.users"
}

// TableName mendapatkan nama tabel
func (UserProfile) TableName() string {
	return "atamlink.user_profiles"
}

// GetPhone mendapatkan phone dengan null handling
func (up *UserProfile) GetPhone() string {
	if up.Phone.Valid {
		return up.Phone.String
	}
	return ""
}

// GetDisplayName mendapatkan display name dengan null handling
func (up *UserProfile) GetDisplayName() string {
	if up.DisplayName.Valid {
		return up.DisplayName.String
	}
	return ""
}

// SetPhone set phone value
func (up *UserProfile) SetPhone(phone string) {
	up.Phone = sql.NullString{
		String: phone,
		Valid:  phone != "",
	}
}

// SetDisplayName set display name value
func (up *UserProfile) SetDisplayName(name string) {
	up.DisplayName = sql.NullString{
		String: name,
		Valid:  name != "",
	}
}