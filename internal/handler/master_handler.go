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
	default:
		utils.InternalServerError(c, constant.ErrMsgInternalServer)
	}
}