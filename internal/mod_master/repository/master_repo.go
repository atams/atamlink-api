package repository

import (
	"database/sql"
	"encoding/json"
	
	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/mod_master/entity"
	"github.com/atam/atamlink/pkg/database"
	"github.com/atam/atamlink/pkg/errors"
)

// MasterRepository interface untuk master repository
type MasterRepository interface {
	// Plan methods
	CreatePlan(tx *sql.Tx, plan *entity.MasterPlan) error
	UpdatePlan(tx *sql.Tx, plan *entity.MasterPlan) error
	DeletePlan(tx *sql.Tx, id int64) error
	ListPlans(filter PlanFilter) ([]*entity.MasterPlan, error)
	GetPlanByID(id int64) (*entity.MasterPlan, error)
	IsPlanNameExists(name string, excludeID int64) (bool, error)
	HasActiveSubscriptions(planID int64) (bool, error)
	
	// Theme methods
	CreateTheme(tx *sql.Tx, theme *entity.MasterTheme) error
	UpdateTheme(tx *sql.Tx, theme *entity.MasterTheme) error
	DeleteTheme(tx *sql.Tx, id int64) error
	ListThemes(filter ThemeFilter) ([]*entity.MasterTheme, error)
	GetThemeByID(id int64) (*entity.MasterTheme, error)
	IsThemeNameExists(name string, excludeID int64) (bool, error)
	HasActiveCatalogs(themeID int64) (bool, error)
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

// CreatePlan membuat plan baru
func (r *masterRepository) CreatePlan(tx *sql.Tx, plan *entity.MasterPlan) error {
	featuresJSON, err := json.Marshal(plan.Features)
	if err != nil {
		return errors.Wrap(err, "failed to marshal features")
	}

	query := `
		INSERT INTO atamlink.master_plans (
			mp_name, mp_price, mp_duration, mp_features, 
			mp_is_active, mp_created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING mp_id`

	err = tx.QueryRow(
		query,
		plan.Name,
		plan.Price,
		plan.Duration,
		featuresJSON,
		plan.IsActive,
		plan.CreatedAt,
	).Scan(&plan.ID)

	if err != nil {
		return errors.Wrap(err, "failed to create plan")
	}

	return nil
}

// UpdatePlan update existing plan
func (r *masterRepository) UpdatePlan(tx *sql.Tx, plan *entity.MasterPlan) error {
	featuresJSON, err := json.Marshal(plan.Features)
	if err != nil {
		return errors.Wrap(err, "failed to marshal features")
	}

	query := `
		UPDATE atamlink.master_plans SET
			mp_name = $2,
			mp_price = $3,
			mp_duration = $4,
			mp_features = $5,
			mp_is_active = $6
		WHERE mp_id = $1`

	result, err := tx.Exec(
		query,
		plan.ID,
		plan.Name,
		plan.Price,
		plan.Duration,
		featuresJSON,
		plan.IsActive,
	)

	if err != nil {
		return errors.Wrap(err, "failed to update plan")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrPlanNotFound, constant.ErrMsgPlanNotFound, 404)
	}

	return nil
}

// DeletePlan soft delete plan (set is_active = false)
func (r *masterRepository) DeletePlan(tx *sql.Tx, id int64) error {
	query := `
		UPDATE atamlink.master_plans 
		SET mp_is_active = false
		WHERE mp_id = $1`

	result, err := tx.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete plan")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrPlanNotFound, constant.ErrMsgPlanNotFound, 404)
	}

	return nil
}

// IsPlanNameExists check if plan name already exists
func (r *masterRepository) IsPlanNameExists(name string, excludeID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM atamlink.master_plans 
			WHERE LOWER(mp_name) = LOWER($1) AND mp_id != $2
		)`

	var exists bool
	err := r.db.QueryRow(query, name, excludeID).Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "failed to check plan name exists")
	}

	return exists, nil
}

// HasActiveSubscriptions check if plan has active subscriptions
func (r *masterRepository) HasActiveSubscriptions(planID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM atamlink.business_subscriptions 
			WHERE bs_mp_id = $1 
			AND bs_status = 'active' 
			AND bs_expires_at > NOW()
		)`

	var exists bool
	err := r.db.QueryRow(query, planID).Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "failed to check active subscriptions")
	}

	return exists, nil
}

// CreateTheme membuat theme baru
func (r *masterRepository) CreateTheme(tx *sql.Tx, theme *entity.MasterTheme) error {
	settingsJSON, err := json.Marshal(theme.DefaultSettings)
	if err != nil {
		return errors.Wrap(err, "failed to marshal default settings")
	}

	query := `
		INSERT INTO atamlink.master_themes (
			mt_name, mt_description, mt_type, mt_default_settings,
			mt_is_premium, mt_is_active, mt_created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING mt_id`

	err = tx.QueryRow(
		query,
		theme.Name,
		theme.Description,
		theme.Type,
		settingsJSON,
		theme.IsPremium,
		theme.IsActive,
		theme.CreatedAt,
	).Scan(&theme.ID)

	if err != nil {
		return errors.Wrap(err, "failed to create theme")
	}

	return nil
}

// UpdateTheme update existing theme
func (r *masterRepository) UpdateTheme(tx *sql.Tx, theme *entity.MasterTheme) error {
	settingsJSON, err := json.Marshal(theme.DefaultSettings)
	if err != nil {
		return errors.Wrap(err, "failed to marshal default settings")
	}

	query := `
		UPDATE atamlink.master_themes SET
			mt_name = $2,
			mt_description = $3,
			mt_type = $4,
			mt_default_settings = $5,
			mt_is_premium = $6,
			mt_is_active = $7
		WHERE mt_id = $1`

	result, err := tx.Exec(
		query,
		theme.ID,
		theme.Name,
		theme.Description,
		theme.Type,
		settingsJSON,
		theme.IsPremium,
		theme.IsActive,
	)

	if err != nil {
		return errors.Wrap(err, "failed to update theme")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrNotFound, "Theme tidak ditemukan", 404)
	}

	return nil
}

// DeleteTheme soft delete theme (set is_active = false)
func (r *masterRepository) DeleteTheme(tx *sql.Tx, id int64) error {
	query := `
		UPDATE atamlink.master_themes 
		SET mt_is_active = false
		WHERE mt_id = $1`

	result, err := tx.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete theme")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrNotFound, "Theme tidak ditemukan", 404)
	}

	return nil
}

// IsThemeNameExists check if theme name already exists
func (r *masterRepository) IsThemeNameExists(name string, excludeID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM atamlink.master_themes 
			WHERE LOWER(mt_name) = LOWER($1) AND mt_id != $2
		)`

	var exists bool
	err := r.db.QueryRow(query, name, excludeID).Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "failed to check theme name exists")
	}

	return exists, nil
}

// HasActiveCatalogs check if theme has active catalogs
func (r *masterRepository) HasActiveCatalogs(themeID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM atamlink.catalogs 
			WHERE c_mt_id = $1 AND c_is_active = true
		)`

	var exists bool
	err := r.db.QueryRow(query, themeID).Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "failed to check active catalogs")
	}

	return exists, nil
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