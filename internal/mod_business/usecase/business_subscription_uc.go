package usecase

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/middleware"
	"github.com/atam/atamlink/internal/mod_business/dto"
	"github.com/atam/atamlink/internal/mod_business/entity"
	"github.com/atam/atamlink/pkg/errors"
)

// ActivateSubscription mengaktifkan langganan untuk bisnis (bypass pembayaran)
func (uc *businessUseCase) ActivateSubscription(ctx *gin.Context, profileID int64, req *dto.ActivateSubscriptionRequest) (*dto.SubscriptionResponse, error) {
	// Check business existence and user permission
	err := uc.checkBusinessPermission(req.BusinessID, profileID, constant.PermSubscriptionUpdate)
	if err != nil {
		return nil, err
	}

	// Get Master Plan details
	plan, err := uc.masterRepo.GetPlanByID(req.PlanID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get plan details")
	}
	if !plan.IsActive {
		return nil, errors.New(errors.ErrInvalidPlan, constant.ErrMsgPlanInactive, 400)
	}

	// Check if business already has an active subscription
	existingSub, err := uc.businessRepo.GetActiveSubscription(req.BusinessID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "failed to check existing subscription")
	}
	if existingSub != nil {
		return nil, errors.New(errors.ErrConflict, "Bisnis sudah memiliki langganan aktif", 409)
	}

	// Calculate subscription period
	now := time.Now()
	// GetDurationDays returns int, convert to time.Duration
	durationDays := time.Duration(plan.GetDurationDays()) * 24 * time.Hour

	newSubscription := &entity.BusinessSubscription{
		BusinessID: req.BusinessID,
		PlanID:     plan.ID,
		Status:     constant.SubscriptionStatusActive, // Bypass payment, set directly to active
		StartsAt:   now,
		ExpiresAt:  now.Add(durationDays),
		CreatedAt:  now,
	}

	// Start transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Create subscription
	if err := uc.businessRepo.CreateSubscription(tx, newSubscription); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	// Fetch the newly created subscription with full details for response
	createdSub, err := uc.businessRepo.GetActiveSubscription(newSubscription.BusinessID)
	if err != nil {
		return nil, err
	}

	// Set audit data untuk SUBSCRIPTION_ACTIVATE
	if ctx != nil {
		middleware.SetAuditBusinessID(ctx, req.BusinessID)
		middleware.SetAuditRecordID(ctx, fmt.Sprintf("%d", createdSub.ID))
		middleware.SetAuditNewData(ctx, map[string]interface{}{
			"plan_id":    createdSub.PlanID,
			"plan_name":  createdSub.Plan.Name,
			"price":      createdSub.Plan.Price,
			"starts_at":  createdSub.StartsAt,
			"expires_at": createdSub.ExpiresAt,
			"status":     createdSub.Status,
		})
	}

	return &dto.SubscriptionResponse{
		ID:        createdSub.ID,
		PlanID:    createdSub.PlanID,
		PlanName:  createdSub.Plan.Name,
		Status:    createdSub.Status,
		StartsAt:  createdSub.StartsAt,
		ExpiresAt: createdSub.ExpiresAt,
		CreatedAt: createdSub.CreatedAt,
		Plan: &dto.PlanResponse{
			ID:       createdSub.Plan.ID,
			Name:     createdSub.Plan.Name,
			Price:    createdSub.Plan.Price,
			Features: createdSub.Plan.Features,
		},
	}, nil
}