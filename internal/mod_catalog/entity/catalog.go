package entity

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Catalog entity untuk tabel catalogs
type Catalog struct {
	ID         int64                  `json:"id" db:"c_id"`
	BusinessID int64                  `json:"business_id" db:"c_b_id"`
	ThemeID    int64                  `json:"theme_id" db:"c_mt_id"`
	Slug       string                 `json:"slug" db:"c_slug"`
	QRUrl      sql.NullString         `json:"qr_url" db:"c_qr_url"`
	Title      string                 `json:"title" db:"c_title"`
	Subtitle   sql.NullString         `json:"subtitle" db:"c_subtitle"`
	IsActive   bool                   `json:"is_active" db:"c_is_active"`
	Settings   map[string]interface{} `json:"settings" db:"c_settings"`
	CreatedBy  int64                  `json:"created_by" db:"c_created_by"`
	CreatedAt  time.Time              `json:"created_at" db:"c_created_at"`
	UpdatedBy  sql.NullInt64          `json:"updated_by" db:"c_updated_by"`
	UpdatedAt  *time.Time             `json:"updated_at" db:"c_updated_at"`

	// Relations
	Business *Business         `json:"business,omitempty"`
	Theme    *MasterTheme      `json:"theme,omitempty"`
	Sections []*CatalogSection `json:"sections,omitempty"`
}

// CatalogSection entity untuk tabel catalog_sections
type CatalogSection struct {
	ID        int64                  `json:"id" db:"cs_id"`
	CatalogID int64                  `json:"catalog_id" db:"cs_c_id"`
	Type      string                 `json:"type" db:"cs_type"`
	IsVisible bool                   `json:"is_visible" db:"cs_is_visible"`
	Config    map[string]interface{} `json:"config" db:"cs_config"`
	CreatedAt time.Time              `json:"created_at" db:"cs_created_at"`
	UpdatedAt *time.Time             `json:"updated_at" db:"cs_updated_at"`

	// Relations - based on type
	Cards        []*CatalogCard        `json:"cards,omitempty"`
	Carousels    []*CatalogCarousel    `json:"carousels,omitempty"`
	FAQs         []*CatalogFAQ         `json:"faqs,omitempty"`
	Links        []*CatalogLink        `json:"links,omitempty"`
	Socials      []*CatalogSocial      `json:"socials,omitempty"`
	Testimonials []*CatalogTestimonial `json:"testimonials,omitempty"`
}

// CatalogCard entity untuk tabel catalog_cards
type CatalogCard struct {
	ID         int64           `json:"id" db:"cc_id"`
	SectionID  int64           `json:"section_id" db:"cc_cs_id"`
	Title      string          `json:"title" db:"cc_title"`
	Subtitle   sql.NullString  `json:"subtitle" db:"cc_subtitle"`
	Type       string          `json:"type" db:"cc_type"`
	URL        sql.NullString  `json:"url" db:"cc_url"`
	IsVisible  bool            `json:"is_visible" db:"cc_is_visible"`
	HasDetail  bool            `json:"has_detail" db:"cc_has_detail"`
	Price      sql.NullInt64   `json:"price" db:"cc_price"`
	Discount   int             `json:"discount" db:"cc_discount"`
	Currency   string          `json:"currency" db:"cc_currency"`
	CreatedBy  int64           `json:"created_by" db:"cc_created_by"`
	CreatedAt  time.Time       `json:"created_at" db:"cc_created_at"`
	UpdatedBy  sql.NullInt64   `json:"updated_by" db:"cc_updated_by"`
	UpdatedAt  *time.Time      `json:"updated_at" db:"cc_updated_at"`

	// Relations
	Detail *CatalogCardDetail  `json:"detail,omitempty"`
	Media  []*CatalogCardMedia `json:"media,omitempty"`
}

// CatalogCardDetail entity untuk tabel catalog_card_details
type CatalogCardDetail struct {
	ID          int64          `json:"id" db:"ccd_id"`
	CardID      int64          `json:"card_id" db:"ccd_cc_id"`
	Slug        string         `json:"slug" db:"ccd_slug"`
	Description sql.NullString `json:"description" db:"ccd_description"`
	IsVisible   bool           `json:"is_visible" db:"ccd_is_visible"`
	CreatedBy   int64          `json:"created_by" db:"ccd_created_by"`
	CreatedAt   time.Time      `json:"created_at" db:"ccd_created_at"`
	UpdatedBy   sql.NullInt64  `json:"updated_by" db:"ccd_updated_by"`
	UpdatedAt   *time.Time     `json:"updated_at" db:"ccd_updated_at"`

	// Relations
	Links []*CatalogCardLink `json:"links,omitempty"`
}

