package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/middleware"
	"github.com/atam/atamlink/internal/mod_business/dto"
	"github.com/atam/atamlink/internal/mod_business/usecase"
	"github.com/atam/atamlink/internal/service" 
	"github.com/atam/atamlink/pkg/errors"
	"github.com/atam/atamlink/pkg/utils"
)

// BusinessHandler handler untuk business endpoints
type BusinessHandler struct {
	businessUC usecase.BusinessUseCase
	uploadService service.UploadService 
	validator  *utils.Validator
}

// NewBusinessHandler membuat instance business handler baru
func NewBusinessHandler(
	businessUC usecase.BusinessUseCase,
	uploadService service.UploadService,
	validator *utils.Validator,
) *BusinessHandler {
	return &BusinessHandler{
		businessUC: businessUC,
		uploadService: uploadService,
		validator:  validator,
	}
}

// Create handler untuk create business
// @Summary Create business
// @Description Create new business
// @Tags businesses
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Business name"
// @Param slug formData string false "Business slug"
// @Param type formData string true "Business type"
// @Param logo formData file false "Logo image file (max 10MB, JPG/PNG)"
// @Success 201 {object} utils.Response{data=dto.BusinessResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /businesses [post]
func (h *BusinessHandler) Create(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Parse multipart form dengan batas maksimal 32MB
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		utils.BadRequest(c, "Gagal memproses form data")
		return
	}

	// Bind request dari form data
	var req dto.CreateBusinessRequest
	
	// Manual binding untuk multipart form
	req.Name = c.PostForm("name")
	req.Slug = c.PostForm("slug")
	req.Type = c.PostForm("type")
	
	// Handle file upload jika ada
	file, err := c.FormFile("logo")
	var logoURL *string = nil

	if err == nil && file != nil {
		// Validate file first
		if err := h.uploadService.ValidateFile(file); err != nil {
			h.handleError(c, err)
			return
		}

		// Upload file to Cloudinary
		uploadedURL, uploadErr := h.uploadService.UploadImageToCloudinary(file, "thumbnail")
		if uploadErr != nil {
			h.handleError(c, uploadErr)
			return
		}
		logoURL = &uploadedURL // Set the URL from upload
	}
	req.LogoURL = logoURL // Set the LogoURL in the request DTO

	// Validate request menggunakan validator
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Create business
	business, err := h.businessUC.Create(c, profileID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.Created(c, "Bisnis berhasil dibuat", business)
}

