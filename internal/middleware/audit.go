package middleware

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/service"
)

// Gin context keys untuk audit
const (
	GinKeyAuditType       = "audit_type"
	GinKeyAuditAction     = "audit_action"
	GinKeyAuditTable      = "audit_table"
	GinKeyAuditRecordID   = "audit_record_id"
	GinKeyAuditBusinessID = "audit_business_id"
	GinKeyAuditCatalogID  = "audit_catalog_id"
	GinKeyAuditOldData    = "audit_old_data"
	GinKeyAuditNewData    = "audit_new_data"
	GinKeyAuditReason     = "audit_reason"
)

// Audit middleware untuk audit logging yang asinkron (renamed from AuditMiddleware)
func Audit(auditService service.AuditService, _ interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get profile ID from middleware
		profileID, _ := GetProfileID(c)
		
		// Set audit metadata berdasarkan route
		setAuditMetadata(c)
		
		// Process request
		c.Next()
		
		// Log audit setelah response selesai (asinkron)
		go func() {
			// Recovery untuk goroutine
			defer func() {
				if r := recover(); r != nil {
					// Log panic tapi jangan crash aplikasi
					fmt.Printf("Audit middleware panic: %v\n", r)
				}
			}()
			
			logAuditAfterResponse(c, auditService, profileID)
		}()
	}
}

