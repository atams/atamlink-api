package usecase

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/mod_business/dto"
	"github.com/atam/atamlink/internal/mod_business/entity"
	"github.com/atam/atamlink/internal/mod_business/repository"
	userRepo "github.com/atam/atamlink/internal/mod_user/repository"
	"github.com/atam/atamlink/internal/service"
	"github.com/atam/atamlink/pkg/database"
	"github.com/atam/atamlink/pkg/errors"
	"github.com/google/uuid"
)

// BusinessUseCase interface untuk business use case
type BusinessUseCase interface {
	Create(profileID int64, req *dto.CreateBusinessRequest) (*dto.BusinessResponse, error)
	GetByID(id int64, profileID int64) (*dto.BusinessResponse, error)
	GetBySlug(slug string) (*dto.BusinessResponse, error)
	List(profileID int64, filter *dto.BusinessFilter, page, perPage int, orderBy string) ([]*dto.BusinessListResponse, int64, error)
	Update(id int64, profileID int64, req *dto.UpdateBusinessRequest) (*dto.BusinessResponse, error)
	Delete(id int64, profileID int64) error

	// User management
	AddUser(businessID int64, profileID int64, req *dto.AddUserRequest) error
	UpdateUserRole(businessID int64, profileID int64, targetProfileID int64, role string) error
	RemoveUser(businessID int64, profileID int64, targetProfileID int64) error

	// Invite management
	CreateInvite(businessID int64, profileID int64, req *dto.CreateInviteRequest) (*dto.InviteResponse, error)
	AcceptInvite(req *dto.AcceptInviteRequest) error
}

type businessUseCase struct {
	db           *sql.DB
	businessRepo repository.BusinessRepository
	userRepo     userRepo.UserRepository
	slugService  service.SlugService
}

// NewBusinessUseCase membuat instance business use case baru
func NewBusinessUseCase(
	db *sql.DB,
	businessRepo repository.BusinessRepository,
	userRepo userRepo.UserRepository,
	slugService service.SlugService,
) BusinessUseCase {
	return &businessUseCase{
		db:           db,
		businessRepo: businessRepo,
		userRepo:     userRepo,
		slugService:  slugService,
	}
}

// Create membuat business baru
func (uc *businessUseCase) Create(profileID int64, req *dto.CreateBusinessRequest) (*dto.BusinessResponse, error) {
	// Validasi business type
	if !constant.IsValidBusinessType(req.Type) {
		return nil, errors.New(errors.ErrValidation, constant.ErrMsgBusinessTypeInvalid, 400)
	}

	// Generate atau validate slug
	var slug string
	if req.Slug != "" {
		if !uc.slugService.IsValid(req.Slug) {
			return nil, errors.New(errors.ErrValidation, "Slug tidak valid", 400)
		}

		// Check if slug exists
		exists, err := uc.businessRepo.IsSlugExists(req.Slug)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New(errors.ErrConflict, constant.ErrMsgBusinessSlugExists, 409)
		}
		slug = req.Slug
	} else {
		// Generate unique slug
		generatedSlug, err := service.GenerateUniqueSlug(
			req.Name,
			uc.slugService,
			func(s string) (bool, error) {
				return uc.businessRepo.IsSlugExists(s)
			},
			5,
		)
		if err != nil {
			return nil, err
		}
		slug = generatedSlug
	}

	// Start transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Create business
	business := &entity.Business{
		Slug:      slug,
		Name:      req.Name,
		Type:      req.Type,
		IsActive:  true,
		CreatedBy: profileID,
		CreatedAt: time.Now(),
	}

	if err := uc.businessRepo.Create(tx, business); err != nil {
		return nil, err
	}

	// Add creator as owner
	businessUser := &entity.BusinessUser{
		BusinessID: business.ID,
		ProfileID:  profileID,
		Role:       constant.RoleOwner,
		IsOwner:    true,
		IsActive:   true,
		CreatedAt:  time.Now(),
	}

	if err := uc.businessRepo.AddUser(tx, businessUser); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	// Get complete business data
	return uc.GetByID(business.ID, profileID)
}

