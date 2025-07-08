package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/middleware"
	"github.com/atam/atamlink/internal/mod_user/dto"
	"github.com/atam/atamlink/internal/mod_user/usecase"
	"github.com/atam/atamlink/pkg/errors"
	"github.com/atam/atamlink/pkg/utils"
)

// UserHandler handler untuk user endpoints
type UserHandler struct {
	userUC    usecase.UserUseCase
	validator *utils.Validator
}

// NewUserHandler membuat instance user handler baru
func NewUserHandler(userUC usecase.UserUseCase, validator *utils.Validator) *UserHandler {
	return &UserHandler{
		userUC:    userUC,
		validator: validator,
	}
}

// GetProfile handler untuk get current user profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Get profile
	profile, err := h.userUC.GetProfileByID(profileID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Profile berhasil diambil", profile)
}

// GetProfileByID handler untuk get profile by ID
func (h *UserHandler) GetProfileByID(c *gin.Context) {
	// Get profile ID from param
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID profile tidak valid")
		return
	}

	// Get profile
	profile, err := h.userUC.GetProfileByID(id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Profile berhasil diambil", profile)
}

// CreateProfile handler untuk create profile
func (h *UserHandler) CreateProfile(c *gin.Context) {
	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Bind request
	var req dto.CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	// Validate request
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Create profile
	profile, err := h.userUC.CreateProfile(userID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.Created(c, "Profile berhasil dibuat", profile)
}

// UpdateProfile handler untuk update profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Bind request
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	// Validate request
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Update profile
	profile, err := h.userUC.UpdateProfile(c, profileID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Profile berhasil diperbarui", profile)
}

// UpdateProfileByID handler untuk update profile by ID (admin only)
func (h *UserHandler) UpdateProfileByID(c *gin.Context) {
	// Get profile ID from param
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID profile tidak valid")
		return
	}

	// Bind request
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	// Validate request
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Update profile
	profile, err := h.userUC.UpdateProfile(c, id, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Profile berhasil diperbarui", profile)
}

// DeleteProfile handler untuk delete profile
func (h *UserHandler) DeleteProfile(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Delete profile
	if err := h.userUC.DeleteProfile(c, profileID); err != nil {
		h.handleError(c, err)
		return
	}

	utils.NoContent(c)
}

// DeleteProfileByID handler untuk delete profile by ID (admin only)
func (h *UserHandler) DeleteProfileByID(c *gin.Context) {
	// Get profile ID from param
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID profile tidak valid")
		return
	}

	// Delete profile
	if err := h.userUC.DeleteProfile(c, id); err != nil {
		h.handleError(c, err)
		return
	}

	utils.NoContent(c)
}

// handleError menangani error dari use case
func (h *UserHandler) handleError(c *gin.Context, err error) {
	// Check if AppError
	if appErr, ok := err.(*errors.AppError); ok {
		utils.Error(c, appErr.StatusCode, appErr.Message)
		return
	}

	// Check known errors
	switch {
	case errors.Is(err, errors.ErrNotFound):
		utils.NotFound(c, constant.ErrMsgProfileNotFound)
	case errors.Is(err, errors.ErrConflict):
		utils.Conflict(c, err.Error())
	case errors.Is(err, errors.ErrAccountInactive):
		utils.Forbidden(c, err.Error())
	case errors.Is(err, errors.ErrValidation):
		utils.BadRequest(c, err.Error())
	default:
		utils.InternalServerError(c, constant.ErrMsgInternalServer)
	}
}