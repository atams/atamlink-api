package middleware

import (
	"bytes"
	"encoding/json"
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

// AuditConfig konfigurasi untuk audit middleware
type AuditConfig struct {
	// Skip audit untuk paths tertentu
	SkipPaths []string
	// Skip audit untuk methods tertentu
	SkipMethods []string
	// Record request body
	RecordRequestBody bool
	// Record response body
	RecordResponseBody bool
	// Max body size to record (bytes)
	MaxBodySize int64
}

// DefaultAuditConfig default config untuk audit
func DefaultAuditConfig() *AuditConfig {
	return &AuditConfig{
		SkipPaths: []string{
			"/health",
			"/health/db",
			"/metrics",
			"/swagger",
		},
		SkipMethods:        []string{"GET", "OPTIONS"},
		RecordRequestBody:  true,
		RecordResponseBody: true,
		MaxBodySize:        1048576, // 1MB
	}
}

// auditResponseWriter custom response writer untuk capture response
type auditResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *auditResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// safeJSONMarshal marshals data to JSON, returns nil if error or data is nil
func safeJSONMarshal(data interface{}) json.RawMessage {
	if data == nil {
		return nil
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	
	// Validate JSON
	if !json.Valid(jsonData) {
		return nil
	}
	
	return jsonData
}

// Audit middleware untuk audit logging
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

		// Get user info from context
		profileID, _ := GetProfileID(c)
		var profileIDPtr *int64
		if profileID > 0 {
			profileIDPtr = &profileID
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

		// After request processed, log audit asynchronously
		go func() {
			// Determine action based on method and path
			action := determineAction(c)

			// Extract business ID if available
			var businessID *int64
			var recordID string
			
			// Try to get ID from URL parameter first
			if bidStr := c.Param("id"); bidStr != "" {
				if bid, err := strconv.ParseInt(bidStr, 10, 64); err == nil {
					businessID = &bid
					recordID = bidStr
				}
			}

			// Extract table name from path
			table := extractTableName(c)

			// Build context
			ctx := map[string]interface{}{
				"method":       c.Request.Method,
				"path":         c.Request.URL.Path,
				"query":        c.Request.URL.RawQuery,
				"status":       c.Writer.Status(),
				"duration_ms":  time.Since(start).Milliseconds(),
				"client_ip":    c.ClientIP(),
				"user_agent":   c.Request.UserAgent(),
				"request_id":   c.GetString("requestID"),
			}

			// Initialize old data dan new data
			var oldDataJSON, newDataJSON json.RawMessage

			// Old data (untuk UPDATE/DELETE operations)
			if oldData, exists := c.Get(GinKeyAuditOldData); exists {
				oldDataJSON = safeJSONMarshal(oldData)
			}

			// New data (untuk CREATE/UPDATE operations)
			if newData, exists := c.Get(GinKeyAuditNewData); exists {
				newDataJSON = safeJSONMarshal(newData)
			} else if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
				// Jika tidak ada data baru dari context, ambil dari response body
				// Tapi hanya ambil bagian "data" saja, bukan keseluruhan response
				if blw.body.Len() > 0 {
					var responseData map[string]interface{}
					if err := json.Unmarshal(blw.body.Bytes(), &responseData); err == nil {
						// Ambil hanya bagian "data" dari response
						if data, exists := responseData["data"]; exists {
							newDataJSON = safeJSONMarshal(data)
						}
					}
				}
			}

			// Extract business ID dan record ID dari response data jika belum ada
			if businessID == nil || recordID == "" {
				extractedBusinessID, extractedRecordID := extractIDsFromData(newDataJSON, oldDataJSON, table)
				if businessID == nil {
					businessID = extractedBusinessID
				}
				if recordID == "" {
					recordID = extractedRecordID
				}
			}

			// Only proceed if we have meaningful data or it's a valid operation
			if shouldLogAudit(action, oldDataJSON, newDataJSON, c.Writer.Status()) {
				// Create audit entry
				entry := &service.AuditEntry{
					UserProfileID: profileIDPtr,
					BusinessID:    businessID,
					Action:        action,
					Table:         table,
					RecordID:      recordID,
					OldData:       oldDataJSON,
					NewData:       newDataJSON,
					Context:       ctx,
				}

				// Log audit entry
				auditService.Log(entry)
			}
		}()
	}
}

