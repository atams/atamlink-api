package usecase

import (
	"github.com/atam/atamlink/internal/mod_user/dto"
	"github.com/atam/atamlink/internal/mod_user/repository"
)

// UserUseCase interface untuk user use case
type UserUseCase interface {
	GetProfileByID(profileID int64) (*dto.ProfileResponse, error)
	GetUserByID(userID string) (*dto.UserResponse, error)
}

type userUseCase struct {
	userRepo repository.UserRepository
}

// NewUserUseCase membuat instance user use case baru
func NewUserUseCase(userRepo repository.UserRepository) UserUseCase {
	return &userUseCase{
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