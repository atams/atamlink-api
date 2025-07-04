package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/atam/atamlink/internal/constant"
	"github.com/atam/atamlink/internal/mod_catalog/entity"
	"github.com/atam/atamlink/pkg/database"
	"github.com/atam/atamlink/pkg/errors"
)

// CatalogRepository interface untuk catalog repository
type CatalogRepository interface {
	// Catalog methods
	Create(tx *sql.Tx, catalog *entity.Catalog) error
	GetByID(id int64) (*entity.Catalog, error)
	GetBySlug(slug string) (*entity.Catalog, error)
	List(filter ListFilter) ([]*entity.Catalog, int64, error)
	Update(tx *sql.Tx, catalog *entity.Catalog) error
	Delete(tx *sql.Tx, id int64) error
	IsSlugExists(slug string) (bool, error)
	
	// Section methods
	CreateSection(tx *sql.Tx, section *entity.CatalogSection) error
	GetSectionsByCatalogID(catalogID int64) ([]*entity.CatalogSection, error)
	GetSectionByID(id int64) (*entity.CatalogSection, error)
	UpdateSection(tx *sql.Tx, section *entity.CatalogSection) error
	DeleteSection(tx *sql.Tx, id int64) error
	
	// Card methods
	CreateCard(tx *sql.Tx, card *entity.CatalogCard) error
	GetCardsBySectionID(sectionID int64) ([]*entity.CatalogCard, error)
	GetCardByID(id int64) (*entity.CatalogCard, error)
	UpdateCard(tx *sql.Tx, card *entity.CatalogCard) error
	DeleteCard(tx *sql.Tx, id int64) error
	
	// Card detail methods
	CreateCardDetail(tx *sql.Tx, detail *entity.CatalogCardDetail) error
	GetCardDetailByCardID(cardID int64) (*entity.CatalogCardDetail, error)
	UpdateCardDetail(tx *sql.Tx, detail *entity.CatalogCardDetail) error
	
	// Card media methods
	CreateCardMedia(tx *sql.Tx, media *entity.CatalogCardMedia) error
	GetCardMediaByCardID(cardID int64) ([]*entity.CatalogCardMedia, error)
	DeleteCardMedia(tx *sql.Tx, id int64) error
	
	// Section content methods (FAQs, Links, etc)
	CreateFAQ(tx *sql.Tx, faq *entity.CatalogFAQ) error
	GetFAQsBySectionID(sectionID int64) ([]*entity.CatalogFAQ, error)
	UpdateFAQ(tx *sql.Tx, faq *entity.CatalogFAQ) error
	DeleteFAQ(tx *sql.Tx, id int64) error
}

type catalogRepository struct {
	db *sql.DB
}

// NewCatalogRepository membuat instance catalog repository baru
func NewCatalogRepository(db *sql.DB) CatalogRepository {
	return &catalogRepository{db: db}
}

// ListFilter filter untuk list catalogs
type ListFilter struct {
	Search     string
	BusinessID int64
	ThemeID    int64
	IsActive   *bool
	Limit      int
	Offset     int
	OrderBy    string
}

// Create membuat catalog baru
func (r *catalogRepository) Create(tx *sql.Tx, catalog *entity.Catalog) error {
	settingsJSON, err := json.Marshal(catalog.Settings)
	if err != nil {
		return errors.Wrap(err, "failed to marshal settings")
	}

	query := `
		INSERT INTO atamlink.catalogs (
			c_b_id, c_mt_id, c_slug, c_title, c_subtitle,
			c_is_active, c_settings, c_created_by, c_created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING c_id`

	err = tx.QueryRow(
		query,
		catalog.BusinessID,
		catalog.ThemeID,
		catalog.Slug,
		catalog.Title,
		catalog.Subtitle,
		catalog.IsActive,
		settingsJSON,
		catalog.CreatedBy,
		catalog.CreatedAt,
	).Scan(&catalog.ID)

	if err != nil {
		return errors.Wrap(err, "failed to create catalog")
	}

	return nil
}

