package utils

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// PaginationParams parameter untuk pagination
type PaginationParams struct {
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	Sort    string `json:"sort"`
	Order   string `json:"order"`
}

// GetOffset menghitung offset untuk query
func (p *PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PerPage
}

// GetLimit mendapatkan limit untuk query
func (p *PaginationParams) GetLimit() int {
	return p.PerPage
}

// Validate validasi dan set default values
func (p *PaginationParams) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}

	if p.PerPage < 1 {
		p.PerPage = 20 // default
	} else if p.PerPage > 100 {
		p.PerPage = 100 // max
	}

	if p.Order != "asc" && p.Order != "desc" {
		p.Order = "desc" // default
	}
}

// GetPaginationParams extract pagination params dari gin context
func GetPaginationParams(c *gin.Context) *PaginationParams {
	params := &PaginationParams{
		Page:    1,
		PerPage: 20,
		Sort:    "created_at",
		Order:   "desc",
	}

	// Parse page
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			params.Page = p
		}
	}

	// Parse per_page
	if perPage := c.Query("per_page"); perPage != "" {
		if pp, err := strconv.Atoi(perPage); err == nil {
			params.PerPage = pp
		}
	}

	// Parse sort
	if sort := c.Query("sort"); sort != "" {
		params.Sort = sort
	}

	// Parse order
	if order := c.Query("order"); order != "" {
		params.Order = order
	}

	// Validate
	params.Validate()

	return params
}

// FilterParams parameter untuk filtering
type FilterParams struct {
	Search  string                 `json:"search"`
	Status  string                 `json:"status"`
	Type    string                 `json:"type"`
	StartDate string               `json:"start_date"`
	EndDate   string               `json:"end_date"`
	Custom  map[string]interface{} `json:"custom"`
}

// GetFilterParams extract filter params dari gin context
func GetFilterParams(c *gin.Context) *FilterParams {
	params := &FilterParams{
		Custom: make(map[string]interface{}),
	}

	// Common filters
	params.Search = c.Query("search")
	params.Status = c.Query("status")
	params.Type = c.Query("type")
	params.StartDate = c.Query("start_date")
	params.EndDate = c.Query("end_date")

	// Additional filters bisa ditambahkan ke Custom
	// Contoh untuk business_id, catalog_id, dll
	if businessID := c.Query("business_id"); businessID != "" {
		params.Custom["business_id"] = businessID
	}

	if catalogID := c.Query("catalog_id"); catalogID != "" {
		params.Custom["catalog_id"] = catalogID
	}

	return params
}

// BuildOrderBy build ORDER BY clause untuk SQL
func BuildOrderBy(sort, order string, allowedSorts map[string]string) string {
	// Check if sort field is allowed
	sortField, allowed := allowedSorts[sort]
	if !allowed {
		// Use default sort
		sortField = allowedSorts["created_at"]
		if sortField == "" {
			sortField = "created_at"
		}
	}

	// Validate order
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	return sortField + " " + strings.ToUpper(order)
}

// CalculateTotalPages menghitung total halaman
func CalculateTotalPages(total int64, perPage int) int {
	if total == 0 || perPage == 0 {
		return 0
	}

	pages := int(total) / perPage
	if int(total)%perPage > 0 {
		pages++
	}

	return pages
}