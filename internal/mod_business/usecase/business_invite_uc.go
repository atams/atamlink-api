package usecase

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/mod_business/dto"
	"github.com/atam/atamlink/internal/mod_business/entity"
	"github.com/atam/atamlink/pkg/errors"
)

// CreateInvite membuat invite link
func (uc *businessUseCase) CreateInvite(businessID int64, profileID int64, req *dto.CreateInviteRequest) (*dto.InviteResponse, error) {
	// Check permission
	if err := uc.checkBusinessPermission(businessID, profileID, constant.PermUserInvite); err != nil {
		return nil, err
	}

	// Validate role
	if !constant.IsValidRole(req.Role) {
		return nil, errors.New(errors.ErrValidation, "Role tidak valid", 400)
	}

	// Owner can only be added directly
	if req.Role == constant.RoleOwner {
		return nil, errors.New(errors.ErrValidation, "Owner harus ditambahkan langsung", 400)
	}

	// Set expiry
	expiresAt := req.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(7 * 24 * time.Hour) // 7 days
	}

	// Generate token
	token := uuid.New().String()

	// Create invite
	tx, err := uc.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	invite := &entity.BusinessInvite{
		BusinessID: businessID,
		Token:      token,
		Role:       req.Role,
		InvitedBy:  profileID,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
	}

	if err := uc.businessRepo.CreateInvite(tx, invite); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	// Return response
	return &dto.InviteResponse{
		ID:        invite.ID,
		Token:     invite.Token,
		Role:      invite.Role,
		InvitedBy: invite.InvitedBy,
		ExpiresAt: invite.ExpiresAt,
		CreatedAt: invite.CreatedAt,
		InviteURL: fmt.Sprintf("/invite/%s", invite.Token), // TODO: Use proper base URL
	}, nil
}

// AcceptInvite accept invite
func (uc *businessUseCase) AcceptInvite(req *dto.AcceptInviteRequest) error {
	// Get invite
	invite, err := uc.businessRepo.GetInviteByToken(req.Token)
	if err != nil {
		return err
	}

	// Validate invite
	if !invite.IsValid() {
		return errors.New(errors.ErrValidation, "Invite tidak valid atau sudah kadaluarsa", 400)
	}

	// Check if user already member
	existingUser, err := uc.businessRepo.GetUserByBusinessAndProfile(invite.BusinessID, req.ProfileID)
	if err != nil {
		return err
	}
	if existingUser != nil && existingUser.IsActive {
		return errors.New(errors.ErrConflict, "Anda sudah menjadi member", 409)
	}

	// Accept invite
	tx, err := uc.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Mark invite as used
	if err := uc.businessRepo.UseInvite(tx, req.Token); err != nil {
		return err
	}

	// Add user to business
	businessUser := &entity.BusinessUser{
		BusinessID: invite.BusinessID,
		ProfileID:  req.ProfileID,
		Role:       invite.Role,
		IsActive:   true,
		CreatedAt:  time.Now(),
	}

	if err := uc.businessRepo.AddUser(tx, businessUser); err != nil {
		return err
	}

	return tx.Commit()
}