package service

import (
	"crypto/rand"
	"fmt"
	"strings"
	"unicode"

	"github.com/atam/atamlink/internal/constant"
)

// SlugService service untuk generate dan validasi slug
type SlugService interface {
	Generate(text string) string
	GenerateUnique(text string, length int) string
	GenerateRandom(length int) string
	Normalize(text string) string
	IsValid(slug string) bool
}

type slugService struct{}

// NewSlugService membuat instance slug service baru
func NewSlugService() SlugService {
	return &slugService{}
}

// Generate membuat slug dari text
func (s *slugService) Generate(text string) string {
	// Normalize text
	slug := s.Normalize(text)
	
	// Jika kosong, generate random
	if slug == "" {
		return s.GenerateRandom(constant.DefaultSlugLength)
	}
	
	return slug
}

// GenerateUnique membuat slug unique dengan suffix random
func (s *slugService) GenerateUnique(text string, length int) string {
	baseSlug := s.Normalize(text)
	
	// Jika base slug kosong, langsung random
	if baseSlug == "" {
		return s.GenerateRandom(length)
	}
	
	// Tambahkan random suffix
	suffix := s.GenerateRandom(4)
	slug := fmt.Sprintf("%s-%s", baseSlug, suffix)
	
	// Trim jika terlalu panjang
	if len(slug) > length {
		// Keep base slug dan adjust suffix
		maxBase := length - 5 // 4 untuk suffix + 1 untuk dash
		if maxBase < 3 {
			return s.GenerateRandom(length)
		}
		baseSlug = baseSlug[:maxBase]
		slug = fmt.Sprintf("%s-%s", baseSlug, suffix)
	}
	
	return slug
}

// GenerateRandom generate random slug
func (s *slugService) GenerateRandom(length int) string {
	if length <= 0 {
		length = constant.DefaultSlugLength
	}
	
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	bytes := make([]byte, length)
	
	// Use crypto/rand for better randomness
	if _, err := rand.Read(bytes); err != nil {
		// Fallback ke simple random
		return s.simpleRandom(length)
	}
	
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	
	// Ensure first character is letter
	if bytes[0] >= '0' && bytes[0] <= '9' {
		bytes[0] = charset[int(bytes[0])%26] // Force to letter
	}
	
	return string(bytes)
}

// Normalize normalize text menjadi slug format
func (s *slugService) Normalize(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)
	
	// Replace spaces and special chars with dash
	var result []rune
	lastWasDash := false
	
	for _, r := range text {
		// Keep alphanumeric
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			result = append(result, r)
			lastWasDash = false
		} else if !lastWasDash && len(result) > 0 {
			// Replace non-alphanumeric with dash
			result = append(result, '-')
			lastWasDash = true
		}
	}
	
	// Trim trailing dash
	slug := strings.Trim(string(result), "-")
	
	// Remove consecutive dashes
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	
	return slug
}

// IsValid check apakah slug valid
func (s *slugService) IsValid(slug string) bool {
	if slug == "" {
		return false
	}
	
	// Check length
	if len(slug) < 3 || len(slug) > 100 {
		return false
	}
	
	// Check format: lowercase alphanumeric dan dash
	for i, r := range slug {
		// First and last char tidak boleh dash
		if (i == 0 || i == len(slug)-1) && r == '-' {
			return false
		}
		
		// Hanya allow lowercase letter, number, dash
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}
	
	// Check no consecutive dashes
	if strings.Contains(slug, "--") {
		return false
	}
	
	return true
}

// simpleRandom fallback random generator
func (s *slugService) simpleRandom(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, length)
	
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	
	return string(result)
}

// SlugValidator untuk validasi slug di repository
type SlugValidator func(slug string) (bool, error)

// GenerateUniqueSlug generate slug yang belum digunakan
func GenerateUniqueSlug(
	baseText string,
	slugService SlugService,
	validator SlugValidator,
	maxRetries int,
) (string, error) {
	if maxRetries <= 0 {
		maxRetries = constant.MaxSlugRetries
	}
	
	// Try base slug first
	slug := slugService.Generate(baseText)
	if slug != "" {
		exists, err := validator(slug)
		if err != nil {
			return "", fmt.Errorf("failed to validate slug: %w", err)
		}
		if !exists {
			return slug, nil
		}
	}
	
	// Try with unique suffix
	for i := 0; i < maxRetries; i++ {
		slug = slugService.GenerateUnique(baseText, 50)
		
		exists, err := validator(slug)
		if err != nil {
			return "", fmt.Errorf("failed to validate slug: %w", err)
		}
		
		if !exists {
			return slug, nil
		}
	}
	
	return "", fmt.Errorf("failed to generate unique slug after %d attempts", maxRetries)
}