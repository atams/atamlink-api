package usecase

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/middleware"
	"github.com/atam/atamlink/internal/mod_business/dto"
	"github.com/atam/atamlink/internal/mod_business/entity"
	"github.com/atam/atamlink/internal/mod_business/repository"
	"github.com/atam/atamlink/internal/service"
	"github.com/atam/atamlink/pkg/errors"
)

// Create membuat business baru
func (uc *businessUseCase) Create(ctx *gin.Context, profileID int64, req *dto.CreateBusinessRequest) (*dto.BusinessResponse, error) {
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

	// Handle logo URL from DTO (already uploaded by handler)
	if req.LogoURL != nil {
		business.LogoURL = sql.NullString{
			String: *req.LogoURL,
			Valid:  true,
		}
	}

	// Create business
	if err := uc.businessRepo.Create(tx, business); err != nil {
		// Jika create gagal dan logo sudah diupload, hapus dari Cloudinary
		if req.LogoURL != nil && *req.LogoURL != "" {
			go func() {
				// Best effort delete, tidak perlu handle error
				_ = uc.uploadService.DeleteFromCloudinary(*req.LogoURL)
			}()
		}
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
		// Rollback upload jika add user gagal
		if req.LogoURL != nil && *req.LogoURL != "" {
			go func() {
				_ = uc.uploadService.DeleteFromCloudinary(*req.LogoURL)
			}()
		}
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		// Rollback upload jika commit gagal
		if req.LogoURL != nil && *req.LogoURL != "" {
			go func() {
				_ = uc.uploadService.DeleteFromCloudinary(*req.LogoURL)
			}()
		}
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	// Get complete business data setelah commit berhasil
	createdBusiness, err := uc.GetByID(business.ID, profileID)
	if err != nil {
		return nil, err
	}

	// Set new data untuk audit (data business setelah create)
	// Untuk CREATE, old_data kosong, new_data berisi data yang dibuat
	if ctx != nil {
		ctx.Set(middleware.GinKeyAuditNewData, createdBusiness)
	}

	return createdBusiness, nil
}

// GetByID mendapatkan business by ID dengan subscription info
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

	// Get active subscription (akan null jika tidak ada)
	subscription, err := uc.businessRepo.GetActiveSubscription(id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Convert to response dengan subscription info (null jika tidak ada)
	return uc.toBusinessResponse(business, users, subscription), nil
}

// GetBySlug mendapatkan business by slug
func (uc *businessUseCase) GetBySlug(slug string) (*dto.BusinessResponse, error) {
	business, err := uc.businessRepo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}

	// Public access, no need to check permission
	// Get subscription info untuk public view
	subscription, err := uc.businessRepo.GetActiveSubscription(business.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return uc.toBusinessResponse(business, nil, subscription), nil
}

// List mendapatkan list businesses dengan user count dan role
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
		repoFilter.IsSuspended = filter.IsSuspended
		if filter.ProfileID > 0 {
			repoFilter.ProfileID = filter.ProfileID
		} else if profileID > 0 {
			repoFilter.ProfileID = profileID
		}
	} else if profileID > 0 {
		repoFilter.ProfileID = profileID
	}

	// Get businesses dengan user count dan role
	businesses, total, err := uc.businessRepo.GetBusinessesWithUserCount(repoFilter)
	if err != nil {
		return nil, 0, err
	}

	// Convert to response
	responses := make([]*dto.BusinessListResponse, len(businesses))
	for i, business := range businesses {
		response := &dto.BusinessListResponse{
			ID:          business.ID,
			Slug:        business.Slug,
			Name:        business.Name,
			Type:        business.Type,
			IsActive:    business.IsActive,
			IsSuspended: business.IsSuspended,
			UserCount:   business.UserCount,
			CreatedAt:   business.CreatedAt,
			UpdatedAt:   business.UpdatedAt,
		}

		// Set logo URL jika ada
		if business.LogoURL.Valid {
			response.LogoURL = &business.LogoURL.String
		}

		// Set user role jika ada
		if business.UserRole != nil {
			response.UserRole = business.UserRole
		}

		responses[i] = response
	}

	return responses, total, nil
}