// shouldLogAudit determines if an audit entry should be logged
func shouldLogAudit(action string, oldData, newData json.RawMessage, status int) bool {
	// Don't log if status is not successful
	if status < 200 || status >= 300 {
		return false
	}

	// Always log these actions
	switch action {
	case "DELETE":
		return oldData != nil
	case "CREATE":
		return newData != nil
	case "UPDATE":
		return oldData != nil && newData != nil
	case "INVITE_SENT":
		return newData != nil
	default:
		return oldData != nil || newData != nil
	}
}

// determineAction menentukan action berdasarkan method dan path
func determineAction(c *gin.Context) string {
	method := c.Request.Method
	path := c.Request.URL.Path

	switch method {
	case "POST":
		if strings.Contains(path, "/invite") {
			return "INVITE_SENT"
		}
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	default:
		return method
	}
}

// extractIDsFromData extracts business ID and record ID from JSON data
func extractIDsFromData(newData, oldData json.RawMessage, table string) (*int64, string) {
	// Try to extract from new data first
	if newData != nil {
		if businessID, recordID := parseIDsFromJSON(newData, table); businessID != nil {
			return businessID, recordID
		}
	}
	
	// Fall back to old data
	if oldData != nil {
		if businessID, recordID := parseIDsFromJSON(oldData, table); businessID != nil {
			return businessID, recordID
		}
	}
	
	return nil, ""
}

// parseIDsFromJSON parses JSON and extracts relevant IDs
func parseIDsFromJSON(data json.RawMessage, table string) (*int64, string) {
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, ""
	}
	
	// Extract record ID (always "id" field)
	var recordID string
	if id, exists := parsed["id"]; exists {
		switch v := id.(type) {
		case float64:
			recordID = strconv.FormatInt(int64(v), 10)
		case int64:
			recordID = strconv.FormatInt(v, 10)
		case string:
			recordID = v
		}
	}
	
	// Extract business ID based on table
	var businessID *int64
	switch table {
	case "businesses":
		// For businesses table, business ID = record ID
		if recordID != "" {
			if bid, err := strconv.ParseInt(recordID, 10, 64); err == nil {
				businessID = &bid
			}
		}
	case "business_users", "business_invites", "business_subscriptions":
		// For business-related tables, look for business_id field
		if bid, exists := parsed["business_id"]; exists {
			switch v := bid.(type) {
			case float64:
				id := int64(v)
				businessID = &id
			case int64:
				businessID = &v
			}
		}
	case "catalogs", "products":
		// For catalog/product tables, might need to look up business_id
		// For now, return nil since we need business context
		businessID = nil
	}
	
	return businessID, recordID
}

// extractTableName mengekstrak nama tabel dari path
func extractTableName(c *gin.Context) string {
	path := c.Request.URL.Path

	// Map path ke table name
	switch {
	case strings.Contains(path, "/businesses"):
		return "businesses"
	case strings.Contains(path, "/catalogs"):
		return "catalogs"
	case strings.Contains(path, "/profile"):
		return "user_profiles"
	case strings.Contains(path, "/users"):
		return "users"
	default:
		return ""
	}
}

// shouldSkipAudit menentukan apakah audit harus di-skip
func shouldSkipAudit(c *gin.Context, config *AuditConfig) bool {
	// Skip berdasarkan path
	for _, path := range config.SkipPaths {
		if strings.HasPrefix(c.Request.URL.Path, path) {
			return true
		}
	}

	// Skip berdasarkan method
	for _, method := range config.SkipMethods {
		if c.Request.Method == method {
			return true
		}
	}

	return false
}