// GetByID mendapatkan catalog by ID
func (r *catalogRepository) GetByID(id int64) (*entity.Catalog, error) {
	query := `
		SELECT 
			c.c_id, c.c_b_id, c.c_mt_id, c.c_slug, c.c_qr_url,
			c.c_title, c.c_subtitle, c.c_is_active, c.c_settings,
			c.c_created_by, c.c_created_at, c.c_updated_by, c.c_updated_at,
			b.b_id, b.b_name, b.b_slug,
			mt.mt_id, mt.mt_name, mt.mt_type
		FROM atamlink.catalogs c
		INNER JOIN atamlink.businesses b ON b.b_id = c.c_b_id
		INNER JOIN atamlink.master_themes mt ON mt.mt_id = c.c_mt_id
		WHERE c.c_id = $1`

	catalog := &entity.Catalog{
		Business: &entity.Business{},
		Theme:    &entity.MasterTheme{},
	}

	var settingsJSON []byte
	err := r.db.QueryRow(query, id).Scan(
		&catalog.ID,
		&catalog.BusinessID,
		&catalog.ThemeID,
		&catalog.Slug,
		&catalog.QRUrl,
		&catalog.Title,
		&catalog.Subtitle,
		&catalog.IsActive,
		&settingsJSON,
		&catalog.CreatedBy,
		&catalog.CreatedAt,
		&catalog.UpdatedBy,
		&catalog.UpdatedAt,
		&catalog.Business.ID,
		&catalog.Business.Name,
		&catalog.Business.Slug,
		&catalog.Theme.ID,
		&catalog.Theme.Name,
		&catalog.Theme.Type,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrCatalogNotFound, constant.ErrMsgCatalogNotFound, 404)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get catalog")
	}

	// Parse settings
	if err := json.Unmarshal(settingsJSON, &catalog.Settings); err != nil {
		return nil, errors.Wrap(err, "failed to parse settings")
	}

	return catalog, nil
}

// GetBySlug mendapatkan catalog by slug
func (r *catalogRepository) GetBySlug(slug string) (*entity.Catalog, error) {
	query := `
		SELECT 
			c.c_id, c.c_b_id, c.c_mt_id, c.c_slug, c.c_qr_url,
			c.c_title, c.c_subtitle, c.c_is_active, c.c_settings,
			c.c_created_by, c.c_created_at, c.c_updated_by, c.c_updated_at,
			b.b_id, b.b_name, b.b_slug,
			mt.mt_id, mt.mt_name, mt.mt_type
		FROM atamlink.catalogs c
		INNER JOIN atamlink.businesses b ON b.b_id = c.c_b_id
		INNER JOIN atamlink.master_themes mt ON mt.mt_id = c.c_mt_id
		WHERE c.c_slug = $1`

	catalog := &entity.Catalog{
		Business: &entity.Business{},
		Theme:    &entity.MasterTheme{},
	}

	var settingsJSON []byte
	err := r.db.QueryRow(query, slug).Scan(
		&catalog.ID,
		&catalog.BusinessID,
		&catalog.ThemeID,
		&catalog.Slug,
		&catalog.Business.Type,
    	&catalog.Business.IsActive,
		&catalog.QRUrl,
		&catalog.Title,
		&catalog.Subtitle,
		&catalog.IsActive,
		&settingsJSON,
		&catalog.CreatedBy,
		&catalog.CreatedAt,
		&catalog.UpdatedBy,
		&catalog.UpdatedAt,
		&catalog.Business.ID,
		&catalog.Business.Name,
		&catalog.Business.Slug,
		&catalog.Theme.ID,
		&catalog.Theme.Name,
		&catalog.Theme.Type,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrCatalogNotFound, constant.ErrMsgCatalogNotFound, 404)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get catalog by slug")
	}

	// Parse settings
	if err := json.Unmarshal(settingsJSON, &catalog.Settings); err != nil {
		return nil, errors.Wrap(err, "failed to parse settings")
	}

	return catalog, nil
}

