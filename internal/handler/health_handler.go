package handler

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/atam/atamlink/pkg/utils"
)

// HealthHandler handler untuk health check
type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler membuat instance health handler baru
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// HealthStatus struktur untuk health status response
type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Uptime    string    `json:"uptime"`
}

// HealthDBStatus struktur untuk database health status
type HealthDBStatus struct {
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	Version     string    `json:"version"`
	Uptime      string    `json:"uptime"`
	Database    string    `json:"database"`
	DBConnected bool      `json:"db_connected"`
	DBVersion   string    `json:"db_version,omitempty"`
	DBError     string    `json:"db_error,omitempty"`
}

var startTime = time.Now()

// Check endpoint untuk basic health check
// @Summary Health check
// @Description Check if service is running
// @Tags health
// @Produce json
// @Success 200 {object} utils.Response{data=HealthStatus}
// @Router /health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	health := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0", // TODO: Get from build info
		Uptime:    time.Since(startTime).String(),
	}

	utils.OK(c, "Service is healthy", health)
}

// CheckDB endpoint untuk database health check
// @Summary Database health check
// @Description Check database connection
// @Tags health
// @Produce json
// @Success 200 {object} utils.Response{data=HealthDBStatus}
// @Failure 503 {object} utils.Response
// @Router /health/db [get]
func (h *HealthHandler) CheckDB(c *gin.Context) {
	health := HealthDBStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    time.Since(startTime).String(),
		Database:  "postgresql",
	}

	// Check database connection
	if err := h.db.Ping(); err != nil {
		health.Status = "unhealthy"
		health.DBConnected = false
		health.DBError = err.Error()
		
		utils.Error(c, 503, "Database tidak terhubung")
		return
	}

	// Get database version
	var dbVersion string
	err := h.db.QueryRow("SELECT version()").Scan(&dbVersion)
	if err == nil {
		health.DBVersion = dbVersion
	}

	health.DBConnected = true
	utils.OK(c, "Service dan database healthy", health)
}