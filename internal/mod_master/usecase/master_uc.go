package usecase

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/middleware"
	"github.com/atam/atamlink/internal/mod_master/dto"
	"github.com/atam/atamlink/internal/mod_master/entity"
	"github.com/atam/atamlink/internal/mod_master/repository"
	"github.com/atam/atamlink/pkg/database"
	"github.com/atam/atamlink/pkg/errors"
)

// MasterUseCase interface untuk master use case
type MasterUseCase interface {
	// Plan methods
	CreatePlan(req *dto.CreatePlanRequest) (*dto.PlanResponse, error)
	UpdatePlan(ctx *gin.Context, id int64, req *dto.UpdatePlanRequest) (*dto.PlanResponse, error)
	DeletePlan(ctx *gin.Context, id int64) error
	ListPlans(filter *dto.PlanFilter) ([]*dto.PlanListResponse, error)
	GetPlanByID(id int64) (*dto.PlanResponse, error)
	
	// Theme methods
	CreateTheme(req *dto.CreateThemeRequest) (*dto.ThemeResponse, error)
	UpdateTheme(ctx *gin.Context, id int64, req *dto.UpdateThemeRequest) (*dto.ThemeResponse, error)
	DeleteTheme(ctx *gin.Context, id int64) error
	ListThemes(filter *dto.ThemeFilter) ([]*dto.ThemeListResponse, error)
	GetThemeByID(id int64) (*dto.ThemeResponse, error)
}

type masterUseCase struct {
	db         *sql.DB
	masterRepo repository.MasterRepository
}

// NewMasterUseCase membuat instance master use case baru
func NewMasterUseCase(db *sql.DB, masterRepo repository.MasterRepository) MasterUseCase {
	return &masterUseCase{
		db:         db,
		masterRepo: masterRepo,
	}
}

// CreatePlan membuat plan baru
func (uc *masterUseCase) CreatePlan(req *dto.CreatePlanRequest) (*dto.PlanResponse, error) {
	// Validasi nama unik
	exists, err := uc.masterRepo.IsPlanNameExists(req.Name, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New(errors.ErrConflict, "Nama plan sudah digunakan", 409)
	}

	// Validasi duration format
	// PostgreSQL interval format: '30 days', '1 mon', '1 year'
	if !isValidInterval(req.Duration) {
		return nil, errors.New(errors.ErrValidation, "Format duration tidak valid", 400)
	}

	// Set default features jika kosong
	if req.Features == nil {
		req.Features = getDefaultPlanFeatures()
	}

	// Start transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Create plan
	plan := &entity.MasterPlan{
		Name:      req.Name,
		Price:     req.Price,
		Duration:  req.Duration,
		Features:  req.Features,
		IsActive:  req.IsActive,
		CreatedAt: time.Now(),
	}

	if err := uc.masterRepo.CreatePlan(tx, plan); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	// Return response
	return &dto.PlanResponse{
		ID:           plan.ID,
		Name:         plan.Name,
		Price:        plan.Price,
		Duration:     plan.Duration,
		DurationDays: plan.GetDurationDays(),
		Features:     plan.Features,
		IsActive:     plan.IsActive,
		CreatedAt:    plan.CreatedAt,
	}, nil
}

