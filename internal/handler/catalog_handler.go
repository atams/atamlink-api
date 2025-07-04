package handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/middleware"
	"github.com/atam/atamlink/internal/mod_catalog/dto"
	"github.com/atam/atamlink/internal/mod_catalog/usecase"
	"github.com/atam/atamlink/internal/service"
	"github.com/atam/atamlink/pkg/errors"
	"github.com/atam/atamlink/pkg/utils"
)

// CatalogHandler handler untuk catalog endpoints
type CatalogHandler struct {
	catalogUC     usecase.CatalogUseCase
	uploadService service.UploadService
	validator     *utils.Validator
}

// NewCatalogHandler membuat instance catalog handler baru
func NewCatalogHandler(
	catalogUC usecase.CatalogUseCase,
	uploadService service.UploadService,
	validator *utils.Validator,
) *CatalogHandler {
	return &CatalogHandler{
		catalogUC:     catalogUC,
		uploadService: uploadService,
		validator:     validator,
	}
}

// Create handler untuk create catalog
// @Summary Create catalog
// @Description Create new catalog
// @Tags catalogs
// @Accept json
// @Produce json
// @Param body body dto.CreateCatalogRequest true "Catalog data"
// @Success 201 {object} utils.Response{data=dto.CatalogResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /catalogs [post]
func (h *CatalogHandler) Create(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Bind request
	var req dto.CreateCatalogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	// Validate request
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Create catalog
	catalog, err := h.catalogUC.Create(profileID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.Created(c, "Katalog berhasil dibuat", catalog)
}

// List handler untuk list catalogs
// @Summary List catalogs
// @Description Get list of catalogs
// @Tags catalogs
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param search query string false "Search keyword"
// @Param business_id query int false "Business ID filter"
// @Param theme_id query int false "Theme ID filter"
// @Param is_active query bool false "Active status filter"
// @Param sort query string false "Sort field" default(created_at)
// @Param order query string false "Sort order" default(desc)
// @Success 200 {object} utils.PaginatedResponse{data=[]dto.CatalogListResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /catalogs [get]
func (h *CatalogHandler) List(c *gin.Context) {
	// Get profile ID from context
	profileID, _ := middleware.GetProfileID(c)

	// Get pagination params
	paginationParams := utils.GetPaginationParams(c)

	// Get filter params
	filterParams := utils.GetFilterParams(c)

	// Build filter
	filter := &dto.CatalogFilter{
		Search: filterParams.Search,
	}

	// Parse business_id filter
	if businessIDStr := c.Query("business_id"); businessIDStr != "" {
		businessID, err := strconv.ParseInt(businessIDStr, 10, 64)
		if err == nil {
			filter.BusinessID = businessID
		}
	}

	// Parse theme_id filter
	if themeIDStr := c.Query("theme_id"); themeIDStr != "" {
		themeID, err := strconv.ParseInt(themeIDStr, 10, 64)
		if err == nil {
			filter.ThemeID = themeID
		}
	}

	// Parse is_active filter
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err == nil {
			filter.IsActive = &isActive
		}
	}

	// Build order by
	allowedSorts := map[string]string{
		"created_at": "c_created_at",
		"updated_at": "c_updated_at",
		"title":      "c_title",
	}
	orderBy := utils.BuildOrderBy(paginationParams.Sort, paginationParams.Order, allowedSorts)

	// Get catalogs
	catalogs, total, err := h.catalogUC.List(
		profileID,
		filter,
		paginationParams.Page,
		paginationParams.PerPage,
		orderBy,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Return paginated response
	meta := utils.GetPaginationMeta(paginationParams.Page, paginationParams.PerPage, total)
	utils.SuccessPaginated(c, 200, "Data katalog berhasil diambil", catalogs, meta)
}

// GetByID handler untuk get catalog by ID
// @Summary Get catalog by ID
// @Description Get catalog details by ID
// @Tags catalogs
// @Accept json
// @Produce json
// @Param id path int true "Catalog ID"
// @Success 200 {object} utils.Response{data=dto.CatalogResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /catalogs/{id} [get]
func (h *CatalogHandler) GetByID(c *gin.Context) {
	// Get profile ID from context
	profileID, _ := middleware.GetProfileID(c)

	// Get catalog ID from param
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID katalog tidak valid")
		return
	}

	// Get catalog
	catalog, err := h.catalogUC.GetByID(id, profileID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Data katalog berhasil diambil", catalog)
}

// Update handler untuk update catalog
// @Summary Update catalog
// @Description Update catalog data
// @Tags catalogs
// @Accept json
// @Produce json
// @Param id path int true "Catalog ID"
// @Param body body dto.UpdateCatalogRequest true "Update data"
// @Success 200 {object} utils.Response{data=dto.CatalogResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /catalogs/{id} [put]
func (h *CatalogHandler) Update(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Get catalog ID from param
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID katalog tidak valid")
		return
	}

	// Bind request
	var req dto.UpdateCatalogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	// Validate request
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Update catalog
	catalog, err := h.catalogUC.Update(id, profileID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Katalog berhasil diperbarui", catalog)
}

// Delete handler untuk delete catalog
// @Summary Delete catalog
// @Description Soft delete catalog
// @Tags catalogs
// @Accept json
// @Produce json
// @Param id path int true "Catalog ID"
// @Success 204 {object} nil
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /catalogs/{id} [delete]
func (h *CatalogHandler) Delete(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Get catalog ID from param
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID katalog tidak valid")
		return
	}

	// Delete catalog
	if err := h.catalogUC.Delete(id, profileID); err != nil {
		h.handleError(c, err)
		return
	}

	utils.NoContent(c)
}