// List mendapatkan list catalogs
func (r *catalogRepository) List(filter ListFilter) ([]*entity.Catalog, int64, error) {
	// Build query
	qb := database.NewQueryBuilder()
	qb.Select(
		"c.c_id", "c.c_b_id", "c.c_mt_id", "c.c_slug", "c.c_qr_url",
		"c.c_title", "c.c_subtitle", "c.c_is_active", "c.c_settings",
		"c.c_created_by", "c.c_created_at", "c.c_updated_by", "c.c_updated_at",
		"b.b_name", "mt.mt_name",
	).From("atamlink.catalogs c")
	qb.InnerJoin("atamlink.businesses b", "b.b_id = c.c_b_id")
	qb.InnerJoin("atamlink.master_themes mt", "mt.mt_id = c.c_mt_id")

	// Apply filters
	if filter.Search != "" {
		qb.Where("(LOWER(c.c_title) LIKE LOWER($1) OR LOWER(c.c_slug) LIKE LOWER($1))", "%"+filter.Search+"%")
	}
	
	if filter.BusinessID > 0 {
		qb.Where("c.c_b_id = ?", filter.BusinessID)
	}
	
	if filter.ThemeID > 0 {
		qb.Where("c.c_mt_id = ?", filter.ThemeID)
	}
	
	if filter.IsActive != nil {
		qb.Where("c.c_is_active = ?", *filter.IsActive)
	}

	// Count total
	countQuery, countArgs := qb.BuildCount()
	var total int64
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to count catalogs")
	}

	// Get data
	qb.OrderBy(filter.OrderBy)
	qb.Limit(filter.Limit)
	qb.Offset(filter.Offset)

	query, args := qb.Build()
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to query catalogs")
	}
	defer rows.Close()

	catalogs := make([]*entity.Catalog, 0)
	for rows.Next() {
		catalog := &entity.Catalog{
			Business: &entity.Business{},
			Theme:    &entity.MasterTheme{},
		}
		var settingsJSON []byte
		
		err := rows.Scan(
			&catalog.ID,
			&catalog.BusinessID,
			&catalog.ThemeID,
			&catalog.Slug,
			&catalog.QRUrl,
			&catalog.Title,
			&catalog.Subtitle,
			&catalog.IsActive,
			&settingsJSON,
			&catalog.CreatedBy,
			&catalog.CreatedAt,
			&catalog.UpdatedBy,
			&catalog.UpdatedAt,
			&catalog.Business.Name,
			&catalog.Theme.Name,
		)
		if err != nil {
			return nil, 0, errors.Wrap(err, "failed to scan catalog row")
		}

		// Parse settings
		if err := json.Unmarshal(settingsJSON, &catalog.Settings); err != nil {
			return nil, 0, errors.Wrap(err, "failed to parse settings")
		}

		catalogs = append(catalogs, catalog)
	}

	return catalogs, total, nil
}

// Update update catalog
func (r *catalogRepository) Update(tx *sql.Tx, catalog *entity.Catalog) error {
	settingsJSON, err := json.Marshal(catalog.Settings)
	if err != nil {
		return errors.Wrap(err, "failed to marshal settings")
	}

	query := `
		UPDATE atamlink.catalogs SET
			c_mt_id = $2,
			c_title = $3,
			c_subtitle = $4,
			c_is_active = $5,
			c_settings = $6,
			c_updated_by = $7,
			c_updated_at = $8
		WHERE c_id = $1`

	result, err := tx.Exec(
		query,
		catalog.ID,
		catalog.ThemeID,
		catalog.Title,
		catalog.Subtitle,
		catalog.IsActive,
		settingsJSON,
		catalog.UpdatedBy,
		time.Now(),
	)

	if err != nil {
		return errors.Wrap(err, "failed to update catalog")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrCatalogNotFound, constant.ErrMsgCatalogNotFound, 404)
	}

	return nil
}

// Delete soft delete catalog
func (r *catalogRepository) Delete(tx *sql.Tx, id int64) error {
	query := `
		UPDATE atamlink.catalogs 
		SET c_is_active = false, c_updated_at = $2
		WHERE c_id = $1`

	result, err := tx.Exec(query, id, time.Now())
	if err != nil {
		return errors.Wrap(err, "failed to delete catalog")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrCatalogNotFound, constant.ErrMsgCatalogNotFound, 404)
	}

	return nil
}

