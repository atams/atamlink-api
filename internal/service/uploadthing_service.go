package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/chai2010/webp"
	"golang.org/x/image/draw"

	"github.com/atam/atamlink/internal/config"
	"github.com/atam/atamlink/pkg/errors"
)

// UploadThingService service untuk handle upload ke UploadThing
type UploadThingService interface {
	Upload(file *multipart.FileHeader, folder string) (*UploadThingResponse, error)
	ValidateFile(file *multipart.FileHeader) error
	ConvertToWebP(file *multipart.FileHeader) (*bytes.Buffer, error)
}

type uploadThingService struct {
	config config.UploadThingConfig
	client *http.Client
}

// UploadThingResponse response dari UploadThing API
type UploadThingResponse struct {
	Key      string    `json:"key"`
	URL      string    `json:"url"`
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Type     string    `json:"type"`
	UploadedAt time.Time `json:"uploadedAt"`
}

// UploadThingRequest request structure untuk UploadThing
type UploadThingRequest struct {
	Files []UploadThingFile `json:"files"`
}

type UploadThingFile struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type string `json:"type"`
}

// NewUploadThingService membuat instance UploadThing service baru
func NewUploadThingService(config config.UploadThingConfig) UploadThingService {
	if config.Secret == "" || config.AppID == "" {
		panic("UploadThing credentials tidak boleh kosong")
	}

	return &uploadThingService{
		config: config,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Upload file ke UploadThing
func (s *uploadThingService) Upload(file *multipart.FileHeader, folder string) (*UploadThingResponse, error) {
	// Validasi file
	if err := s.ValidateFile(file); err != nil {
		return nil, err
	}

	// Convert ke WebP
	webpBuffer, err := s.ConvertToWebP(file)
	if err != nil {
		return nil, errors.Wrap(err, "gagal convert image ke WebP")
	}

	// Generate filename dengan folder
	filename := s.generateFilename(file.Filename, folder)

	// Upload ke UploadThing
	response, err := s.uploadToUploadThing(webpBuffer, filename)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// ValidateFile validasi file yang diupload
func (s *uploadThingService) ValidateFile(file *multipart.FileHeader) error {
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
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".png", ".jpg", ".jpeg", ".heic"}
	
	allowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			allowed = true
			break
		}
	}

	if !allowed {
		return errors.New(
			errors.ErrInvalidFileType,
			"hanya menerima file PNG, JPG, JPEG, dan HEIC",
			400,
		)
	}

	return nil
}

// ConvertToWebP convert image ke format WebP
func (s *uploadThingService) ConvertToWebP(file *multipart.FileHeader) (*bytes.Buffer, error) {
	// Buka file
	src, err := file.Open()
	if err != nil {
		return nil, errors.Wrap(err, "gagal membuka file")
	}
	defer src.Close()

	// Decode image berdasarkan extension
	var img image.Image
	ext := strings.ToLower(filepath.Ext(file.Filename))
	
	switch ext {
	case ".png":
		img, err = png.Decode(src)
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(src)
	case ".heic":
		// Untuk HEIC, kita perlu library tambahan atau convert manual
		// Untuk sementara, anggap sudah di-handle oleh library external
		return nil, errors.New(errors.ErrInvalidFileType, "HEIC belum didukung", 400)
	default:
		return nil, errors.New(errors.ErrInvalidFileType, "format file tidak didukung", 400)
	}

	if err != nil {
		return nil, errors.Wrap(err, "gagal decode image")
	}

	// Resize jika terlalu besar (optional - untuk optimasi)
	img = s.resizeImage(img, 1920, 1080) // Max 1920x1080

	// Convert ke WebP
	webpData, err := webp.EncodeRGBA(img, 85.0)
	if err != nil {
		return nil, errors.Wrap(err, "gagal convert ke WebP")
	}

	// Create a new buffer from the resulting bytes as required by the function signature.
	buf := bytes.NewBuffer(webpData)

	return buf, nil
}

// resizeImage resize image jika melebihi dimensi maksimal
func (s *uploadThingService) resizeImage(src image.Image, maxWidth, maxHeight int) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Jika tidak perlu resize
	if width <= maxWidth && height <= maxHeight {
		return src
	}

	// Hitung ratio
	ratio := float64(width) / float64(height)
	
	var newWidth, newHeight int
	if float64(maxWidth)/float64(maxHeight) > ratio {
		newHeight = maxHeight
		newWidth = int(float64(newHeight) * ratio)
	} else {
		newWidth = maxWidth
		newHeight = int(float64(newWidth) / ratio)
	}

	// Resize
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

	return dst
}

