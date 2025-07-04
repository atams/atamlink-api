package usecase

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/mod_business/repository"
	"github.com/atam/atamlink/internal/mod_catalog/dto"
	"github.com/atam/atamlink/internal/mod_catalog/entity"
	catalogRepo "github.com/atam/atamlink/internal/mod_catalog/repository"
	"github.com/atam/atamlink/internal/service"
	"github.com/atam/atamlink/pkg/database"
	"github.com/atam/atamlink/pkg/errors"
)

// CatalogUseCase interface untuk catalog use case
type CatalogUseCase interface {
	Create(profileID int64, req *dto.CreateCatalogRequest) (*dto.CatalogResponse, error)
	GetByID(id int64, profileID int64) (*dto.CatalogResponse, error)
	GetBySlug(slug string) (*dto.PublicCatalogResponse, error)
	List(profileID int64, filter *dto.CatalogFilter, page, perPage int, orderBy string) ([]*dto.CatalogListResponse, int64, error)
	Update(id int64, profileID int64, req *dto.UpdateCatalogRequest) (*dto.CatalogResponse, error)
	Delete(id int64, profileID int64) error
	
	// Section management
	CreateSection(catalogID int64, profileID int64, req *dto.CreateSectionRequest) error
	UpdateSection(sectionID int64, profileID int64, req *dto.UpdateSectionRequest) error
	DeleteSection(sectionID int64, profileID int64) error
	
	// Card management
	CreateCard(sectionID int64, profileID int64, req *dto.CreateCardRequest) error
	UpdateCard(cardID int64, profileID int64, req *dto.UpdateCardRequest) error
	DeleteCard(cardID int64, profileID int64) error
}

type catalogUseCase struct {
	db           *sql.DB
	catalogRepo  catalogRepo.CatalogRepository
	businessRepo repository.BusinessRepository
	slugService  service.SlugService
}

// NewCatalogUseCase membuat instance catalog use case baru
func NewCatalogUseCase(
	db *sql.DB,
	catalogRepo catalogRepo.CatalogRepository,
	businessRepo repository.BusinessRepository,
	slugService service.SlugService,
) CatalogUseCase {
	return &catalogUseCase{
		db:           db,
		catalogRepo:  catalogRepo,
		businessRepo: businessRepo,
		slugService:  slugService,
	}
}