// setAuditMetadata sets audit metadata berdasarkan HTTP method dan path
func setAuditMetadata(c *gin.Context) {
	method := c.Request.Method
	path := c.Request.URL.Path
	
	// Business endpoints
	if strings.Contains(path, "/api/v1/businesses") {
		c.Set(GinKeyAuditType, constant.AuditTypeBusiness)
		
		switch {
		case method == "POST" && strings.HasSuffix(path, "/businesses"):
			c.Set(GinKeyAuditAction, constant.AuditActionCreate)
			c.Set(GinKeyAuditTable, constant.AuditTableBusinesses)
			
		case method == "PUT" && strings.Contains(path, "/businesses/"):
			c.Set(GinKeyAuditAction, constant.AuditActionUpdate)
			c.Set(GinKeyAuditTable, constant.AuditTableBusinesses)
			
		case method == "DELETE" && strings.Contains(path, "/businesses/"):
			c.Set(GinKeyAuditAction, constant.AuditActionDelete)
			c.Set(GinKeyAuditTable, constant.AuditTableBusinesses)
			
		case method == "POST" && strings.Contains(path, "/users"):
			c.Set(GinKeyAuditAction, constant.AuditActionUserAdd)
			c.Set(GinKeyAuditTable, constant.AuditTableBusinessUsers)
			
		case method == "PUT" && strings.Contains(path, "/users/"):
			c.Set(GinKeyAuditAction, constant.AuditActionUserRoleUpdate)
			c.Set(GinKeyAuditTable, constant.AuditTableBusinessUsers)
			
		case method == "DELETE" && strings.Contains(path, "/users/"):
			c.Set(GinKeyAuditAction, constant.AuditActionUserRemove)
			c.Set(GinKeyAuditTable, constant.AuditTableBusinessUsers)
			
		case method == "POST" && strings.Contains(path, "/invites"):
			c.Set(GinKeyAuditAction, constant.AuditActionInviteCreate)
			c.Set(GinKeyAuditTable, constant.AuditTableBusinessInvites)
			
		case method == "POST" && strings.Contains(path, "/invites/accept"):
			c.Set(GinKeyAuditAction, constant.AuditActionInviteAccept)
			c.Set(GinKeyAuditTable, constant.AuditTableBusinessInvites)
			
		case method == "POST" && strings.Contains(path, "/subscriptions/activate"):
			c.Set(GinKeyAuditAction, constant.AuditActionSubscriptionActivate)
			c.Set(GinKeyAuditTable, constant.AuditTableBusinessSubscriptions)
		}
	}
	
	// Catalog endpoints
	if strings.Contains(path, "/api/v1/catalogs") {
		c.Set(GinKeyAuditType, constant.AuditTypeCatalog)
		
		switch {
		case method == "POST" && strings.HasSuffix(path, "/catalogs"):
			c.Set(GinKeyAuditAction, constant.AuditActionCreate)
			c.Set(GinKeyAuditTable, constant.AuditTableCatalogs)
			
		case method == "PUT" && strings.Contains(path, "/catalogs/"):
			c.Set(GinKeyAuditAction, constant.AuditActionUpdate)
			c.Set(GinKeyAuditTable, constant.AuditTableCatalogs)
			
		case method == "DELETE" && strings.Contains(path, "/catalogs/"):
			c.Set(GinKeyAuditAction, constant.AuditActionDelete)
			c.Set(GinKeyAuditTable, constant.AuditTableCatalogs)
			
		case method == "POST" && strings.Contains(path, "/sections"):
			c.Set(GinKeyAuditAction, constant.AuditActionSectionAdd)
			c.Set(GinKeyAuditTable, constant.AuditTableCatalogSections)
			
		case method == "PUT" && strings.Contains(path, "/sections/"):
			c.Set(GinKeyAuditAction, constant.AuditActionUpdate)
			c.Set(GinKeyAuditTable, constant.AuditTableCatalogSections)
			
		case method == "DELETE" && strings.Contains(path, "/sections/"):
			c.Set(GinKeyAuditAction, constant.AuditActionSectionRemove)
			c.Set(GinKeyAuditTable, constant.AuditTableCatalogSections)
			
		case method == "POST" && strings.Contains(path, "/cards"):
			c.Set(GinKeyAuditAction, constant.AuditActionCardAdd)
			c.Set(GinKeyAuditTable, constant.AuditTableCatalogCards)
			
		case method == "PUT" && strings.Contains(path, "/cards/"):
			c.Set(GinKeyAuditAction, constant.AuditActionUpdate)
			c.Set(GinKeyAuditTable, constant.AuditTableCatalogCards)
			
		case method == "DELETE" && strings.Contains(path, "/cards/"):
			c.Set(GinKeyAuditAction, constant.AuditActionCardRemove)
			c.Set(GinKeyAuditTable, constant.AuditTableCatalogCards)
		}
	}
	
	// Profile endpoints
	if strings.Contains(path, "/api/v1/profile") {
		c.Set(GinKeyAuditType, constant.AuditTypeBusiness)
		c.Set(GinKeyAuditTable, constant.AuditTableUserProfiles)
		
		switch method {
		case "POST":
			c.Set(GinKeyAuditAction, constant.AuditActionCreate)
		case "PUT":
			c.Set(GinKeyAuditAction, constant.AuditActionUpdate)
		case "DELETE":
			c.Set(GinKeyAuditAction, constant.AuditActionDelete)
		}
	}
}

