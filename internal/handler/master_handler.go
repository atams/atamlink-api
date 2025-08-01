package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/mod_master/dto"
	"github.com/atam/atamlink/internal/mod_master/usecase"
	"github.com/atam/atamlink/pkg/errors"
	"github.com/atam/atamlink/pkg/utils"
)

// MasterHandler handler untuk master data endpoints
type MasterHandler struct {
	masterUC  usecase.MasterUseCase
	validator *utils.Validator
}

// NewMasterHandler membuat instance master handler baru
func NewMasterHandler(masterUC usecase.MasterUseCase, validator *utils.Validator) *MasterHandler {
	return &MasterHandler{
		masterUC:  masterUC,
		validator: validator,
	}
}

// CreatePlan handler untuk create plan
// @Summary Create plan
// @Description Create new subscription plan
// @Tags masters
// @Accept json
// @Produce json
// @Param body body dto.CreatePlanRequest true "Plan data"
// @Success 201 {object} utils.Response{data=dto.PlanResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /masters/plans [post]
func (h *MasterHandler) CreatePlan(c *gin.Context) {
	// Bind request
	var req dto.CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	// Validate request
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Create plan
	plan, err := h.masterUC.CreatePlan(&req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.Created(c, "Plan berhasil dibuat", plan)
}

// UpdatePlan handler untuk update plan
// @Summary Update plan
// @Description Update existing subscription plan
// @Tags masters
// @Accept json
// @Produce json
// @Param id path int true "Plan ID"
// @Param body body dto.UpdatePlanRequest true "Update data"
// @Success 200 {object} utils.Response{data=dto.PlanResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /masters/plans/{id} [put]
func (h *MasterHandler) UpdatePlan(c *gin.Context) {
	// Get plan ID from param
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID plan tidak valid")
		return
	}

	// Bind request
	var req dto.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	// Validate request
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Update plan
	plan, err := h.masterUC.UpdatePlan(c, id, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Plan berhasil diperbarui", plan)
}

// DeletePlan handler untuk delete plan
// @Summary Delete plan
// @Description Soft delete subscription plan
// @Tags masters
// @Accept json
// @Produce json
// @Param id path int true "Plan ID"
// @Success 204 {object} nil
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /masters/plans/{id} [delete]
func (h *MasterHandler) DeletePlan(c *gin.Context) {
	// Get plan ID from param
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID plan tidak valid")
		return
	}

	// Delete plan
	if err := h.masterUC.DeletePlan(c, id); err != nil {
		h.handleError(c, err)
		return
	}

	utils.NoContent(c)
}

// ListPlans handler untuk list plans
// @Summary List plans
// @Description Get list of subscription plans
// @Tags masters
// @Accept json
// @Produce json
// @Param is_active query bool false "Active status filter"
// @Param is_free query bool false "Free plan filter"
// @Param max_price query int false "Maximum price filter"
// @Param min_price query int false "Minimum price filter"
// @Success 200 {object} utils.Response{data=[]dto.PlanListResponse}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /masters/plans [get]
func (h *MasterHandler) ListPlans(c *gin.Context) {
	// Build filter
	filter := &dto.PlanFilter{}

	// Parse is_active filter
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err == nil {
			filter.IsActive = &isActive
		}
	}

	// Parse is_free filter
	if isFreeStr := c.Query("is_free"); isFreeStr != "" {
		isFree, err := strconv.ParseBool(isFreeStr)
		if err == nil {
			filter.IsFree = &isFree
		}
	}

	// Parse max_price filter
	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		maxPrice, err := strconv.Atoi(maxPriceStr)
		if err == nil && maxPrice >= 0 {
			filter.MaxPrice = &maxPrice
		}
	}

	// Parse min_price filter
	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		minPrice, err := strconv.Atoi(minPriceStr)
		if err == nil && minPrice >= 0 {
			filter.MinPrice = &minPrice
		}
	}

	// Get plans
	plans, err := h.masterUC.ListPlans(filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Data plan berhasil diambil", plans)
}

// GetPlanByID handler untuk get plan by ID
// @Summary Get plan by ID
// @Description Get plan details by ID
// @Tags masters
// @Accept json
// @Produce json
// @Param id path int true "Plan ID"
// @Success 200 {object} utils.Response{data=dto.PlanResponse}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /masters/plans/{id} [get]
func (h *MasterHandler) GetPlanByID(c *gin.Context) {
	// Get plan ID from param
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID plan tidak valid")
		return
	}

	// Get plan
	plan, err := h.masterUC.GetPlanByID(id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Data plan berhasil diambil", plan)
}