// Update update business
func (uc *businessUseCase) Update(ctx *gin.Context, id int64, profileID int64, req *dto.UpdateBusinessRequest) (*dto.BusinessResponse, error) {
	// Get existing business
	business, err := uc.businessRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Create deep copy untuk old data audit SEBELUM modifikasi
	oldBusinessData := uc.deepCopyBusiness(business)

	// Set old data untuk audit dengan deep copy
	if ctx != nil {
		ctx.Set(middleware.GinKeyAuditOldData, oldBusinessData)
	}

	// Check permission
	if err := uc.checkBusinessPermission(id, profileID, constant.PermBusinessUpdate); err != nil {
		return nil, err
	}

	// Validate updates
	if req.Type != "" && !constant.IsValidBusinessType(req.Type) {
		return nil, errors.New(errors.ErrValidation, constant.ErrMsgBusinessTypeInvalid, 400)
	}

	// Start transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Track old logo URL untuk potential cleanup
	oldLogoURL := ""
	if business.LogoURL.Valid {
		oldLogoURL = business.LogoURL.String
	}

	// Update fields pada business object (setelah deep copy dibuat)
	if req.Name != "" {
		business.Name = req.Name
	}
	if req.Type != "" {
		business.Type = req.Type
	}
	if req.IsActive != nil {
		business.IsActive = *req.IsActive
	}

	// Handle logo URL from DTO (already uploaded by handler)
	if req.LogoURL != nil { 
		if *req.LogoURL == "" { 
			business.LogoURL = sql.NullString{Valid: false}
		} else {
			business.LogoURL = sql.NullString{
				String: *req.LogoURL,
				Valid:  true,
			}
		}
	}

	// Update metadata
	business.UpdatedBy = sql.NullInt64{
		Int64: profileID,
		Valid: true,
	}
	business.UpdatedAt = &[]time.Time{time.Now()}[0]

	// Execute update
	if err := uc.businessRepo.Update(tx, business); err != nil {
		// Rollback new upload if update fails
		if req.LogoURL != nil && *req.LogoURL != "" {
			go func() {
				_ = uc.uploadService.DeleteFromCloudinary(*req.LogoURL)
			}()
		}
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		// Rollback new upload if commit fails
		if req.LogoURL != nil && *req.LogoURL != "" {
			go func() {
				_ = uc.uploadService.DeleteFromCloudinary(*req.LogoURL)
			}()
		}
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	// Delete old logo if replaced or removed
	if oldLogoURL != "" && (req.LogoURL != nil && *req.LogoURL == "" || (req.LogoURL != nil && *req.LogoURL != oldLogoURL)) {
		go func() {
			_ = uc.uploadService.DeleteFromCloudinary(oldLogoURL)
		}()
	}

	// Get updated business setelah commit berhasil
	updatedBusiness, err := uc.GetByID(id, profileID)
	if err != nil {
		return nil, err
	}

	// Set new data untuk audit (data business setelah update)
	if ctx != nil {
		ctx.Set(middleware.GinKeyAuditNewData, updatedBusiness)
	}

	return updatedBusiness, nil
}

// Delete soft delete business
func (uc *businessUseCase) Delete(ctx *gin.Context, id int64, profileID int64) error {
	// Get existing business
	business, err := uc.businessRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Create deep copy untuk old data audit
	oldBusinessData := uc.deepCopyBusiness(business)

	// Set old data untuk audit
	if ctx != nil {
		ctx.Set(middleware.GinKeyAuditOldData, oldBusinessData)
	}

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

	// Untuk DELETE, new_data adalah null (tidak ada data baru)
	if ctx != nil {
		ctx.Set(middleware.GinKeyAuditNewData, nil)
	}

	return nil
}