// UpdatePlan update existing plan
func (uc *masterUseCase) UpdatePlan(ctx *gin.Context, id int64, req *dto.UpdatePlanRequest) (*dto.PlanResponse, error) {
	// Get existing plan
	plan, err := uc.masterRepo.GetPlanByID(id)
	if err != nil {
		return nil, err
	}

	// Inject old_data ke audit context
	if ctx != nil {
		ctx.Set(middleware.GinKeyAuditOldData, plan)
	}

	// Check if has active subscriptions
	hasActive, err := uc.masterRepo.HasActiveSubscriptions(id)
	if err != nil {
		return nil, err
	}
	if hasActive {
		return nil, errors.New(errors.ErrConflict, "Tidak dapat mengubah plan yang memiliki subscription aktif", 409)
	}

	// Validate name uniqueness if changed
	if req.Name != "" && req.Name != plan.Name {
		exists, err := uc.masterRepo.IsPlanNameExists(req.Name, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New(errors.ErrConflict, "Nama plan sudah digunakan", 409)
		}
	}

	// Validate duration if changed
	if req.Duration != "" && !isValidInterval(req.Duration) {
		return nil, errors.New(errors.ErrValidation, "Format duration tidak valid", 400)
	}

	// Update fields
	if req.Name != "" {
		plan.Name = req.Name
	}
	if req.Price != nil {
		plan.Price = *req.Price
	}
	if req.Duration != "" {
		plan.Duration = req.Duration
	}
	if req.Features != nil {
		plan.Features = req.Features
	}
	if req.IsActive != nil {
		plan.IsActive = *req.IsActive
	}

	// Update in transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.masterRepo.UpdatePlan(tx, plan); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	// Return response
	return &dto.PlanResponse{
		ID:           plan.ID,
		Name:         plan.Name,
		Price:        plan.Price,
		Duration:     plan.Duration,
		DurationDays: plan.GetDurationDays(),
		Features:     plan.Features,
		IsActive:     plan.IsActive,
		CreatedAt:    plan.CreatedAt,
	}, nil
}

