package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"

	"github.com/atam/atamlink/internal/config"
	"github.com/atam/atamlink/pkg/errors"
)

// UploadService service untuk handle file upload
type UploadService interface {
	Upload(file *multipart.FileHeader, folder string) (string, error)
	UploadMultiple(files []*multipart.FileHeader, folder string) ([]string, error)
	Delete(filePath string) error
	ValidateFile(file *multipart.FileHeader) error
	GetFullPath(relativePath string) string
	GetRelativePath(fullPath string) string
	
	// New methods for Cloudinary integration
	UploadImageToCloudinary(file *multipart.FileHeader, imageType string) (string, error)
	DeleteFromCloudinary(publicID string) error
}

type uploadService struct {
	config     config.UploadConfig
	cloudinary *cloudinary.Cloudinary
}

// NewUploadService membuat instance upload service baru
func NewUploadService(config config.UploadConfig) UploadService {
	// Ensure upload directory exists for local storage
	if err := os.MkdirAll(config.Path, os.ModePerm); err != nil {
		panic(fmt.Sprintf("failed to create upload directory: %v", err))
	}

	// Initialize Cloudinary
	cld, err := cloudinary.NewFromParams(
		config.Cloudinary.CloudName,
		config.Cloudinary.APIKey,
		config.Cloudinary.APISecret,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize Cloudinary: %v", err))
	}

	return &uploadService{
		config:     config,
		cloudinary: cld,
	}
}

// UploadImageToCloudinary upload image ke Cloudinary dengan proses kompresi dan konversi
// UploadImageToCloudinary upload image ke Cloudinary dengan proses kompresi dan konversi
func (s *uploadService) UploadImageToCloudinary(file *multipart.FileHeader, imageType string) (string, error) {
	// 1. Validasi tipe gambar internal
	if !isValidImageType(imageType) {
		return "", errors.New(errors.ErrValidation, "Tipe gambar tidak valid", 400)
	}

	// 2. Validasi ukuran file
	if file.Size > 10*1024*1024 {
		return "", errors.New(errors.ErrFileTooLarge, "Ukuran file maksimal 10MB", 400)
	}

	// 3. Buka file dan baca seluruh konten ke memory
	src, err := file.Open()
	if err != nil {
		return "", errors.Wrap(err, "failed to open uploaded file")
	}
	defer src.Close()

	// Baca seluruh file ke memory
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return "", errors.Wrap(err, "failed to read file")
	}

	// 4. Validasi MIME type
	contentType := http.DetectContentType(fileBytes)
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
	}

	if !allowedTypes[contentType] {
		return "", errors.New(errors.ErrInvalidFileType, "Hanya file PNG, JPG, dan JPEG yang diizinkan", 400)
	}

	// 5. Decode image dari bytes
	reader := bytes.NewReader(fileBytes)
	img, format, err := image.Decode(reader)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode image")
	}

	// 6. Resize dan kompresi berdasarkan tipe
	processedImg, err := s.processImage(img, imageType)
	if err != nil {
		return "", err
	}

	// 7. Convert to optimized format (JPEG dengan kompresi tinggi)
	imageData, uploadFormat, err := s.convertToOptimizedFormat(processedImg, imageType, format)
	if err != nil {
		return "", err
	}

	// 8. Generate unique filename
	filename := s.generateCloudinaryFilename(imageType)

	// 9. Upload ke Cloudinary
	ctx := context.Background()
	uploadResult, err := s.cloudinary.Upload.Upload(ctx, bytes.NewReader(imageData), uploader.UploadParams{
		PublicID: filename,
		Folder:   s.config.Cloudinary.Folder,
		Format:   uploadFormat,
		Overwrite: func() *bool { b := true; return &b }(),
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to upload to Cloudinary")
	}

	return uploadResult.SecureURL, nil
}

// DeleteFromCloudinary hapus file dari Cloudinary
func (s *uploadService) DeleteFromCloudinary(publicID string) error {
	ctx := context.Background()
	_, err := s.cloudinary.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete from Cloudinary")
	}
	return nil
}

