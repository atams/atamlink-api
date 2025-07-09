package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/atam/atamlink/pkg/utils"
)

const GinKeyUserID = "user_id"

// AuthUser data user yang sudah terautentikasi
type AuthUser struct {
	UserID    string `json:"user_id"`
	ProfileID int64  `json:"profile_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
}

// Auth middleware untuk autentikasi (placeholder untuk integrasi dengan auth service)
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Abort(c, 401, "Token tidak ditemukan")
			return
		}

		// Check Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Abort(c, 401, "Format token tidak valid")
			return
		}

		token := parts[1]
		if token == "" {
			utils.Abort(c, 401, "Token tidak boleh kosong")
			return
		}

		// TODO: Validate token dengan auth service
		// Untuk sekarang, return unauthorized
		utils.Abort(c, 401, "Silakan gunakan AUTH_BYPASS=true untuk development")
	}
}

// AuthBypass middleware untuk bypass auth di development
func AuthBypass(userID string, profileID int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set dummy auth user
		authUser := AuthUser{
			UserID:    userID,
			ProfileID: profileID,
			Email:     "dev@atamlink.com",
			Username:  "developer",
		}

		// Set to context
		c.Set("auth_user", authUser)
		c.Set("user_id", userID)
		c.Set("profile_id", profileID)

		c.Next()
	}
}

// GetAuthUser mendapatkan auth user dari context
func GetAuthUser(c *gin.Context) (*AuthUser, bool) {
	user, exists := c.Get("auth_user")
	if !exists {
		return nil, false
	}

	authUser, ok := user.(AuthUser)
	if !ok {
		return nil, false
	}

	return &authUser, true
}

// GetUserID mendapatkan user ID dari context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(GinKeyUserID)
	if !exists {
		return "", false
	}

	id, ok := userID.(string)
	return id, ok
}

// GetProfileID mendapatkan profile ID dari context
func GetProfileID(c *gin.Context) (int64, bool) {
	profileID, exists := c.Get("profile_id")
	if !exists {
		return 0, false
	}

	id, ok := profileID.(int64)
	return id, ok
}

// RequireRole middleware untuk check role (akan diimplementasi nanti)
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement role checking
		// Untuk sekarang, allow all
		c.Next()
	}
}

// RequireBusinessAccess middleware untuk check akses ke business
func RequireBusinessAccess(minRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement business access checking
		// Check apakah user punya akses ke business_id dari param
		// dengan role minimal sesuai parameter
		c.Next()
	}
}