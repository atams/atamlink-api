package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	
	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/mod_audit/entity"
	"github.com/atam/atamlink/internal/mod_audit/repository"
	"github.com/atam/atamlink/pkg/logger"
)

// AuditService interface untuk audit service
type AuditService interface {
	Start()
	Stop()
	LogBusinessAction(ctx context.Context, profileID *int64, businessID *int64, action, table, recordID string, oldData, newData interface{}, reason string)
	LogCatalogAction(ctx context.Context, profileID *int64, catalogID *int64, action, table, recordID string, oldData, newData interface{}, reason string)
	LogAsync(entry *AuditEntry)
	LogBusinessActionAsync(profileID *int64, businessID *int64, action, table, recordID string, oldData, newData interface{}, reason string)
	LogCatalogActionAsync(profileID *int64, catalogID *int64, action, table, recordID string, oldData, newData interface{}, reason string)
}

// AuditEntry entry untuk audit log
type AuditEntry struct {
	Type          string                 // "business" atau "catalog"
	UserProfileID *int64
	BusinessID    *int64
	CatalogID     *int64
	Action        string
	Table         string
	RecordID      string
	OldData       json.RawMessage
	NewData       json.RawMessage
	Context       map[string]interface{}
	Reason        string
}

type auditService struct {
	businessRepo repository.AuditRepository
	catalogRepo  repository.AuditCatalogRepository
	logger       logger.Logger
	channel      chan *AuditEntry
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	isRunning    bool
	mu           sync.RWMutex
}

// NewAuditService membuat instance audit service baru
func NewAuditService(businessRepo repository.AuditRepository, catalogRepo repository.AuditCatalogRepository, log logger.Logger) AuditService {
	ctx, cancel := context.WithCancel(context.Background())
	return &auditService{
		businessRepo: businessRepo,
		catalogRepo:  catalogRepo,
		logger:       log,
		channel:      make(chan *AuditEntry, 1000), // Buffer 1000 entries
		ctx:          ctx,
		cancel:       cancel,
		isRunning:    false,
	}
}

// Start memulai audit service worker
func (s *auditService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.isRunning {
		return
	}
	
	s.isRunning = true
	s.wg.Add(1)
	
	go s.worker()
	s.logger.Info("Audit service started")
}

// Stop menghentikan audit service
func (s *auditService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.isRunning {
		return
	}
	
	s.cancel()
	close(s.channel)
	s.wg.Wait()
	s.isRunning = false
	s.logger.Info("Audit service stopped")
}

// worker goroutine untuk memproses audit entries
func (s *auditService) worker() {
	defer s.wg.Done()
	
	batchSize := 50
	batchTimeout := 5 * time.Second
	
	businessBatch := make([]*entity.AuditLogBusiness, 0, batchSize)
	catalogBatch := make([]*entity.AuditLogCatalog, 0, batchSize)
	
	ticker := time.NewTicker(batchTimeout)
	defer ticker.Stop()
	
	for {
		select {
		case entry, ok := <-s.channel:
			if !ok {
				// Channel closed, process remaining entries
				s.processBatches(businessBatch, catalogBatch)
				return
			}
			
			if entry.Type == constant.AuditTypeBusiness {
				businessLog := s.convertToBusinessLog(entry)
				businessBatch = append(businessBatch, businessLog)
				
				if len(businessBatch) >= batchSize {
					s.processBusinessBatch(businessBatch)
					businessBatch = businessBatch[:0]
				}
			} else {
				catalogLog := s.convertToCatalogLog(entry)
				catalogBatch = append(catalogBatch, catalogLog)
				
				if len(catalogBatch) >= batchSize {
					s.processCatalogBatch(catalogBatch)
					catalogBatch = catalogBatch[:0]
				}
			}
			
		case <-ticker.C:
			// Process batches periodically
			s.processBatches(businessBatch, catalogBatch)
			businessBatch = businessBatch[:0]
			catalogBatch = catalogBatch[:0]
			
		case <-s.ctx.Done():
			// Context cancelled, process remaining entries
			s.processBatches(businessBatch, catalogBatch)
			return
		}
	}
}

// processBatches memproses kedua batch
func (s *auditService) processBatches(businessBatch []*entity.AuditLogBusiness, catalogBatch []*entity.AuditLogCatalog) {
	if len(businessBatch) > 0 {
		s.processBusinessBatch(businessBatch)
	}
	if len(catalogBatch) > 0 {
		s.processCatalogBatch(catalogBatch)
	}
}