// GetPublicCatalog handler untuk get public catalog by slug
// @Summary Get public catalog
// @Description Get public catalog by slug
// @Tags catalogs
// @Accept json
// @Produce json
// @Param slug path string true "Catalog slug"
// @Success 200 {object} utils.Response{data=dto.PublicCatalogResponse}
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /c/{slug} [get]
func (h *CatalogHandler) GetPublicCatalog(c *gin.Context) {
	// Get slug from param
	slug := c.Param("slug")
	if slug == "" {
		utils.BadRequest(c, "Slug katalog tidak valid")
		return
	}

	// Get public catalog
	catalog, err := h.catalogUC.GetBySlug(slug)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Data katalog berhasil diambil", catalog)
}

// CreateSection handler untuk create section
// @Summary Create catalog section
// @Description Create new section in catalog
// @Tags catalogs
// @Accept json
// @Produce json
// @Param id path int true "Catalog ID"
// @Param body body dto.CreateSectionRequest true "Section data"
// @Success 201 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /catalogs/{id}/sections [post]
func (h *CatalogHandler) CreateSection(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Get catalog ID from param
	catalogID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID katalog tidak valid")
		return
	}

	// Bind request
	var req dto.CreateSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	// Validate request
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Create section
	if err := h.catalogUC.CreateSection(catalogID, profileID, &req); err != nil {
		h.handleError(c, err)
		return
	}

	utils.Created(c, "Section berhasil dibuat", nil)
}

// UpdateSection handler untuk update section
// @Summary Update catalog section
// @Description Update section data
// @Tags catalogs
// @Accept json
// @Produce json
// @Param section_id path int true "Section ID"
// @Param body body dto.UpdateSectionRequest true "Update data"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /catalogs/sections/{section_id} [put]
func (h *CatalogHandler) UpdateSection(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Get section ID from param
	sectionID, err := strconv.ParseInt(c.Param("section_id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID section tidak valid")
		return
	}

	// Bind request
	var req dto.UpdateSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	// Validate request
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Update section
	if err := h.catalogUC.UpdateSection(sectionID, profileID, &req); err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Section berhasil diperbarui", nil)
}

// DeleteSection handler untuk delete section
// @Summary Delete catalog section
// @Description Delete section from catalog
// @Tags catalogs
// @Accept json
// @Produce json
// @Param section_id path int true "Section ID"
// @Success 204 {object} nil
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /catalogs/sections/{section_id} [delete]
func (h *CatalogHandler) DeleteSection(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Get section ID from param
	sectionID, err := strconv.ParseInt(c.Param("section_id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID section tidak valid")
		return
	}

	// Delete section
	if err := h.catalogUC.DeleteSection(sectionID, profileID); err != nil {
		h.handleError(c, err)
		return
	}

	utils.NoContent(c)
}

