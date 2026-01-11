package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/csic-platform/services/security/forensic-tools/internal/config"
	"github.com/csic-platform/services/security/forensic-tools/internal/core/ports"
)

// LocalStorage implements BlobStorage using local filesystem
type LocalStorage struct {
	basePath   string
	maxFileSize int64
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(cfg *config.LocalStorage) (*LocalStorage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(cfg.BasePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalStorage{
		basePath:   cfg.BasePath,
		maxFileSize: cfg.MaxFileSize,
	}, nil
}

// Upload uploads a file to local storage
func (s *LocalStorage) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	// Check file size
	if size > s.maxFileSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", size, s.maxFileSize)
	}

	// Create full path
	fullPath := filepath.Join(s.basePath, objectName)

	// Create directory structure
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data
	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// UploadFromFile copies a local file to storage
func (s *LocalStorage) UploadFromFile(ctx context.Context, sourcePath, objectName string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer source.Close()

	sourceInfo, err := source.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	return s.Upload(ctx, objectName, source, sourceInfo.Size(), "")
}

// Download downloads a file from local storage
func (s *LocalStorage) Download(ctx context.Context, objectName string, writer io.Writer) error {
	fullPath := filepath.Join(s.basePath, objectName)

	file, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(writer, file)
	return err
}

// GetFile returns a file reader for the given object
func (s *LocalStorage) GetFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, objectName)

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// Delete deletes a file from local storage
func (s *LocalStorage) Delete(ctx context.Context, objectName string) error {
	fullPath := filepath.Join(s.basePath, objectName)

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// DeleteMultiple deletes multiple files from local storage
func (s *LocalStorage) DeleteMultiple(ctx context.Context, objectNames []string) error {
	for _, name := range objectNames {
		if err := s.Delete(ctx, name); err != nil {
			return err
		}
	}
	return nil
}

// GetObjectInfo returns information about a stored object
func (s *LocalStorage) GetObjectInfo(ctx context.Context, objectName string) (*ports.BlobObjectInfo, error) {
	fullPath := filepath.Join(s.basePath, objectName)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &ports.BlobObjectInfo{
		Name:       objectName,
		Size:       info.Size(),
		CreatedAt:  info.ModTime(),
		ModifiedAt: info.ModTime(),
	}, nil
}

// ObjectExists checks if an object exists
func (s *LocalStorage) ObjectExists(ctx context.Context, objectName string) (bool, error) {
	fullPath := filepath.Join(s.basePath, objectName)

	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// GetSignedURL returns a signed URL for accessing an object
// For local storage, this returns a file:// URL
func (s *LocalStorage) GetSignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	fullPath := filepath.Join(s.basePath, objectName)
	return fmt.Sprintf("file://%s", fullPath), nil
}

// Ensure LocalStorage implements BlobStorage
var _ ports.BlobStorage = (*LocalStorage)(nil)