// // validateImageFile validasi file gambar
// func (s *uploadService) validateImageFile(file *multipart.FileHeader) error {
// 	// Check ukuran maksimal 10MB
// 	if file.Size > 10*1024*1024 {
// 		return errors.New(errors.ErrFileTooLarge, "Ukuran file maksimal 10MB", 400)
// 	}

// 	// Validasi MIME type
// 	src, err := file.Open()
// 	if err != nil {
// 		return errors.Wrap(err, "failed to open file for validation")
// 	}
// 	defer src.Close()

// 	// Read first 512 bytes untuk deteksi MIME type
// 	buffer := make([]byte, 512)
// 	_, err = src.Read(buffer)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to read file")
// 	}

// 	// Detect content type
// 	contentType := http.DetectContentType(buffer)
	
// 	// Check allowed types
// 	allowedTypes := map[string]bool{
// 		"image/jpeg": true,
// 		"image/jpg":  true,
// 		"image/png":  true,
// 	}

// 	if !allowedTypes[contentType] {
// 		return errors.New(errors.ErrInvalidFileType, "Hanya file PNG, JPG, dan JPEG yang diizinkan", 400)
// 	}

// 	return nil
// }

// processImage resize dan optimize image berdasarkan tipe
func (s *uploadService) processImage(img image.Image, imageType string) (image.Image, error) {
	var maxWidth, maxHeight int
	
	switch imageType {
	case "thumbnail":
		maxWidth = 400
		maxHeight = 400
	case "gallery", "cover":
		maxWidth = 1200
		maxHeight = 800
	default:
		return nil, errors.New(errors.ErrValidation, "Tipe gambar tidak valid", 400)
	}

	// Resize dengan maintain aspect ratio
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate new dimensions
	if width > maxWidth || height > maxHeight {
		img = imaging.Fit(img, maxWidth, maxHeight, imaging.Lanczos)
	}

	return img, nil
}

// convertToOptimizedFormat konversi image ke format yang optimal (JPEG/PNG)
func (s *uploadService) convertToOptimizedFormat(img image.Image, imageType string, originalFormat string) ([]byte, string, error) {
	var quality int
	var maxSize int64
	var targetFormat string

	switch imageType {
	case "thumbnail":
		quality = 85
		maxSize = 100 * 1024 // 100KB
		targetFormat = "jpg"
	case "gallery", "cover":
		quality = 90
		maxSize = 500 * 1024 // 500KB
		targetFormat = "jpg"
	default:
		return nil, "", errors.New(errors.ErrValidation, "Tipe gambar tidak valid", 400)
	}

	// Jika original format adalah PNG dan memiliki transparency, pertahankan PNG
	if originalFormat == "png" && hasTransparency(img) {
		targetFormat = "png"
		return s.encodePNG(img, maxSize)
	}

	// Encode ke JPEG dengan quality yang ditentukan
	data, err := s.encodeJPEG(img, quality)
	if err != nil {
		return nil, "", err
	}

	// Check size dan kurangi quality jika perlu
	for len(data) > int(maxSize) && quality > 60 {
		quality -= 10
		data, err = s.encodeJPEG(img, quality)
		if err != nil {
			return nil, "", err
		}
	}

	// Jika masih terlalu besar, resize image
	if len(data) > int(maxSize) {
		// Reduce dimensions by 20%
		bounds := img.Bounds()
		newWidth := int(float64(bounds.Dx()) * 0.8)
		newHeight := int(float64(bounds.Dy()) * 0.8)
		
		resizedImg := imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
		
		data, err = s.encodeJPEG(resizedImg, quality)
		if err != nil {
			return nil, "", err
		}
	}

	return data, targetFormat, nil
}

