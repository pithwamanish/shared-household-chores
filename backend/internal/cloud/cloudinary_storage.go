package cloud

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// CloudinaryConfig holds configuration for Cloudinary image hosting.
type CloudinaryConfig struct {
	CloudName string
	APIKey    string
	APISecret string
	Folder    string
	Enabled   bool
}

// LoadCloudinaryConfigFromEnv reads Cloudinary configuration from environment variables.
// Supports both CLOUDINARY_URL format (cloudinary://API_KEY:API_SECRET@CLOUD_NAME)
// and individual CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY, CLOUDINARY_API_SECRET variables.
func LoadCloudinaryConfigFromEnv() CloudinaryConfig {
	rawURL := strings.TrimSpace(os.Getenv("CLOUDINARY_URL"))
	cloudName := strings.TrimSpace(os.Getenv("CLOUDINARY_CLOUD_NAME"))
	apiKey := strings.TrimSpace(os.Getenv("CLOUDINARY_API_KEY"))
	apiSecret := strings.TrimSpace(os.Getenv("CLOUDINARY_API_SECRET"))
	folder := strings.TrimSpace(os.Getenv("CLOUDINARY_FOLDER"))
	if folder == "" {
		folder = "choresync/proofs"
	}

	if rawURL != "" {
		parsed, err := url.Parse(rawURL)
		if err == nil {
			cloudName = parsed.Host
			if parsed.User != nil {
				apiKey = parsed.User.Username()
				apiSecret, _ = parsed.User.Password()
			}
		}
	}

	enabled := cloudName != "" && apiKey != "" && apiSecret != ""

	return CloudinaryConfig{
		CloudName: cloudName,
		APIKey:    apiKey,
		APISecret: apiSecret,
		Folder:    folder,
		Enabled:   enabled,
	}
}

// CloudinaryStorageService implements StorageService using Cloudinary REST API.
type CloudinaryStorageService struct {
	cfg        CloudinaryConfig
	httpClient *http.Client
}

// NewCloudinaryStorageService creates a new Cloudinary-backed storage service.
func NewCloudinaryStorageService(cfg CloudinaryConfig) *CloudinaryStorageService {
	return &CloudinaryStorageService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsEnabled returns true if Cloudinary is fully configured with credentials.
func (c *CloudinaryStorageService) IsEnabled() bool {
	return c.cfg.Enabled
}

type cloudinaryUploadResponse struct {
	PublicID  string `json:"public_id"`
	SecureURL string `json:"secure_url"`
	Bytes     int64  `json:"bytes"`
	Format    string `json:"format"`
	Error     *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// UploadProofPhoto uploads an image to Cloudinary and returns the public ID and secure CDN URL.
func (c *CloudinaryStorageService) UploadProofPhoto(ctx context.Context, filename string, contentType string, reader io.Reader) (string, string, int64, error) {
	if !c.IsEnabled() {
		return "", "", 0, fmt.Errorf("Cloudinary storage is not enabled or credentials are missing")
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// Build sorted parameters to sign
	params := map[string]string{
		"timestamp": timestamp,
	}
	if c.cfg.Folder != "" {
		params["folder"] = c.cfg.Folder
	}

	sig := c.generateSignature(params)

	// Build multipart request body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("api_key", c.cfg.APIKey)
	_ = writer.WriteField("timestamp", timestamp)
	_ = writer.WriteField("signature", sig)
	if c.cfg.Folder != "" {
		_ = writer.WriteField("folder", c.cfg.Folder)
	}

	ext := filepath.Ext(filename)
	if ext == "" {
		filename += ".jpg"
	}

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to create multipart form file: %w", err)
	}
	if _, err := io.Copy(part, reader); err != nil {
		return "", "", 0, fmt.Errorf("failed to copy file data: %w", err)
	}
	_ = writer.Close()

	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", c.cfg.CloudName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, body)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to create Cloudinary HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to execute Cloudinary upload request: %w", err)
	}
	defer resp.Body.Close()

	var cldResp cloudinaryUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&cldResp); err != nil {
		return "", "", 0, fmt.Errorf("failed to parse Cloudinary response: %w", err)
	}

	if cldResp.Error != nil {
		return "", "", 0, fmt.Errorf("Cloudinary upload failed: %s", cldResp.Error.Message)
	}

	if resp.StatusCode >= 400 {
		return "", "", 0, fmt.Errorf("Cloudinary upload returned status %d", resp.StatusCode)
	}

	log.Printf("[Cloudinary] Uploaded photo proof: PublicID='%s', URL='%s', Size=%d bytes", cldResp.PublicID, cldResp.SecureURL, cldResp.Bytes)
	return cldResp.PublicID, cldResp.SecureURL, cldResp.Bytes, nil
}

// GetProofPhoto fetches a photo from Cloudinary via its public ID.
func (c *CloudinaryStorageService) GetProofPhoto(ctx context.Context, key string) (string, io.ReadCloser, int64, error) {
	if !c.IsEnabled() {
		return "", nil, 0, fmt.Errorf("Cloudinary storage is not enabled")
	}

	imageURL := fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/%s", c.cfg.CloudName, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", nil, 0, fmt.Errorf("failed to create Cloudinary download request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", nil, 0, fmt.Errorf("failed to fetch photo from Cloudinary: %w", err)
	}

	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return "", nil, 0, fmt.Errorf("Cloudinary image not found or inaccessible (status %d)", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/jpeg"
	}

	return ct, resp.Body, resp.ContentLength, nil
}

// DeleteProofPhoto removes an image from Cloudinary.
func (c *CloudinaryStorageService) DeleteProofPhoto(ctx context.Context, key string) error {
	if !c.IsEnabled() {
		return fmt.Errorf("Cloudinary storage is not enabled")
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	params := map[string]string{
		"public_id": key,
		"timestamp": timestamp,
	}

	sig := c.generateSignature(params)

	data := url.Values{}
	data.Set("public_id", key)
	data.Set("api_key", c.cfg.APIKey)
	data.Set("timestamp", timestamp)
	data.Set("signature", sig)

	destroyURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/destroy", c.cfg.CloudName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, destroyURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create Cloudinary destroy request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute Cloudinary destroy request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Cloudinary delete returned status %d", resp.StatusCode)
	}

	log.Printf("[Cloudinary] Deleted photo proof: PublicID='%s'", key)
	return nil
}

// generateSignature computes Cloudinary SHA-1 signature over alphabetically sorted parameters.
func (c *CloudinaryStorageService) generateSignature(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}

	toSign := strings.Join(parts, "&") + c.cfg.APISecret
	hasher := sha1.New()
	hasher.Write([]byte(toSign))
	return hex.EncodeToString(hasher.Sum(nil))
}
