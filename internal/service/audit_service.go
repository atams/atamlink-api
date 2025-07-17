package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/atam/atamlink/internal/mod_audit/entity"
	"github.com/atam/atamlink/internal/mod_audit/repository"
	"github.com/atam/atamlink/pkg/logger"
)

// AuditService service untuk handle audit logging
type AuditService interface {
	Start()
	Stop()
	Log(entry *AuditEntry)
	// Tambahan untuk async processing yang aman
	LogAsync(entry *AuditEntry)
}

// AuditEntry entry untuk audit log
type AuditEntry struct {
	UserProfileID *int64
	BusinessID    *int64
	Action        string
	Table         string
	RecordID      string
	OldData       json.RawMessage
	NewData       json.RawMessage
	Context       map[string]interface{}
	Reason        string
}

// DeepCopyAuditEntry membuat deep copy dari audit entry
func DeepCopyAuditEntry(entry *AuditEntry) *AuditEntry {
	if entry == nil {
		return nil
	}

	copied := &AuditEntry{
		Action:   entry.Action,
		Table:    entry.Table,
		RecordID: entry.RecordID,
		Reason:   entry.Reason,
	}

	// Copy pointer fields
	if entry.UserProfileID != nil {
		profileID := *entry.UserProfileID
		copied.UserProfileID = &profileID
	}

	if entry.BusinessID != nil {
		businessID := *entry.BusinessID
		copied.BusinessID = &businessID
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

	// Deep copy context map
	if entry.Context != nil {
		copied.Context = make(map[string]interface{})
		for k, v := range entry.Context {
			// Handle nested maps/slices if needed
			switch val := v.(type) {
			case map[string]interface{}:
				// Deep copy nested map
				nestedMap := make(map[string]interface{})
				for nk, nv := range val {
					nestedMap[nk] = nv
				}
				copied.Context[k] = nestedMap
			case []interface{}:
				// Deep copy slice
				slice := make([]interface{}, len(val))
				copy(slice, val)
				copied.Context[k] = slice
			default:
				// Primitive types are copied by value
				copied.Context[k] = v
			}
		}
	}

	return copied
}

type auditService struct {
	repo      repository.AuditRepository
	log       logger.Logger
	queue     chan *entity.AuditLog
	batchSize int
	flushTime time.Duration
	wg        sync.WaitGroup
	stop      chan bool
	// Pool untuk mengurangi alokasi memory
	entryPool sync.Pool
}

// NewAuditService membuat instance audit service baru
func NewAuditService(repo repository.AuditRepository, log logger.Logger) AuditService {
	s := &auditService{
		repo:      repo,
		log:       log,
		queue:     make(chan *entity.AuditLog, 1000),
		batchSize: 10,
		flushTime: 5 * time.Second,
		stop:      make(chan bool),
	}

	// Initialize pool
	s.entryPool = sync.Pool{
		New: func() interface{} {
			return &entity.AuditLog{}
		},
	}

	return s
}

// Start memulai audit service worker
func (s *auditService) Start() {
	s.wg.Add(1)
	go s.worker()
	s.log.Info("Audit service started")
}

// Stop menghentikan audit service dengan graceful shutdown
func (s *auditService) Stop() {
	s.log.Info("Stopping audit service...")
	
	// Signal stop
	close(s.stop)
	
	// Wait for worker to finish with timeout
	done := make(chan bool)
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.log.Info("Audit service stopped gracefully")
	case <-time.After(30 * time.Second):
		s.log.Error("Audit service stop timeout")
	}

	close(s.queue)
}

// Log menambahkan entry ke queue (synchronous)
func (s *auditService) Log(entry *AuditEntry) {
	auditLog := s.convertToAuditLog(entry)
	
	// Non-blocking send dengan timeout
	select {
	case s.queue <- auditLog:
		// Successfully queued
	case <-time.After(100 * time.Millisecond):
		// Timeout, log error
		s.log.Error("Audit queue timeout",
			logger.String("action", entry.Action),
			logger.String("table", entry.Table),
			logger.String("record", entry.RecordID),
		)
	default:
		// Queue full, log error
		s.log.Error("Audit queue full",
			logger.String("action", entry.Action),
			logger.String("table", entry.Table),
			logger.String("record", entry.RecordID),
		)
	}
}

// LogAsync menambahkan entry ke queue secara asynchronous dengan deep copy
func (s *auditService) LogAsync(entry *AuditEntry) {
	// Deep copy entry untuk menghindari data race
	copiedEntry := DeepCopyAuditEntry(entry)
	
	// Process in goroutine
	go func() {
		s.Log(copiedEntry)
	}()
}

// convertToAuditLog konversi AuditEntry ke entity.AuditLog
func (s *auditService) convertToAuditLog(entry *AuditEntry) *entity.AuditLog {
	// Get from pool or create new
	auditLog := s.entryPool.Get().(*entity.AuditLog)
	
	// Reset and populate
	*auditLog = entity.AuditLog{
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

	return auditLog
}

// worker process audit logs dari queue
func (s *auditService) worker() {
	defer s.wg.Done()

	batch := make([]*entity.AuditLog, 0, s.batchSize)
	ticker := time.NewTicker(s.flushTime)
	defer ticker.Stop()

	for {
		select {
		case log := <-s.queue:
			if log != nil {
				batch = append(batch, log)
				if len(batch) >= s.batchSize {
					s.flush(batch)
					// Reset batch dan return logs ke pool
					for _, l := range batch {
						s.entryPool.Put(l)
					}
					batch = batch[:0]
				}
			}

		case <-ticker.C:
			if len(batch) > 0 {
				s.flush(batch)
				// Return logs ke pool
				for _, l := range batch {
					s.entryPool.Put(l)
				}
				batch = batch[:0]
			}

		case <-s.stop:
			// Flush remaining logs before stopping
			if len(batch) > 0 {
				s.flush(batch)
				for _, l := range batch {
					s.entryPool.Put(l)
				}
			}
			
			// Process remaining queued items with timeout
			timeout := time.After(5 * time.Second)
			for {
				select {
				case log := <-s.queue:
					if log != nil {
						batch = append(batch, log)
						if len(batch) >= s.batchSize {
							s.flush(batch)
							for _, l := range batch {
								s.entryPool.Put(l)
							}
							batch = batch[:0]
						}
					}
				case <-timeout:
					// Final flush
					if len(batch) > 0 {
						s.flush(batch)
						for _, l := range batch {
							s.entryPool.Put(l)
						}
					}
					return
				default:
					// No more items
					if len(batch) > 0 {
						s.flush(batch)
						for _, l := range batch {
							s.entryPool.Put(l)
						}
					}
					return
				}
			}
		}
	}
}

// flush menyimpan batch audit logs ke database dengan retry
func (s *auditService) flush(batch []*entity.AuditLog) {
	if len(batch) == 0 {
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Retry logic
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		// Try to save
		err := s.saveWithContext(ctx, batch)
		if err == nil {
			s.log.Debug("Audit logs saved",
				logger.Int("batch_size", len(batch)),
				logger.Int("attempt", attempt+1),
			)
			return
		}

		s.log.Error("Failed to save audit logs",
			logger.Error(err),
			logger.Int("batch_size", len(batch)),
			logger.Int("attempt", attempt+1),
		)
	}

	// All retries failed
	s.log.Error("Failed to save audit logs after retries",
		logger.Int("batch_size", len(batch)),
	)
}

// saveWithContext save batch dengan context
func (s *auditService) saveWithContext(ctx context.Context, batch []*entity.AuditLog) error {
	done := make(chan error, 1)
	
	go func() {
		done <- s.repo.BatchCreate(batch)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}