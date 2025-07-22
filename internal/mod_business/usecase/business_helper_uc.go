package usecase

import (
	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/mod_business/dto"
	"github.com/atam/atamlink/internal/mod_business/entity"
	"github.com/atam/atamlink/pkg/errors"
)

// checkBusinessPermission check permission user dalam business
func (uc *businessUseCase) checkBusinessPermission(businessID, profileID int64, permission string) error {
	// Get user role in business
	user, err := uc.businessRepo.GetUserByBusinessAndProfile(businessID, profileID)
	if err != nil {
		return err
	}

	if user == nil || !user.IsActive {
		return errors.New(errors.ErrForbidden, constant.ErrMsgBusinessAccessDenied, 403)
	}

	// Check permission
	if !constant.HasPermission(user.Role, permission) {
		return errors.New(errors.ErrForbidden, "Anda tidak memiliki izin untuk aksi ini", 403)
	}

	return nil
}

// toBusinessResponse convert entity to DTO response
func (uc *businessUseCase) toBusinessResponse(business *entity.Business, users []*entity.BusinessUser, subscription *entity.BusinessSubscription) *dto.BusinessResponse {
	resp := &dto.BusinessResponse{
		ID:               business.ID,
		Slug:             business.Slug,
		Name:             business.Name,
		Type:             business.Type,
		IsActive:         business.IsActive,
		IsSuspended:      business.IsSuspended,
		SuspensionReason: business.GetSuspensionReason(),
		CreatedBy:        business.CreatedBy,
		CreatedAt:        business.CreatedAt,
		UpdatedAt:        business.UpdatedAt,
		ActivePlan:       nil, // Default null
	}

	// Add LogoURL if valid
	if business.LogoURL.Valid {
		resp.LogoURL = &business.LogoURL.String
	}

	// Add users
	if users != nil {
		resp.Users = make([]dto.BusinessUserResponse, len(users))
		for i, user := range users {
			resp.Users[i] = dto.BusinessUserResponse{
				ID:        user.ID,
				ProfileID: user.ProfileID,
				Role:      user.Role,
				IsOwner:   user.IsOwner,
				IsActive:  user.IsActive,
				JoinedAt:  user.CreatedAt,
			}

			if user.Profile != nil {
				resp.Users[i].Profile = &dto.ProfileResponse{
					ID:          user.Profile.ID,
					DisplayName: user.Profile.GetDisplayName(),
				}
			}
		}
	}

	// Add subscription info (akan null jika tidak ada)
	if subscription != nil {
		resp.ActivePlan = &dto.SubscriptionResponse{
			ID:        subscription.ID,
			PlanID:    subscription.PlanID,
			Status:    subscription.Status,
			StartsAt:  subscription.StartsAt,
			ExpiresAt: subscription.ExpiresAt,
			CreatedAt: subscription.CreatedAt,
		}

		// Add plan details if available
		if subscription.Plan != nil {
			resp.ActivePlan.PlanName = subscription.Plan.Name
			resp.ActivePlan.Plan = &dto.PlanResponse{
				ID:       subscription.Plan.ID,
				Name:     subscription.Plan.Name,
				Price:    subscription.Plan.Price,
				Features: subscription.Plan.Features,
			}
		}
	}

	return resp
}

// deepCopyBusinessForAudit creates a deep copy of business data for audit purposes
// This returns a simplified version suitable for audit logs
func (uc *businessUseCase) deepCopyBusinessForAudit(original *entity.Business) map[string]interface{} {
	if original == nil {
		return nil
	}

	data := map[string]interface{}{
		"id":           original.ID,
		"slug":         original.Slug,
		"name":         original.Name,
		"type":         original.Type,
		"is_active":    original.IsActive,
		"is_suspended": original.IsSuspended,
		"created_by":   original.CreatedBy,
		"created_at":   original.CreatedAt,
	}

	// Add optional fields if they have values
	if original.LogoURL.Valid {
		data["logo_url"] = original.LogoURL.String
	}

	if original.SuspensionReason.Valid {
		data["suspension_reason"] = original.SuspensionReason.String
	}

	if original.SuspendedBy.Valid {
		data["suspended_by"] = original.SuspendedBy.Int64
	}

	if original.SuspendedAt != nil {
		data["suspended_at"] = original.SuspendedAt
	}

	if original.UpdatedBy.Valid {
		data["updated_by"] = original.UpdatedBy.Int64
	}

	if original.UpdatedAt != nil {
		data["updated_at"] = original.UpdatedAt
	}

	return data
}