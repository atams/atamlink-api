package dto

import (
	"time"
)

// CreateCatalogRequest request untuk create catalog
type CreateCatalogRequest struct {
	BusinessID int64                  `json:"business_id" validate:"required,gt=0"`
	ThemeID    int64                  `json:"theme_id" validate:"required,gt=0"`
	Slug       string                 `json:"slug,omitempty" validate:"omitempty,slug,min=3,max=100"`
	Title      string                 `json:"title" validate:"required,min=3,max=200"`
	Subtitle   string                 `json:"subtitle,omitempty" validate:"max=300"`
	Settings   map[string]interface{} `json:"settings,omitempty"`
	Sections   []CreateSectionRequest `json:"sections,omitempty"`
}

// UpdateCatalogRequest request untuk update catalog
type UpdateCatalogRequest struct {
	ThemeID  int64                  `json:"theme_id,omitempty" validate:"omitempty,gt=0"`
	Title    string                 `json:"title,omitempty" validate:"omitempty,min=3,max=200"`
	Subtitle string                 `json:"subtitle,omitempty" validate:"max=300"`
	IsActive *bool                  `json:"is_active,omitempty"`
	Settings map[string]interface{} `json:"settings,omitempty"`
}

// CatalogResponse response untuk catalog
type CatalogResponse struct {
	ID         int64                  `json:"id"`
	BusinessID int64                  `json:"business_id"`
	ThemeID    int64                  `json:"theme_id"`
	Slug       string                 `json:"slug"`
	QRUrl      string                 `json:"qr_url,omitempty"`
	Title      string                 `json:"title"`
	Subtitle   string                 `json:"subtitle,omitempty"`
	IsActive   bool                   `json:"is_active"`
	Settings   map[string]interface{} `json:"settings"`
	CreatedBy  int64                  `json:"created_by"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  *time.Time             `json:"updated_at,omitempty"`
	PublicURL  string                 `json:"public_url"`
	Business   *BusinessResponse      `json:"business,omitempty"`
	Theme      *ThemeResponse         `json:"theme,omitempty"`
	Sections   []SectionResponse      `json:"sections,omitempty"`
}

// CatalogListResponse response untuk list catalogs
type CatalogListResponse struct {
	ID           int64      `json:"id"`
	BusinessID   int64      `json:"business_id"`
	BusinessName string     `json:"business_name"`
	Slug         string     `json:"slug"`
	Title        string     `json:"title"`
	Subtitle     string     `json:"subtitle,omitempty"`
	IsActive     bool       `json:"is_active"`
	ThemeName    string     `json:"theme_name"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
	PublicURL    string     `json:"public_url"`
}

// BusinessResponse simple business response
type BusinessResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// ThemeResponse simple theme response
type ThemeResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// CreateSectionRequest request untuk create section
type CreateSectionRequest struct {
	Type      string                 `json:"type" validate:"required,oneof=hero cards carousel faqs links socials testimonials cta text video"`
	IsVisible bool                   `json:"is_visible"`
	Config    map[string]interface{} `json:"config,omitempty"`
	Content   interface{}            `json:"content,omitempty"` // Specific content based on type
}

// UpdateSectionRequest request untuk update section
type UpdateSectionRequest struct {
	Type      string                 `json:"type,omitempty" validate:"omitempty,oneof=hero cards carousel faqs links socials testimonials cta text video"`
	IsVisible *bool                  `json:"is_visible,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`
}

// SectionResponse response untuk section
type SectionResponse struct {
	ID        int64                  `json:"id"`
	CatalogID int64                  `json:"catalog_id"`
	Type      string                 `json:"type"`
	IsVisible bool                   `json:"is_visible"`
	Config    map[string]interface{} `json:"config"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt *time.Time             `json:"updated_at,omitempty"`
	Content   interface{}            `json:"content,omitempty"` // Based on type
}

// CreateCardRequest request untuk create card
type CreateCardRequest struct {
	Title     string   `json:"title" validate:"required,min=1,max=200"`
	Subtitle  string   `json:"subtitle,omitempty" validate:"max=300"`
	Type      string   `json:"type" validate:"required,oneof=product service portfolio article event offer"`
	URL       string   `json:"url,omitempty" validate:"omitempty,url"`
	IsVisible bool     `json:"is_visible"`
	HasDetail bool     `json:"has_detail"`
	Price     int64    `json:"price,omitempty" validate:"omitempty,gte=0"`
	Discount  int      `json:"discount,omitempty" validate:"omitempty,gte=0,lte=100"`
	Currency  string   `json:"currency,omitempty" validate:"omitempty,oneof=IDR"`
	Detail    *CardDetailRequest `json:"detail,omitempty"`
	MediaURLs []string `json:"media_urls,omitempty"`
}

// UpdateCardRequest request untuk update card
type UpdateCardRequest struct {
	Title     string   `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Subtitle  string   `json:"subtitle,omitempty" validate:"max=300"`
	Type      string   `json:"type,omitempty" validate:"omitempty,oneof=product service portfolio article event offer"`
	URL       string   `json:"url,omitempty" validate:"omitempty,url"`
	IsVisible *bool    `json:"is_visible,omitempty"`
	HasDetail *bool    `json:"has_detail,omitempty"`
	Price     *int64   `json:"price,omitempty" validate:"omitempty,gte=0"`
	Discount  *int     `json:"discount,omitempty" validate:"omitempty,gte=0,lte=100"`
	Currency  string   `json:"currency,omitempty" validate:"omitempty,oneof=IDR"`
}