// uploadToUploadThing upload file ke UploadThing API
func (s *uploadThingService) uploadToUploadThing(fileBuffer *bytes.Buffer, filename string) (*UploadThingResponse, error) {
	// Step 1: Request presigned URL
	presignedURL, err := s.requestPresignedURL(filename, int64(fileBuffer.Len()))
	if err != nil {
		return nil, err
	}

	// Step 2: Upload file menggunakan presigned URL
	if err := s.uploadWithPresignedURL(presignedURL, fileBuffer); err != nil {
		return nil, err
	}

	// Step 3: Confirm upload
	response, err := s.confirmUpload(presignedURL, filename)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// requestPresignedURL request presigned URL dari UploadThing
func (s *uploadThingService) requestPresignedURL(filename string, size int64) (string, error) {
	url := "https://uploadthing.com/api/uploadthing"
	
	reqBody := UploadThingRequest{
		Files: []UploadThingFile{
			{
				Name: filename,
				Size: size,
				Type: "image/webp",
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", errors.Wrap(err, "gagal marshal request")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", errors.Wrap(err, "gagal membuat request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Uploadthing-Api-Key", s.config.Secret)
	req.Header.Set("X-Uploadthing-Version", "6.4.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "gagal request ke UploadThing")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", errors.New(
			errors.ErrFileUploadFailed,
			fmt.Sprintf("UploadThing error: %s", string(body)),
			resp.StatusCode,
		)
	}

	// Parse response untuk mendapatkan presigned URL
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", errors.Wrap(err, "gagal parse response UploadThing")
	}

	// Extract presigned URL from response
	// Struktur response UploadThing bisa berbeda, sesuaikan dengan dokumentasi terbaru
	presignedURL, ok := result["presignedUrl"].(string)
	if !ok {
		return "", errors.New(errors.ErrFileUploadFailed, "presigned URL tidak ditemukan", 500)
	}

	return presignedURL, nil
}

// uploadWithPresignedURL upload file menggunakan presigned URL
func (s *uploadThingService) uploadWithPresignedURL(presignedURL string, fileBuffer *bytes.Buffer) error {
	req, err := http.NewRequest("PUT", presignedURL, fileBuffer)
	if err != nil {
		return errors.Wrap(err, "gagal membuat upload request")
	}

	req.Header.Set("Content-Type", "image/webp")

	resp, err := s.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "gagal upload file")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return errors.New(
			errors.ErrFileUploadFailed,
			fmt.Sprintf("upload gagal: %s", string(body)),
			resp.StatusCode,
		)
	}

	return nil
}

// confirmUpload konfirmasi upload ke UploadThing
func (s *uploadThingService) confirmUpload(presignedURL, filename string) (*UploadThingResponse, error) {
	// Implementasi confirm upload sesuai dengan UploadThing API
	// Biasanya ada endpoint terpisah untuk konfirmasi
	
	// Untuk sementara, return response sederhana
	// Sesuaikan dengan dokumentasi UploadThing terbaru
	return &UploadThingResponse{
		Key:        s.extractKeyFromURL(presignedURL),
		URL:        s.generatePublicURL(filename),
		Name:       filename,
		Type:       "image/webp",
		UploadedAt: time.Now(),
	}, nil
}

// generateFilename generate nama file unik
func (s *uploadThingService) generateFilename(originalName, folder string) string {
	// Ambil nama tanpa extension
	name := strings.TrimSuffix(originalName, filepath.Ext(originalName))
	
	// Generate timestamp
	timestamp := time.Now().Format("20060102150405")
	
	// Combine dengan folder
	if folder != "" {
		return fmt.Sprintf("%s/%s_%s.webp", folder, name, timestamp)
	}
	
	return fmt.Sprintf("%s_%s.webp", name, timestamp)
}

// extractKeyFromURL extract key dari presigned URL
func (s *uploadThingService) extractKeyFromURL(presignedURL string) string {
	// Extract key dari URL, implementasi tergantung format URL UploadThing
	parts := strings.Split(presignedURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// generatePublicURL generate public URL file
func (s *uploadThingService) generatePublicURL(filename string) string {
	// Format URL public UploadThing
	return fmt.Sprintf("https://uploadthing-prod.s3.us-west-2.amazonaws.com/%s", filename)
}