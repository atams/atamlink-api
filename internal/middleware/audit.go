package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/atam/atamlink/internal/service"
)

const (
	GinKeyAuditOldData = "audit_old_data"
	GinKeyAuditNewData = "audit_new_data"
)

// Audit configuration
type AuditConfig struct {
	SkipPaths      []string
	SkipMethods    []string
	SkipStatusCode []int
}

// Default audit configuration
func DefaultAuditConfig() *AuditConfig {
	return &AuditConfig{
		SkipPaths: []string{
			"/health",
			"/health/db",
			"/swagger",
			"/metrics",
			"/favicon.ico",
		},
		SkipMethods: []string{
			"OPTIONS",
			"HEAD",
		},
		SkipStatusCode: []int{},
	}
}

// auditResponseWriter wraps gin.ResponseWriter untuk capture response
type auditResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *auditResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// AuditData struktur untuk menyimpan data audit yang akan di-process
type AuditData struct {
	ProfileID     *int64
	BusinessID    *int64
	Action        string
	Table         string
	RecordID      string
	OldData       json.RawMessage
	NewData       json.RawMessage
	Context       map[string]interface{}
	RequestBody   []byte
	ResponseBody  []byte
	Status        int
}

// Audit middleware untuk audit logging yang aman
func Audit(auditService service.AuditService, config *AuditConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultAuditConfig()
	}

	return func(c *gin.Context) {
		// Check if should skip
		if shouldSkipAudit(c, config) {
			c.Next()
			return
		}

		// Prepare audit data collection
		auditData := &AuditData{
			Context: make(map[string]interface{}),
		}

		// Get user info from context
		if profileID, exists := GetProfileID(c); exists && profileID > 0 {
			auditData.ProfileID = &profileID
		}

		// Capture request body jika ada
		if c.Request.Body != nil && c.Request.Method != "GET" {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			// Batasi ukuran body yang disimpan (max 10KB)
			if len(bodyBytes) <= 10240 {
				auditData.RequestBody = bodyBytes
			}
		}

		// Start timer
		start := time.Now()

		// Wrap response writer
		blw := &auditResponseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		// Process request
		c.Next()

		// Collect data setelah request selesai TAPI SEBELUM response dikirim
		auditData.Status = c.Writer.Status()
		
		// Copy response body (batasi 10KB)
		if blw.body.Len() > 0 && blw.body.Len() <= 10240 {
			auditData.ResponseBody = make([]byte, blw.body.Len())
			copy(auditData.ResponseBody, blw.body.Bytes())
		}

		// Determine action based on method and path
		auditData.Action = determineAction(c)
		auditData.Table = extractTableName(c)

		// Extract IDs
		if bidStr := c.Param("id"); bidStr != "" {
			if bid, err := strconv.ParseInt(bidStr, 10, 64); err == nil {
				auditData.BusinessID = &bid
				auditData.RecordID = bidStr
			}
		}

		// Build context dengan data yang aman untuk di-copy
		auditData.Context = map[string]interface{}{
			"method":       c.Request.Method,
			"path":         c.Request.URL.Path,
			"query":        c.Request.URL.RawQuery,
			"status":       auditData.Status,
			"duration_ms":  time.Since(start).Milliseconds(),
			"client_ip":    c.ClientIP(),
			"user_agent":   c.Request.UserAgent(),
			"request_id":   c.GetString("requestID"),
		}

		// Get old data dari context (sudah di-set oleh usecase)
		if oldData, exists := c.Get(GinKeyAuditOldData); exists && oldData != nil {
			// Deep copy old data
			if jsonData, err := json.Marshal(oldData); err == nil {
				auditData.OldData = make(json.RawMessage, len(jsonData))
				copy(auditData.OldData, jsonData)
			}
		}

		// Get new data dari context atau response
		if newData, exists := c.Get(GinKeyAuditNewData); exists && newData != nil {
			// Deep copy new data dari context
			if jsonData, err := json.Marshal(newData); err == nil {
				auditData.NewData = make(json.RawMessage, len(jsonData))
				copy(auditData.NewData, jsonData)
			}
		} else if auditData.Status >= 200 && auditData.Status < 300 && len(auditData.ResponseBody) > 0 {
			// Extract dari response body jika sukses
			var responseData map[string]interface{}
			if err := json.Unmarshal(auditData.ResponseBody, &responseData); err == nil {
				if data, exists := responseData["data"]; exists && data != nil {
					if jsonData, err := json.Marshal(data); err == nil {
						auditData.NewData = make(json.RawMessage, len(jsonData))
						copy(auditData.NewData, jsonData)
					}
				}
			}
		}

		// Extract business ID dan record ID dari data jika belum ada
		if auditData.BusinessID == nil || auditData.RecordID == "" {
			extractedBusinessID, extractedRecordID := extractIDsFromAuditData(auditData)
			if auditData.BusinessID == nil {
				auditData.BusinessID = extractedBusinessID
			}
			if auditData.RecordID == "" {
				auditData.RecordID = extractedRecordID
			}
		}

		// Check if should log
		if shouldLogAudit(auditData) {
			// Create audit entry dengan data yang sudah di-copy
			entry := &service.AuditEntry{
				UserProfileID: auditData.ProfileID,
				BusinessID:    auditData.BusinessID,
				Action:        auditData.Action,
				Table:         auditData.Table,
				RecordID:      auditData.RecordID,
				OldData:       auditData.OldData,
				NewData:       auditData.NewData,
				Context:       auditData.Context,
			}

			// Log secara asynchronous dengan data yang sudah aman
			auditService.LogAsync(entry)
		}
	}
}