// CreateCard handler untuk create card
// @Summary Create catalog card
// @Description Create new card in section
// @Tags catalogs
// @Accept json
// @Produce json
// @Param section_id path int true "Section ID"
// @Param body body dto.CreateCardRequest true "Card data"
// @Success 201 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /catalogs/sections/{section_id}/cards [post]
func (h *CatalogHandler) CreateCard(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Get section ID from param
	sectionID, err := strconv.ParseInt(c.Param("section_id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID section tidak valid")
		return
	}

	// Bind request
	var req dto.CreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	// Validate request
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Create card
	if err := h.catalogUC.CreateCard(sectionID, profileID, &req); err != nil {
		h.handleError(c, err)
		return
	}

	utils.Created(c, "Card berhasil dibuat", nil)
}

// UpdateCard handler untuk update card
// @Summary Update catalog card
// @Description Update card data
// @Tags catalogs
// @Accept json
// @Produce json
// @Param card_id path int true "Card ID"
// @Param body body dto.UpdateCardRequest true "Update data"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /catalogs/cards/{card_id} [put]
func (h *CatalogHandler) UpdateCard(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Get card ID from param
	cardID, err := strconv.ParseInt(c.Param("card_id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID card tidak valid")
		return
	}

	// Bind request
	var req dto.UpdateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	// Validate request
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Update card
	if err := h.catalogUC.UpdateCard(cardID, profileID, &req); err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Card berhasil diperbarui", nil)
}

// DeleteCard handler untuk delete card
// @Summary Delete catalog card
// @Description Delete card from section
// @Tags catalogs
// @Accept json
// @Produce json
// @Param card_id path int true "Card ID"
// @Success 204 {object} nil
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /catalogs/cards/{card_id} [delete]
func (h *CatalogHandler) DeleteCard(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Get card ID from param
	cardID, err := strconv.ParseInt(c.Param("card_id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID card tidak valid")
		return
	}

	// Delete card
	if err := h.catalogUC.DeleteCard(cardID, profileID); err != nil {
		h.handleError(c, err)
		return
	}

	utils.NoContent(c)
}

// UploadCardImage handler untuk upload card image
// @Summary Upload card image
// @Description Upload image for catalog card
// @Tags catalogs
// @Accept multipart/form-data
// @Produce json
// @Param card_id path int true "Card ID"
// @Param file formData file true "Image file"
// @Success 200 {object} utils.Response{data=service.FileInfo}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 413 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /catalogs/cards/{card_id}/images [post]
func (h *CatalogHandler) UploadCardImage(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Get card ID from param
	cardID, err := strconv.ParseInt(c.Param("card_id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID card tidak valid")
		return
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, constant.ErrMsgFileRequired)
		return
	}

	// Validate file
	if err := h.uploadService.ValidateFile(file); err != nil {
		h.handleError(c, err)
		return
	}

	// Upload file
	folder := fmt.Sprintf("catalogs/cards/%d", cardID)
	fileInfo, err := service.ProcessUpload(
		file,
		folder,
		h.uploadService,
		c.Request.Host,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// TODO: Save media URL to database

	utils.OK(c, "Gambar berhasil diupload", fileInfo)
}

// handleError menangani error dari use case
func (h *CatalogHandler) handleError(c *gin.Context, err error) {
	// Check if AppError
	if appErr, ok := err.(*errors.AppError); ok {
		utils.Error(c, appErr.StatusCode, appErr.Message)
		return
	}

	// Check known errors
	switch {
	case errors.Is(err, errors.ErrCatalogNotFound):
		utils.NotFound(c, constant.ErrMsgCatalogNotFound)
	case errors.Is(err, errors.ErrBusinessNotFound):
		utils.NotFound(c, constant.ErrMsgBusinessNotFound)
	case errors.Is(err, errors.ErrSectionNotFound):
		utils.NotFound(c, constant.ErrMsgSectionNotFound)
	case errors.Is(err, errors.ErrCardNotFound):
		utils.NotFound(c, constant.ErrMsgCardNotFound)
	case errors.Is(err, errors.ErrDuplicateSlug):
		utils.Conflict(c, constant.ErrMsgCatalogSlugExists)
	case errors.Is(err, errors.ErrForbidden):
		utils.Forbidden(c, constant.ErrMsgForbidden)
	case errors.Is(err, errors.ErrValidation):
		utils.BadRequest(c, err.Error())
	case errors.Is(err, errors.ErrFileTooLarge):
		utils.Error(c, 413, constant.ErrMsgFileTooLarge)
	case errors.Is(err, errors.ErrInvalidFileType):
		utils.BadRequest(c, constant.ErrMsgFileTypeInvalid)
	default:
		utils.InternalServerError(c, constant.ErrMsgInternalServer)
	}
}