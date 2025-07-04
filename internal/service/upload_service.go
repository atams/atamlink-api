package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

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
}

type uploadService struct {
	config config.UploadConfig
}

// NewUploadService membuat instance upload service baru
func NewUploadService(config config.UploadConfig) UploadService {
	// Ensure upload directory exists
	if err := os.MkdirAll(config.Path, os.ModePerm); err != nil {
		panic(fmt.Sprintf("failed to create upload directory: %v", err))
	}

	return &uploadService{
		config: config,
	}
}

// Upload single file
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

// Delete file
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

// ValidateFile validate uploaded file
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

// generateFilename generate unique filename
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

// ProcessUpload process upload dan return file info
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
		URL:         fmt.Sprintf("%s/uploads/%s", strings.TrimSuffix(baseURL, "/"), path),
		UploadedAt:  time.Now(),
	}

	return info, nil
}