// logAuditAfterResponse melakukan audit logging setelah response dikirim
func logAuditAfterResponse(c *gin.Context, auditService service.AuditService, profileID int64) {
	// Skip jika bukan successful operation atau GET request
	if c.Writer.Status() >= 400 || c.Request.Method == "GET" {
		return
	}
	
	// Get audit metadata
	auditType, exists := c.Get(GinKeyAuditType)
	if !exists {
		return
	}
	
	action, _ := c.Get(GinKeyAuditAction)
	table, _ := c.Get(GinKeyAuditTable)
	recordID, _ := c.Get(GinKeyAuditRecordID)
	reason, _ := c.Get(GinKeyAuditReason)
	
	// Convert to strings
	actionStr, _ := action.(string)
	tableStr, _ := table.(string)
	recordIDStr, _ := recordID.(string)
	reasonStr, _ := reason.(string)
	
	if actionStr == "" || tableStr == "" {
		return
	}
	
	// Get old and new data
	oldData, _ := c.Get(GinKeyAuditOldData)
	newData, _ := c.Get(GinKeyAuditNewData)
	
	// Extract IDs from path if not set
	businessID := extractBusinessIDFromPath(c.Request.URL.Path)
	catalogID := extractCatalogIDFromPath(c.Request.URL.Path)
	
	// Override with context values if available
	if ctxBusinessID, exists := c.Get(GinKeyAuditBusinessID); exists {
		if id, ok := ctxBusinessID.(int64); ok {
			businessID = &id
		}
	}
	
	if ctxCatalogID, exists := c.Get(GinKeyAuditCatalogID); exists {
		if id, ok := ctxCatalogID.(int64); ok {
			catalogID = &id
		}
	}
	
	// Extract record ID from response if not set
	if recordIDStr == "" {
		recordIDStr = extractRecordIDFromResponse(c, actionStr)
	}
	
	var profileIDPtr *int64
	if profileID > 0 {
		profileIDPtr = &profileID
	}
	
	// Log based on type
	switch auditType {
	case constant.AuditTypeBusiness:
		auditService.LogBusinessActionAsync(
			profileIDPtr,
			businessID,
			actionStr,
			tableStr,
			recordIDStr,
			oldData,
			newData,
			reasonStr,
		)
		
	case constant.AuditTypeCatalog:
		auditService.LogCatalogActionAsync(
			profileIDPtr,
			catalogID,
			actionStr,
			tableStr,
			recordIDStr,
			oldData,
			newData,
			reasonStr,
		)
	}
}

// extractBusinessIDFromPath extracts business ID from URL path
func extractBusinessIDFromPath(path string) *int64 {
	// Pattern: /api/v1/businesses/{id}
	parts := strings.Split(path, "/")
	
	for i, part := range parts {
		if part == "businesses" && i+1 < len(parts) {
			if id, err := strconv.ParseInt(parts[i+1], 10, 64); err == nil {
				return &id
			}
		}
	}
	
	return nil
}

// extractCatalogIDFromPath extracts catalog ID from URL path
func extractCatalogIDFromPath(path string) *int64 {
	// Pattern: /api/v1/catalogs/{id}
	parts := strings.Split(path, "/")
	
	for i, part := range parts {
		if part == "catalogs" && i+1 < len(parts) {
			if id, err := strconv.ParseInt(parts[i+1], 10, 64); err == nil {
				return &id
			}
		}
	}
	
	return nil
}

// extractRecordIDFromResponse extracts record ID from response body
func extractRecordIDFromResponse(c *gin.Context, action string) string {
	// Only extract for CREATE actions
	if action != constant.AuditActionCreate {
		return ""
	}
	
	// You could store the ID in context during creation
	if recordID, exists := c.Get("created_record_id"); exists {
		if id, ok := recordID.(int64); ok {
			return fmt.Sprintf("%d", id)
		}
		if idStr, ok := recordID.(string); ok {
			return idStr
		}
	}
	
	return ""
}

// SetAuditData helper function untuk set audit data di handler
func SetAuditData(c *gin.Context, key string, value interface{}) {
	c.Set(key, value)
}

// SetAuditBusinessID sets business ID for audit
func SetAuditBusinessID(c *gin.Context, businessID int64) {
	c.Set(GinKeyAuditBusinessID, businessID)
}

// SetAuditCatalogID sets catalog ID for audit
func SetAuditCatalogID(c *gin.Context, catalogID int64) {
	c.Set(GinKeyAuditCatalogID, catalogID)
}

// SetAuditRecordID sets record ID for audit
func SetAuditRecordID(c *gin.Context, recordID string) {
	c.Set(GinKeyAuditRecordID, recordID)
}

// SetAuditOldData sets old data for audit
func SetAuditOldData(c *gin.Context, data interface{}) {
	c.Set(GinKeyAuditOldData, data)
}

// SetAuditNewData sets new data for audit
func SetAuditNewData(c *gin.Context, data interface{}) {
	c.Set(GinKeyAuditNewData, data)
}

// SetAuditReason sets reason for audit
func SetAuditReason(c *gin.Context, reason string) {
	c.Set(GinKeyAuditReason, reason)
}