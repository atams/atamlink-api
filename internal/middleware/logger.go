package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/atam/atamlink/pkg/logger"
)

// bodyLogWriter untuk capture response body
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Logger middleware untuk logging request dan response
func Logger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("requestID", requestID)

		// Log request
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		// Read request body untuk logging (skip binary content)
		var requestBody string
		if shouldLogRequestBody(c) {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			if isSafeForLogging(bodyBytes) {
				requestBody = string(bodyBytes)
			} else {
				requestBody = "[BINARY_DATA]"
			}
			// Restore body untuk handler
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Replace writer untuk capture response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get response body (hati-hati dengan large response)
		responseBody := ""
		if blw.body.Len() < 1048576 && isSafeForLogging(blw.body.Bytes()) { // < 1MB dan safe
			responseBody = blw.body.String()
		}

		// Log fields
		fields := []logger.Field{
			logger.String("request_id", requestID),
			logger.String("method", c.Request.Method),
			logger.String("path", path),
			logger.Int("status", c.Writer.Status()),
			logger.Duration("latency", latency),
			logger.String("client_ip", c.ClientIP()),
			logger.String("user_agent", c.Request.UserAgent()),
			logger.Int64("content_length", c.Request.ContentLength),
		}

		// Add error if exists
		if len(c.Errors) > 0 {
			fields = append(fields, logger.String("error", c.Errors.String()))
		}

		// Add request/response body for debug mode
		if log != nil {
			if requestBody != "" {
				fields = append(fields, logger.String("request_body", maskSensitiveData(requestBody)))
			}
			if responseBody != "" && c.Writer.Status() >= 400 {
				fields = append(fields, logger.String("response_body", responseBody))
			}
		}

		// Log based on status code
		switch {
		case c.Writer.Status() >= 500:
			log.Error("Server error", fields...)
		case c.Writer.Status() >= 400:
			log.Warn("Client error", fields...)
		case c.Writer.Status() >= 300:
			log.Info("Redirection", fields...)
		default:
			log.Info("Request processed", fields...)
		}
	}
}

// shouldLogRequestBody determines if request body should be logged
func shouldLogRequestBody(c *gin.Context) bool {
	// Skip jika tidak ada body
	if c.Request.Body == nil || c.Request.ContentLength == 0 {
		return false
	}

	// Skip jika body terlalu besar (> 512KB)
	if c.Request.ContentLength > 524288 {
		return false
	}

	// Skip untuk content types yang biasanya binary
	contentType := c.GetHeader("Content-Type")
	
	// Skip multipart/form-data (file uploads)
	if strings.Contains(contentType, "multipart/form-data") {
		return false
	}
	
	// Skip binary content types
	binaryTypes := []string{
		"image/",
		"video/",
		"audio/",
		"application/octet-stream",
		"application/pdf",
		"application/zip",
		"application/gzip",
	}
	
	for _, binaryType := range binaryTypes {
		if strings.Contains(contentType, binaryType) {
			return false
		}
	}

	return true
}

// isSafeForLogging checks if data is safe to log (no binary/null bytes)
func isSafeForLogging(data []byte) bool {
	// Check for null bytes (binary data indicator)
	if bytes.Contains(data, []byte{0}) {
		return false
	}
	
	// Check for excessive non-printable characters
	nonPrintable := 0
	for _, b := range data {
		if b < 32 && b != 9 && b != 10 && b != 13 { // Allow tab, LF, CR
			nonPrintable++
		}
	}
	
	// If more than 10% non-printable, consider it binary
	if len(data) > 0 && float64(nonPrintable)/float64(len(data)) > 0.1 {
		return false
	}
	
	return true
}

// generateRequestID generate unique request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + generateRandomString(8)
}

// generateRandomString generate random string
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// maskSensitiveData mask sensitive data dalam log
func maskSensitiveData(data string) string {
	// Simple masking untuk password
	// Dalam produksi, gunakan regex yang lebih comprehensive
	masked := data
	sensitiveFields := []string{"password", "token", "secret", "api_key"}
	
	for _, field := range sensitiveFields {
		// Basic implementation - enhance sesuai kebutuhan
		if contains(data, field) {
			masked = "[MASKED]"
			break
		}
	}
	
	return masked
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}