// IsSlugExists check if slug exists
func (r *catalogRepository) IsSlugExists(slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM atamlink.catalogs WHERE c_slug = $1)`
	
	var exists bool
	err := r.db.QueryRow(query, slug).Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "failed to check slug exists")
	}

	return exists, nil
}

// CreateSection create catalog section
func (r *catalogRepository) CreateSection(tx *sql.Tx, section *entity.CatalogSection) error {
	configJSON, err := json.Marshal(section.Config)
	if err != nil {
		return errors.Wrap(err, "failed to marshal config")
	}

	query := `
		INSERT INTO atamlink.catalog_sections (
			cs_c_id, cs_type, cs_is_visible, cs_config, cs_created_at
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING cs_id`

	err = tx.QueryRow(
		query,
		section.CatalogID,
		section.Type,
		section.IsVisible,
		configJSON,
		section.CreatedAt,
	).Scan(&section.ID)

	if err != nil {
		return errors.Wrap(err, "failed to create section")
	}

	return nil
}

// GetSectionsByCatalogID get sections by catalog ID
func (r *catalogRepository) GetSectionsByCatalogID(catalogID int64) ([]*entity.CatalogSection, error) {
	query := `
		SELECT cs_id, cs_c_id, cs_type, cs_is_visible, cs_config, cs_created_at, cs_updated_at
		FROM atamlink.catalog_sections
		WHERE cs_c_id = $1
		ORDER BY cs_id ASC`

	rows, err := r.db.Query(query, catalogID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get sections")
	}
	defer rows.Close()

	sections := make([]*entity.CatalogSection, 0)
	for rows.Next() {
		section := &entity.CatalogSection{}
		var configJSON []byte
		
		err := rows.Scan(
			&section.ID,
			&section.CatalogID,
			&section.Type,
			&section.IsVisible,
			&configJSON,
			&section.CreatedAt,
			&section.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan section")
		}

		// Parse config
		if err := json.Unmarshal(configJSON, &section.Config); err != nil {
			return nil, errors.Wrap(err, "failed to parse config")
		}

		sections = append(sections, section)
	}

	return sections, nil
}

// GetSectionByID get section by ID
func (r *catalogRepository) GetSectionByID(id int64) (*entity.CatalogSection, error) {
	query := `
		SELECT cs_id, cs_c_id, cs_type, cs_is_visible, cs_config, cs_created_at, cs_updated_at
		FROM atamlink.catalog_sections
		WHERE cs_id = $1`

	section := &entity.CatalogSection{}
	var configJSON []byte
	
	err := r.db.QueryRow(query, id).Scan(
		&section.ID,
		&section.CatalogID,
		&section.Type,
		&section.IsVisible,
		&configJSON,
		&section.CreatedAt,
		&section.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrSectionNotFound, constant.ErrMsgSectionNotFound, 404)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get section")
	}

	// Parse config
	if err := json.Unmarshal(configJSON, &section.Config); err != nil {
		return nil, errors.Wrap(err, "failed to parse config")
	}

	return section, nil
}

// UpdateSection update section
func (r *catalogRepository) UpdateSection(tx *sql.Tx, section *entity.CatalogSection) error {
	configJSON, err := json.Marshal(section.Config)
	if err != nil {
		return errors.Wrap(err, "failed to marshal config")
	}

	query := `
		UPDATE atamlink.catalog_sections SET
			cs_type = $2,
			cs_is_visible = $3,
			cs_config = $4,
			cs_updated_at = $5
		WHERE cs_id = $1`

	result, err := tx.Exec(
		query,
		section.ID,
		section.Type,
		section.IsVisible,
		configJSON,
		time.Now(),
	)

	if err != nil {
		return errors.Wrap(err, "failed to update section")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrSectionNotFound, constant.ErrMsgSectionNotFound, 404)
	}

	return nil
}

// DeleteSection delete section
func (r *catalogRepository) DeleteSection(tx *sql.Tx, id int64) error {
	query := `DELETE FROM atamlink.catalog_sections WHERE cs_id = $1`

	result, err := tx.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete section")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrSectionNotFound, constant.ErrMsgSectionNotFound, 404)
	}

	return nil
}

// CreateCard create catalog card
func (r *catalogRepository) CreateCard(tx *sql.Tx, card *entity.CatalogCard) error {
	query := `
		INSERT INTO atamlink.catalog_cards (
			cc_cs_id, cc_title, cc_subtitle, cc_type, cc_url,
			cc_is_visible, cc_has_detail, cc_price, cc_discount,
			cc_currency, cc_created_by, cc_created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING cc_id`

	err := tx.QueryRow(
		query,
		card.SectionID,
		card.Title,
		card.Subtitle,
		card.Type,
		card.URL,
		card.IsVisible,
		card.HasDetail,
		card.Price,
		card.Discount,
		card.Currency,
		card.CreatedBy,
		card.CreatedAt,
	).Scan(&card.ID)

	if err != nil {
		return errors.Wrap(err, "failed to create card")
	}

	return nil
}

// GetCardsBySectionID get cards by section ID
func (r *catalogRepository) GetCardsBySectionID(sectionID int64) ([]*entity.CatalogCard, error) {
	query := `
		SELECT 
			cc_id, cc_cs_id, cc_title, cc_subtitle, cc_type, cc_url,
			cc_is_visible, cc_has_detail, cc_price, cc_discount,
			cc_currency, cc_created_by, cc_created_at, cc_updated_by, cc_updated_at
		FROM atamlink.catalog_cards
		WHERE cc_cs_id = $1
		ORDER BY cc_id ASC`

	rows, err := r.db.Query(query, sectionID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cards")
	}
	defer rows.Close()

	cards := make([]*entity.CatalogCard, 0)
	for rows.Next() {
		card := &entity.CatalogCard{}
		err := rows.Scan(
			&card.ID,
			&card.SectionID,
			&card.Title,
			&card.Subtitle,
			&card.Type,
			&card.URL,
			&card.IsVisible,
			&card.HasDetail,
			&card.Price,
			&card.Discount,
			&card.Currency,
			&card.CreatedBy,
			&card.CreatedAt,
			&card.UpdatedBy,
			&card.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan card")
		}
		cards = append(cards, card)
	}

	return cards, nil
}

// GetCardByID get card by ID
func (r *catalogRepository) GetCardByID(id int64) (*entity.CatalogCard, error) {
	query := `
		SELECT 
			cc_id, cc_cs_id, cc_title, cc_subtitle, cc_type, cc_url,
			cc_is_visible, cc_has_detail, cc_price, cc_discount,
			cc_currency, cc_created_by, cc_created_at, cc_updated_by, cc_updated_at
		FROM atamlink.catalog_cards
		WHERE cc_id = $1`

	card := &entity.CatalogCard{}
	err := r.db.QueryRow(query, id).Scan(
		&card.ID,
		&card.SectionID,
		&card.Title,
		&card.Subtitle,
		&card.Type,
		&card.URL,
		&card.IsVisible,
		&card.HasDetail,
		&card.Price,
		&card.Discount,
		&card.Currency,
		&card.CreatedBy,
		&card.CreatedAt,
		&card.UpdatedBy,
		&card.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrCardNotFound, constant.ErrMsgCardNotFound, 404)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get card")
	}

	return card, nil
}

// UpdateCard update card
func (r *catalogRepository) UpdateCard(tx *sql.Tx, card *entity.CatalogCard) error {
	query := `
		UPDATE atamlink.catalog_cards SET
			cc_title = $2,
			cc_subtitle = $3,
			cc_type = $4,
			cc_url = $5,
			cc_is_visible = $6,
			cc_has_detail = $7,
			cc_price = $8,
			cc_discount = $9,
			cc_currency = $10,
			cc_updated_by = $11,
			cc_updated_at = $12
		WHERE cc_id = $1`

	result, err := tx.Exec(
		query,
		card.ID,
		card.Title,
		card.Subtitle,
		card.Type,
		card.URL,
		card.IsVisible,
		card.HasDetail,
		card.Price,
		card.Discount,
		card.Currency,
		card.UpdatedBy,
		time.Now(),
	)

	if err != nil {
		return errors.Wrap(err, "failed to update card")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrCardNotFound, constant.ErrMsgCardNotFound, 404)
	}

	return nil
}

// DeleteCard delete card
func (r *catalogRepository) DeleteCard(tx *sql.Tx, id int64) error {
	query := `DELETE FROM atamlink.catalog_cards WHERE cc_id = $1`

	result, err := tx.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete card")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrCardNotFound, constant.ErrMsgCardNotFound, 404)
	}

	return nil
}

// CreateCardDetail create card detail
func (r *catalogRepository) CreateCardDetail(tx *sql.Tx, detail *entity.CatalogCardDetail) error {
	query := `
		INSERT INTO atamlink.catalog_card_details (
			ccd_cc_id, ccd_slug, ccd_description, ccd_is_visible,
			ccd_created_by, ccd_created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ccd_id`

	err := tx.QueryRow(
		query,
		detail.CardID,
		detail.Slug,
		detail.Description,
		detail.IsVisible,
		detail.CreatedBy,
		detail.CreatedAt,
	).Scan(&detail.ID)

	if err != nil {
		return errors.Wrap(err, "failed to create card detail")
	}

	return nil
}

// GetCardDetailByCardID get card detail by card ID
func (r *catalogRepository) GetCardDetailByCardID(cardID int64) (*entity.CatalogCardDetail, error) {
	query := `
		SELECT 
			ccd_id, ccd_cc_id, ccd_slug, ccd_description, ccd_is_visible,
			ccd_created_by, ccd_created_at, ccd_updated_by, ccd_updated_at
		FROM atamlink.catalog_card_details
		WHERE ccd_cc_id = $1`

	detail := &entity.CatalogCardDetail{}
	err := r.db.QueryRow(query, cardID).Scan(
		&detail.ID,
		&detail.CardID,
		&detail.Slug,
		&detail.Description,
		&detail.IsVisible,
		&detail.CreatedBy,
		&detail.CreatedAt,
		&detail.UpdatedBy,
		&detail.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get card detail")
	}

	return detail, nil
}

// UpdateCardDetail update card detail
func (r *catalogRepository) UpdateCardDetail(tx *sql.Tx, detail *entity.CatalogCardDetail) error {
	query := `
		UPDATE atamlink.catalog_card_details SET
			ccd_slug = $2,
			ccd_description = $3,
			ccd_is_visible = $4,
			ccd_updated_by = $5,
			ccd_updated_at = $6
		WHERE ccd_id = $1`

	result, err := tx.Exec(
		query,
		detail.ID,
		detail.Slug,
		detail.Description,
		detail.IsVisible,
		detail.UpdatedBy,
		time.Now(),
	)

	if err != nil {
		return errors.Wrap(err, "failed to update card detail")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrNotFound, "Detail card tidak ditemukan", 404)
	}

	return nil
}

// CreateCardMedia create card media
func (r *catalogRepository) CreateCardMedia(tx *sql.Tx, media *entity.CatalogCardMedia) error {
	query := `
		INSERT INTO atamlink.catalog_card_media (
			ccm_cc_id, ccm_type, ccm_url, ccm_created_by, ccm_created_at
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING ccm_id`

	err := tx.QueryRow(
		query,
		media.CardID,
		media.Type,
		media.URL,
		media.CreatedBy,
		media.CreatedAt,
	).Scan(&media.ID)

	if err != nil {
		return errors.Wrap(err, "failed to create card media")
	}

	return nil
}

// GetCardMediaByCardID get card media by card ID
func (r *catalogRepository) GetCardMediaByCardID(cardID int64) ([]*entity.CatalogCardMedia, error) {
	query := `
		SELECT 
			ccm_id, ccm_cc_id, ccm_type, ccm_url,
			ccm_created_by, ccm_created_at, ccm_updated_by, ccm_updated_at
		FROM atamlink.catalog_card_media
		WHERE ccm_cc_id = $1
		ORDER BY ccm_id ASC`

	rows, err := r.db.Query(query, cardID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get card media")
	}
	defer rows.Close()

	mediaList := make([]*entity.CatalogCardMedia, 0)
	for rows.Next() {
		media := &entity.CatalogCardMedia{}
		err := rows.Scan(
			&media.ID,
			&media.CardID,
			&media.Type,
			&media.URL,
			&media.CreatedBy,
			&media.CreatedAt,
			&media.UpdatedBy,
			&media.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan card media")
		}
		mediaList = append(mediaList, media)
	}

	return mediaList, nil
}

// DeleteCardMedia delete card media
func (r *catalogRepository) DeleteCardMedia(tx *sql.Tx, id int64) error {
	query := `DELETE FROM atamlink.catalog_card_media WHERE ccm_id = $1`

	result, err := tx.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete card media")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrNotFound, "Media tidak ditemukan", 404)
	}

	return nil
}

// CreateFAQ create FAQ
func (r *catalogRepository) CreateFAQ(tx *sql.Tx, faq *entity.CatalogFAQ) error {
	query := `
		INSERT INTO atamlink.catalog_faqs (
			cf_cs_id, cf_question, cf_answer, cf_is_visible,
			cf_created_by, cf_created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING cf_id`

	err := tx.QueryRow(
		query,
		faq.SectionID,
		faq.Question,
		faq.Answer,
		faq.IsVisible,
		faq.CreatedBy,
		faq.CreatedAt,
	).Scan(&faq.ID)

	if err != nil {
		return errors.Wrap(err, "failed to create FAQ")
	}

	return nil
}

// GetFAQsBySectionID get FAQs by section ID
func (r *catalogRepository) GetFAQsBySectionID(sectionID int64) ([]*entity.CatalogFAQ, error) {
	query := `
		SELECT 
			cf_id, cf_cs_id, cf_question, cf_answer, cf_is_visible,
			cf_created_by, cf_created_at, cf_updated_by, cf_updated_at
		FROM atamlink.catalog_faqs
		WHERE cf_cs_id = $1
		ORDER BY cf_id ASC`

	rows, err := r.db.Query(query, sectionID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get FAQs")
	}
	defer rows.Close()

	faqs := make([]*entity.CatalogFAQ, 0)
	for rows.Next() {
		faq := &entity.CatalogFAQ{}
		err := rows.Scan(
			&faq.ID,
			&faq.SectionID,
			&faq.Question,
			&faq.Answer,
			&faq.IsVisible,
			&faq.CreatedBy,
			&faq.CreatedAt,
			&faq.UpdatedBy,
			&faq.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan FAQ")
		}
		faqs = append(faqs, faq)
	}

	return faqs, nil
}

// UpdateFAQ update FAQ
func (r *catalogRepository) UpdateFAQ(tx *sql.Tx, faq *entity.CatalogFAQ) error {
	query := `
		UPDATE atamlink.catalog_faqs SET
			cf_question = $2,
			cf_answer = $3,
			cf_is_visible = $4,
			cf_updated_by = $5,
			cf_updated_at = $6
		WHERE cf_id = $1`

	result, err := tx.Exec(
		query,
		faq.ID,
		faq.Question,
		faq.Answer,
		faq.IsVisible,
		faq.UpdatedBy,
		time.Now(),
	)

	if err != nil {
		return errors.Wrap(err, "failed to update FAQ")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrNotFound, "FAQ tidak ditemukan", 404)
	}

	return nil
}

// DeleteFAQ delete FAQ
func (r *catalogRepository) DeleteFAQ(tx *sql.Tx, id int64) error {
	query := `DELETE FROM atamlink.catalog_faqs WHERE cf_id = $1`

	result, err := tx.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete FAQ")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to check rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrNotFound, "FAQ tidak ditemukan", 404)
	}

	return nil
}