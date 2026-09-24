package cloud

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Config encapsulates configuration for AWS / Floci cloud emulator integration.
type Config struct {
	Region          string
	EndpointURL     string // e.g. "http://floci:4566" or "http://localhost:4566"
	AccessKeyID     string
	SecretAccessKey string
	S3BucketName    string
	SQSQueueName    string
	Enabled         bool
}

// LoadConfigFromEnv reads AWS / local cloud configuration from environment variables.
func LoadConfigFromEnv() Config {
	endpoint := strings.TrimSpace(os.Getenv("AWS_ENDPOINT_URL"))
	if endpoint == "" {
		endpoint = strings.TrimSpace(os.Getenv("LOCALSTACK_ENDPOINT"))
	}

	region := strings.TrimSpace(os.Getenv("AWS_REGION"))
	if region == "" {
		region = "us-east-1"
	}

	accessKey := strings.TrimSpace(os.Getenv("AWS_ACCESS_KEY_ID"))
	if accessKey == "" {
		accessKey = "test"
	}

	secretKey := strings.TrimSpace(os.Getenv("AWS_SECRET_ACCESS_KEY"))
	if secretKey == "" {
		secretKey = "test"
	}

	bucket := strings.TrimSpace(os.Getenv("S3_BUCKET_NAME"))
	if bucket == "" {
		bucket = "choresync-proofs"
	}

	queue := strings.TrimSpace(os.Getenv("SQS_QUEUE_NAME"))
	if queue == "" {
		queue = "choresync-reminders"
	}

	enabledStr := strings.ToLower(strings.TrimSpace(os.Getenv("ENABLE_CLOUD_SERVICES")))
	enabled := true
	if enabledStr == "false" || enabledStr == "0" || enabledStr == "no" {
		enabled = false
	}

	return Config{
		Region:          region,
		EndpointURL:     endpoint,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		S3BucketName:    bucket,
		SQSQueueName:    queue,
		Enabled:         enabled,
	}
}

// ClientManager holds initialized AWS / Floci clients.
type ClientManager struct {
	cfg       Config
	awsCfg    aws.Config
	s3Client  *s3.Client
	sqsClient *sqs.Client
	queueURL  string
	mu        sync.RWMutex
}

// NewClientManager initializes AWS clients with support for local cloud emulators (Floci, MinIO, LocalStack).
func NewClientManager(ctx context.Context, cfg Config) (*ClientManager, error) {
	if !cfg.Enabled {
		return &ClientManager{cfg: cfg}, nil
	}

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	// Initialize S3 client with path-style addressing for emulators
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.EndpointURL != "" {
			o.BaseEndpoint = aws.String(cfg.EndpointURL)
			o.UsePathStyle = true
		}
	})

	// Initialize SQS client
	sqsClient := sqs.NewFromConfig(awsCfg, func(o *sqs.Options) {
		if cfg.EndpointURL != "" {
			o.BaseEndpoint = aws.String(cfg.EndpointURL)
		}
	})

	mgr := &ClientManager{
		cfg:       cfg,
		awsCfg:    awsCfg,
		s3Client:  s3Client,
		sqsClient: sqsClient,
	}

	// Auto-provision bucket and queue if running against local cloud emulator or enabled
	if cfg.EndpointURL != "" {
		initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := mgr.ensureInfrastructure(initCtx); err != nil {
			log.Printf("[Local Cloud] Warning: Infrastructure provisioning deferred or failed: %v", err)
		}
	}

	return mgr, nil
}

// IsEnabled returns true if cloud integration is active.
func (m *ClientManager) IsEnabled() bool {
	return m.cfg.Enabled && m.s3Client != nil && m.sqsClient != nil
}

// Config returns the configuration.
func (m *ClientManager) Config() Config {
	return m.cfg
}

// S3Client returns the S3 client instance.
func (m *ClientManager) S3Client() *s3.Client {
	return m.s3Client
}

// SQSClient returns the SQS client instance.
func (m *ClientManager) SQSClient() *sqs.Client {
	return m.sqsClient
}

// QueueURL returns the resolved SQS queue URL.
func (m *ClientManager) QueueURL() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.queueURL
}

func (m *ClientManager) ensureInfrastructure(ctx context.Context) error {
	// 1. Ensure S3 bucket exists
	_, err := m.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(m.cfg.S3BucketName),
	})
	if err != nil {
		log.Printf("[Local Cloud] Creating S3 bucket '%s' on %s...", m.cfg.S3BucketName, m.cfg.EndpointURL)
		_, err = m.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(m.cfg.S3BucketName),
		})
		if err != nil && !strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") && !strings.Contains(err.Error(), "BucketAlreadyExists") {
			return fmt.Errorf("failed to create S3 bucket %s: %w", m.cfg.S3BucketName, err)
		}
		log.Printf("[Local Cloud] S3 bucket '%s' ready.", m.cfg.S3BucketName)
	}

	// 2. Ensure SQS queue exists
	getOut, err := m.sqsClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(m.cfg.SQSQueueName),
	})
	if err == nil && getOut.QueueUrl != nil {
		m.mu.Lock()
		m.queueURL = *getOut.QueueUrl
		m.mu.Unlock()
		log.Printf("[Local Cloud] Found existing SQS queue '%s': %s", m.cfg.SQSQueueName, m.queueURL)
		return nil
	}

	log.Printf("[Local Cloud] Creating SQS queue '%s' on %s...", m.cfg.SQSQueueName, m.cfg.EndpointURL)
	createOut, err := m.sqsClient.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(m.cfg.SQSQueueName),
	})
	if err != nil {
		return fmt.Errorf("failed to create SQS queue %s: %w", m.cfg.SQSQueueName, err)
	}

	m.mu.Lock()
	if createOut.QueueUrl != nil {
		m.queueURL = *createOut.QueueUrl
	}
	m.mu.Unlock()
	log.Printf("[Local Cloud] SQS queue '%s' ready at %s", m.cfg.SQSQueueName, m.queueURL)

	return nil
}
