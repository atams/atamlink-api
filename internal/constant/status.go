package constant

// Business types
const (
	BusinessTypeRetail        = "retail"
	BusinessTypeService       = "service"
	BusinessTypeManufacturing = "manufacturing"
	BusinessTypeTechnology    = "technology"
	BusinessTypeHospitality   = "hospitality"
	BusinessTypeHealthcare    = "healthcare"
	BusinessTypeEducation     = "education"
	BusinessTypeOther         = "other"
)

// Subscription status
const (
	SubscriptionStatusActive    = "active"
	SubscriptionStatusExpired   = "expired"
	SubscriptionStatusCancelled = "cancelled"
	SubscriptionStatusPending   = "pending"
	SubscriptionStatusSuspended = "suspended"
)

// Section types
const (
	SectionTypeHero         = "hero"
	SectionTypeCards        = "cards"
	SectionTypeCarousel     = "carousel"
	SectionTypeFAQs         = "faqs"
	SectionTypeLinks        = "links"
	SectionTypeSocials      = "socials"
	SectionTypeTestimonials = "testimonials"
	SectionTypeCTA          = "cta"
	SectionTypeText         = "text"
	SectionTypeVideo        = "video"
)

// Card types
const (
	CardTypeProduct   = "product"
	CardTypeService   = "service"
	CardTypePortfolio = "portfolio"
	CardTypeArticle   = "article"
	CardTypeEvent     = "event"
	CardTypeOffer     = "offer"
)

// Link types
const (
	LinkTypeWhatsApp   = "whatsapp"
	LinkTypeShopee     = "shopee"
	LinkTypeTokopedia  = "tokopedia"
	LinkTypeWebsite    = "website"
	LinkTypeTiktokShop = "tiktokshop"
	LinkTypeFacebook   = "facebook"
	LinkTypeInstagram  = "instagram"
	LinkTypeTelegram   = "telegram"
	LinkTypeEmail      = "email"
	LinkTypePhone      = "phone"
	LinkTypeCustom     = "custom"
)

// Media types
const (
	MediaTypeThumbnail = "thumbnail"
	MediaTypeCover     = "cover"
	MediaTypeGallery   = "gallery"
	MediaTypeVideo     = "video"
	MediaTypeDocument  = "document"
)

// Social platforms
const (
	SocialFacebook  = "facebook"
	SocialInstagram = "instagram"
	SocialTwitter   = "twitter"
	SocialLinkedIn  = "linkedin"
	SocialYouTube   = "youtube"
	SocialTikTok    = "tiktok"
	SocialWhatsApp  = "whatsapp"
	SocialTelegram  = "telegram"
	SocialPinterest = "pinterest"
	SocialGitHub    = "github"
)

// Theme types
const (
	ThemeMinimal      = "minimal"
	ThemeModern       = "modern"
	ThemeClassic      = "classic"
	ThemeBold         = "bold"
	ThemeElegant      = "elegant"
	ThemePlayful      = "playful"
	ThemeProfessional = "professional"
	ThemeCreative     = "creative"
)

// Currency types
const (
	CurrencyIDR = "IDR"
)

// Default values
const (
	DefaultPageSize     = 20
	MaxPageSize         = 100
	DefaultCacheExpiry  = 300 // 5 minutes
	DefaultSlugLength   = 8
	MaxSlugRetries      = 5
	DefaultTokenExpiry  = 86400 // 24 hours
)

// Validation functions

// IsValidBusinessType check apakah business type valid
func IsValidBusinessType(t string) bool {
	validTypes := []string{
		BusinessTypeRetail, BusinessTypeService, BusinessTypeManufacturing,
		BusinessTypeTechnology, BusinessTypeHospitality, BusinessTypeHealthcare,
		BusinessTypeEducation, BusinessTypeOther,
	}
	return contains(validTypes, t)
}

// IsValidSubscriptionStatus check apakah subscription status valid
func IsValidSubscriptionStatus(s string) bool {
	validStatuses := []string{
		SubscriptionStatusActive, SubscriptionStatusExpired,
		SubscriptionStatusCancelled, SubscriptionStatusPending,
		SubscriptionStatusSuspended,
	}
	return contains(validStatuses, s)
}

// IsValidSectionType check apakah section type valid
func IsValidSectionType(t string) bool {
	validTypes := []string{
		SectionTypeHero, SectionTypeCards, SectionTypeCarousel,
		SectionTypeFAQs, SectionTypeLinks, SectionTypeSocials,
		SectionTypeTestimonials, SectionTypeCTA, SectionTypeText,
		SectionTypeVideo,
	}
	return contains(validTypes, t)
}

// IsValidCardType check apakah card type valid
func IsValidCardType(t string) bool {
	validTypes := []string{
		CardTypeProduct, CardTypeService, CardTypePortfolio,
		CardTypeArticle, CardTypeEvent, CardTypeOffer,
	}
	return contains(validTypes, t)
}

// IsValidLinkType check apakah link type valid
func IsValidLinkType(t string) bool {
	validTypes := []string{
		LinkTypeWhatsApp, LinkTypeShopee, LinkTypeTokopedia,
		LinkTypeWebsite, LinkTypeTiktokShop, LinkTypeFacebook,
		LinkTypeInstagram, LinkTypeTelegram, LinkTypeEmail,
		LinkTypePhone, LinkTypeCustom,
	}
	return contains(validTypes, t)
}

// IsValidMediaType check apakah media type valid
func IsValidMediaType(t string) bool {
	validTypes := []string{
		MediaTypeThumbnail, MediaTypeCover, MediaTypeGallery,
		MediaTypeVideo, MediaTypeDocument,
	}
	return contains(validTypes, t)
}

// IsValidSocialPlatform check apakah social platform valid
func IsValidSocialPlatform(p string) bool {
	validPlatforms := []string{
		SocialFacebook, SocialInstagram, SocialTwitter,
		SocialLinkedIn, SocialYouTube, SocialTikTok,
		SocialWhatsApp, SocialTelegram, SocialPinterest,
		SocialGitHub,
	}
	return contains(validPlatforms, p)
}

// IsValidThemeType check apakah theme type valid
func IsValidThemeType(t string) bool {
	validTypes := []string{
		ThemeMinimal, ThemeModern, ThemeClassic,
		ThemeBold, ThemeElegant, ThemePlayful,
		ThemeProfessional, ThemeCreative,
	}
	return contains(validTypes, t)
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}