// shouldSkipAudit check if should skip audit logging
func shouldSkipAudit(c *gin.Context, config *AuditConfig) bool {
	// Skip by path
	for _, path := range config.SkipPaths {
		if strings.HasPrefix(c.Request.URL.Path, path) {
			return true
		}
	}

	// Skip by method
	for _, method := range config.SkipMethods {
		if c.Request.Method == method {
			return true
		}
	}

	// Skip by status code
	for _, status := range config.SkipStatusCode {
		if c.Writer.Status() == status {
			return true
		}
	}

	return false
}

// determineAction determine audit action from HTTP method
func determineAction(c *gin.Context) string {
	switch c.Request.Method {
	case "POST":
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	case "GET":
		if c.Param("id") != "" {
			return "VIEW"
		}
		return "LIST"
	default:
		return c.Request.Method
	}
}

// extractTableName extract table name from path
func extractTableName(c *gin.Context) string {
	path := c.Request.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	
	// Skip prefix (api/v1)
	if len(parts) > 2 {
		// Get resource name (usually after api/v1)
		resourceName := parts[2]
		
		// Convert to singular form for table name
		return toSingular(resourceName)
	}
	
	return "unknown"
}

// toSingular convert plural to singular (simple implementation)
func toSingular(plural string) string {
	// Simple rules, bisa diperluas
	switch plural {
	case "businesses":
		return "business"
	case "users":
		return "user"
	case "catalogs":
		return "catalog"
	case "categories":
		return "category"
	default:
		// Remove trailing 's' if exists
		if strings.HasSuffix(plural, "s") {
			return strings.TrimSuffix(plural, "s")
		}
		return plural
	}
}

// extractIDsFromAuditData extract IDs from audit data
func extractIDsFromAuditData(data *AuditData) (*int64, string) {
	var businessID *int64
	var recordID string

	// Try from new data first
	if data.NewData != nil {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(data.NewData, &jsonData); err == nil {
			// Extract business_id
			if bid, ok := jsonData["business_id"].(float64); ok {
				bidInt := int64(bid)
				businessID = &bidInt
			}
			
			// Extract record ID based on table
			idField := fmt.Sprintf("%s_id", data.Table)
			if id, ok := jsonData["id"].(float64); ok {
				recordID = fmt.Sprintf("%.0f", id)
			} else if id, ok := jsonData[idField].(float64); ok {
				recordID = fmt.Sprintf("%.0f", id)
			}
		}
	}

	// Try from old data if not found
	if (businessID == nil || recordID == "") && data.OldData != nil {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(data.OldData, &jsonData); err == nil {
			if businessID == nil {
				if bid, ok := jsonData["business_id"].(float64); ok {
					bidInt := int64(bid)
					businessID = &bidInt
				}
			}
			
			if recordID == "" {
				idField := fmt.Sprintf("%s_id", data.Table)
				if id, ok := jsonData["id"].(float64); ok {
					recordID = fmt.Sprintf("%.0f", id)
				} else if id, ok := jsonData[idField].(float64); ok {
					recordID = fmt.Sprintf("%.0f", id)
				}
			}
		}
	}

	return businessID, recordID
}

// shouldLogAudit check if should create audit log
func shouldLogAudit(data *AuditData) bool {
	// Skip jika tidak ada perubahan data yang signifikan
	if data.Action == "LIST" || data.Action == "VIEW" {
		return false
	}

	// Skip jika error (kecuali untuk operasi tertentu)
	if data.Status >= 400 {
		// Log error hanya untuk operasi yang mengubah data
		if data.Action == "UPDATE" || data.Action == "DELETE" {
			return data.OldData != nil
		}
		return false
	}

	// Harus ada data untuk di-log
	return data.OldData != nil || data.NewData != nil
}