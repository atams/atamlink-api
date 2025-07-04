package errors

import (
	"errors"
	"fmt"
)

// Domain errors
var (
	// General errors
	ErrInternalServer = errors.New("terjadi kesalahan internal server")
	ErrBadRequest     = errors.New("permintaan tidak valid")
	ErrUnauthorized   = errors.New("autentikasi diperlukan")
	ErrForbidden      = errors.New("akses ditolak")
	ErrNotFound       = errors.New("data tidak ditemukan")
	ErrConflict       = errors.New("terjadi konflik data")
	ErrValidation     = errors.New("validasi gagal")

	// Database errors
	ErrDatabaseConnection = errors.New("koneksi database gagal")
	ErrDatabaseQuery      = errors.New("query database gagal")
	ErrDatabaseTx         = errors.New("transaksi database gagal")
	ErrNoRowsAffected     = errors.New("tidak ada data yang terpengaruh")

	// Business logic errors
	ErrDuplicateEntry     = errors.New("data sudah ada")
	ErrInvalidCredentials = errors.New("kredensial tidak valid")
	ErrInvalidToken       = errors.New("token tidak valid")
	ErrTokenExpired       = errors.New("token sudah kadaluarsa")
	ErrAccountLocked      = errors.New("akun terkunci")
	ErrAccountInactive    = errors.New("akun tidak aktif")
	ErrInsufficientRole   = errors.New("role tidak mencukupi")

	// Business specific
	ErrBusinessNotFound   = errors.New("bisnis tidak ditemukan")
	ErrBusinessInactive   = errors.New("bisnis tidak aktif")
	ErrBusinessSuspended  = errors.New("bisnis ditangguhkan")
	ErrDuplicateSlug      = errors.New("slug sudah digunakan")
	ErrInvalidBusinessType = errors.New("tipe bisnis tidak valid")

	// Catalog specific
	ErrCatalogNotFound    = errors.New("katalog tidak ditemukan")
	ErrCatalogInactive    = errors.New("katalog tidak aktif")
	ErrSectionNotFound    = errors.New("section tidak ditemukan")
	ErrCardNotFound       = errors.New("card tidak ditemukan")
	ErrInvalidSectionType = errors.New("tipe section tidak valid")
	ErrInvalidCardType    = errors.New("tipe card tidak valid")

	// File upload errors
	ErrFileTooLarge      = errors.New("ukuran file terlalu besar")
	ErrInvalidFileType   = errors.New("tipe file tidak diizinkan")
	ErrFileUploadFailed  = errors.New("upload file gagal")

	// Subscription errors
	ErrSubscriptionExpired = errors.New("subscription sudah kadaluarsa")
	ErrPlanNotFound        = errors.New("plan tidak ditemukan")
	ErrInvalidPlan         = errors.New("plan tidak valid")
)

// AppError struktur error aplikasi dengan context tambahan
type AppError struct {
	Err        error
	Message    string
	StatusCode int
	Context    map[string]interface{}
}

// Error implementasi error interface
func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

// Unwrap untuk mendapatkan error asli
func (e *AppError) Unwrap() error {
	return e.Err
}

// New membuat AppError baru
func New(err error, message string, statusCode int) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
		Context:    make(map[string]interface{}),
	}
}

// WithContext menambahkan context ke error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	e.Context[key] = value
	return e
}

// Wrap membungkus error dengan message tambahan
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Is mengecek apakah error adalah tipe tertentu
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As mengecek dan convert error ke tipe tertentu
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}