// CreateTheme handler untuk create theme
// @Summary Create theme
// @Description Create new catalog theme
// @Tags masters
// @Accept json
// @Produce json
// @Param body body dto.CreateThemeRequest true "Theme data"
// @Success 201 {object} utils.Response{data=dto.ThemeResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /masters/themes [post]
func (h *MasterHandler) CreateTheme(c *gin.Context) {
	// Bind request
	var req dto.CreateThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	// Validate request
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Create theme
	theme, err := h.masterUC.CreateTheme(&req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.Created(c, "Theme berhasil dibuat", theme)
}

// UpdateTheme handler untuk update theme
// @Summary Update theme
// @Description Update existing catalog theme
// @Tags masters
// @Accept json
// @Produce json
// @Param id path int true "Theme ID"
// @Param body body dto.UpdateThemeRequest true "Update data"
// @Success 200 {object} utils.Response{data=dto.ThemeResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /masters/themes/{id} [put]
func (h *MasterHandler) UpdateTheme(c *gin.Context) {
	// Get theme ID from param
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID theme tidak valid")
		return
	}

	// Bind request
	var req dto.UpdateThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	// Validate request
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Update theme
	theme, err := h.masterUC.UpdateTheme(c, id, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Theme berhasil diperbarui", theme)
}

// DeleteTheme handler untuk delete theme
// @Summary Delete theme
// @Description Soft delete catalog theme
// @Tags masters
// @Accept json
// @Produce json
// @Param id path int true "Theme ID"
// @Success 204 {object} nil
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /masters/themes/{id} [delete]
func (h *MasterHandler) DeleteTheme(c *gin.Context) {
	// Get theme ID from param
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID theme tidak valid")
		return
	}

	// Delete theme
	if err := h.masterUC.DeleteTheme(c, id); err != nil {
		h.handleError(c, err)
		return
	}

	utils.NoContent(c)
}

// ListThemes handler untuk list themes
// @Summary List themes
// @Description Get list of catalog themes
// @Tags masters
// @Accept json
// @Produce json
// @Param search query string false "Search keyword"
// @Param type query string false "Theme type filter"
// @Param is_active query bool false "Active status filter"
// @Param is_premium query bool false "Premium theme filter"
// @Success 200 {object} utils.Response{data=[]dto.ThemeListResponse}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /masters/themes [get]
func (h *MasterHandler) ListThemes(c *gin.Context) {
	// Build filter
	filter := &dto.ThemeFilter{
		Search: c.Query("search"),
		Type:   c.Query("type"),
	}

	// Parse is_active filter
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err == nil {
			filter.IsActive = &isActive
		}
	}

	// Parse is_premium filter
	if isPremiumStr := c.Query("is_premium"); isPremiumStr != "" {
		isPremium, err := strconv.ParseBool(isPremiumStr)
		if err == nil {
			filter.IsPremium = &isPremium
		}
	}

	// Validate theme type if provided
	if filter.Type != "" && !constant.IsValidThemeType(filter.Type) {
		utils.BadRequest(c, "Tipe theme tidak valid")
		return
	}

	// Get themes
	themes, err := h.masterUC.ListThemes(filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Data theme berhasil diambil", themes)
}

// GetThemeByID handler untuk get theme by ID
// @Summary Get theme by ID
// @Description Get theme details by ID
// @Tags masters
// @Accept json
// @Produce json
// @Param id path int true "Theme ID"
// @Success 200 {object} utils.Response{data=dto.ThemeResponse}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /masters/themes/{id} [get]
func (h *MasterHandler) GetThemeByID(c *gin.Context) {
	// Get theme ID from param
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID theme tidak valid")
		return
	}

	// Get theme
	theme, err := h.masterUC.GetThemeByID(id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Data theme berhasil diambil", theme)
}

// handleError menangani error dari use case
func (h *MasterHandler) handleError(c *gin.Context, err error) {
	// Check if AppError
	if appErr, ok := err.(*errors.AppError); ok {
		utils.Error(c, appErr.StatusCode, appErr.Message)
		return
	}

	// Check known errors
	switch {
	case errors.Is(err, errors.ErrPlanNotFound):
		utils.NotFound(c, constant.ErrMsgPlanNotFound)
	case errors.Is(err, errors.ErrNotFound):
		utils.NotFound(c, constant.ErrMsgNotFound)
	case errors.Is(err, errors.ErrConflict):
		utils.Conflict(c, err.Error())
	case errors.Is(err, errors.ErrValidation):
		utils.BadRequest(c, err.Error())
	default:
		utils.InternalServerError(c, constant.ErrMsgInternalServer)
	}
}