// List handler untuk list businesses
// @Summary List businesses
// @Description Get list of active businesses only
// @Tags businesses
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param search query string false "Search keyword"
// @Param type query string false "Business type filter"
// @Param sort query string false "Sort field" default(created_at)
// @Param order query string false "Sort order" default(desc)
// @Success 200 {object} utils.PaginatedResponse{data=[]dto.BusinessListResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /businesses [get]
func (h *BusinessHandler) List(c *gin.Context) {
	// Get profile ID from context
	profileID, _ := middleware.GetProfileID(c)

	// Get pagination params
	paginationParams := utils.GetPaginationParams(c)

	// Get filter params
	filterParams := utils.GetFilterParams(c)

	// Build filter
	filter := &dto.BusinessFilter{
		Search: filterParams.Search,
		Type:   filterParams.Type,
	}

	// Parse is_suspended filter (tetap pertahankan untuk admin)
	if isSuspendedStr := c.Query("is_suspended"); isSuspendedStr != "" {
		isSuspended, err := strconv.ParseBool(isSuspendedStr)
		if err == nil {
			filter.IsSuspended = &isSuspended
		}
	}

	// Build order by
	allowedSorts := map[string]string{
		"created_at": "b_created_at",
		"updated_at": "b_updated_at",
		"name":       "b_name",
		"type":       "b_type",
	}
	orderBy := utils.BuildOrderBy(paginationParams.Sort, paginationParams.Order, allowedSorts)

	// Get businesses
	businesses, total, err := h.businessUC.List(
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
	utils.SuccessPaginated(c, 200, "Data bisnis berhasil diambil", businesses, meta)
}

// GetByID handler untuk get business by ID
// @Summary Get business by ID
// @Description Get business details by ID
// @Tags businesses
// @Accept json
// @Produce json
// @Param id path int true "Business ID"
// @Success 200 {object} utils.Response{data=dto.BusinessResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /businesses/{id} [get]
func (h *BusinessHandler) GetByID(c *gin.Context) {
	// Get profile ID from context
	profileID, _ := middleware.GetProfileID(c)

	// Get business ID from param
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID bisnis tidak valid")
		return
	}

	// Get business
	business, err := h.businessUC.GetByID(id, profileID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Data bisnis berhasil diambil", business)
}

// Update handler untuk update business
// @Summary Update business
// @Description Update business data
// @Tags businesses
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Business ID"
// @Param name formData string false "Business name"
// @Param type formData string false "Business type"
// @Param is_active formData bool false "Active status"
// @Param logo formData file false "Logo image file (max 10MB, JPG/PNG)"
// @Success 200 {object} utils.Response{data=dto.BusinessResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /businesses/{id} [put]
func (h *BusinessHandler) Update(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Get business ID from param
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID bisnis tidak valid")
		return
	}

	// Deteksi content type
	contentType := c.GetHeader("Content-Type")
	var req dto.UpdateBusinessRequest
	var logoURL *string = nil 

	// Handle berdasarkan content type
	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Parse multipart form
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			utils.BadRequest(c, "Gagal memproses form data")
			return
		}

		// Manual binding untuk multipart form
		req.Name = c.PostForm("name")
		req.Type = c.PostForm("type")
		
		// Handle is_active boolean
		if isActiveStr := c.PostForm("is_active"); isActiveStr != "" {
			isActive, err := strconv.ParseBool(isActiveStr)
			if err != nil {
				utils.BadRequest(c, "Field 'is_active' harus berupa boolean")
				return
			}
			req.IsActive = &isActive
		}
		
		// Handle file upload jika ada
		file, err := c.FormFile("logo")
		if err == nil && file != nil {
			// Validate file first
			if err := h.uploadService.ValidateFile(file); err != nil {
				h.handleError(c, err)
				return
			}

			// Upload file to Cloudinary
			uploadedURL, uploadErr := h.uploadService.UploadImageToCloudinary(file, "thumbnail")
			if uploadErr != nil {
				h.handleError(c, uploadErr)
				return
			}
			logoURL = &uploadedURL // Set the URL from upload
		}
	} else {
		// Handle JSON request
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.BadRequest(c, constant.ErrMsgBadRequest)
			return
		}
		// If JSON, req.LogoURL will be directly parsed from JSON body if present
		// No file upload logic here
	}

	req.LogoURL = logoURL // Set the LogoURL in the request DTO (from multipart or potentially null)

	// Validate request jika ada field yang diisi
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Update business
	business, err := h.businessUC.Update(c, id, profileID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Bisnis berhasil diperbarui", business)
}

// Delete handler untuk delete business
// @Summary Delete business
// @Description Soft delete business
// @Tags businesses
// @Accept json
// @Produce json
// @Param id path int true "Business ID"
// @Success 204 {object} nil
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /businesses/{id} [delete]
func (h *BusinessHandler) Delete(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Get business ID from param
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID bisnis tidak valid")
		return
	}

	// Delete business
	if err := h.businessUC.Delete(c, id, profileID); err != nil {
		h.handleError(c, err)
		return
	}

	utils.NoContent(c)
}

// AddUser handler untuk add user to business
// @Summary Add user to business
// @Description Add user as member of business
// @Tags businesses
// @Accept json
// @Produce json
// @Param id path int true "Business ID"
// @Param body body dto.AddUserRequest true "User data"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /businesses/{id}/users [post]
func (h *BusinessHandler) AddUser(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Get business ID from param
	businessID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID bisnis tidak valid")
		return
	}

	// Bind request
	var req dto.AddUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	// Validate request
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Add user - pass context untuk audit
	if err := h.businessUC.AddUser(c, businessID, profileID, &req); err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "User berhasil ditambahkan", nil)
}

// UpdateUserRole handler untuk update user role
// @Summary Update user role
// @Description Update user role in business
// @Tags businesses
// @Accept json
// @Produce json
// @Param id path int true "Business ID"
// @Param user_id path int true "User Profile ID"
// @Param body body dto.UpdateUserRoleRequest true "Role data"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /businesses/{id}/users/{user_id} [put]
func (h *BusinessHandler) UpdateUserRole(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Get business ID from param
	businessID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID bisnis tidak valid")
		return
	}

	// Get target user ID from param
	targetProfileID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID user tidak valid")
		return
	}

	// Bind request
	var req dto.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	// Validate request
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Update user role - pass context untuk audit
	if err := h.businessUC.UpdateUserRole(c, businessID, profileID, targetProfileID, req.Role); err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Role user berhasil diperbarui", nil)
}

