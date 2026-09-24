package cloud

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// StorageService defines object storage operations for chore proofs and attachments.
type StorageService interface {
	UploadProofPhoto(ctx context.Context, filename string, contentType string, reader io.Reader) (key string, url string, size int64, err error)
	GetProofPhoto(ctx context.Context, key string) (contentType string, reader io.ReadCloser, size int64, err error)
	DeleteProofPhoto(ctx context.Context, key string) error
	IsEnabled() bool
}

// S3StorageService implements StorageService using AWS S3 / Floci S3 emulator.
type S3StorageService struct {
	mgr *ClientManager
}

// NewS3StorageService creates a new S3-backed storage service.
func NewS3StorageService(mgr *ClientManager) *S3StorageService {
	return &S3StorageService{mgr: mgr}
}

func (s *S3StorageService) IsEnabled() bool {
	return s.mgr != nil && s.mgr.IsEnabled()
}

// UploadProofPhoto uploads a file to the S3 bucket and returns the unique key and public/API URL.
func (s *S3StorageService) UploadProofPhoto(ctx context.Context, filename string, contentType string, reader io.Reader) (string, string, int64, error) {
	// Read payload into memory to determine size and allow re-read
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to read upload payload: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".jpg"
	}

	randomID := generateRandomHex(8)
	timestamp := time.Now().Format("20060102150405")
	key := fmt.Sprintf("proofs/%s-%s%s", timestamp, randomID, ext)

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	client := s.mgr.S3Client()
	bucket := s.mgr.Config().S3BucketName

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to upload object to S3 (%s/%s): %w", bucket, key, err)
	}

	apiURL := fmt.Sprintf("/api/v1/uploads/photo/%s", key)
	return key, apiURL, int64(len(data)), nil
}

// GetProofPhoto fetches an object from S3.
func (s *S3StorageService) GetProofPhoto(ctx context.Context, key string) (string, io.ReadCloser, int64, error) {
	client := s.mgr.S3Client()
	bucket := s.mgr.Config().S3BucketName

	out, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", nil, 0, fmt.Errorf("failed to retrieve object from S3 (%s/%s): %w", bucket, key, err)
	}

	contentType := "application/octet-stream"
	if out.ContentType != nil {
		contentType = *out.ContentType
	}

	var size int64 = 0
	if out.ContentLength != nil {
		size = *out.ContentLength
	}

	return contentType, out.Body, size, nil
}

// DeleteProofPhoto deletes an object from S3.
func (s *S3StorageService) DeleteProofPhoto(ctx context.Context, key string) error {
	client := s.mgr.S3Client()
	bucket := s.mgr.Config().S3BucketName

	_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from S3 (%s/%s): %w", bucket, key, err)
	}
	return nil
}

// MockStorageService provides an in-memory implementation for offline unit tests.
type MockStorageService struct {
	mu    sync.RWMutex
	files map[string][]byte
	types map[string]string
}

// NewMockStorageService initializes an in-memory mock storage service.
func NewMockStorageService() *MockStorageService {
	return &MockStorageService{
		files: make(map[string][]byte),
		types: make(map[string]string),
	}
}

func (m *MockStorageService) IsEnabled() bool {
	return true
}

func (m *MockStorageService) UploadProofPhoto(ctx context.Context, filename string, contentType string, reader io.Reader) (string, string, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", "", 0, err
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".jpg"
	}
	key := fmt.Sprintf("proofs/%s-%s%s", time.Now().Format("20060102150405"), generateRandomHex(8), ext)
	m.files[key] = data
	m.types[key] = contentType

	return key, fmt.Sprintf("/api/v1/uploads/photo/%s", key), int64(len(data)), nil
}

func (m *MockStorageService) GetProofPhoto(ctx context.Context, key string) (string, io.ReadCloser, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, ok := m.files[key]
	if !ok {
		return "", nil, 0, fmt.Errorf("file not found: %s", key)
	}

	ct := m.types[key]
	if ct == "" {
		ct = "image/jpeg"
	}

	return ct, io.NopCloser(bytes.NewReader(data)), int64(len(data)), nil
}

func (m *MockStorageService) DeleteProofPhoto(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.files, key)
	delete(m.types, key)
	return nil
}

func generateRandomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
