package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config menyimpan semua konfigurasi aplikasi
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Log      LogConfig
	CORS     CORSConfig
	Upload   UploadConfig
	UploadThing UploadThingConfig
	API      APIConfig
	Auth     AuthConfig
}

// ServerConfig konfigurasi server HTTP
type ServerConfig struct {
	Port         string
	Mode         string        // debug, release, test
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DatabaseConfig konfigurasi koneksi database
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// LogConfig konfigurasi logging
type LogConfig struct {
	Level  string // debug, info, warn, error, fatal
	Format string // json, text
}

// CORSConfig konfigurasi CORS
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

// UploadConfig konfigurasi upload file
type UploadConfig struct {
	MaxSize      int64
	AllowedTypes []string
	Path         string
}

// UploadThingConfig konfigurasi upload file
type UploadThingConfig struct {
	Secret  string
	AppID   string
	MaxSize int64
}

// APIConfig konfigurasi API
type APIConfig struct {
	Prefix  string
	Timeout time.Duration
}

// AuthConfig konfigurasi autentikasi
type AuthConfig struct {
	Bypass          bool
	BypassUserID    string
	BypassProfileID int64
}

// Load membaca konfigurasi dari environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", ""),
			Mode:         getEnv("SERVER_MODE", ""),
			ReadTimeout:  getDuration("SERVER_READ_TIMEOUT", ""),
			WriteTimeout: getDuration("SERVER_WRITE_TIMEOUT", ""),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", ""),
			Port:            getEnv("DB_PORT", ""),
			User:            getEnv("DB_USER", ""),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", ""),
			SSLMode:         getEnv("DB_SSLMODE", ""),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 0),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 0),
			ConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME", ""),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", ""),
			Format: getEnv("LOG_FORMAT", ""),
		},
		CORS: CORSConfig{
			AllowedOrigins:   getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			AllowedMethods:   getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders:   getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization"}),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
		},
		Upload: UploadConfig{
			MaxSize:      getEnvAsInt64("UPLOAD_MAX_SIZE", 0), // 10MB
			AllowedTypes: getEnvAsSlice("UPLOAD_ALLOWED_TYPES", []string{"image/jpeg", "image/png", "image/gif", "image/webp"}),
			Path:         getEnv("UPLOAD_PATH", ""),
		},
		UploadThing: UploadThingConfig{
			Secret:  getEnv("UPLOADTHING_SECRET", ""),
			AppID:   getEnv("UPLOADTHING_APP_ID", ""),
		},
		API: APIConfig{
			Prefix:  getEnv("API_PREFIX", ""),
			Timeout: getDuration("API_TIMEOUT", ""),
		},
		Auth: AuthConfig{
			Bypass:          getEnvAsBool("AUTH_BYPASS", false),
			BypassUserID:    getEnv("AUTH_BYPASS_USER_ID", ""),
			BypassProfileID: getEnvAsInt64("AUTH_BYPASS_PROFILE_ID", 0),
		},
	}
}

// Helper functions untuk membaca environment variables

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	strValue := getEnv(key, "")
	if value, err := strconv.ParseInt(strValue, 10, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	strValue := getEnv(key, "")
	if value, err := strconv.ParseBool(strValue); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	strValue := getEnv(key, "")
	if strValue == "" {
		return defaultValue
	}
	return strings.Split(strValue, ",")
}

func getDuration(key, defaultValue string) time.Duration {
	strValue := getEnv(key, defaultValue)
	if duration, err := time.ParseDuration(strValue); err == nil {
		return duration
	}
	// Jika parsing gagal, coba parse default value
	duration, _ := time.ParseDuration(defaultValue)
	return duration
}