// RemoveUser handler untuk remove user from business
// @Summary Remove user from business
// @Description Remove user from business
// @Tags businesses
// @Accept json
// @Produce json
// @Param id path int true "Business ID"
// @Param user_id path int true "User Profile ID"
// @Success 204 {object} nil
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /businesses/{id}/users/{user_id} [delete]
func (h *BusinessHandler) RemoveUser(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Get business ID from param
	businessID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID bisnis tidak valid")
		return
	}

	// Get target user ID from param
	targetProfileID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID user tidak valid")
		return
	}

	// Remove user - pass context untuk audit
	if err := h.businessUC.RemoveUser(c, businessID, profileID, targetProfileID); err != nil {
		h.handleError(c, err)
		return
	}

	utils.NoContent(c)
}

// CreateInvite handler untuk create invite
// @Summary Create business invite
// @Description Create invite link for business
// @Tags businesses
// @Accept json
// @Produce json
// @Param id path int true "Business ID"
// @Param body body dto.CreateInviteRequest true "Invite data"
// @Success 201 {object} utils.Response{data=dto.InviteResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /businesses/{id}/invites [post]
func (h *BusinessHandler) CreateInvite(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Get business ID from param
	businessID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "ID bisnis tidak valid")
		return
	}

	// Bind request
	var req dto.CreateInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	// Validate request
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Create invite - pass context untuk audit
	invite, err := h.businessUC.CreateInvite(c, businessID, profileID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.Created(c, "Invite berhasil dibuat", invite)
}

// AcceptInvite handler untuk accept invite
// @Summary Accept business invite
// @Description Accept invite to join business
// @Tags businesses
// @Accept json
// @Produce json
// @Param body body dto.AcceptInviteRequest true "Accept invite data"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /businesses/invites/accept [post]
func (h *BusinessHandler) AcceptInvite(c *gin.Context) {
	// Get profile ID from context
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	// Bind request
	var req dto.AcceptInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	// Set profile ID from context
	req.ProfileID = profileID

	// Validate request
	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Accept invite - pass context untuk audit
	if err := h.businessUC.AcceptInvite(c, &req); err != nil {
		h.handleError(c, err)
		return
	}

	utils.OK(c, "Berhasil bergabung ke bisnis", nil)
}

// NEW: ActivateSubscription handler untuk mengaktifkan langganan bisnis (bypass pembayaran)
// @Summary Activate business subscription (Bypass Payment)
// @Description Activates a business subscription immediately for testing/development without payment.
// @Tags businesses
// @Accept json
// @Produce json
// @Param body body dto.ActivateSubscriptionRequest true "Activation data"
// @Success 201 {object} utils.Response{data=dto.SubscriptionResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /businesses/subscriptions/activate [post]
func (h *BusinessHandler) ActivateSubscription(c *gin.Context) {
	profileID, exists := middleware.GetProfileID(c)
	if !exists {
		utils.Unauthorized(c, constant.ErrMsgUnauthorized)
		return
	}

	var req dto.ActivateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, constant.ErrMsgBadRequest)
		return
	}

	if errors := h.validator.Validate(req); len(errors) > 0 {
		utils.ValidationError(c, errors)
		return
	}

	// Activate subscription - pass context untuk audit
	subscription, err := h.businessUC.ActivateSubscription(c, profileID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	utils.Created(c, "Langganan berhasil diaktifkan", subscription)
}


// handleError menangani error dari use case
func (h *BusinessHandler) handleError(c *gin.Context, err error) {
	// Check if AppError
	if appErr, ok := err.(*errors.AppError); ok {
		utils.Error(c, appErr.StatusCode, appErr.Message)
		return
	}

	// Check known errors
	switch {
	case errors.Is(err, errors.ErrBusinessNotFound):
		utils.NotFound(c, constant.ErrMsgBusinessNotFound)
	case errors.Is(err, errors.ErrDuplicateSlug):
		utils.Conflict(c, constant.ErrMsgBusinessSlugExists)
	case errors.Is(err, errors.ErrForbidden):
		utils.Forbidden(c, constant.ErrMsgForbidden)
	case errors.Is(err, errors.ErrValidation):
		utils.BadRequest(c, err.Error())
	case errors.Is(err, errors.ErrFileTooLarge): 
		utils.Error(c, 413, constant.ErrMsgFileTooLarge)
	case errors.Is(err, errors.ErrInvalidFileType):
		utils.BadRequest(c, constant.ErrMsgFileTypeInvalid)
	case errors.Is(err, errors.ErrPlanNotFound): // Handle plan not found
		utils.NotFound(c, constant.ErrMsgPlanNotFound)
	case errors.Is(err, errors.ErrInvalidPlan): // Handle invalid plan
		utils.BadRequest(c, constant.ErrMsgInvalidPlan)
	case errors.Is(err, errors.ErrConflict): // Handle conflicts like existing active sub
		utils.Conflict(c, err.Error())
	default:
		utils.InternalServerError(c, constant.ErrMsgInternalServer)
	}
}