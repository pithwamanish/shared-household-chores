package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/choresync/backend/internal/email"
)

// ReminderJob represents an asynchronous chore reminder notification payload.
type ReminderJob struct {
	ChoreID        string    `json:"chore_id"`
	ChoreTitle     string    `json:"chore_title"`
	DueDate        string    `json:"due_date"`
	AssigneeEmail  string    `json:"assignee_email"`
	AssigneeName   string    `json:"assignee_name"`
	SenderName     string    `json:"sender_name"`
	HouseholdName  string    `json:"household_name"`
	EnqueuedAt     time.Time `json:"enqueued_at"`
}

// QueueService defines message queuing operations for asynchronous notifications and background workers.
type QueueService interface {
	PublishReminder(ctx context.Context, job ReminderJob) (messageID string, err error)
	StartReminderWorker(ctx context.Context, emailService email.Service)
	IsEnabled() bool
}

// SQSQueueService implements QueueService using AWS SQS / Floci SQS emulator.
type SQSQueueService struct {
	mgr *ClientManager
}

// NewSQSQueueService creates an SQS-backed queue service.
func NewSQSQueueService(mgr *ClientManager) *SQSQueueService {
	return &SQSQueueService{mgr: mgr}
}

func (q *SQSQueueService) IsEnabled() bool {
	return q.mgr != nil && q.mgr.IsEnabled() && q.mgr.QueueURL() != ""
}

// PublishReminder serializes and enqueues a ReminderJob into SQS.
func (q *SQSQueueService) PublishReminder(ctx context.Context, job ReminderJob) (string, error) {
	if !q.IsEnabled() {
		return "", fmt.Errorf("SQS queue service is not enabled or queue URL is missing")
	}

	if job.EnqueuedAt.IsZero() {
		job.EnqueuedAt = time.Now().UTC()
	}

	body, err := json.Marshal(job)
	if err != nil {
		return "", fmt.Errorf("failed to serialize reminder job: %w", err)
	}

	queueURL := q.mgr.QueueURL()
	client := q.mgr.SQSClient()

	out, err := client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(body)),
	})
	if err != nil {
		return "", fmt.Errorf("failed to send message to SQS (%s): %w", queueURL, err)
	}

	msgID := ""
	if out.MessageId != nil {
		msgID = *out.MessageId
	}

	log.Printf("[Local Cloud SQS] Enqueued chore reminder for '%s' (MessageId: %s)", job.ChoreTitle, msgID)
	return msgID, nil
}

// StartReminderWorker launches a resilient background consumer loop listening on SQS.
func (q *SQSQueueService) StartReminderWorker(ctx context.Context, emailService email.Service) {
	if !q.IsEnabled() {
		log.Println("[Local Cloud SQS] Worker not started: SQS is not enabled")
		return
	}

	go func() {
		log.Printf("[Local Cloud SQS] Starting background reminder worker for queue: %s", q.mgr.QueueURL())
		client := q.mgr.SQSClient()
		queueURL := q.mgr.QueueURL()

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("[Local Cloud SQS] Stopping background reminder worker...")
				return
			default:
				receiveCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				out, err := client.ReceiveMessage(receiveCtx, &sqs.ReceiveMessageInput{
					QueueUrl:            aws.String(queueURL),
					MaxNumberOfMessages: 5,
					WaitTimeSeconds:     5, // Long polling
				})
				cancel()

				if err != nil {
					if ctx.Err() != nil {
						return
					}
					// Backoff on transient error
					time.Sleep(1 * time.Second)
					continue
				}

				for _, msg := range out.Messages {
					if msg.Body == nil {
						continue
					}

					var job ReminderJob
					if err := json.Unmarshal([]byte(*msg.Body), &job); err != nil {
						log.Printf("[Local Cloud SQS] Error deserializing job (%s): %v", *msg.MessageId, err)
						continue
					}

					log.Printf("[Local Cloud SQS] Consuming reminder job: chore='%s', to='%s'", job.ChoreTitle, job.AssigneeEmail)
					if emailService != nil && job.AssigneeEmail != "" {
						err := emailService.SendChoreReminder(
							job.AssigneeEmail,
							job.AssigneeName,
							job.SenderName,
							job.ChoreTitle,
							job.DueDate,
							job.HouseholdName,
						)
						if err != nil {
							log.Printf("[Local Cloud SQS] Failed to dispatch email for job: %v", err)
						} else {
							log.Printf("[Local Cloud SQS] Email successfully dispatched for chore '%s'", job.ChoreTitle)
						}
					}

					// Acknowledge and delete message
					delCtx, delCancel := context.WithTimeout(ctx, 5*time.Second)
					_, _ = client.DeleteMessage(delCtx, &sqs.DeleteMessageInput{
						QueueUrl:      aws.String(queueURL),
						ReceiptHandle: msg.ReceiptHandle,
					})
					delCancel()
				}
			}
		}
	}()
}

// MockQueueService provides an in-memory queue for offline unit testing.
type MockQueueService struct {
	mu       sync.Mutex
	jobs     []ReminderJob
	dispatched []ReminderJob
}

// NewMockQueueService initializes an in-memory mock queue service.
func NewMockQueueService() *MockQueueService {
	return &MockQueueService{
		jobs:       make([]ReminderJob, 0),
		dispatched: make([]ReminderJob, 0),
	}
}

func (m *MockQueueService) IsEnabled() bool {
	return true
}

func (m *MockQueueService) PublishReminder(ctx context.Context, job ReminderJob) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	job.EnqueuedAt = time.Now().UTC()
	m.jobs = append(m.jobs, job)
	return fmt.Sprintf("mock-msg-%d", len(m.jobs)), nil
}

func (m *MockQueueService) StartReminderWorker(ctx context.Context, emailService email.Service) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(100 * time.Millisecond):
				m.mu.Lock()
				if len(m.jobs) > 0 {
					job := m.jobs[0]
					m.jobs = m.jobs[1:]
					m.dispatched = append(m.dispatched, job)
					m.mu.Unlock()

					if emailService != nil && job.AssigneeEmail != "" {
						_ = emailService.SendChoreReminder(
							job.AssigneeEmail,
							job.AssigneeName,
							job.SenderName,
							job.ChoreTitle,
							job.DueDate,
							job.HouseholdName,
						)
					}
				} else {
					m.mu.Unlock()
				}
			}
		}
	}()
}

// GetQueuedJobs returns all pending queued jobs.
func (m *MockQueueService) GetQueuedJobs() []ReminderJob {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]ReminderJob, len(m.jobs))
	copy(copied, m.jobs)
	return copied
}

// GetDispatchedJobs returns all dispatched jobs.
func (m *MockQueueService) GetDispatchedJobs() []ReminderJob {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]ReminderJob, len(m.dispatched))
	copy(copied, m.dispatched)
	return copied
}