// CatalogCardMedia entity untuk tabel catalog_card_media
type CatalogCardMedia struct {
	ID        int64         `json:"id" db:"ccm_id"`
	CardID    int64         `json:"card_id" db:"ccm_cc_id"`
	Type      string        `json:"type" db:"ccm_type"`
	URL       string        `json:"url" db:"ccm_url"`
	CreatedBy int64         `json:"created_by" db:"ccm_created_by"`
	CreatedAt time.Time     `json:"created_at" db:"ccm_created_at"`
	UpdatedBy sql.NullInt64 `json:"updated_by" db:"ccm_updated_by"`
	UpdatedAt *time.Time    `json:"updated_at" db:"ccm_updated_at"`
}

// CatalogCardLink entity untuk tabel catalog_card_links
type CatalogCardLink struct {
	ID        int64         `json:"id" db:"ccl_id"`
	DetailID  int64         `json:"detail_id" db:"ccl_ccd_id"`
	Type      string        `json:"type" db:"ccl_type"`
	URL       string        `json:"url" db:"ccl_url"`
	IsVisible bool          `json:"is_visible" db:"ccl_is_visible"`
	CreatedBy int64         `json:"created_by" db:"ccl_created_by"`
	CreatedAt time.Time     `json:"created_at" db:"ccl_created_at"`
	UpdatedBy sql.NullInt64 `json:"updated_by" db:"ccl_updated_by"`
	UpdatedAt *time.Time    `json:"updated_at" db:"ccl_updated_at"`
}

// CatalogCarousel entity untuk tabel catalog_carousels
type CatalogCarousel struct {
	ID        int64          `json:"id" db:"cr_id"`
	SectionID int64          `json:"section_id" db:"cr_cs_id"`
	Title     sql.NullString `json:"title" db:"cr_title"`
	IsVisible bool           `json:"is_visible" db:"cr_is_visible"`
	CreatedBy int64          `json:"created_by" db:"cr_created_by"`
	CreatedAt time.Time      `json:"created_at" db:"cr_created_at"`
	UpdatedBy sql.NullInt64  `json:"updated_by" db:"cr_updated_by"`
	UpdatedAt *time.Time     `json:"updated_at" db:"cr_updated_at"`

	// Relations
	Items []*CatalogCarouselItem `json:"items,omitempty"`
}

// CatalogCarouselItem entity untuk tabel catalog_carousel_items
type CatalogCarouselItem struct {
	ID          int64          `json:"id" db:"cci_id"`
	CarouselID  int64          `json:"carousel_id" db:"cci_cr_id"`
	ImageURL    string         `json:"image_url" db:"cci_image_url"`
	Caption     sql.NullString `json:"caption" db:"cci_caption"`
	Description sql.NullString `json:"description" db:"cci_description"`
	LinkURL     sql.NullString `json:"link_url" db:"cci_link_url"`
	CreatedBy   int64          `json:"created_by" db:"cci_created_by"`
	CreatedAt   time.Time      `json:"created_at" db:"cci_created_at"`
	UpdatedBy   sql.NullInt64  `json:"updated_by" db:"cci_updated_by"`
	UpdatedAt   *time.Time     `json:"updated_at" db:"cci_updated_at"`
}

// CatalogFAQ entity untuk tabel catalog_faqs
type CatalogFAQ struct {
	ID        int64         `json:"id" db:"cf_id"`
	SectionID int64         `json:"section_id" db:"cf_cs_id"`
	Question  string        `json:"question" db:"cf_question"`
	Answer    string        `json:"answer" db:"cf_answer"`
	IsVisible bool          `json:"is_visible" db:"cf_is_visible"`
	CreatedBy int64         `json:"created_by" db:"cf_created_by"`
	CreatedAt time.Time     `json:"created_at" db:"cf_created_at"`
	UpdatedBy sql.NullInt64 `json:"updated_by" db:"cf_updated_by"`
	UpdatedAt *time.Time    `json:"updated_at" db:"cf_updated_at"`
}

// CatalogLink entity untuk tabel catalog_links
type CatalogLink struct {
	ID          int64         `json:"id" db:"cl_id"`
	SectionID   int64         `json:"section_id" db:"cl_cs_id"`
	URL         string        `json:"url" db:"cl_url"`
	DisplayName string        `json:"display_name" db:"cl_display_name"`
	IsVisible   bool          `json:"is_visible" db:"cl_is_visible"`
	CreatedBy   int64         `json:"created_by" db:"cl_created_by"`
	CreatedAt   time.Time     `json:"created_at" db:"cl_created_at"`
	UpdatedBy   sql.NullInt64 `json:"updated_by" db:"cl_updated_by"`
	UpdatedAt   *time.Time    `json:"updated_at" db:"cl_updated_at"`
}