// Create membuat catalog baru
func (uc *catalogUseCase) Create(profileID int64, req *dto.CreateCatalogRequest) (*dto.CatalogResponse, error) {
	// Check business access
	if err := uc.checkBusinessAccess(req.BusinessID, profileID, constant.PermCatalogCreate); err != nil {
		return nil, err
	}

	// Check if business exists and active
	business, err := uc.businessRepo.GetByID(req.BusinessID)
	if err != nil {
		return nil, err
	}
	if !business.IsAccessible() {
		return nil, errors.New(errors.ErrBusinessInactive, constant.ErrMsgBusinessInactive, 400)
	}

	// Generate atau validate slug
	var slug string
	if req.Slug != "" {
		if !uc.slugService.IsValid(req.Slug) {
			return nil, errors.New(errors.ErrValidation, "Slug tidak valid", 400)
		}
		
		exists, err := uc.catalogRepo.IsSlugExists(req.Slug)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New(errors.ErrConflict, constant.ErrMsgCatalogSlugExists, 409)
		}
		slug = req.Slug
	} else {
		generatedSlug, err := service.GenerateUniqueSlug(
			req.Title,
			uc.slugService,
			func(s string) (bool, error) {
				return uc.catalogRepo.IsSlugExists(s)
			},
			5,
		)
		if err != nil {
			return nil, err
		}
		slug = generatedSlug
	}

	// Set default settings if empty
	if req.Settings == nil {
		req.Settings = make(map[string]interface{})
	}

	// Start transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Create catalog
	catalog := &entity.Catalog{
		BusinessID: req.BusinessID,
		ThemeID:    req.ThemeID,
		Slug:       slug,
		Title:      req.Title,
		Subtitle:   database.NullString(req.Subtitle),
		IsActive:   true,
		Settings:   req.Settings,
		CreatedBy:  profileID,
		CreatedAt:  time.Now(),
	}

	if err := uc.catalogRepo.Create(tx, catalog); err != nil {
		return nil, err
	}

	// Create sections if provided
	if len(req.Sections) > 0 {
		for _, sectionReq := range req.Sections {
			if err := uc.createSectionInternal(tx, catalog.ID, profileID, &sectionReq); err != nil {
				return nil, err
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	// Get complete catalog data
	return uc.GetByID(catalog.ID, profileID)
}

// GetByID mendapatkan catalog by ID
func (uc *catalogUseCase) GetByID(id int64, profileID int64) (*dto.CatalogResponse, error) {
	// Get catalog
	catalog, err := uc.catalogRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check access
	if profileID > 0 {
		if err := uc.checkBusinessAccess(catalog.BusinessID, profileID, constant.PermCatalogView); err != nil {
			return nil, err
		}
	}

	// Get sections
	sections, err := uc.catalogRepo.GetSectionsByCatalogID(id)
	if err != nil {
		return nil, err
	}

	// Convert to response
	return uc.toCatalogResponse(catalog, sections), nil
}

// GetBySlug mendapatkan public catalog by slug
func (uc *catalogUseCase) GetBySlug(slug string) (*dto.PublicCatalogResponse, error) {
	// Get catalog
	catalog, err := uc.catalogRepo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}

	// Check if catalog is active
	if !catalog.IsActive {
		return nil, errors.New(errors.ErrCatalogInactive, constant.ErrMsgCatalogInactive, 404)
	}

	// Check if business is accessible
	if !catalog.Business.IsActive {
		return nil, errors.New(errors.ErrBusinessInactive, constant.ErrMsgBusinessInactive, 404)
	}

	// Get sections
	sections, err := uc.catalogRepo.GetSectionsByCatalogID(catalog.ID)
	if err != nil {
		return nil, err
	}

	// Load section content
	for _, section := range sections {
		if !section.IsVisible {
			continue
		}

		switch section.Type {
		case constant.SectionTypeCards:
			cards, err := uc.catalogRepo.GetCardsBySectionID(section.ID)
			if err != nil {
				return nil, err
			}
			// Load card details and media
			for _, card := range cards {
				if card.HasDetail {
					detail, err := uc.catalogRepo.GetCardDetailByCardID(card.ID)
					if err != nil {
						return nil, err
					}
					card.Detail = detail
				}
				media, err := uc.catalogRepo.GetCardMediaByCardID(card.ID)
				if err != nil {
					return nil, err
				}
				card.Media = media
			}
			section.Cards = cards

		case constant.SectionTypeFAQs:
			faqs, err := uc.catalogRepo.GetFAQsBySectionID(section.ID)
			if err != nil {
				return nil, err
			}
			section.FAQs = faqs

		// TODO: Implement other section types
		}
	}

	// Convert to public response
	return uc.toPublicCatalogResponse(catalog, sections), nil
}

// List mendapatkan list catalogs
func (uc *catalogUseCase) List(profileID int64, filter *dto.CatalogFilter, page, perPage int, orderBy string) ([]*dto.CatalogListResponse, int64, error) {
	// If profileID provided, filter by user's businesses
	businessIDs := []int64{}
	if profileID > 0 && (filter == nil || filter.BusinessID == 0) {
		// Get user businesses
		businesses, _, err := uc.businessRepo.List(repository.ListFilter{
			ProfileID: profileID,
			Limit:     100, // Get all user businesses
		})
		if err != nil {
			return nil, 0, err
		}

		for _, b := range businesses {
			businessIDs = append(businessIDs, b.ID)
		}
	}

	// Build filter
	repoFilter := catalogRepo.ListFilter{
		Limit:   perPage,
		Offset:  (page - 1) * perPage,
		OrderBy: orderBy,
	}

	if filter != nil {
		repoFilter.Search = filter.Search
		repoFilter.BusinessID = filter.BusinessID
		repoFilter.ThemeID = filter.ThemeID
		repoFilter.IsActive = filter.IsActive
	}

	// Apply business filter
	if len(businessIDs) > 0 && repoFilter.BusinessID == 0 {
		// TODO: Implement multiple business IDs filter in repository
		// For now, we'll filter manually after fetching
	}

	// Get catalogs
	catalogs, total, err := uc.catalogRepo.List(repoFilter)
	if err != nil {
		return nil, 0, err
	}

	// Convert to response
	responses := make([]*dto.CatalogListResponse, 0)
	for _, catalog := range catalogs {
		// Manual filter for multiple businesses (temporary)
		if len(businessIDs) > 0 && repoFilter.BusinessID == 0 {
			found := false
			for _, bid := range businessIDs {
				if catalog.BusinessID == bid {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		responses = append(responses, &dto.CatalogListResponse{
			ID:           catalog.ID,
			BusinessID:   catalog.BusinessID,
			BusinessName: catalog.Business.Name,
			Slug:         catalog.Slug,
			Title:        catalog.Title,
			Subtitle:     catalog.GetSubtitle(),
			IsActive:     catalog.IsActive,
			ThemeName:    catalog.Theme.Name,
			CreatedAt:    catalog.CreatedAt,
			UpdatedAt:    catalog.UpdatedAt,
			PublicURL:    fmt.Sprintf("/c/%s", catalog.Slug),
		})
	}

	return responses, int64(len(responses)), nil
}

// Update update catalog
func (uc *catalogUseCase) Update(id int64, profileID int64, req *dto.UpdateCatalogRequest) (*dto.CatalogResponse, error) {
	// Get existing catalog
	catalog, err := uc.catalogRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check permission
	if err := uc.checkBusinessAccess(catalog.BusinessID, profileID, constant.PermCatalogUpdate); err != nil {
		return nil, err
	}

	// Update fields
	if req.ThemeID > 0 {
		catalog.ThemeID = req.ThemeID
	}
	if req.Title != "" {
		catalog.Title = req.Title
	}
	if req.Subtitle != "" {
		catalog.Subtitle = database.NullString(req.Subtitle)
	}
	if req.IsActive != nil {
		catalog.IsActive = *req.IsActive
	}
	if req.Settings != nil {
		catalog.Settings = req.Settings
	}

	catalog.UpdatedBy = database.NullInt64(profileID)
	catalog.UpdatedAt = &[]time.Time{time.Now()}[0]

	// Update in transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.catalogRepo.Update(tx, catalog); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	// Return updated catalog
	return uc.GetByID(id, profileID)
}

// Delete soft delete catalog
func (uc *catalogUseCase) Delete(id int64, profileID int64) error {
	// Get catalog
	catalog, err := uc.catalogRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Check permission
	if err := uc.checkBusinessAccess(catalog.BusinessID, profileID, constant.PermCatalogDelete); err != nil {
		return err
	}

	// Delete in transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.catalogRepo.Delete(tx, id); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

// CreateSection membuat section baru
func (uc *catalogUseCase) CreateSection(catalogID int64, profileID int64, req *dto.CreateSectionRequest) error {
	// Get catalog
	catalog, err := uc.catalogRepo.GetByID(catalogID)
	if err != nil {
		return err
	}

	// Check permission
	if err := uc.checkBusinessAccess(catalog.BusinessID, profileID, constant.PermCatalogUpdate); err != nil {
		return err
	}

	// Validate section type
	if !constant.IsValidSectionType(req.Type) {
		return errors.New(errors.ErrValidation, constant.ErrMsgSectionTypeInvalid, 400)
	}

	// Create in transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.createSectionInternal(tx, catalogID, profileID, req); err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateSection update section
func (uc *catalogUseCase) UpdateSection(sectionID int64, profileID int64, req *dto.UpdateSectionRequest) error {
	// Get section
	section, err := uc.catalogRepo.GetSectionByID(sectionID)
	if err != nil {
		return err
	}

	// Get catalog for permission check
	catalog, err := uc.catalogRepo.GetByID(section.CatalogID)
	if err != nil {
		return err
	}

	// Check permission
	if err := uc.checkBusinessAccess(catalog.BusinessID, profileID, constant.PermCatalogUpdate); err != nil {
		return err
	}

	// Validate updates
	if req.Type != "" && !constant.IsValidSectionType(req.Type) {
		return errors.New(errors.ErrValidation, constant.ErrMsgSectionTypeInvalid, 400)
	}

	// Update fields
	if req.Type != "" {
		section.Type = req.Type
	}
	if req.IsVisible != nil {
		section.IsVisible = *req.IsVisible
	}
	if req.Config != nil {
		section.Config = req.Config
	}

	// Update in transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.catalogRepo.UpdateSection(tx, section); err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteSection delete section
func (uc *catalogUseCase) DeleteSection(sectionID int64, profileID int64) error {
	// Get section
	section, err := uc.catalogRepo.GetSectionByID(sectionID)
	if err != nil {
		return err
	}

	// Get catalog for permission check
	catalog, err := uc.catalogRepo.GetByID(section.CatalogID)
	if err != nil {
		return err
	}

	// Check permission
	if err := uc.checkBusinessAccess(catalog.BusinessID, profileID, constant.PermCatalogUpdate); err != nil {
		return err
	}

	// Delete in transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.catalogRepo.DeleteSection(tx, sectionID); err != nil {
		return err
	}

	return tx.Commit()
}

// CreateCard membuat card baru
func (uc *catalogUseCase) CreateCard(sectionID int64, profileID int64, req *dto.CreateCardRequest) error {
	// Get section
	section, err := uc.catalogRepo.GetSectionByID(sectionID)
	if err != nil {
		return err
	}

	// Validate section type
	if section.Type != constant.SectionTypeCards {
		return errors.New(errors.ErrValidation, "Section bukan tipe cards", 400)
	}

	// Get catalog for permission check
	catalog, err := uc.catalogRepo.GetByID(section.CatalogID)
	if err != nil {
		return err
	}

	// Check permission
	if err := uc.checkBusinessAccess(catalog.BusinessID, profileID, constant.PermCatalogUpdate); err != nil {
		return err
	}

	// Validate card type
	if !constant.IsValidCardType(req.Type) {
		return errors.New(errors.ErrValidation, constant.ErrMsgCardTypeInvalid, 400)
	}

	// Start transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Create card
	card := &entity.CatalogCard{
		SectionID: sectionID,
		Title:     req.Title,
		Subtitle:  database.NullString(req.Subtitle),
		Type:      req.Type,
		URL:       database.NullString(req.URL),
		IsVisible: req.IsVisible,
		HasDetail: req.HasDetail,
		Price:     database.NullInt64(req.Price),
		Discount:  req.Discount,
		Currency:  req.Currency,
		CreatedBy: profileID,
		CreatedAt: time.Now(),
	}

	if card.Currency == "" {
		card.Currency = constant.CurrencyIDR
	}

	if err := uc.catalogRepo.CreateCard(tx, card); err != nil {
		return err
	}

	// Create detail if requested
	if req.HasDetail && req.Detail != nil {
		var detailSlug string
		if req.Detail.Slug != "" {
			detailSlug = req.Detail.Slug
		} else {
			detailSlug = uc.slugService.GenerateUnique(req.Title, 50)
		}

		detail := &entity.CatalogCardDetail{
			CardID:      card.ID,
			Slug:        detailSlug,
			Description: database.NullString(req.Detail.Description),
			IsVisible:   req.Detail.IsVisible,
			CreatedBy:   profileID,
			CreatedAt:   time.Now(),
		}

		if err := uc.catalogRepo.CreateCardDetail(tx, detail); err != nil {
			return err
		}
	}

	// Create media if provided
	for _, mediaURL := range req.MediaURLs {
		media := &entity.CatalogCardMedia{
			CardID:    card.ID,
			Type:      constant.MediaTypeThumbnail,
			URL:       mediaURL,
			CreatedBy: profileID,
			CreatedAt: time.Now(),
		}

		if err := uc.catalogRepo.CreateCardMedia(tx, media); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateCard update card
func (uc *catalogUseCase) UpdateCard(cardID int64, profileID int64, req *dto.UpdateCardRequest) error {
	// Get card
	card, err := uc.catalogRepo.GetCardByID(cardID)
	if err != nil {
		return err
	}

	// Get section
	section, err := uc.catalogRepo.GetSectionByID(card.SectionID)
	if err != nil {
		return err
	}

	// Get catalog for permission check
	catalog, err := uc.catalogRepo.GetByID(section.CatalogID)
	if err != nil {
		return err
	}

	// Check permission
	if err := uc.checkBusinessAccess(catalog.BusinessID, profileID, constant.PermCatalogUpdate); err != nil {
		return err
	}

	// Validate updates
	if req.Type != "" && !constant.IsValidCardType(req.Type) {
		return errors.New(errors.ErrValidation, constant.ErrMsgCardTypeInvalid, 400)
	}

	// Update fields
	if req.Title != "" {
		card.Title = req.Title
	}
	if req.Subtitle != "" {
		card.Subtitle = database.NullString(req.Subtitle)
	}
	if req.Type != "" {
		card.Type = req.Type
	}
	if req.URL != "" {
		card.URL = database.NullString(req.URL)
	}
	if req.IsVisible != nil {
		card.IsVisible = *req.IsVisible
	}
	if req.HasDetail != nil {
		card.HasDetail = *req.HasDetail
	}
	if req.Price != nil {
		card.Price = database.NullInt64(*req.Price)
	}
	if req.Discount != nil {
		card.Discount = *req.Discount
	}
	if req.Currency != "" {
		card.Currency = req.Currency
	}

	card.UpdatedBy = database.NullInt64(profileID)
	card.UpdatedAt = &[]time.Time{time.Now()}[0]

	// Update in transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.catalogRepo.UpdateCard(tx, card); err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteCard delete card
func (uc *catalogUseCase) DeleteCard(cardID int64, profileID int64) error {
	// Get card
	card, err := uc.catalogRepo.GetCardByID(cardID)
	if err != nil {
		return err
	}

	// Get section
	section, err := uc.catalogRepo.GetSectionByID(card.SectionID)
	if err != nil {
		return err
	}

	// Get catalog for permission check
	catalog, err := uc.catalogRepo.GetByID(section.CatalogID)
	if err != nil {
		return err
	}

	// Check permission
	if err := uc.checkBusinessAccess(catalog.BusinessID, profileID, constant.PermCatalogUpdate); err != nil {
		return err
	}

	// Delete in transaction
	tx, err := uc.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := uc.catalogRepo.DeleteCard(tx, cardID); err != nil {
		return err
	}

	return tx.Commit()
}

// Helper methods

func (uc *catalogUseCase) checkBusinessAccess(businessID, profileID int64, permission string) error {
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

func (uc *catalogUseCase) createSectionInternal(tx *sql.Tx, catalogID int64, profileID int64, req *dto.CreateSectionRequest) error {
	// Set default config if empty
	if req.Config == nil {
		req.Config = make(map[string]interface{})
	}

	// Create section
	section := &entity.CatalogSection{
		CatalogID: catalogID,
		Type:      req.Type,
		IsVisible: req.IsVisible,
		Config:    req.Config,
		CreatedAt: time.Now(),
	}

	if err := uc.catalogRepo.CreateSection(tx, section); err != nil {
		return err
	}

	// Create initial content based on type
	switch req.Type {
	case constant.SectionTypeCards:
		// Cards will be added separately
		
	case constant.SectionTypeFAQs:
		if faqs, ok := req.Content.([]dto.FAQRequest); ok {
			for _, faqReq := range faqs {
				faq := &entity.CatalogFAQ{
					SectionID: section.ID,
					Question:  faqReq.Question,
					Answer:    faqReq.Answer,
					IsVisible: faqReq.IsVisible,
					CreatedBy: profileID,
					CreatedAt: time.Now(),
				}
				if err := uc.catalogRepo.CreateFAQ(tx, faq); err != nil {
					return err
				}
			}
		}

	// TODO: Implement other section types
	}

	return nil
}

func (uc *catalogUseCase) toCatalogResponse(catalog *entity.Catalog, sections []*entity.CatalogSection) *dto.CatalogResponse {
	resp := &dto.CatalogResponse{
		ID:         catalog.ID,
		BusinessID: catalog.BusinessID,
		ThemeID:    catalog.ThemeID,
		Slug:       catalog.Slug,
		QRUrl:      catalog.GetQRUrl(),
		Title:      catalog.Title,
		Subtitle:   catalog.GetSubtitle(),
		IsActive:   catalog.IsActive,
		Settings:   catalog.Settings,
		CreatedBy:  catalog.CreatedBy,
		CreatedAt:  catalog.CreatedAt,
		UpdatedAt:  catalog.UpdatedAt,
		PublicURL:  fmt.Sprintf("/c/%s", catalog.Slug),
	}

	// Add business info
	if catalog.Business != nil {
		resp.Business = &dto.BusinessResponse{
			ID:   catalog.Business.ID,
			Name: catalog.Business.Name,
			Slug: catalog.Business.Slug,
		}
	}

	// Add theme info
	if catalog.Theme != nil {
		resp.Theme = &dto.ThemeResponse{
			ID:   catalog.Theme.ID,
			Name: catalog.Theme.Name,
			Type: catalog.Theme.Type,
		}
	}

	// Add sections
	if sections != nil {
		resp.Sections = make([]dto.SectionResponse, len(sections))
		for i, section := range sections {
			resp.Sections[i] = dto.SectionResponse{
				ID:        section.ID,
				CatalogID: section.CatalogID,
				Type:      section.Type,
				IsVisible: section.IsVisible,
				Config:    section.Config,
				CreatedAt: section.CreatedAt,
				UpdatedAt: section.UpdatedAt,
			}
		}
	}

	return resp
}

func (uc *catalogUseCase) toPublicCatalogResponse(catalog *entity.Catalog, sections []*entity.CatalogSection) *dto.PublicCatalogResponse {
	resp := &dto.PublicCatalogResponse{
		ID:       catalog.ID,
		Slug:     catalog.Slug,
		Title:    catalog.Title,
		Subtitle: catalog.GetSubtitle(),
		Settings: catalog.Settings,
		Business: dto.PublicBusinessInfo{
			Name: catalog.Business.Name,
			Type: catalog.Business.Type,
		},
		Theme: dto.ThemeResponse{
			ID:   catalog.Theme.ID,
			Name: catalog.Theme.Name,
			Type: catalog.Theme.Type,
		},
	}

	// Add visible sections only
	resp.Sections = make([]dto.PublicSectionResponse, 0)
	for _, section := range sections {
		if !section.IsVisible {
			continue
		}

		publicSection := dto.PublicSectionResponse{
			Type:   section.Type,
			Config: section.Config,
		}

		// Convert content based on type
		switch section.Type {
		case constant.SectionTypeCards:
			cards := make([]dto.CardResponse, 0)
			for _, card := range section.Cards {
				if !card.IsVisible {
					continue
				}

				cardResp := dto.CardResponse{
					ID:              card.ID,
					SectionID:       card.SectionID,
					Title:           card.Title,
					Subtitle:        card.Subtitle.String,
					Type:            card.Type,
					URL:             card.URL.String,
					IsVisible:       card.IsVisible,
					HasDetail:       card.HasDetail,
					Price:           card.Price.Int64,
					Discount:        card.Discount,
					Currency:        card.Currency,
					DiscountedPrice: card.GetDiscountedPrice(),
					CreatedAt:       card.CreatedAt,
					UpdatedAt:       card.UpdatedAt,
				}

				// Add media
				if card.Media != nil {
					cardResp.Media = make([]dto.MediaResponse, len(card.Media))
					for j, media := range card.Media {
						cardResp.Media[j] = dto.MediaResponse{
							ID:        media.ID,
							Type:      media.Type,
							URL:       media.URL,
							CreatedAt: media.CreatedAt,
						}
					}
				}

				cards = append(cards, cardResp)
			}
			publicSection.Content = cards

		case constant.SectionTypeFAQs:
			faqs := make([]map[string]interface{}, 0)
			for _, faq := range section.FAQs {
				if !faq.IsVisible {
					continue
				}
				faqs = append(faqs, map[string]interface{}{
					"question": faq.Question,
					"answer":   faq.Answer,
				})
			}
			publicSection.Content = faqs

		// TODO: Implement other section types
		}

		resp.Sections = append(resp.Sections, publicSection)
	}

	return resp
}