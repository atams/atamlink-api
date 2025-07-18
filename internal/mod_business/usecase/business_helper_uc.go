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

// deepCopyBusiness creates a deep copy of business entity for audit purposes
func (uc *businessUseCase) deepCopyBusiness(original *entity.Business) *entity.Business {
	if original == nil {
		return nil
	}

	copy := &entity.Business{
		ID:               original.ID,
		Slug:             original.Slug,
		Name:             original.Name,
		LogoURL:          original.LogoURL,
		Type:             original.Type,
		IsActive:         original.IsActive,
		IsSuspended:      original.IsSuspended,
		SuspensionReason: original.SuspensionReason,
		SuspendedBy:      original.SuspendedBy,
		SuspendedAt:      original.SuspendedAt,
		CreatedBy:        original.CreatedBy,
		CreatedAt:        original.CreatedAt,
		UpdatedBy:        original.UpdatedBy,
		UpdatedAt:        original.UpdatedAt,
	}

	// Deep copy time pointer if exists
	if original.UpdatedAt != nil {
		updatedAt := *original.UpdatedAt
		copy.UpdatedAt = &updatedAt
	}

	if original.SuspendedAt != nil {
		suspendedAt := *original.SuspendedAt
		copy.SuspendedAt = &suspendedAt
	}

	return copy
}