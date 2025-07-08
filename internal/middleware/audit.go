package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/atam/atamlink/internal/service"
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

		// Capture request body
		var requestBody interface{}
		if config.RecordRequestBody && c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			
			if len(bodyBytes) > 0 && int64(len(bodyBytes)) <= config.MaxBodySize {
				var body interface{}
				if err := json.Unmarshal(bodyBytes, &body); err == nil {
					requestBody = body
				} else {
					requestBody = string(bodyBytes)
				}
			}
		}

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
			if bidStr := c.Param("id"); bidStr != "" {
				if bid, err := strconv.ParseInt(bidStr, 10, 64); err == nil {
					businessID = &bid
				}
			}

			// Extract table name and record ID from path
			table, recordID := extractTableAndRecord(c)

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

			// Add response body if configured
			var responseBody interface{}
			if config.RecordResponseBody && blw.body.Len() > 0 && 
			   int64(blw.body.Len()) <= config.MaxBodySize &&
			   c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
				var body interface{}
				if err := json.Unmarshal(blw.body.Bytes(), &body); err == nil {
					responseBody = body
				}
			}

			// Create audit entry
			entry := &service.AuditEntry{
				UserProfileID: profileIDPtr,
				BusinessID:    businessID,
				Action:        action,
				Table:         table,
				RecordID:      recordID,
				OldData:       nil, // Will be set for UPDATE operations
				NewData:       requestBody,
				Context:       ctx,
			}

			// For successful responses, record response as new data
			if c.Writer.Status() >= 200 && c.Writer.Status() < 300 && responseBody != nil {
				if action == "CREATE" || action == "UPDATE" {
					entry.NewData = responseBody
				}
			}

			// Log to audit service
			auditService.Log(entry)
		}()
	}
}

// shouldSkipAudit check apakah request harus di-skip dari audit
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

	return false
}

// determineAction menentukan action berdasarkan method dan path
func determineAction(c *gin.Context) string {
	method := c.Request.Method
	path := c.Request.URL.Path

	switch method {
	case "POST":
		if strings.Contains(path, "/invites/accept") {
			return "INVITE_USED"
		} else if strings.Contains(path, "/invites") {
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

// extractTableAndRecord extract table name dan record ID dari path
func extractTableAndRecord(c *gin.Context) (string, string) {
	path := c.Request.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	
	if len(parts) < 3 {
		return "", ""
	}

	// Skip prefix (api/v1)
	if parts[0] == "api" {
		parts = parts[2:]
	}

	if len(parts) == 0 {
		return "", ""
	}

	// Get table name from first part
	table := parts[0]
	
	// Get record ID if exists
	recordID := ""
	if len(parts) > 1 && isNumeric(parts[1]) {
		recordID = parts[1]
	}

	return table, recordID
}

// isNumeric check apakah string adalah numeric
func isNumeric(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}