// GetByID mendapatkan business by ID
func (uc *businessUseCase) GetByID(id int64, profileID int64) (*dto.BusinessResponse, error) {
	// Get business
	business, err := uc.businessRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check access jika profileID > 0
	if profileID > 0 {
		user, err := uc.businessRepo.GetUserByBusinessAndProfile(id, profileID)
		if err != nil {
			return nil, err
		}
		if user == nil || !user.IsActive {
			return nil, errors.New(errors.ErrForbidden, constant.ErrMsgBusinessAccessDenied, 403)
		}
	}

	// Get users
	users, err := uc.businessRepo.GetUsersByBusinessID(id)
	if err != nil {
		return nil, err
	}

	// Get active subscription
	subscription, err := uc.businessRepo.GetActiveSubscription(id)
	if err != nil {
		return nil, err
	}

	// Convert to response
	return uc.toBusinessResponse(business, users, subscription), nil
}

// GetBySlug mendapatkan business by slug
func (uc *businessUseCase) GetBySlug(slug string) (*dto.BusinessResponse, error) {
	business, err := uc.businessRepo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}

	// Public access, no need to check permission
	return uc.toBusinessResponse(business, nil, nil), nil
}

// List mendapatkan list businesses
func (uc *businessUseCase) List(profileID int64, filter *dto.BusinessFilter, page, perPage int, orderBy string) ([]*dto.BusinessListResponse, int64, error) {
	// Build filter
	repoFilter := repository.ListFilter{
		Limit:   perPage,
		Offset:  (page - 1) * perPage,
		OrderBy: orderBy,
	}

	if filter != nil {
		repoFilter.Search = filter.Search
		repoFilter.Type = filter.Type
		repoFilter.IsActive = filter.IsActive
		repoFilter.IsSuspended = filter.IsSuspended
		if filter.ProfileID > 0 {
			repoFilter.ProfileID = filter.ProfileID
		} else if profileID > 0 {
			repoFilter.ProfileID = profileID
		}
	} else if profileID > 0 {
		repoFilter.ProfileID = profileID
	}

	// Get businesses
	businesses, total, err := uc.businessRepo.List(repoFilter)
	if err != nil {
		return nil, 0, err
	}

	// Convert to response
	responses := make([]*dto.BusinessListResponse, len(businesses))
	for i, business := range businesses {
		responses[i] = &dto.BusinessListResponse{
			ID:          business.ID,
			Slug:        business.Slug,
			Name:        business.Name,
			Type:        business.Type,
			IsActive:    business.IsActive,
			IsSuspended: business.IsSuspended,
			CreatedAt:   business.CreatedAt,
			UpdatedAt:   business.UpdatedAt,
		}
	}

	return responses, total, nil
}

// Update update business
func (uc *businessUseCase) Update(id int64, profileID int64, req *dto.UpdateBusinessRequest) (*dto.BusinessResponse, error) {
	// Get existing business
	business, err := uc.businessRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check permission
	if err := uc.checkBusinessPermission(id, profileID, constant.PermBusinessUpdate); err != nil {
		return nil, err
	}

	// Validate updates
	if req.Type != "" && !constant.IsValidBusinessType(req.Type) {
		return nil, errors.New(errors.ErrValidation, constant.ErrMsgBusinessTypeInvalid, 400)
	}

	// Update fields
	if req.Name != "" {
		business.Name = req.Name
	}
	if req.Type != "" {
		business.Type = req.Type
	}
	if req.IsActive != nil {
		business.IsActive = *req.IsActive
	}

	business.UpdatedBy = database.NullInt64(profileID)
	business.UpdatedAt = &[]time.Time{time.Now()}[0]

	// Update in transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.businessRepo.Update(tx, business); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	// Return updated business
	return uc.GetByID(id, profileID)
}

