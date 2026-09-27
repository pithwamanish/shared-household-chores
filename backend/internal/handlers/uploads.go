package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/choresync/backend/internal/cloud"
	"github.com/choresync/backend/internal/telemetry"
)

// UploadHandler manages photo proof and attachment uploads to S3/Floci.
type UploadHandler struct {
	Storage cloud.StorageService
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(storage cloud.StorageService) *UploadHandler {
	return &UploadHandler{Storage: storage}
}

// UploadProofPhotoResponse represents the JSON response after uploading a photo proof.
type UploadProofPhotoResponse struct {
	Success     bool   `json:"success"`
	URL         string `json:"url"`
	S3Key       string `json:"s3_key"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

// UploadPhoto handles multipart form upload of a chore photo proof.
func (h *UploadHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	if h.Storage == nil || !h.Storage.IsEnabled() {
		respondError(w, "SERVICE_UNAVAILABLE", "Cloud object storage service is not available", http.StatusServiceUnavailable, nil)
		return
	}

	// Limit upload size to 10MB
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, "PAYLOAD_TOO_LARGE", "Upload exceeds maximum allowed size of 10MB", http.StatusRequestEntityTooLarge, nil)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		file, header, err = r.FormFile("photo")
	}
	if err != nil {
		respondError(w, "BAD_REQUEST", "Form field 'file' or 'photo' is required", http.StatusBadRequest, nil)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true,
	}
	if !validExts[ext] {
		respondError(w, "UNSUPPORTED_MEDIA_TYPE", fmt.Sprintf("Unsupported file format '%s'. Allowed: jpg, png, webp, gif", ext), http.StatusUnsupportedMediaType, nil)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		switch ext {
		case ".png":
			contentType = "image/png"
		case ".webp":
			contentType = "image/webp"
		case ".gif":
			contentType = "image/gif"
		default:
			contentType = "image/jpeg"
		}
	}

	key, url, size, err := h.Storage.UploadProofPhoto(r.Context(), header.Filename, contentType, file)
	if err != nil {
		telemetry.Error(r.Context(), "Failed to upload photo to S3", "error", err.Error(), "filename", header.Filename)
		respondError(w, "INTERNAL_SERVER_ERROR", fmt.Sprintf("Failed to store photo in cloud storage: %v", err), http.StatusInternalServerError, nil)
		return
	}

	telemetry.Info(r.Context(), "Chore photo proof uploaded to S3", "s3_key", key, "size_bytes", size)

	respondJSON(w, http.StatusOK, UploadProofPhotoResponse{
		Success:     true,
		URL:         url,
		S3Key:       key,
		ContentType: contentType,
		SizeBytes:   size,
	})
}

// GetPhoto retrieves a photo proof from S3 by key.
func (h *UploadHandler) GetPhoto(w http.ResponseWriter, r *http.Request) {
	if h.Storage == nil || !h.Storage.IsEnabled() {
		respondError(w, "SERVICE_UNAVAILABLE", "Cloud object storage service is not available", http.StatusServiceUnavailable, nil)
		return
	}

	// Chi wildcard param e.g. /uploads/photo/*
	key := chi.URLParam(r, "*")
	if key == "" {
		key = chi.URLParam(r, "key")
	}
	key = strings.TrimPrefix(key, "/")
	if !strings.HasPrefix(key, "proofs/") && !strings.HasPrefix(key, "choresync/") {
		key = "proofs/" + key
	}

	contentType, reader, size, err := h.Storage.GetProofPhoto(r.Context(), key)
	if err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Photo proof '%s' not found in storage", key), http.StatusNotFound, nil)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", contentType)
	if size > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")

	_, _ = io.Copy(w, reader)
}