// CardResponse response untuk card
type CardResponse struct {
	ID              int64              `json:"id"`
	SectionID       int64              `json:"section_id"`
	Title           string             `json:"title"`
	Subtitle        string             `json:"subtitle,omitempty"`
	Type            string             `json:"type"`
	URL             string             `json:"url,omitempty"`
	IsVisible       bool               `json:"is_visible"`
	HasDetail       bool               `json:"has_detail"`
	Price           int64              `json:"price,omitempty"`
	Discount        int                `json:"discount,omitempty"`
	DiscountedPrice int64              `json:"discounted_price,omitempty"`
	Currency        string             `json:"currency,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       *time.Time         `json:"updated_at,omitempty"`
	Detail          *CardDetailResponse `json:"detail,omitempty"`
	Media           []MediaResponse     `json:"media,omitempty"`
}

// CardDetailRequest request untuk card detail
type CardDetailRequest struct {
	Slug        string      `json:"slug,omitempty" validate:"omitempty,slug"`
	Description string      `json:"description,omitempty"`
	IsVisible   bool        `json:"is_visible"`
	Links       []LinkRequest `json:"links,omitempty"`
}

// CardDetailResponse response untuk card detail
type CardDetailResponse struct {
	ID          int64           `json:"id"`
	CardID      int64           `json:"card_id"`
	Slug        string          `json:"slug"`
	Description string          `json:"description,omitempty"`
	IsVisible   bool            `json:"is_visible"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   *time.Time      `json:"updated_at,omitempty"`
	Links       []LinkResponse  `json:"links,omitempty"`
}

// LinkRequest request untuk links
type LinkRequest struct {
	Type      string `json:"type" validate:"required,oneof=whatsapp shopee tokopedia website tiktokshop facebook instagram telegram email phone custom"`
	URL       string `json:"url" validate:"required,max=500"`
	IsVisible bool   `json:"is_visible"`
}

// LinkResponse response untuk links
type LinkResponse struct {
	ID        int64      `json:"id"`
	Type      string     `json:"type"`
	URL       string     `json:"url"`
	IsVisible bool       `json:"is_visible"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// MediaResponse response untuk media
type MediaResponse struct {
	ID        int64      `json:"id"`
	Type      string     `json:"type"`
	URL       string     `json:"url"`
	CreatedAt time.Time  `json:"created_at"`
}

// CreateCarouselRequest request untuk create carousel
type CreateCarouselRequest struct {
	Title     string                    `json:"title,omitempty" validate:"max=200"`
	IsVisible bool                      `json:"is_visible"`
	Items     []CreateCarouselItemRequest `json:"items"`
}

// CreateCarouselItemRequest request untuk carousel item
type CreateCarouselItemRequest struct {
	ImageURL    string `json:"image_url" validate:"required,max=500"`
	Caption     string `json:"caption,omitempty" validate:"max=200"`
	Description string `json:"description,omitempty"`
	LinkURL     string `json:"link_url,omitempty" validate:"omitempty,url"`
}

// FAQRequest request untuk FAQ
type FAQRequest struct {
	Question  string `json:"question" validate:"required"`
	Answer    string `json:"answer" validate:"required"`
	IsVisible bool   `json:"is_visible"`
}

// SocialRequest request untuk social link
type SocialRequest struct {
	Platform  string `json:"platform" validate:"required,oneof=facebook instagram twitter linkedin youtube tiktok whatsapp telegram pinterest github"`
	URL       string `json:"url" validate:"required,url,max=500"`
	IsVisible bool   `json:"is_visible"`
}

// TestimonialRequest request untuk testimonial
type TestimonialRequest struct {
	Message   string `json:"message" validate:"required"`
	Author    string `json:"author" validate:"required,max=200"`
	IsVisible bool   `json:"is_visible"`
}

// CatalogFilter filter untuk query catalogs
type CatalogFilter struct {
	Search     string     `json:"search,omitempty"`
	BusinessID int64      `json:"business_id,omitempty"`
	ThemeID    int64      `json:"theme_id,omitempty"`
	IsActive   *bool      `json:"is_active,omitempty"`
	CreatedFrom *time.Time `json:"created_from,omitempty"`
	CreatedTo   *time.Time `json:"created_to,omitempty"`
}

// PublicCatalogResponse response untuk public catalog view
type PublicCatalogResponse struct {
	ID         int64                  `json:"id"`
	Slug       string                 `json:"slug"`
	Title      string                 `json:"title"`
	Subtitle   string                 `json:"subtitle,omitempty"`
	Settings   map[string]interface{} `json:"settings"`
	Business   PublicBusinessInfo     `json:"business"`
	Theme      ThemeResponse          `json:"theme"`
	Sections   []PublicSectionResponse `json:"sections"`
}

// PublicBusinessInfo public business info
type PublicBusinessInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// PublicSectionResponse public section response
type PublicSectionResponse struct {
	Type      string                 `json:"type"`
	Config    map[string]interface{} `json:"config"`
	Content   interface{}            `json:"content"`
}