// encodeJPEG encode image ke JPEG dengan quality tertentu
func (s *uploadService) encodeJPEG(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	
	options := &jpeg.Options{
		Quality: quality,
	}
	
	if err := jpeg.Encode(&buf, img, options); err != nil {
		return nil, errors.Wrap(err, "failed to encode JPEG")
	}
	
	return buf.Bytes(), nil
}

// encodePNG encode image ke PNG dengan kompresi
func (s *uploadService) encodePNG(img image.Image, maxSize int64) ([]byte, string, error) {
	var buf bytes.Buffer
	
	if err := png.Encode(&buf, img); err != nil {
		return nil, "", errors.Wrap(err, "failed to encode PNG")
	}
	
	data := buf.Bytes()
	
	// Jika PNG terlalu besar, resize
	if len(data) > int(maxSize) {
		// Reduce dimensions by 20%
		bounds := img.Bounds()
		newWidth := int(float64(bounds.Dx()) * 0.8)
		newHeight := int(float64(bounds.Dy()) * 0.8)
		
		resizedImg := imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
		
		buf.Reset()
		if err := png.Encode(&buf, resizedImg); err != nil {
			return nil, "", errors.Wrap(err, "failed to encode resized PNG")
		}
		data = buf.Bytes()
	}
	
	return data, "png", nil
}

// hasTransparency check apakah image memiliki transparency
func hasTransparency(img image.Image) bool {
	switch img.(type) {
	case *image.NRGBA, *image.RGBA:
		return true
	default:
		return false
	}
}

// generateCloudinaryFilename generate unique filename untuk Cloudinary
func (s *uploadService) generateCloudinaryFilename(imageType string) string {
	now := time.Now()
	
	// Format: ss-mm-hh-dd-mm-yyyy-ms_[tipe]
	filename := fmt.Sprintf("%02d-%02d-%02d-%02d-%02d-%04d-%03d_%s",
		now.Second(),
		now.Minute(),
		now.Hour(),
		now.Day(),
		now.Month(),
		now.Year(),
		now.Nanosecond()/1000000, // Convert to milliseconds
		imageType,
	)
	
	return filename
}

// isValidImageType check apakah image type valid
func isValidImageType(t string) bool {
	validTypes := []string{"thumbnail", "gallery", "cover"}
	for _, valid := range validTypes {
		if t == valid {
			return true
		}
	}
	return false
}

// Legacy methods untuk backward compatibility

// Upload single file ke local storage
func (s *uploadService) Upload(file *multipart.FileHeader, folder string) (string, error) {
	// Validate file
	if err := s.ValidateFile(file); err != nil {
		return "", err
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := s.generateFilename(ext)

	// Create folder path
	folderPath := filepath.Join(s.config.Path, folder)
	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
		return "", errors.Wrap(err, "failed to create folder")
	}

	// Full file path
	fullPath := filepath.Join(folderPath, filename)

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return "", errors.Wrap(err, "failed to open uploaded file")
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to create destination file")
	}
	defer dst.Close()

	// Copy file
	if _, err := io.Copy(dst, src); err != nil {
		// Clean up on error
		os.Remove(fullPath)
		return "", errors.Wrap(err, "failed to save file")
	}

	// Return relative path
	relativePath := filepath.Join(folder, filename)
	return s.normalizePathSeparator(relativePath), nil
}

// UploadMultiple upload multiple files
func (s *uploadService) UploadMultiple(files []*multipart.FileHeader, folder string) ([]string, error) {
	paths := make([]string, 0, len(files))

	for _, file := range files {
		path, err := s.Upload(file, folder)
		if err != nil {
			// Rollback: delete already uploaded files
			for _, p := range paths {
				s.Delete(p)
			}
			return nil, err
		}
		paths = append(paths, path)
	}

	return paths, nil
}

// Delete file dari local storage
func (s *uploadService) Delete(filePath string) error {
	// Security check: ensure path is within upload directory
	fullPath := filepath.Join(s.config.Path, filePath)
	cleanPath := filepath.Clean(fullPath)

	// Check if path is within upload directory
	if !strings.HasPrefix(cleanPath, s.config.Path) {
		return errors.New(errors.ErrForbidden, "invalid file path", 403)
	}

	// Check if file exists
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		// File doesn't exist, consider it success
		return nil
	}

	// Delete file
	if err := os.Remove(cleanPath); err != nil {
		return errors.Wrap(err, "failed to delete file")
	}

	return nil
}