// DeletePlan soft delete plan
func (uc *masterUseCase) DeletePlan(ctx *gin.Context, id int64) error {
	// Get existing plan
	plan, err := uc.masterRepo.GetPlanByID(id)
	if err != nil {
		return err
	}

	// Inject old_data ke audit context
	if ctx != nil {
		ctx.Set(middleware.GinKeyAuditOldData, plan)
	}

	// Check if has active subscriptions
	hasActive, err := uc.masterRepo.HasActiveSubscriptions(id)
	if err != nil {
		return err
	}
	if hasActive {
		return errors.New(errors.ErrConflict, "Tidak dapat menghapus plan yang memiliki subscription aktif", 409)
	}

	// Delete in transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.masterRepo.DeletePlan(tx, id); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

// CreateTheme membuat theme baru
func (uc *masterUseCase) CreateTheme(req *dto.CreateThemeRequest) (*dto.ThemeResponse, error) {
	// Validate theme type
	if !constant.IsValidThemeType(req.Type) {
		return nil, errors.New(errors.ErrValidation, "Tipe theme tidak valid", 400)
	}

	// Validate name uniqueness
	exists, err := uc.masterRepo.IsThemeNameExists(req.Name, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New(errors.ErrConflict, "Nama theme sudah digunakan", 409)
	}

	// Set default settings if empty
	if req.DefaultSettings == nil {
		req.DefaultSettings = getDefaultThemeSettings(req.Type)
	}

	// Start transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Create theme
	theme := &entity.MasterTheme{
		Name:            req.Name,
		Description:     database.NullString(req.Description),
		Type:            req.Type,
		DefaultSettings: req.DefaultSettings,
		IsPremium:       req.IsPremium,
		IsActive:        req.IsActive,
		CreatedAt:       time.Now(),
	}

	if err := uc.masterRepo.CreateTheme(tx, theme); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	// Return response
	return &dto.ThemeResponse{
		ID:              theme.ID,
		Name:            theme.Name,
		Description:     theme.GetDescription(),
		Type:            theme.Type,
		DefaultSettings: theme.DefaultSettings,
		IsPremium:       theme.IsPremium,
		IsActive:        theme.IsActive,
		CreatedAt:       theme.CreatedAt,
		PreviewURL:      fmt.Sprintf("/themes/%d/preview", theme.ID),
	}, nil
}

// UpdateTheme update existing theme
func (uc *masterUseCase) UpdateTheme(ctx *gin.Context, id int64, req *dto.UpdateThemeRequest) (*dto.ThemeResponse, error) {
	// Get existing theme
	theme, err := uc.masterRepo.GetThemeByID(id)
	if err != nil {
		return nil, err
	}

	// Inject old_data ke audit context
	if ctx != nil {
		ctx.Set(middleware.GinKeyAuditOldData, theme)
	}

	// Validate name uniqueness if changed
	if req.Name != "" && req.Name != theme.Name {
		exists, err := uc.masterRepo.IsThemeNameExists(req.Name, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New(errors.ErrConflict, "Nama theme sudah digunakan", 409)
		}
	}

	// Validate theme type if changed
	if req.Type != "" && !constant.IsValidThemeType(req.Type) {
		return nil, errors.New(errors.ErrValidation, "Tipe theme tidak valid", 400)
	}

	// Update fields
	if req.Name != "" {
		theme.Name = req.Name
	}
	if req.Description != "" {
		theme.Description = database.NullString(req.Description)
	}
	if req.Type != "" {
		theme.Type = req.Type
	}
	if req.DefaultSettings != nil {
		theme.DefaultSettings = req.DefaultSettings
	}
	if req.IsPremium != nil {
		theme.IsPremium = *req.IsPremium
	}
	if req.IsActive != nil {
		theme.IsActive = *req.IsActive
	}

	// Update in transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.masterRepo.UpdateTheme(tx, theme); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	// Return response
	return &dto.ThemeResponse{
		ID:              theme.ID,
		Name:            theme.Name,
		Description:     theme.GetDescription(),
		Type:            theme.Type,
		DefaultSettings: theme.DefaultSettings,
		IsPremium:       theme.IsPremium,
		IsActive:        theme.IsActive,
		CreatedAt:       theme.CreatedAt,
		PreviewURL:      fmt.Sprintf("/themes/%d/preview", theme.ID),
	}, nil
}

// DeleteTheme soft delete theme
func (uc *masterUseCase) DeleteTheme(ctx *gin.Context, id int64) error {
	// Get existing theme
	theme, err := uc.masterRepo.GetThemeByID(id)
	if err != nil {
		return err
	}

	// Inject old_data ke audit context
	if ctx != nil {
		ctx.Set(middleware.GinKeyAuditOldData, theme)
	}

	// Check if has active catalogs
	hasActive, err := uc.masterRepo.HasActiveCatalogs(id)
	if err != nil {
		return err
	}
	if hasActive {
		return errors.New(errors.ErrConflict, "Tidak dapat menghapus theme yang sedang digunakan", 409)
	}

	// Delete in transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.masterRepo.DeleteTheme(tx, id); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

// ListPlans mendapatkan list plans
func (uc *masterUseCase) ListPlans(filter *dto.PlanFilter) ([]*dto.PlanListResponse, error) {
	// Build repository filter
	repoFilter := repository.PlanFilter{}
	
	if filter != nil {
		repoFilter.IsActive = filter.IsActive
		repoFilter.IsFree = filter.IsFree
		repoFilter.MaxPrice = filter.MaxPrice
		repoFilter.MinPrice = filter.MinPrice
	}

	// Get plans
	plans, err := uc.masterRepo.ListPlans(repoFilter)
	if err != nil {
		return nil, err
	}

	// Convert to response
	responses := make([]*dto.PlanListResponse, len(plans))
	for i, plan := range plans {
		responses[i] = &dto.PlanListResponse{
			ID:           plan.ID,
			Name:         plan.Name,
			Price:        plan.Price,
			PriceDisplay: formatPrice(plan.Price),
			Duration:     plan.Duration,
			DurationDays: plan.GetDurationDays(),
			IsActive:     plan.IsActive,
			IsFree:       plan.IsFree(),
			CreatedAt:    plan.CreatedAt,
		}
	}

	return responses, nil
}

// GetPlanByID mendapatkan plan by ID
func (uc *masterUseCase) GetPlanByID(id int64) (*dto.PlanResponse, error) {
	// Get plan
	plan, err := uc.masterRepo.GetPlanByID(id)
	if err != nil {
		return nil, err
	}

	// Convert to response
	return &dto.PlanResponse{
		ID:           plan.ID,
		Name:         plan.Name,
		Price:        plan.Price,
		Duration:     plan.Duration,
		DurationDays: plan.GetDurationDays(),
		Features:     plan.Features,
		IsActive:     plan.IsActive,
		CreatedAt:    plan.CreatedAt,
	}, nil
}

// ListThemes mendapatkan list themes
func (uc *masterUseCase) ListThemes(filter *dto.ThemeFilter) ([]*dto.ThemeListResponse, error) {
	// Build repository filter
	repoFilter := repository.ThemeFilter{}
	
	if filter != nil {
		repoFilter.Search = filter.Search
		repoFilter.Type = filter.Type
		repoFilter.IsActive = filter.IsActive
		repoFilter.IsPremium = filter.IsPremium
	}

	// Get themes
	themes, err := uc.masterRepo.ListThemes(repoFilter)
	if err != nil {
		return nil, err
	}

	// Convert to response
	responses := make([]*dto.ThemeListResponse, len(themes))
	for i, theme := range themes {
		responses[i] = &dto.ThemeListResponse{
			ID:          theme.ID,
			Name:        theme.Name,
			Description: theme.GetDescription(),
			Type:        theme.Type,
			IsPremium:   theme.IsPremium,
			IsActive:    theme.IsActive,
			CreatedAt:   theme.CreatedAt,
			PreviewURL:  fmt.Sprintf("/themes/%d/preview", theme.ID),
		}
	}

	return responses, nil
}

// GetThemeByID mendapatkan theme by ID
func (uc *masterUseCase) GetThemeByID(id int64) (*dto.ThemeResponse, error) {
	// Get theme
	theme, err := uc.masterRepo.GetThemeByID(id)
	if err != nil {
		return nil, err
	}

	// Convert to response
	return &dto.ThemeResponse{
		ID:              theme.ID,
		Name:            theme.Name,
		Description:     theme.GetDescription(),
		Type:            theme.Type,
		DefaultSettings: theme.DefaultSettings,
		IsPremium:       theme.IsPremium,
		IsActive:        theme.IsActive,
		CreatedAt:       theme.CreatedAt,
		PreviewURL:      fmt.Sprintf("/themes/%d/preview", theme.ID),
	}, nil
}

// Helper functions

func formatPrice(price int) string {
	if price == 0 {
		return "Gratis"
	}
	
	// Format to Indonesian Rupiah
	// Simple implementation - enhance as needed
	return fmt.Sprintf("Rp %d", price)
}

func isValidInterval(interval string) bool {
	// Simple validation for PostgreSQL interval format
	// Examples: '30 days', '1 mon', '1 year', '6 months'
	validUnits := []string{"day", "days", "mon", "month", "months", "year", "years"}
	
	for _, unit := range validUnits {
		if len(interval) > len(unit) && interval[len(interval)-len(unit):] == unit {
			return true
		}
	}
	
	return false
}

func getDefaultPlanFeatures() map[string]interface{} {
	return map[string]interface{}{
		"max_catalogs":      1,
		"max_products":      50,
		"max_users":         1,
		"custom_domain":     false,
		"analytics":         false,
		"priority_support":  false,
		"remove_watermark":  false,
		"advanced_themes":   false,
		"api_access":        false,
	}
}

func getDefaultThemeSettings(themeType string) map[string]interface{} {
	// Default settings based on theme type
	baseSettings := map[string]interface{}{
		"colors": map[string]string{
			"primary":    "#1a73e8",
			"secondary":  "#f8f9fa",
			"background": "#ffffff",
			"text":       "#202124",
		},
		"typography": map[string]interface{}{
			"font_family": "Inter, sans-serif",
			"font_size_base": "16px",
		},
		"layout": map[string]string{
			"container_width": "1200px",
			"spacing": "1rem",
		},
	}

	// Customize based on theme type
	switch themeType {
	case constant.ThemeMinimal:
		baseSettings["colors"].(map[string]string)["primary"] = "#000000"
	case constant.ThemeBold:
		baseSettings["colors"].(map[string]string)["primary"] = "#ff4757"
	case constant.ThemeElegant:
		baseSettings["colors"].(map[string]string)["primary"] = "#6c5ce7"
	case constant.ThemePlayful:
		baseSettings["colors"].(map[string]string)["primary"] = "#00b894"
	}

	return baseSettings
}