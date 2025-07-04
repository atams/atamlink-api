package dto

import (
	"time"
)

// CreatePlanRequest request untuk create plan
type CreatePlanRequest struct {
	Name     string                 `json:"name" validate:"required,min=3,max=100"`
	Price    int                    `json:"price" validate:"required,gte=0"`
	Duration string                 `json:"duration" validate:"required"` // PostgreSQL interval format
	Features map[string]interface{} `json:"features" validate:"required"`
	IsActive bool                   `json:"is_active"`
}

// UpdatePlanRequest request untuk update plan
type UpdatePlanRequest struct {
	Name     string                 `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Price    *int                   `json:"price,omitempty" validate:"omitempty,gte=0"`
	Duration string                 `json:"duration,omitempty"`
	Features map[string]interface{} `json:"features,omitempty"`
	IsActive *bool                  `json:"is_active,omitempty"`
}

// PlanResponse response untuk plan
type PlanResponse struct {
	ID           int64                  `json:"id"`
	Name         string                 `json:"name"`
	Price        int                    `json:"price"`
	Duration     string                 `json:"duration"`
	DurationDays int                    `json:"duration_days"`
	Features     map[string]interface{} `json:"features"`
	IsActive     bool                   `json:"is_active"`
	CreatedAt    time.Time              `json:"created_at"`
}

// PlanListResponse response untuk list plans
type PlanListResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Price        int       `json:"price"`
	PriceDisplay string    `json:"price_display"`
	Duration     string    `json:"duration"`
	DurationDays int       `json:"duration_days"`
	IsActive     bool      `json:"is_active"`
	IsFree       bool      `json:"is_free"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateThemeRequest request untuk create theme
type CreateThemeRequest struct {
	Name            string                 `json:"name" validate:"required,min=3,max=100"`
	Description     string                 `json:"description,omitempty"`
	Type            string                 `json:"type" validate:"required,oneof=minimal modern classic bold elegant playful professional creative"`
	DefaultSettings map[string]interface{} `json:"default_settings" validate:"required"`
	IsPremium       bool                   `json:"is_premium"`
	IsActive        bool                   `json:"is_active"`
}

// UpdateThemeRequest request untuk update theme
type UpdateThemeRequest struct {
	Name            string                 `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description     string                 `json:"description,omitempty"`
	Type            string                 `json:"type,omitempty" validate:"omitempty,oneof=minimal modern classic bold elegant playful professional creative"`
	DefaultSettings map[string]interface{} `json:"default_settings,omitempty"`
	IsPremium       *bool                  `json:"is_premium,omitempty"`
	IsActive        *bool                  `json:"is_active,omitempty"`
}

// ThemeResponse response untuk theme
type ThemeResponse struct {
	ID              int64                  `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	Type            string                 `json:"type"`
	DefaultSettings map[string]interface{} `json:"default_settings"`
	IsPremium       bool                   `json:"is_premium"`
	IsActive        bool                   `json:"is_active"`
	CreatedAt       time.Time              `json:"created_at"`
	PreviewURL      string                 `json:"preview_url,omitempty"`
}

// ThemeListResponse response untuk list themes
type ThemeListResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Type        string    `json:"type"`
	IsPremium   bool      `json:"is_premium"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	PreviewURL  string    `json:"preview_url,omitempty"`
}

// PlanFilter filter untuk query plans
type PlanFilter struct {
	IsActive  *bool  `json:"is_active,omitempty"`
	IsFree    *bool  `json:"is_free,omitempty"`
	MaxPrice  *int   `json:"max_price,omitempty"`
	MinPrice  *int   `json:"min_price,omitempty"`
}

// ThemeFilter filter untuk query themes
type ThemeFilter struct {
	Search    string `json:"search,omitempty"`
	Type      string `json:"type,omitempty"`
	IsActive  *bool  `json:"is_active,omitempty"`
	IsPremium *bool  `json:"is_premium,omitempty"`
}

// PlanFeatures struktur standard features plan
type PlanFeatures struct {
	MaxCatalogs      int  `json:"max_catalogs"`
	MaxProducts      int  `json:"max_products"`
	MaxUsers         int  `json:"max_users"`
	MaxStorage       int  `json:"max_storage"` // in MB
	CustomDomain     bool `json:"custom_domain"`
	Analytics        bool `json:"analytics"`
	PrioritySupport  bool `json:"priority_support"`
	RemoveWatermark  bool `json:"remove_watermark"`
	AdvancedThemes   bool `json:"advanced_themes"`
	APIAccess        bool `json:"api_access"`
	ExportData       bool `json:"export_data"`
	CustomBranding   bool `json:"custom_branding"`
}

// ThemeSettings struktur standard theme settings
type ThemeSettings struct {
	Colors      ThemeColors      `json:"colors"`
	Typography  ThemeTypography  `json:"typography"`
	Layout      ThemeLayout      `json:"layout"`
	Components  map[string]interface{} `json:"components"`
	CustomCSS   string          `json:"custom_css,omitempty"`
}

// ThemeColors color settings
type ThemeColors struct {
	Primary     string `json:"primary"`
	Secondary   string `json:"secondary"`
	Background  string `json:"background"`
	Surface     string `json:"surface"`
	Text        string `json:"text"`
	TextMuted   string `json:"text_muted"`
	Border      string `json:"border"`
	Error       string `json:"error"`
	Warning     string `json:"warning"`
	Success     string `json:"success"`
	Info        string `json:"info"`
}

// ThemeTypography typography settings
type ThemeTypography struct {
	FontFamily    string `json:"font_family"`
	FontSizeBase  string `json:"font_size_base"`
	FontWeightNormal int `json:"font_weight_normal"`
	FontWeightBold   int `json:"font_weight_bold"`
	LineHeight    string `json:"line_height"`
	HeadingFont   string `json:"heading_font,omitempty"`
}

// ThemeLayout layout settings
type ThemeLayout struct {
	ContainerWidth string `json:"container_width"`
	Spacing        string `json:"spacing"`
	BorderRadius   string `json:"border_radius"`
	BoxShadow      string `json:"box_shadow"`
}