// ValidateFile validate uploaded file untuk local storage
func (s *uploadService) ValidateFile(file *multipart.FileHeader) error {
	// Check file size
	if file.Size > s.config.MaxSize {
		maxSizeMB := s.config.MaxSize / (1024 * 1024)
		return errors.New(
			errors.ErrFileTooLarge,
			fmt.Sprintf("ukuran file maksimal %d MB", maxSizeMB),
			400,
		)
	}

	// Check file type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		// Try to detect from extension
		ext := strings.ToLower(filepath.Ext(file.Filename))
		contentType = s.getContentTypeFromExt(ext)
	}

	// Validate content type
	allowed := false
	for _, allowedType := range s.config.AllowedTypes {
		if contentType == allowedType {
			allowed = true
			break
		}
		// Check wildcard (e.g., image/*)
		if strings.HasSuffix(allowedType, "/*") {
			prefix := strings.TrimSuffix(allowedType, "/*")
			if strings.HasPrefix(contentType, prefix) {
				allowed = true
				break
			}
		}
	}

	if !allowed {
		return errors.New(
			errors.ErrInvalidFileType,
			fmt.Sprintf("tipe file %s tidak diizinkan", contentType),
			400,
		)
	}

	return nil
}

// GetFullPath get full file path
func (s *uploadService) GetFullPath(relativePath string) string {
	return filepath.Join(s.config.Path, relativePath)
}

// GetRelativePath get relative path from full path
func (s *uploadService) GetRelativePath(fullPath string) string {
	return strings.TrimPrefix(fullPath, s.config.Path+string(filepath.Separator))
}

// generateFilename generate unique filename untuk local storage
func (s *uploadService) generateFilename(ext string) string {
	// Format: year/month/uuid.ext
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	
	// Generate UUID without dashes
	uid := strings.ReplaceAll(uuid.New().String(), "-", "")
	
	return filepath.Join(year, month, uid+ext)
}

// normalizePathSeparator normalize path separator untuk URL
func (s *uploadService) normalizePathSeparator(path string) string {
	return strings.ReplaceAll(path, string(filepath.Separator), "/")
}

// getContentTypeFromExt get content type from file extension
func (s *uploadService) getContentTypeFromExt(ext string) string {
	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
	}

	if ct, exists := contentTypes[ext]; exists {
		return ct
	}

	return "application/octet-stream"
}

// FileInfo informasi file yang diupload
type FileInfo struct {
	Filename     string    `json:"filename"`
	Size         int64     `json:"size"`
	ContentType  string    `json:"content_type"`
	Path         string    `json:"path"`
	URL          string    `json:"url"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

// ProcessUpload process upload dan return file info untuk local storage
func ProcessUpload(
	file *multipart.FileHeader,
	folder string,
	uploadService UploadService,
	baseURL string,
) (*FileInfo, error) {
	// Upload file
	path, err := uploadService.Upload(file, folder)
	if err != nil {
		return nil, err
	}

	// Build file info
	info := &FileInfo{
		Filename:    file.Filename,
		Size:        file.Size,
		ContentType: file.Header.Get("Content-Type"),
		Path:        path,
		URL:         fmt.Sprintf("%s/uploads/%s", baseURL, path),
		UploadedAt:  time.Now(),
	}

	return info, nil
}

// ProcessImageUpload process image upload ke Cloudinary
func ProcessImageUpload(
	file *multipart.FileHeader,
	imageType string,
	uploadService UploadService,
) (string, error) {
	// Upload image ke Cloudinary
	url, err := uploadService.UploadImageToCloudinary(file, imageType)
	if err != nil {
		return "", err
	}

	return url, nil
}