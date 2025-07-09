package usecase

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/atam/atamlink/internal/middleware"
	"github.com/atam/atamlink/internal/mod_user/dto"
	"github.com/atam/atamlink/internal/mod_user/entity"
	"github.com/atam/atamlink/internal/mod_user/repository"
	"github.com/atam/atamlink/pkg/database"
	"github.com/atam/atamlink/pkg/errors"
)

// UserUseCase interface untuk user use case
type UserUseCase interface {
	GetProfileByID(profileID int64) (*dto.ProfileResponse, error)
	GetUserByID(userID string) (*dto.UserResponse, error)
	
	// Profile CRUD methods
	CreateProfile(userID string, req *dto.CreateProfileRequest) (*dto.ProfileResponse, error)
	UpdateProfile(ctx *gin.Context, profileID int64, req *dto.UpdateProfileRequest) (*dto.ProfileResponse, error)
	DeleteProfile(ctx *gin.Context, profileID int64) error
}

type userUseCase struct {
	db       *sql.DB
	userRepo repository.UserRepository
}

// NewUserUseCase membuat instance user use case baru
func NewUserUseCase(db *sql.DB, userRepo repository.UserRepository) UserUseCase {
	return &userUseCase{
		db:       db,
		userRepo: userRepo,
	}
}

// GetProfileByID mendapatkan profile by ID
func (uc *userUseCase) GetProfileByID(profileID int64) (*dto.ProfileResponse, error) {
	// Get profile
	profile, err := uc.userRepo.GetProfileByID(profileID)
	if err != nil {
		return nil, err
	}

	// Convert to response
	return &dto.ProfileResponse{
		ID:          profile.ID,
		UserID:      profile.UserID,
		DisplayName: profile.GetDisplayName(),
		Phone:       profile.GetPhone(),
		CreatedAt:   profile.CreatedAt,
		UpdatedAt:   profile.UpdatedAt,
	}, nil
}

// GetUserByID mendapatkan user by ID
func (uc *userUseCase) GetUserByID(userID string) (*dto.UserResponse, error) {
	// Get user
	user, err := uc.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Get profile
	profile, _ := uc.userRepo.GetProfileByUserID(userID)

	// Convert to response
	resp := &dto.UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		Username:   user.Username,
		IsActive:   user.IsActive,
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}

	// Add profile if exists
	if profile != nil {
		resp.Profile = &dto.ProfileResponse{
			ID:          profile.ID,
			UserID:      profile.UserID,
			DisplayName: profile.GetDisplayName(),
			Phone:       profile.GetPhone(),
			CreatedAt:   profile.CreatedAt,
			UpdatedAt:   profile.UpdatedAt,
		}
	}

	return resp, nil
}

// CreateProfile membuat profile baru
func (uc *userUseCase) CreateProfile(userID string, req *dto.CreateProfileRequest) (*dto.ProfileResponse, error) {
	// Check if user exists
	user, err := uc.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New(errors.ErrAccountInactive, "User tidak aktif", 400)
	}

	// Check if profile already exists
	existingProfile, _ := uc.userRepo.GetProfileByUserID(userID)
	if existingProfile != nil {
		return nil, errors.New(errors.ErrConflict, "Profile sudah ada untuk user ini", 409)
	}

	// Validate phone uniqueness if provided
	if req.Phone != "" {
		exists, err := uc.userRepo.IsPhoneExists(req.Phone, 0)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New(errors.ErrConflict, "Nomor telepon sudah digunakan", 409)
		}
	}

	// Start transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Create profile
	profile := &entity.UserProfile{
		UserID:      userID,
		Phone:       database.NullString(req.Phone),
		DisplayName: database.NullString(req.DisplayName),
		CreatedAt:   time.Now(),
	}

	if err := uc.userRepo.CreateProfile(tx, profile); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	// Return profile
	return uc.GetProfileByID(profile.ID)
}

// UpdateProfile update profile
func (uc *userUseCase) UpdateProfile(ctx *gin.Context, profileID int64, req *dto.UpdateProfileRequest) (*dto.ProfileResponse, error) {
	// Get existing profile
	profile, err := uc.userRepo.GetProfileByID(profileID)
	if err != nil {
		return nil, err
	}

	// Inject old_data ke audit context
	if ctx != nil {
		ctx.Set(middleware.GinKeyAuditOldData, profile)
	}

	// Validate phone uniqueness if changed
	if req.Phone != "" && req.Phone != profile.GetPhone() {
		exists, err := uc.userRepo.IsPhoneExists(req.Phone, profileID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New(errors.ErrConflict, "Nomor telepon sudah digunakan", 409)
		}
	}

	// Update fields
	hasChanges := false

	if req.Phone != "" && req.Phone != profile.GetPhone() {
		profile.SetPhone(req.Phone)
		hasChanges = true
	}

	if req.DisplayName != "" && req.DisplayName != profile.GetDisplayName() {
		profile.SetDisplayName(req.DisplayName)
		hasChanges = true
	}

	// Jika tidak ada perubahan
	if !hasChanges {
		return uc.GetProfileByID(profileID)
	}

	// Update in transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.userRepo.UpdateProfile(tx, profile); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	// Return updated profile
	return uc.GetProfileByID(profileID)
}

// DeleteProfile delete profile
func (uc *userUseCase) DeleteProfile(ctx *gin.Context, profileID int64) error {
	// Get existing profile
	profile, err := uc.userRepo.GetProfileByID(profileID)
	if err != nil {
		return err
	}

	// Inject old_data ke audit context
	if ctx != nil {
		ctx.Set(middleware.GinKeyAuditOldData, profile)
	}

	// Delete in transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.userRepo.DeleteProfile(tx, profileID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}