// Delete soft delete business
func (uc *businessUseCase) Delete(id int64, profileID int64) error {
	// Check permission
	if err := uc.checkBusinessPermission(id, profileID, constant.PermBusinessDelete); err != nil {
		return err
	}

	// Delete in transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.businessRepo.Delete(tx, id); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

// AddUser menambahkan user ke business
func (uc *businessUseCase) AddUser(businessID int64, profileID int64, req *dto.AddUserRequest) error {
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

		return tx.Commit()
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

	return tx.Commit()
}

// UpdateUserRole update role user
func (uc *businessUseCase) UpdateUserRole(businessID int64, profileID int64, targetProfileID int64, role string) error {
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

	// Update role
	tx, err := uc.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.businessRepo.UpdateUserRole(tx, businessID, targetProfileID, role); err != nil {
		return err
	}

	return tx.Commit()
}

// RemoveUser hapus user dari business
func (uc *businessUseCase) RemoveUser(businessID int64, profileID int64, targetProfileID int64) error {
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

	// Remove user
	tx, err := uc.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.businessRepo.RemoveUser(tx, businessID, targetProfileID); err != nil {
		return err
	}

	return tx.Commit()
}

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

// Helper methods

func (uc *businessUseCase) checkBusinessPermission(businessID, profileID int64, permission string) error {
	// Get user role in business
	user, err := uc.businessRepo.GetUserByBusinessAndProfile(businessID, profileID)
	if err != nil {
		return err
	}

	if user == nil || !user.IsActive {
		return errors.New(errors.ErrForbidden, constant.ErrMsgBusinessAccessDenied, 403)
	}

	// Check permission
	if !constant.HasPermission(user.Role, permission) {
		return errors.New(errors.ErrForbidden, "Anda tidak memiliki izin untuk aksi ini", 403)
	}

	return nil
}

func (uc *businessUseCase) toBusinessResponse(business *entity.Business, users []*entity.BusinessUser, subscription *entity.BusinessSubscription) *dto.BusinessResponse {
	resp := &dto.BusinessResponse{
		ID:               business.ID,
		Slug:             business.Slug,
		Name:             business.Name,
		Type:             business.Type,
		IsActive:         business.IsActive,
		IsSuspended:      business.IsSuspended,
		SuspensionReason: business.GetSuspensionReason(),
		CreatedBy:        business.CreatedBy,
		CreatedAt:        business.CreatedAt,
		UpdatedAt:        business.UpdatedAt,
	}

	// Add users
	if users != nil {
		resp.Users = make([]dto.BusinessUserResponse, len(users))
		for i, user := range users {
			resp.Users[i] = dto.BusinessUserResponse{
				ID:        user.ID,
				ProfileID: user.ProfileID,
				Role:      user.Role,
				IsOwner:   user.IsOwner,
				IsActive:  user.IsActive,
				JoinedAt:  user.CreatedAt,
			}

			if user.Profile != nil {
				resp.Users[i].Profile = &dto.ProfileResponse{
					ID:          user.Profile.ID,
					DisplayName: user.Profile.GetDisplayName(),
				}
			}
		}
	}

	// Add subscription
	if subscription != nil {
		resp.ActivePlan = &dto.SubscriptionResponse{
			ID:        subscription.ID,
			PlanID:    subscription.PlanID,
			Status:    subscription.Status,
			StartsAt:  subscription.StartsAt,
			ExpiresAt: subscription.ExpiresAt,
			CreatedAt: subscription.CreatedAt,
		}

		if subscription.Plan != nil {
			resp.ActivePlan.PlanName = subscription.Plan.Name
			resp.ActivePlan.Plan = &dto.PlanResponse{
				ID:       subscription.Plan.ID,
				Name:     subscription.Plan.Name,
				Price:    subscription.Plan.Price,
				Features: subscription.Plan.Features,
			}
		}
	}

	return resp
}