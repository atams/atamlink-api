package entity

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// MasterPlan entity untuk tabel master_plans
type MasterPlan struct {
	ID        int64                  `json:"id" db:"mp_id"`
	Name      string                 `json:"name" db:"mp_name"`
	Price     int                    `json:"price" db:"mp_price"`
	Duration  string                 `json:"duration" db:"mp_duration"` // PostgreSQL interval
	Features  map[string]interface{} `json:"features" db:"mp_features"`
	IsActive  bool                   `json:"is_active" db:"mp_is_active"`
	CreatedAt time.Time              `json:"created_at" db:"mp_created_at"`
}

// MasterTheme entity untuk tabel master_themes
type MasterTheme struct {
	ID              int64                  `json:"id" db:"mt_id"`
	Name            string                 `json:"name" db:"mt_name"`
	Description     sql.NullString         `json:"description" db:"mt_description"`
	Type            string                 `json:"type" db:"mt_type"`
	DefaultSettings map[string]interface{} `json:"default_settings" db:"mt_default_settings"`
	IsPremium       bool                   `json:"is_premium" db:"mt_is_premium"`
	IsActive        bool                   `json:"is_active" db:"mt_is_active"`
	CreatedAt       time.Time              `json:"created_at" db:"mt_created_at"`
}

// TableName methods
func (MasterPlan) TableName() string {
	return "atamlink.master_plans"
}

func (MasterTheme) TableName() string {
	return "atamlink.master_themes"
}

// Helper methods

// GetDescription get description dengan null handling
func (mt *MasterTheme) GetDescription() string {
	if mt.Description.Valid {
		return mt.Description.String
	}
	return ""
}

// SetDescription set description value
func (mt *MasterTheme) SetDescription(desc string) {
	mt.Description = sql.NullString{
		String: desc,
		Valid:  desc != "",
	}
}

// GetDurationDays konversi PostgreSQL interval ke days
func (mp *MasterPlan) GetDurationDays() int {
	// Parse PostgreSQL interval format
	// Contoh: "30 days", "1 mon", "1 year"
	// Simplified parsing - enhance sesuai kebutuhan
	
	// Default 30 days jika parsing gagal
	return 30
}

// PlanFeatures struktur untuk features plan
type PlanFeatures struct {
	MaxCatalogs      int  `json:"max_catalogs"`
	MaxProducts      int  `json:"max_products"`
	MaxUsers         int  `json:"max_users"`
	CustomDomain     bool `json:"custom_domain"`
	Analytics        bool `json:"analytics"`
	PrioritySupport  bool `json:"priority_support"`
	RemoveWatermark  bool `json:"remove_watermark"`
	AdvancedThemes   bool `json:"advanced_themes"`
	APIAccess        bool `json:"api_access"`
}

// GetPlanFeatures parse features ke struct
func (mp *MasterPlan) GetPlanFeatures() (*PlanFeatures, error) {
	data, err := json.Marshal(mp.Features)
	if err != nil {
		return nil, err
	}

	var features PlanFeatures
	if err := json.Unmarshal(data, &features); err != nil {
		return nil, err
	}

	return &features, nil
}

// SetPlanFeatures set features dari struct
func (mp *MasterPlan) SetPlanFeatures(features *PlanFeatures) error {
	data, err := json.Marshal(features)
	if err != nil {
		return err
	}

	mp.Features = make(map[string]interface{})
	return json.Unmarshal(data, &mp.Features)
}

// ThemeSettings struktur untuk default settings theme
type ThemeSettings struct {
	PrimaryColor     string                 `json:"primary_color"`
	SecondaryColor   string                 `json:"secondary_color"`
	BackgroundColor  string                 `json:"background_color"`
	TextColor        string                 `json:"text_color"`
	FontFamily       string                 `json:"font_family"`
	BorderRadius     string                 `json:"border_radius"`
	CustomCSS        string                 `json:"custom_css"`
	Layout           string                 `json:"layout"`
	Components       map[string]interface{} `json:"components"`
}

// GetThemeSettings parse default settings ke struct
func (mt *MasterTheme) GetThemeSettings() (*ThemeSettings, error) {
	data, err := json.Marshal(mt.DefaultSettings)
	if err != nil {
		return nil, err
	}

	var settings ThemeSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// SetThemeSettings set default settings dari struct
func (mt *MasterTheme) SetThemeSettings(settings *ThemeSettings) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	mt.DefaultSettings = make(map[string]interface{})
	return json.Unmarshal(data, &mt.DefaultSettings)
}

// IsFree check apakah plan gratis
func (mp *MasterPlan) IsFree() bool {
	return mp.Price == 0
}

// GetFormattedPrice format harga ke currency
func (mp *MasterPlan) GetFormattedPrice() string {
	// Format ke Rupiah
	// TODO: Implement proper currency formatting
	if mp.Price == 0 {
		return "Gratis"
	}
	return fmt.Sprintf("Rp %d", mp.Price)
}