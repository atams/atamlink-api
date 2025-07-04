package repository

import (
	"database/sql"
	"encoding/json"

	"github.com/atam/atamlink/internal/mod_master/entity"
	"github.com/atam/atamlink/pkg/database"
	"github.com/atam/atamlink/pkg/errors"
)

// MasterRepository interface untuk master repository
type MasterRepository interface {
	// Plan methods
	ListPlans(filter PlanFilter) ([]*entity.MasterPlan, error)
	GetPlanByID(id int64) (*entity.MasterPlan, error)
	
	// Theme methods
	ListThemes(filter ThemeFilter) ([]*entity.MasterTheme, error)
	GetThemeByID(id int64) (*entity.MasterTheme, error)
}

type masterRepository struct {
	db *sql.DB
}

// NewMasterRepository membuat instance master repository baru
func NewMasterRepository(db *sql.DB) MasterRepository {
	return &masterRepository{db: db}
}

// PlanFilter filter untuk list plans
type PlanFilter struct {
	IsActive *bool
	IsFree   *bool
	MaxPrice *int
	MinPrice *int
}

// ThemeFilter filter untuk list themes
type ThemeFilter struct {
	Search    string
	Type      string
	IsActive  *bool
	IsPremium *bool
}

// ListPlans mendapatkan list plans
func (r *masterRepository) ListPlans(filter PlanFilter) ([]*entity.MasterPlan, error) {
	// Build query
	qb := database.NewQueryBuilder()
	qb.Select(
		"mp_id", "mp_name", "mp_price", "mp_duration",
		"mp_features", "mp_is_active", "mp_created_at",
	).From("atamlink.master_plans")

	// Apply filters
	if filter.IsActive != nil {
		qb.Where("mp_is_active = ?", *filter.IsActive)
	}
	
	if filter.IsFree != nil {
		if *filter.IsFree {
			qb.Where("mp_price = 0")
		} else {
			qb.Where("mp_price > 0")
		}
	}
	
	if filter.MaxPrice != nil {
		qb.Where("mp_price <= ?", *filter.MaxPrice)
	}
	
	if filter.MinPrice != nil {
		qb.Where("mp_price >= ?", *filter.MinPrice)
	}

	// Order by price
	qb.OrderBy("mp_price ASC, mp_id ASC")

	query, args := qb.Build()
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query plans")
	}
	defer rows.Close()

	plans := make([]*entity.MasterPlan, 0)
	for rows.Next() {
		plan := &entity.MasterPlan{}
		var featuresJSON []byte
		
		err := rows.Scan(
			&plan.ID,
			&plan.Name,
			&plan.Price,
			&plan.Duration,
			&featuresJSON,
			&plan.IsActive,
			&plan.CreatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan plan")
		}

		// Parse features
		if err := json.Unmarshal(featuresJSON, &plan.Features); err != nil {
			return nil, errors.Wrap(err, "failed to parse features")
		}

		plans = append(plans, plan)
	}

	return plans, nil
}

// GetPlanByID mendapatkan plan by ID
func (r *masterRepository) GetPlanByID(id int64) (*entity.MasterPlan, error) {
	query := `
		SELECT 
			mp_id, mp_name, mp_price, mp_duration,
			mp_features, mp_is_active, mp_created_at
		FROM atamlink.master_plans
		WHERE mp_id = $1`

	plan := &entity.MasterPlan{}
	var featuresJSON []byte
	
	err := r.db.QueryRow(query, id).Scan(
		&plan.ID,
		&plan.Name,
		&plan.Price,
		&plan.Duration,
		&featuresJSON,
		&plan.IsActive,
		&plan.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrPlanNotFound, "Plan tidak ditemukan", 404)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get plan")
	}

	// Parse features
	if err := json.Unmarshal(featuresJSON, &plan.Features); err != nil {
		return nil, errors.Wrap(err, "failed to parse features")
	}

	return plan, nil
}

// ListThemes mendapatkan list themes
func (r *masterRepository) ListThemes(filter ThemeFilter) ([]*entity.MasterTheme, error) {
	// Build query
	qb := database.NewQueryBuilder()
	qb.Select(
		"mt_id", "mt_name", "mt_description", "mt_type",
		"mt_default_settings", "mt_is_premium", "mt_is_active", "mt_created_at",
	).From("atamlink.master_themes")

	// Apply filters
	if filter.Search != "" {
		qb.Where("(LOWER(mt_name) LIKE LOWER($1) OR LOWER(mt_description) LIKE LOWER($1))", "%"+filter.Search+"%")
	}
	
	if filter.Type != "" {
		qb.Where("mt_type = ?", filter.Type)
	}
	
	if filter.IsActive != nil {
		qb.Where("mt_is_active = ?", *filter.IsActive)
	}
	
	if filter.IsPremium != nil {
		qb.Where("mt_is_premium = ?", *filter.IsPremium)
	}

	// Order by name
	qb.OrderBy("mt_name ASC")

	query, args := qb.Build()
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query themes")
	}
	defer rows.Close()

	themes := make([]*entity.MasterTheme, 0)
	for rows.Next() {
		theme := &entity.MasterTheme{}
		var settingsJSON []byte
		
		err := rows.Scan(
			&theme.ID,
			&theme.Name,
			&theme.Description,
			&theme.Type,
			&settingsJSON,
			&theme.IsPremium,
			&theme.IsActive,
			&theme.CreatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan theme")
		}

		// Parse default settings
		if err := json.Unmarshal(settingsJSON, &theme.DefaultSettings); err != nil {
			return nil, errors.Wrap(err, "failed to parse default settings")
		}

		themes = append(themes, theme)
	}

	return themes, nil
}

// GetThemeByID mendapatkan theme by ID
func (r *masterRepository) GetThemeByID(id int64) (*entity.MasterTheme, error) {
	query := `
		SELECT 
			mt_id, mt_name, mt_description, mt_type,
			mt_default_settings, mt_is_premium, mt_is_active, mt_created_at
		FROM atamlink.master_themes
		WHERE mt_id = $1`

	theme := &entity.MasterTheme{}
	var settingsJSON []byte
	
	err := r.db.QueryRow(query, id).Scan(
		&theme.ID,
		&theme.Name,
		&theme.Description,
		&theme.Type,
		&settingsJSON,
		&theme.IsPremium,
		&theme.IsActive,
		&theme.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrNotFound, "Theme tidak ditemukan", 404)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get theme")
	}

	// Parse default settings
	if err := json.Unmarshal(settingsJSON, &theme.DefaultSettings); err != nil {
		return nil, errors.Wrap(err, "failed to parse default settings")
	}

	return theme, nil
}