package constant

// Error messages
const (
	// General errors
	ErrMsgInternalServer = "Terjadi kesalahan pada server"
	ErrMsgBadRequest     = "Permintaan tidak valid"
	ErrMsgUnauthorized   = "Anda tidak memiliki akses"
	ErrMsgForbidden      = "Akses ditolak"
	ErrMsgNotFound       = "Data tidak ditemukan"
	ErrMsgValidation     = "Data tidak valid"

	// Auth errors
	ErrMsgTokenNotFound   = "Token tidak ditemukan"
	ErrMsgTokenInvalid    = "Token tidak valid"
	ErrMsgTokenExpired    = "Token sudah kadaluarsa"
	ErrMsgInvalidLogin    = "Email atau password salah"
	ErrMsgAccountLocked   = "Akun Anda terkunci"
	ErrMsgAccountInactive = "Akun Anda tidak aktif"

	// Business errors
	ErrMsgBusinessNotFound      = "Bisnis tidak ditemukan"
	ErrMsgBusinessNameRequired  = "Nama bisnis wajib diisi"
	ErrMsgBusinessSlugExists    = "Slug bisnis sudah digunakan"
	ErrMsgBusinessTypeInvalid   = "Tipe bisnis tidak valid"
	ErrMsgBusinessAccessDenied  = "Anda tidak memiliki akses ke bisnis ini"
	ErrMsgBusinessSuspended     = "Bisnis ini sedang ditangguhkan"
	ErrMsgBusinessInactive      = "Bisnis tidak aktif"
	ErrMsgBusinessOwnerRequired = "Bisnis harus memiliki minimal satu owner"

	// Catalog errors
	ErrMsgCatalogNotFound     = "Katalog tidak ditemukan"
	ErrMsgCatalogTitleRequired = "Judul katalog wajib diisi"
	ErrMsgCatalogSlugExists   = "Slug katalog sudah digunakan"
	ErrMsgCatalogInactive     = "Katalog tidak aktif"
	ErrMsgThemeNotFound       = "Tema tidak ditemukan"
	ErrMsgThemeInactive       = "Tema tidak tersedia"

	// Section errors
	ErrMsgSectionNotFound  = "Section tidak ditemukan"
	ErrMsgSectionTypeInvalid = "Tipe section tidak valid"
	ErrMsgSectionRequired  = "Section wajib diisi"

	// Card errors
	ErrMsgCardNotFound      = "Card tidak ditemukan"
	ErrMsgCardTitleRequired = "Judul card wajib diisi"
	ErrMsgCardTypeInvalid   = "Tipe card tidak valid"
	ErrMsgCardPriceInvalid  = "Harga tidak valid"

	// File upload errors
	ErrMsgFileRequired    = "File wajib diupload"
	ErrMsgFileTooLarge    = "Ukuran file terlalu besar"
	ErrMsgFileTypeInvalid = "Tipe file tidak diizinkan"
	ErrMsgFileUploadFailed = "Upload file gagal"

	// Subscription errors
	ErrMsgSubscriptionExpired = "Subscription Anda sudah habis"
	ErrMsgSubscriptionRequired = "Fitur ini memerlukan subscription"
	ErrMsgPlanNotFound        = "Plan tidak ditemukan"
	ErrMsgInvalidPlan		 = "Plan tidak valid"
	ErrMsgPlanInactive        = "Plan tidak tersedia"

	// User errors
	ErrMsgUserNotFound     = "User tidak ditemukan"
	ErrMsgEmailExists      = "Email sudah terdaftar"
	ErrMsgUsernameExists   = "Username sudah digunakan"
	ErrMsgProfileNotFound  = "Profile tidak ditemukan"
)