// processBusinessBatch memproses batch business audit logs
func (s *auditService) processBusinessBatch(batch []*entity.AuditLogBusiness) {
	if len(batch) == 0 {
		return
	}
	
	if err := s.businessRepo.BatchCreate(batch); err != nil {
		s.logger.Error("Failed to batch create business audit logs", 
			logger.Error(err), 
			logger.Int("count", len(batch)))
		// Fallback: try to save individually
		for _, log := range batch {
			if err := s.businessRepo.Create(log); err != nil {
				s.logger.Error("Failed to create business audit log", 
					logger.Error(err), 
					logger.Int64("log_id", log.ID))
			}
		}
	} else {
		s.logger.Debug("Business audit logs batch processed", logger.Int("count", len(batch)))
	}
}

// processCatalogBatch memproses batch catalog audit logs
func (s *auditService) processCatalogBatch(batch []*entity.AuditLogCatalog) {
	if len(batch) == 0 {
		return
	}
	
	if err := s.catalogRepo.BatchCreate(batch); err != nil {
		s.logger.Error("Failed to batch create catalog audit logs", 
			logger.Error(err), 
			logger.Int("count", len(batch)))
		// Fallback: try to save individually
		for _, log := range batch {
			if err := s.catalogRepo.Create(log); err != nil {
				s.logger.Error("Failed to create catalog audit log", 
					logger.Error(err), 
					logger.Int64("log_id", log.ID))
			}
		}
	} else {
		s.logger.Debug("Catalog audit logs batch processed", logger.Int("count", len(batch)))
	}
}

// LogBusinessAction logs business-related action (synchronous)
func (s *auditService) LogBusinessAction(ctx context.Context, profileID *int64, businessID *int64, action, table, recordID string, oldData, newData interface{}, reason string) {
	entry := s.createAuditEntry(constant.AuditTypeBusiness, profileID, businessID, nil, action, table, recordID, oldData, newData, reason)
	s.addContextFromGin(ctx, entry)
	
	// Process immediately for critical actions
	if s.isCriticalAction(action) {
		businessLog := s.convertToBusinessLog(entry)
		if err := s.businessRepo.Create(businessLog); err != nil {
			s.logger.Error("Failed to create critical business audit log", logger.Error(err))
		}
		return
	}
	
	s.LogAsync(entry)
}

// LogCatalogAction logs catalog-related action (synchronous)
func (s *auditService) LogCatalogAction(ctx context.Context, profileID *int64, catalogID *int64, action, table, recordID string, oldData, newData interface{}, reason string) {
	entry := s.createAuditEntry(constant.AuditTypeCatalog, profileID, nil, catalogID, action, table, recordID, oldData, newData, reason)
	s.addContextFromGin(ctx, entry)
	
	// Process immediately for critical actions
	if s.isCriticalAction(action) {
		catalogLog := s.convertToCatalogLog(entry)
		if err := s.catalogRepo.Create(catalogLog); err != nil {
			s.logger.Error("Failed to create critical catalog audit log", logger.Error(err))
		}
		return
	}
	
	s.LogAsync(entry)
}

// LogBusinessActionAsync logs business action asynchronously
func (s *auditService) LogBusinessActionAsync(profileID *int64, businessID *int64, action, table, recordID string, oldData, newData interface{}, reason string) {
	entry := s.createAuditEntry(constant.AuditTypeBusiness, profileID, businessID, nil, action, table, recordID, oldData, newData, reason)
	s.LogAsync(entry)
}

// LogCatalogActionAsync logs catalog action asynchronously
func (s *auditService) LogCatalogActionAsync(profileID *int64, catalogID *int64, action, table, recordID string, oldData, newData interface{}, reason string) {
	entry := s.createAuditEntry(constant.AuditTypeCatalog, profileID, nil, catalogID, action, table, recordID, oldData, newData, reason)
	s.LogAsync(entry)
}

// LogAsync menambahkan entry ke channel untuk diproses asinkron
func (s *auditService) LogAsync(entry *AuditEntry) {
	s.mu.RLock()
	isRunning := s.isRunning
	s.mu.RUnlock()
	
	if !isRunning {
		s.logger.Warn("Audit service not running, skipping log entry")
		return
	}
	
	// Deep copy entry untuk menghindari race condition
	copiedEntry := s.deepCopyEntry(entry)
	
	select {
	case s.channel <- copiedEntry:
		// Successfully queued
	default:
		// Channel full, log error and drop entry
		s.logger.Error("Audit channel full, dropping entry", 
			logger.String("action", entry.Action), 
			logger.String("table", entry.Table))
	}
}

