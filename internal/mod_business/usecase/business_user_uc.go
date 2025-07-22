package usecase

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/middleware"
	"github.com/atam/atamlink/internal/mod_business/dto"
	"github.com/atam/atamlink/internal/mod_business/entity"
	"github.com/atam/atamlink/pkg/errors"
)

// AddUser menambahkan user ke business
func (uc *businessUseCase) AddUser(ctx *gin.Context, businessID int64, profileID int64, req *dto.AddUserRequest) error {
	// Check permission
	if err := uc.checkBusinessPermission(businessID, profileID, constant.PermUserInvite); err != nil {
		return err
	}

	// Validate role
	if !constant.IsValidRole(req.Role) {
		return errors.New(errors.ErrValidation, "Role tidak valid", 400)
	}

	// Check if user exists
	targetProfile, err := uc.userRepo.GetProfileByID(req.ProfileID)
	if err != nil {
		return err
	}
	if targetProfile == nil {
		return errors.New(errors.ErrNotFound, constant.ErrMsgProfileNotFound, 404)
	}

	// Check if already member
	existingUser, err := uc.businessRepo.GetUserByBusinessAndProfile(businessID, req.ProfileID)
	if err != nil {
		return err
	}
	if existingUser != nil {
		if existingUser.IsActive {
			return errors.New(errors.ErrConflict, "User sudah menjadi member", 409)
		}
		// Reactivate user
		existingUser.IsActive = true
		existingUser.Role = req.Role
		tx, err := uc.db.Begin()
		if err != nil {
			return errors.Wrap(err, "failed to begin transaction")
		}
		defer tx.Rollback()

		if err := uc.businessRepo.UpdateUserRole(tx, businessID, req.ProfileID, req.Role); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return errors.Wrap(err, "failed to commit transaction")
		}

		// Set audit data untuk USER_ACTIVATE
		if ctx != nil {
			middleware.SetAuditBusinessID(ctx, businessID)
			middleware.SetAuditRecordID(ctx, fmt.Sprintf("%d", existingUser.ID))
			middleware.SetAuditNewData(ctx, map[string]interface{}{
				"profile_id": req.ProfileID,
				"role":       req.Role,
				"is_active":  true,
			})
			// Override action di middleware
			ctx.Set(middleware.GinKeyAuditAction, constant.AuditActionUserActivate)
		}

		return nil
	}

	// Add user
	tx, err := uc.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	businessUser := &entity.BusinessUser{
		BusinessID: businessID,
		ProfileID:  req.ProfileID,
		Role:       req.Role,
		IsOwner:    req.Role == constant.RoleOwner,
		IsActive:   true,
		CreatedAt:  time.Now(),
	}

	if err := uc.businessRepo.AddUser(tx, businessUser); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	// Set audit data untuk USER_ADD
	if ctx != nil {
		middleware.SetAuditBusinessID(ctx, businessID)
		middleware.SetAuditRecordID(ctx, fmt.Sprintf("%d", businessUser.ID))
		middleware.SetAuditNewData(ctx, map[string]interface{}{
			"profile_id":    req.ProfileID,
			"role":          req.Role,
			"display_name":  targetProfile.DisplayName,
		})
	}

	return nil
}

// UpdateUserRole update role user
func (uc *businessUseCase) UpdateUserRole(ctx *gin.Context, businessID int64, profileID int64, targetProfileID int64, role string) error {
	// Check permission
	if err := uc.checkBusinessPermission(businessID, profileID, constant.PermUserUpdate); err != nil {
		return err
	}

	// Validate role
	if !constant.IsValidRole(role) {
		return errors.New(errors.ErrValidation, "Role tidak valid", 400)
	}

	// Can't update own role
	if profileID == targetProfileID {
		return errors.New(errors.ErrValidation, "Tidak dapat mengubah role sendiri", 400)
	}

	// Check target user exists
	targetUser, err := uc.businessRepo.GetUserByBusinessAndProfile(businessID, targetProfileID)
	if err != nil {
		return err
	}
	if targetUser == nil || !targetUser.IsActive {
		return errors.New(errors.ErrNotFound, "User tidak ditemukan dalam bisnis", 404)
	}

	// Store old role for audit
	oldRole := targetUser.Role

	// Update role
	tx, err := uc.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.businessRepo.UpdateUserRole(tx, businessID, targetProfileID, role); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	// Set audit data untuk USER_ROLE_UPDATE
	if ctx != nil {
		middleware.SetAuditBusinessID(ctx, businessID)
		middleware.SetAuditRecordID(ctx, fmt.Sprintf("%d", targetUser.ID))
		middleware.SetAuditOldData(ctx, map[string]interface{}{
			"role": oldRole,
		})
		middleware.SetAuditNewData(ctx, map[string]interface{}{
			"role": role,
		})
	}

	return nil
}

// RemoveUser hapus user dari business
func (uc *businessUseCase) RemoveUser(ctx *gin.Context, businessID int64, profileID int64, targetProfileID int64) error {
	// Check permission
	if err := uc.checkBusinessPermission(businessID, profileID, constant.PermUserRemove); err != nil {
		return err
	}

	// Can't remove self
	if profileID == targetProfileID {
		return errors.New(errors.ErrValidation, "Tidak dapat menghapus diri sendiri", 400)
	}

	// Check if target is the only owner
	users, err := uc.businessRepo.GetUsersByBusinessID(businessID)
	if err != nil {
		return err
	}

	ownerCount := 0
	var targetUser *entity.BusinessUser
	for _, user := range users {
		if user.IsOwner && user.IsActive {
			ownerCount++
		}
		if user.ProfileID == targetProfileID {
			targetUser = user
		}
	}

	if targetUser == nil || !targetUser.IsActive {
		return errors.New(errors.ErrNotFound, "User tidak ditemukan dalam bisnis", 404)
	}

	if targetUser.IsOwner && ownerCount <= 1 {
		return errors.New(errors.ErrValidation, constant.ErrMsgBusinessOwnerRequired, 400)
	}

	// Store data for audit
	removedUserData := map[string]interface{}{
		"profile_id": targetProfileID,
		"role":       targetUser.Role,
		"was_owner":  targetUser.IsOwner,
	}

	// Remove user
	tx, err := uc.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.businessRepo.RemoveUser(tx, businessID, targetProfileID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	// Set audit data untuk USER_REMOVE
	if ctx != nil {
		middleware.SetAuditBusinessID(ctx, businessID)
		middleware.SetAuditRecordID(ctx, fmt.Sprintf("%d", targetUser.ID))
		middleware.SetAuditOldData(ctx, removedUserData)
		middleware.SetAuditNewData(ctx, nil)
	}

	return nil
}