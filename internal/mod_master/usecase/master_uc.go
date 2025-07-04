package usecase

import (
	"fmt"

	"github.com/atam/atamlink/internal/mod_master/dto"
	"github.com/atam/atamlink/internal/mod_master/repository"
)

// MasterUseCase interface untuk master use case
type MasterUseCase interface {
	// Plan methods
	ListPlans(filter *dto.PlanFilter) ([]*dto.PlanListResponse, error)
	GetPlanByID(id int64) (*dto.PlanResponse, error)
	
	// Theme methods
	ListThemes(filter *dto.ThemeFilter) ([]*dto.ThemeListResponse, error)
	GetThemeByID(id int64) (*dto.ThemeResponse, error)
}

type masterUseCase struct {
	masterRepo repository.MasterRepository
}

// NewMasterUseCase membuat instance master use case baru
func NewMasterUseCase(masterRepo repository.MasterRepository) MasterUseCase {
	return &masterUseCase{
		masterRepo: masterRepo,
	}
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