package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/choresync/backend/internal/cloud"
	"github.com/choresync/backend/internal/email"
	"github.com/choresync/backend/internal/server"
	"github.com/choresync/backend/internal/store"
	"github.com/choresync/backend/internal/telemetry"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// Initialize OpenTelemetry Tracing and Metrics
	tracerShutdown, err := telemetry.InitTracer(context.Background())
	if err != nil {
		log.Printf("[OpenTelemetry] Warning: Tracer initialization error: %v", err)
	} else {
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := tracerShutdown(shutdownCtx); err != nil {
				log.Printf("[OpenTelemetry] Error shutting down tracer: %v", err)
			}
		}()
	}

	meterShutdown, err := telemetry.InitMeter(context.Background())
	if err != nil {
		log.Printf("[OpenTelemetry] Warning: Meter initialization error: %v", err)
	} else {
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := meterShutdown(shutdownCtx); err != nil {
				log.Printf("[OpenTelemetry] Error shutting down meter: %v", err)
			}
		}()
	}

	var appStore store.Store
	var pgStore *store.PostgresStore
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		var err error
		pgStore, err = store.NewPostgresStore(databaseURL)
		if err != nil {
			log.Fatalf("Failed to initialize PostgreSQL store: %v", err)
		}
		defer pgStore.Close()
		appStore = pgStore
		log.Println("Database Engine: PostgreSQL (sqlc + pgx/v5 persistent store active)")
	} else {
		appStore = store.NewStore()
		log.Println("Database Engine: In-Memory (ephemeral store active)")
	}

	emailService := email.NewServiceFromEnv()

	// -------------------------------------------------------------------------
	// Initialize Cloud Services (Storage & Queue Providers)
	// -------------------------------------------------------------------------
	cloudCfg := cloud.LoadConfigFromEnv()
	cldCfg := cloud.LoadCloudinaryConfigFromEnv()

	var storageService cloud.StorageService
	var queueService cloud.QueueService

	storageProvider := strings.ToLower(strings.TrimSpace(os.Getenv("STORAGE_PROVIDER")))
	queueProvider := strings.ToLower(strings.TrimSpace(os.Getenv("QUEUE_PROVIDER")))

	// 1. Storage Provider Resolution (Cloudinary vs S3 vs Mock)
	if storageProvider == "cloudinary" || (storageProvider == "" && cldCfg.Enabled) {
		if cldCfg.Enabled {
			log.Printf("[Storage] Initialized Cloudinary Storage Service (Cloud: %s, Folder: %s)", cldCfg.CloudName, cldCfg.Folder)
			storageService = cloud.NewCloudinaryStorageService(cldCfg)
		} else {
			log.Println("[Storage] Warning: Cloudinary requested but credentials not provided. Falling back to mock storage.")
			storageService = cloud.NewMockStorageService()
		}
	} else if storageProvider == "s3" || (storageProvider == "" && cloudCfg.Enabled && (cloudCfg.EndpointURL != "" || cloudCfg.AccessKeyID != "test")) {
		cloudMgr, err := cloud.NewClientManager(context.Background(), cloudCfg)
		if err != nil {
			log.Printf("[Storage] Warning: Failed to initialize S3 client manager: %v. Using mock storage.", err)
			storageService = cloud.NewMockStorageService()
		} else if cloudMgr.IsEnabled() {
			log.Printf("[Storage] Initialized S3 Storage Service (Endpoint: %s, Bucket: %s)", cloudCfg.EndpointURL, cloudCfg.S3BucketName)
			storageService = cloud.NewS3StorageService(cloudMgr)
		} else {
			storageService = cloud.NewMockStorageService()
		}
	} else {
		log.Println("[Storage] Storage running in mock mode.")
		storageService = cloud.NewMockStorageService()
	}

	// 2. Queue Provider Resolution (Neon/Postgres Queue vs SQS vs Mock)
	if queueProvider == "neon" || queueProvider == "postgres" || (queueProvider == "" && pgStore != nil && pgStore.Pool() != nil && cloudCfg.EndpointURL == "") {
		if pgStore != nil && pgStore.Pool() != nil {
			log.Println("[Queue] Initialized Neon / PostgreSQL Queue Service (ACID SKIP LOCKED worker active)")
			queueService = cloud.NewPostgresQueueService(pgStore.Pool())
		} else {
			log.Println("[Queue] Warning: Neon/Postgres queue requested but PostgreSQL not connected. Falling back to mock queue.")
			queueService = cloud.NewMockQueueService()
		}
	} else if queueProvider == "sqs" || (queueProvider == "" && cloudCfg.Enabled && cloudCfg.EndpointURL != "") {
		cloudMgr, err := cloud.NewClientManager(context.Background(), cloudCfg)
		if err != nil {
			log.Printf("[Queue] Warning: Failed to initialize SQS client manager: %v. Using mock queue.", err)
			queueService = cloud.NewMockQueueService()
		} else if cloudMgr.IsEnabled() {
			log.Printf("[Queue] Initialized SQS Queue Service (Endpoint: %s, Queue: %s)", cloudCfg.EndpointURL, cloudCfg.SQSQueueName)
			queueService = cloud.NewSQSQueueService(cloudMgr)
		} else {
			queueService = cloud.NewMockQueueService()
		}
	} else {
		log.Println("[Queue] Queue running in mock mode.")
		queueService = cloud.NewMockQueueService()
	}

	handler := server.NewRouter(appStore, emailService, storageService, queueService)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Start background SQS reminder worker
	queueService.StartReminderWorker(serverCtx, emailService)

	// Listen for syscall signals for graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, shutdownCancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer shutdownCancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("Graceful shutdown timed out... forcing exit.")
			}
		}()

		log.Println("Shutting down ChoreSync HTTP server...")
		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	log.Printf("ChoreSync API backend running at http://localhost:%s", port)
	log.Printf("Health check: http://localhost:%s/healthz", port)
	log.Printf("API Base URL: http://localhost:%s/api", port)

	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server startup failed: %v", err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
	log.Println("ChoreSync server successfully terminated.")
}