// CatalogSocial entity untuk tabel catalog_socials
type CatalogSocial struct {
	ID        int64         `json:"id" db:"csoc_id"`
	SectionID int64         `json:"section_id" db:"csoc_cs_id"`
	Platform  string        `json:"platform" db:"csoc_platform"`
	URL       string        `json:"url" db:"csoc_url"`
	IsVisible bool          `json:"is_visible" db:"csoc_is_visible"`
	CreatedBy int64         `json:"created_by" db:"csoc_created_by"`
	CreatedAt time.Time     `json:"created_at" db:"csoc_created_at"`
	UpdatedBy sql.NullInt64 `json:"updated_by" db:"csoc_updated_by"`
	UpdatedAt *time.Time    `json:"updated_at" db:"csoc_updated_at"`
}

// CatalogTestimonial entity untuk tabel catalog_testimonials
type CatalogTestimonial struct {
	ID        int64         `json:"id" db:"ct_id"`
	SectionID int64         `json:"section_id" db:"ct_cs_id"`
	Message   string        `json:"message" db:"ct_message"`
	Author    string        `json:"author" db:"ct_author"`
	IsVisible bool          `json:"is_visible" db:"ct_is_visible"`
	CreatedBy int64         `json:"created_by" db:"ct_created_by"`
	CreatedAt time.Time     `json:"created_at" db:"ct_created_at"`
	UpdatedBy sql.NullInt64 `json:"updated_by" db:"ct_updated_by"`
	UpdatedAt *time.Time    `json:"updated_at" db:"ct_updated_at"`
}

// Relations dari module lain
// type Business struct {
// 	ID   int64  `json:"id" db:"b_id"`
// 	Name string `json:"name" db:"b_name"`
// 	Slug string `json:"slug" db:"b_slug"`
// }

type Business struct {
	ID       int64  `json:"id" db:"b_id"`
	Name     string `json:"name" db:"b_name"`
	Slug     string `json:"slug" db:"b_slug"`
	Type     string `json:"type" db:"b_type"`
	IsActive bool   `json:"is_active" db:"b_is_active"`
}

type MasterTheme struct {
	ID   int64  `json:"id" db:"mt_id"`
	Name string `json:"name" db:"mt_name"`
	Type string `json:"type" db:"mt_type"`
}

// TableName methods
func (Catalog) TableName() string               { return "atamlink.catalogs" }
func (CatalogSection) TableName() string        { return "atamlink.catalog_sections" }
func (CatalogCard) TableName() string           { return "atamlink.catalog_cards" }
func (CatalogCardDetail) TableName() string     { return "atamlink.catalog_card_details" }
func (CatalogCardMedia) TableName() string      { return "atamlink.catalog_card_media" }
func (CatalogCardLink) TableName() string       { return "atamlink.catalog_card_links" }
func (CatalogCarousel) TableName() string       { return "atamlink.catalog_carousels" }
func (CatalogCarouselItem) TableName() string   { return "atamlink.catalog_carousel_items" }
func (CatalogFAQ) TableName() string            { return "atamlink.catalog_faqs" }
func (CatalogLink) TableName() string           { return "atamlink.catalog_links" }
func (CatalogSocial) TableName() string         { return "atamlink.catalog_socials" }
func (CatalogTestimonial) TableName() string    { return "atamlink.catalog_testimonials" }

// Helper methods

// GetSubtitle get subtitle dengan null handling
func (c *Catalog) GetSubtitle() string {
	if c.Subtitle.Valid {
		return c.Subtitle.String
	}
	return ""
}

// GetQRUrl get QR URL dengan null handling
func (c *Catalog) GetQRUrl() string {
	if c.QRUrl.Valid {
		return c.QRUrl.String
	}
	return ""
}

// GetDiscountedPrice menghitung harga setelah diskon
func (cc *CatalogCard) GetDiscountedPrice() int64 {
	if !cc.Price.Valid || cc.Discount <= 0 {
		return 0
	}

	discountAmount := (cc.Price.Int64 * int64(cc.Discount)) / 100
	return cc.Price.Int64 - discountAmount
}

// MarshalSettings marshal settings to JSON
func (c *Catalog) MarshalSettings() ([]byte, error) {
	return json.Marshal(c.Settings)
}

// UnmarshalSettings unmarshal settings from JSON
func (c *Catalog) UnmarshalSettings(data []byte) error {
	return json.Unmarshal(data, &c.Settings)
}