// createAuditEntry membuat audit entry baru
func (s *auditService) createAuditEntry(entryType string, profileID, businessID, catalogID *int64, action, table, recordID string, oldData, newData interface{}, reason string) *AuditEntry {
	entry := &AuditEntry{
		Type:          entryType,
		UserProfileID: profileID,
		BusinessID:    businessID,
		CatalogID:     catalogID,
		Action:        action,
		Table:         table,
		RecordID:      recordID,
		Context:       make(map[string]interface{}),
		Reason:        reason,
	}
	
	// Convert data to JSON
	if oldData != nil {
		if jsonData, err := json.Marshal(oldData); err == nil {
			entry.OldData = jsonData
		}
	}
	
	if newData != nil {
		if jsonData, err := json.Marshal(newData); err == nil {
			entry.NewData = jsonData
		}
	}
	
	// Add default reason if empty
	if entry.Reason == "" {
		entry.Reason = constant.GetAuditMessage(action)
	}
	
	return entry
}

// addContextFromGin menambahkan context dari Gin context
func (s *auditService) addContextFromGin(ctx context.Context, entry *AuditEntry) {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		entry.Context["request_id"] = ginCtx.GetString("request_id")
		entry.Context["user_agent"] = ginCtx.GetHeader("User-Agent")
		entry.Context["ip_address"] = ginCtx.ClientIP()
		entry.Context["method"] = ginCtx.Request.Method
		entry.Context["path"] = ginCtx.Request.URL.Path
	}
}

// isCriticalAction menentukan apakah action perlu diproses segera
func (s *auditService) isCriticalAction(action string) bool {
	criticalActions := map[string]bool{
		constant.AuditActionDelete:             true,
		constant.AuditActionBusinessSuspend:    true,
		constant.AuditActionUserRemove:         true,
		constant.AuditActionSubscriptionCancel: true,
		constant.AuditActionInviteCancel:       true,
	}
	return criticalActions[action]
}

// convertToBusinessLog mengkonversi AuditEntry ke AuditLogBusiness
func (s *auditService) convertToBusinessLog(entry *AuditEntry) *entity.AuditLogBusiness {
	return &entity.AuditLogBusiness{
		Timestamp:     time.Now(),
		UserProfileID: entry.UserProfileID,
		BusinessID:    entry.BusinessID,
		Action:        entry.Action,
		Table:         entry.Table,
		RecordID:      entry.RecordID,
		OldData:       entry.OldData,
		NewData:       entry.NewData,
		Context:       entry.Context,
		Reason:        entry.Reason,
	}
}

// convertToCatalogLog mengkonversi AuditEntry ke AuditLogCatalog
func (s *auditService) convertToCatalogLog(entry *AuditEntry) *entity.AuditLogCatalog {
	return &entity.AuditLogCatalog{
		Timestamp:     time.Now(),
		UserProfileID: entry.UserProfileID,
		CatalogID:     entry.CatalogID,
		Action:        entry.Action,
		Table:         entry.Table,
		RecordID:      entry.RecordID,
		OldData:       entry.OldData,
		NewData:       entry.NewData,
		Context:       entry.Context,
		Reason:        entry.Reason,
	}
}

// deepCopyEntry membuat deep copy dari audit entry
func (s *auditService) deepCopyEntry(entry *AuditEntry) *AuditEntry {
	copied := &AuditEntry{
		Type:     entry.Type,
		Action:   entry.Action,
		Table:    entry.Table,
		RecordID: entry.RecordID,
		Reason:   entry.Reason,
	}
	
	// Copy pointers
	if entry.UserProfileID != nil {
		profileID := *entry.UserProfileID
		copied.UserProfileID = &profileID
	}
	
	if entry.BusinessID != nil {
		businessID := *entry.BusinessID
		copied.BusinessID = &businessID
	}
	
	if entry.CatalogID != nil {
		catalogID := *entry.CatalogID
		copied.CatalogID = &catalogID
	}
	
	// Deep copy JSON data
	if entry.OldData != nil {
		copied.OldData = make(json.RawMessage, len(entry.OldData))
		copy(copied.OldData, entry.OldData)
	}
	
	if entry.NewData != nil {
		copied.NewData = make(json.RawMessage, len(entry.NewData))
		copy(copied.NewData, entry.NewData)
	}
	
	// Deep copy context
	if entry.Context != nil {
		copied.Context = make(map[string]interface{})
		for k, v := range entry.Context {
			copied.Context[k] = v
		}
	}
	
	return copied
}