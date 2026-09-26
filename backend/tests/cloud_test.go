package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/choresync/backend/internal/cloud"
	"github.com/choresync/backend/internal/email"
	"github.com/choresync/backend/internal/handlers"
	"github.com/choresync/backend/internal/models"
	"github.com/choresync/backend/internal/server"
	"github.com/choresync/backend/internal/store"
)

func TestStorageServiceMock(t *testing.T) {
	ctx := context.Background()
	storage := cloud.NewMockStorageService()

	sampleData := []byte("fake-image-bytes-jpeg")
	filename := "oven-cleaned.jpg"
	contentType := "image/jpeg"

	// 1. Upload proof photo
	key, url, size, err := storage.UploadProofPhoto(ctx, filename, contentType, bytes.NewReader(sampleData))
	if err != nil {
		t.Fatalf("unexpected error uploading proof photo: %v", err)
	}
	if !strings.HasPrefix(key, "proofs/") {
		t.Errorf("expected key to start with 'proofs/', got '%s'", key)
	}
	if !strings.HasPrefix(url, "/api/v1/uploads/photo/") {
		t.Errorf("expected url to start with '/api/v1/uploads/photo/', got '%s'", url)
	}
	if size != int64(len(sampleData)) {
		t.Errorf("expected size %d, got %d", len(sampleData), size)
	}

	// 2. Retrieve proof photo
	ct, reader, rSize, err := storage.GetProofPhoto(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error getting proof photo: %v", err)
	}
	defer reader.Close()

	if ct != contentType {
		t.Errorf("expected content type %s, got %s", contentType, ct)
	}
	if rSize != int64(len(sampleData)) {
		t.Errorf("expected size %d, got %d", len(sampleData), rSize)
	}

	downloaded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read retrieved photo: %v", err)
	}
	if !bytes.Equal(downloaded, sampleData) {
		t.Errorf("downloaded content mismatch")
	}

	// 3. Delete proof photo
	if err := storage.DeleteProofPhoto(ctx, key); err != nil {
		t.Fatalf("unexpected error deleting photo: %v", err)
	}

	_, _, _, err = storage.GetProofPhoto(ctx, key)
	if err == nil {
		t.Fatal("expected error getting deleted photo, got nil")
	}
}

func TestQueueServiceMock(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	queue := cloud.NewMockQueueService()
	emailSvc := email.NewMockService(email.Config{})

	// Start reminder worker
	queue.StartReminderWorker(ctx, emailSvc)

	job := cloud.ReminderJob{
		ChoreID:       "c-dishes",
		ChoreTitle:    "Wash Dinner Dishes",
		DueDate:       "today",
		AssigneeEmail: "liam@roommates.local",
		AssigneeName:  "Liam Vance",
		SenderName:    "Sarah Chen",
		HouseholdName: "Oakwood Flatmates",
	}

	msgID, err := queue.PublishReminder(ctx, job)
	if err != nil {
		t.Fatalf("unexpected error publishing reminder: %v", err)
	}
	if msgID == "" {
		t.Fatal("expected non-empty message ID")
	}

	// Wait for worker to consume job and dispatch email
	time.Sleep(300 * time.Millisecond)

	dispatched := queue.GetDispatchedJobs()
	if len(dispatched) == 0 {
		t.Fatalf("expected job to be dispatched by worker, got 0")
	}
	if dispatched[0].ChoreTitle != "Wash Dinner Dishes" {
		t.Errorf("expected chore title 'Wash Dinner Dishes', got '%s'", dispatched[0].ChoreTitle)
	}

	emails := emailSvc.GetRecentDevEmails()
	if len(emails) == 0 {
		t.Fatalf("expected 1 email in mock mailbox, got 0")
	}
	if emails[0].To != "liam@roommates.local" {
		t.Errorf("expected email to liam@roommates.local, got %s", emails[0].To)
	}
}

