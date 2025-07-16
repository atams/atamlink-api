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

type auditService struct {
	repo      repository.AuditRepository
	log       logger.Logger
	queue     chan *entity.AuditLog
	batchSize int
	flushTime time.Duration
	wg        sync.WaitGroup
	stop      chan bool
}

// NewAuditService membuat instance audit service baru
func NewAuditService(repo repository.AuditRepository, log logger.Logger) AuditService {
	return &auditService{
		repo:      repo,
		log:       log,
		queue:     make(chan *entity.AuditLog, 1000),
		batchSize: 10,
		flushTime: 5 * time.Second,
		stop:      make(chan bool),
	}
}

// Start memulai audit service worker
func (s *auditService) Start() {
	s.wg.Add(1)
	go s.worker()
}

// Stop menghentikan audit service
func (s *auditService) Stop() {
	close(s.stop)
	s.wg.Wait()
	close(s.queue)
}

// Log menambahkan entry ke queue
func (s *auditService) Log(entry *AuditEntry) {
	// Convert data to JSON
	// var oldDataJSON, newDataJSON json.RawMessage
	
	// if entry.OldData != nil {
	// 	data, err := json.Marshal(entry.OldData)
	// 	if err != nil {
	// 		s.log.Error("Failed to marshal old data", logger.Error(err))
	// 		return
	// 	}
	// 	oldDataJSON = data
	// }
	
	// if entry.NewData != nil {
	// 	data, err := json.Marshal(entry.NewData)
	// 	if err != nil {
	// 		s.log.Error("Failed to marshal new data", logger.Error(err))
	// 		return
	// 	}
	// 	newDataJSON = data
	// }

	// Create audit log
	auditLog := &entity.AuditLog{
		Timestamp:     time.Now(),
		UserProfileID: entry.UserProfileID,
		BusinessID:    entry.BusinessID,
		Action:        entry.Action,
		Table:     	   entry.Table,
		RecordID:      entry.RecordID,
		// OldData:       oldDataJSON,
		// NewData:       newDataJSON,
		OldData:       entry.OldData,
		NewData:       entry.NewData,
		Context:       entry.Context,
		Reason:        entry.Reason,
	}

	// Non-blocking send to queue
	select {
	case s.queue <- auditLog:
		// Successfully queued
	default:
		// Queue full, log error
		s.log.Error("Audit queue full, dropping log entry",
			logger.String("action", entry.Action),
			logger.String("table", entry.Table),
			logger.String("record", entry.RecordID),
		)
	}
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
					batch = batch[:0]
				}
			}

		case <-ticker.C:
			if len(batch) > 0 {
				s.flush(batch)
				batch = batch[:0]
			}

		case <-s.stop:
			// Flush remaining logs before stopping
			if len(batch) > 0 {
				s.flush(batch)
			}
			// Process remaining queued items
			for len(s.queue) > 0 {
				select {
				case log := <-s.queue:
					batch = append(batch, log)
					if len(batch) >= s.batchSize {
						s.flush(batch)
						batch = batch[:0]
					}
				default:
					// No more items
				}
			}
			// Final flush
			if len(batch) > 0 {
				s.flush(batch)
			}
			return
		}
	}
}

// flush menyimpan batch audit logs ke database
func (s *auditService) flush(batch []*entity.AuditLog) {
	if len(batch) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use goroutine dengan context untuk timeout
	done := make(chan error, 1)
	go func() {
		done <- s.repo.BatchCreate(batch)
	}()

	select {
	case err := <-done:
		if err != nil {
			s.log.Error("Failed to save audit logs",
				logger.Error(err),
				logger.Int("batch_size", len(batch)),
			)
		} else {
			s.log.Debug("Audit logs saved",
				logger.Int("batch_size", len(batch)),
			)
		}
	case <-ctx.Done():
		s.log.Error("Audit log save timeout",
			logger.Int("batch_size", len(batch)),
		)
	}
}