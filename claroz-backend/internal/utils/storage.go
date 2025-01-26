package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lukelittle/claroz/claroz-backend/internal/config"
)

// FileStorage handles file upload operations
type FileStorage struct {
	config *config.StorageConfig
}

// NewFileStorage creates a new FileStorage instance
func NewFileStorage(cfg *config.StorageConfig) (FileStorageInterface, error) {
	if cfg.Provider == "local" {
		// Create uploads directory if it doesn't exist
		err := os.MkdirAll(cfg.LocalPath, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create upload directory: %w", err)
		}
	}
	return &FileStorage{config: cfg}, nil
}

// allowedImageTypes defines the allowed image MIME types
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
}

// SaveFile saves an uploaded file and returns its URL
func (fs *FileStorage) SaveFile(file *multipart.FileHeader) (string, error) {
	// Validate file size
	if file.Size > fs.config.MaxFileSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", fs.config.MaxFileSize)
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return "", fmt.Errorf("unsupported file type: %s", contentType)
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s-%s%s",
		time.Now().Format("20060102"),
		uuid.New().String(),
		strings.ToLower(ext),
	)

	if fs.config.Provider == "local" {
		return fs.saveLocal(file, filename)
	}

	return "", fmt.Errorf("unsupported storage provider: %s", fs.config.Provider)
}

// saveLocal saves file to local storage
func (fs *FileStorage) saveLocal(file *multipart.FileHeader, filename string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	filepath := filepath.Join(fs.config.LocalPath, filename)
	dst, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	// Return relative path that can be used in URLs
	return fmt.Sprintf("/uploads/%s", filename), nil
}

// DeleteFile removes a file from storage
func (fs *FileStorage) DeleteFile(fileURL string) error {
	if fs.config.Provider != "local" {
		return fmt.Errorf("unsupported storage provider: %s", fs.config.Provider)
	}

	// Extract filename from URL
	filename := filepath.Base(fileURL)
	filepath := filepath.Join(fs.config.LocalPath, filename)

	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filepath)
	}

	// Delete file
	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