func TestUploadPhotoEndpoint(t *testing.T) {
	s := store.NewStore()
	em := email.NewMockService(email.Config{})
	storage := cloud.NewMockStorageService()
	queue := cloud.NewMockQueueService()

	router := server.NewRouter(s, em, storage, queue)

	// 1. Prepare multipart upload body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "proof.jpg")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	fakeImage := []byte("JPEG_EXIF_IMAGE_DATA_12345")
	_, _ = part.Write(fakeImage)
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/uploads/photo", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK uploading photo, got %d: %s", w.Code, w.Body.String())
	}

	var resp handlers.UploadProofPhotoResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse upload response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success: true")
	}
	if resp.S3Key == "" || !strings.HasPrefix(resp.S3Key, "proofs/") {
		t.Errorf("invalid s3_key in response: %s", resp.S3Key)
	}
	if resp.SizeBytes != int64(len(fakeImage)) {
		t.Errorf("expected size_bytes %d, got %d", len(fakeImage), resp.SizeBytes)
	}

	// 2. Fetch the uploaded image through GET /api/v1/uploads/photo/{key}
	reqGet := httptest.NewRequest("GET", resp.URL, nil)
	wGet := httptest.NewRecorder()
	router.ServeHTTP(wGet, reqGet)

	if wGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK fetching photo, got %d: %s", wGet.Code, wGet.Body.String())
	}
	if !bytes.Equal(wGet.Body.Bytes(), fakeImage) {
		t.Errorf("fetched photo content does not match uploaded bytes")
	}
	if wGet.Header().Get("Content-Type") != "image/jpeg" {
		t.Errorf("expected Content-Type image/jpeg, got %s", wGet.Header().Get("Content-Type"))
	}
}

func TestNudgeChoreWithSQSIntegration(t *testing.T) {
	s := store.NewStore()
	em := email.NewMockService(email.Config{})
	storage := cloud.NewMockStorageService()
	queue := cloud.NewMockQueueService()

	router := server.NewRouter(s, em, storage, queue)

	// Send friendly nudge for chore c-vacuum-living (assigned to Liam Vance: liam@roommates.local)
	nudgeReq := models.ChoreNudgeRequest{
		SenderMemberID: "m-sarah",
	}
	reqBody, _ := json.Marshal(nudgeReq)
	req := httptest.NewRequest("POST", "/api/v1/chores/c-vacuum-living/nudge", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for nudge, got %d: %s", w.Code, w.Body.String())
	}

	var nudgeResp models.ChoreNudgeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &nudgeResp); err != nil {
		t.Fatalf("failed to decode nudge response: %v", err)
	}

	if !nudgeResp.Success {
		t.Errorf("expected success: true")
	}
	if !nudgeResp.EmailDispatched {
		t.Errorf("expected email_dispatched: true")
	}
	if !nudgeResp.SQSQueued {
		t.Errorf("expected sqs_queued: true")
	}

	// Verify job was enqueued in mock SQS
	queued := queue.GetQueuedJobs()
	if len(queued) != 1 {
		t.Fatalf("expected 1 job in queue, got %d", len(queued))
	}
	if queued[0].AssigneeEmail != "liam@example.com" {
		t.Errorf("expected assignee email liam@example.com, got %s", queued[0].AssigneeEmail)
	}
}

func TestCloudinaryConfigParsing(t *testing.T) {
	os.Setenv("CLOUDINARY_URL", "cloudinary://123456789:abcdefghijk@my-cloud-name")
	defer os.Unsetenv("CLOUDINARY_URL")

	cfg := cloud.LoadCloudinaryConfigFromEnv()
	if !cfg.Enabled {
		t.Fatal("expected Cloudinary to be enabled when CLOUDINARY_URL is set")
	}
	if cfg.CloudName != "my-cloud-name" {
		t.Errorf("expected CloudName 'my-cloud-name', got '%s'", cfg.CloudName)
	}
	if cfg.APIKey != "123456789" {
		t.Errorf("expected APIKey '123456789', got '%s'", cfg.APIKey)
	}
	if cfg.APISecret != "abcdefghijk" {
		t.Errorf("expected APISecret 'abcdefghijk', got '%s'", cfg.APISecret)
	}
	if cfg.Folder != "choresync/proofs" {
		t.Errorf("expected default Folder 'choresync/proofs', got '%s'", cfg.Folder)
	}
}

func TestCloudinaryServiceEnabled(t *testing.T) {
	cfg := cloud.CloudinaryConfig{
		CloudName: "test-cloud",
		APIKey:    "test-key",
		APISecret: "my_api_secret",
		Folder:    "choresync/proofs",
		Enabled:   true,
	}
	svc := cloud.NewCloudinaryStorageService(cfg)
	if !svc.IsEnabled() {
		t.Fatal("expected service to be enabled")
	}

	disabledSvc := cloud.NewCloudinaryStorageService(cloud.CloudinaryConfig{Enabled: false})
	if disabledSvc.IsEnabled() {
		t.Fatal("expected service to be disabled")
	}
}
