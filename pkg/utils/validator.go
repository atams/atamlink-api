package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/atam/atamlink/internal/constant"
)

// Validator wrapper untuk go-playground/validator
type Validator struct {
	validator *validator.Validate
}

// ValidationErr struktur untuk error validasi
type ValidationErr struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// NewValidator membuat validator instance baru
func NewValidator() *Validator {
	v := validator.New()

	// Register custom tag name function
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validators
	registerCustomValidators(v)

	// Set validator untuk Gin binding
	binding.Validator = &defaultValidator{validator: v}

	return &Validator{validator: v}
}

// Validate melakukan validasi struct
func (v *Validator) Validate(data interface{}) []ValidationErr {
	var errors []ValidationErr

	err := v.validator.Struct(data)
	if err == nil {
		return errors
	}

	// Type assert ke validator.ValidationErrors
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		errors = append(errors, ValidationErr{
			Field:   "general",
			Message: "Validasi gagal",
		})
		return errors
	}

	// Convert ke format yang lebih user friendly
	for _, e := range validationErrors {
		errors = append(errors, ValidationErr{
			Field:   e.Field(),
			Message: getErrorMessage(e),
		})
	}

	return errors
}

// ValidateVar validasi single variable
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validator.Var(field, tag)
}

// defaultValidator untuk Gin binding
type defaultValidator struct {
	validator *validator.Validate
}

// ValidateStruct implements binding.StructValidator
func (v *defaultValidator) ValidateStruct(obj interface{}) error {
	return v.validator.Struct(obj)
}

// Engine returns the underlying validator engine
func (v *defaultValidator) Engine() interface{} {
	return v.validator
}

// registerCustomValidators register custom validation rules
func registerCustomValidators(v *validator.Validate) {
	// Slug validator
	v.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		slug := fl.Field().String()
		if slug == "" {
			return true
		}
		// Slug harus lowercase, alphanumeric, dan dash
		for _, c := range slug {
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
				return false
			}
		}
		// Tidak boleh mulai atau akhir dengan dash
		return slug[0] != '-' && slug[len(slug)-1] != '-'
	})

	// Phone validator untuk Indonesia
	v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		if phone == "" {
			return true
		}
		// Format: +62xxx atau 08xxx
		if strings.HasPrefix(phone, "+62") {
			return len(phone) >= 10 && len(phone) <= 15
		}
		if strings.HasPrefix(phone, "08") {
			return len(phone) >= 10 && len(phone) <= 13
		}
		return false
	})

	// Username validator
	v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		username := fl.Field().String()
		if username == "" {
			return true
		}
		// Username harus alphanumeric dan underscore
		for _, c := range username {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
				(c >= '0' && c <= '9') || c == '_') {
				return false
			}
		}
		return true
	})

	// No spaces validator
	v.RegisterValidation("nospaces", func(fl validator.FieldLevel) bool {
		return !strings.Contains(fl.Field().String(), " ")
	})

	// Business_role validator 🐞
	v.RegisterValidation("business_role", func(fl validator.FieldLevel) bool {
		role := fl.Field().String()
		return constant.IsValidRole(role) // Use the existing helper function
	})
}

// getErrorMessage mendapatkan pesan error yang user friendly
func getErrorMessage(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()
	param := e.Param()

	// Custom messages untuk tags tertentu
	messages := map[string]string{
		"required":  fmt.Sprintf("%s wajib diisi", field),
		"email":     fmt.Sprintf("%s harus berupa email yang valid", field),
		"min":       fmt.Sprintf("%s minimal %s karakter", field, param),
		"max":       fmt.Sprintf("%s maksimal %s karakter", field, param),
		"len":       fmt.Sprintf("%s harus %s karakter", field, param),
		"gte":       fmt.Sprintf("%s harus lebih besar atau sama dengan %s", field, param),
		"lte":       fmt.Sprintf("%s harus lebih kecil atau sama dengan %s", field, param),
		"gt":        fmt.Sprintf("%s harus lebih besar dari %s", field, param),
		"lt":        fmt.Sprintf("%s harus lebih kecil dari %s", field, param),
		"alpha":     fmt.Sprintf("%s hanya boleh berisi huruf", field),
		"alphanum":  fmt.Sprintf("%s hanya boleh berisi huruf dan angka", field),
		"numeric":   fmt.Sprintf("%s harus berupa angka", field),
		"slug":      fmt.Sprintf("%s harus berupa slug yang valid (lowercase, alphanumeric, dash)", field),
		"phone":     fmt.Sprintf("%s harus berupa nomor telepon yang valid", field),
		"username":  fmt.Sprintf("%s hanya boleh berisi huruf, angka, dan underscore", field),
		"nospaces":  fmt.Sprintf("%s tidak boleh mengandung spasi", field),
		"url":       fmt.Sprintf("%s harus berupa URL yang valid", field),
		"uuid":      fmt.Sprintf("%s harus berupa UUID yang valid", field),
		"oneof":     fmt.Sprintf("%s harus salah satu dari: %s", field, param),
	}

	if msg, exists := messages[tag]; exists {
		return msg
	}

	// Default message
	return fmt.Sprintf("%s tidak valid", field)
}

// Common validation rules yang bisa digunakan

// IsEmail check apakah string adalah email valid
func IsEmail(email string) bool {
	v := validator.New()
	err := v.Var(email, "required,email")
	return err == nil
}

// IsURL check apakah string adalah URL valid
func IsURL(url string) bool {
	v := validator.New()
	err := v.Var(url, "required,url")
	return err == nil
}

// IsUUID check apakah string adalah UUID valid
func IsUUID(uuid string) bool {
	v := validator.New()
	err := v.Var(uuid, "required,uuid")
	return err == nil
}

// IsSlug check apakah string adalah slug valid
func IsSlug(slug string) bool {
	v := validator.New()
	registerCustomValidators(v)
	err := v.Var(slug, "required,slug")
	return err == nil
}

// IsPhone check apakah string adalah phone number valid
func IsPhone(phone string) bool {
	v := validator.New()
	registerCustomValidators(v)
	err := v.Var(phone, "required,phone")
	return err == nil
}