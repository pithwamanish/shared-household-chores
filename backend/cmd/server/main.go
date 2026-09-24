package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		pgStore, err := store.NewPostgresStore(databaseURL)
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

	// Initialize Cloud Services (AWS / Floci local cloud emulator)
	cloudCfg := cloud.LoadConfigFromEnv()
	var storageService cloud.StorageService
	var queueService cloud.QueueService

	cloudMgr, err := cloud.NewClientManager(context.Background(), cloudCfg)
	if err != nil {
		log.Printf("[Cloud] Warning: Failed to initialize AWS/Floci client manager: %v. Using in-memory cloud fallbacks.", err)
		storageService = cloud.NewMockStorageService()
		queueService = cloud.NewMockQueueService()
	} else if cloudMgr.IsEnabled() {
		log.Printf("[Cloud] Initialized Local Cloud Emulator / AWS (Endpoint: %s, S3: %s, SQS: %s)", cloudCfg.EndpointURL, cloudCfg.S3BucketName, cloudCfg.SQSQueueName)
		storageService = cloud.NewS3StorageService(cloudMgr)
		queueService = cloud.NewSQSQueueService(cloudMgr)
	} else {
		log.Println("[Cloud] Cloud services disabled or running in mock mode.")
		storageService = cloud.NewMockStorageService()
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
