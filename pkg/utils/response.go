package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response struktur standar untuk semua API response
type Response struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PaginationMeta metadata untuk pagination
type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// PaginatedResponse response dengan pagination
type PaginatedResponse struct {
	Code    int            `json:"code"`
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Data    interface{}    `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

// Success mengirim response sukses
func Success(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, Response{
		Code:    code,
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// SuccessPaginated mengirim response sukses dengan pagination
func SuccessPaginated(c *gin.Context, code int, message string, data interface{}, meta PaginationMeta) {
	c.JSON(code, PaginatedResponse{
		Code:    code,
		Status:  "success",
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Error mengirim response error
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Status:  "error",
		Message: message,
		Data:    nil,
	})
}

// ValidationError mengirim response validation error
// func ValidationError(c *gin.Context, errors interface{}) {
// 	c.JSON(http.StatusBadRequest, Response{
// 		Code:    http.StatusBadRequest,
// 		Status:  "error",
// 		Message: "Validasi gagal",
// 		Data:    errors,
// 	})
// }

// Abort menghentikan request dengan error
func Abort(c *gin.Context, code int, message string) {
	Error(c, code, message)
	c.Abort()
}

// Helper functions untuk response standar

// OK response untuk 200 OK
func OK(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusOK, message, data)
}

// Created response untuk 201 Created
func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data)
}

// NoContent response untuk 204 No Content
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// BadRequest response untuk 400 Bad Request
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized response untuk 401 Unauthorized
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden response untuk 403 Forbidden
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// NotFound response untuk 404 Not Found
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// MethodNotAllowed response untuk 405 Method Not Allowed
func MethodNotAllowed(c *gin.Context, message string) {
	Error(c, http.StatusMethodNotAllowed, message)
}

// Conflict response untuk 409 Conflict
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, message)
}

// InternalServerError response untuk 500 Internal Server Error
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// GetPaginationMeta membuat pagination meta dari parameter
func GetPaginationMeta(page, perPage int, total int64) PaginationMeta {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}