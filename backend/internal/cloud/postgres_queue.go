package cloud

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/choresync/backend/internal/email"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresQueueService implements QueueService using PostgreSQL / Neon.
// It leverages row-level locks (FOR UPDATE SKIP LOCKED) to achieve ACID-safe,
// lock-free background job processing across multiple distributed replicas.
type PostgresQueueService struct {
	pool *pgxpool.Pool
}

// NewPostgresQueueService creates a PostgreSQL-backed queue service.
func NewPostgresQueueService(pool *pgxpool.Pool) *PostgresQueueService {
	p := &PostgresQueueService{pool: pool}
	if pool != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := p.ensureSchema(ctx); err != nil {
			log.Printf("[Neon Queue] Warning: Failed to ensure reminder_jobs table: %v", err)
		}
	}
	return p
}

func (p *PostgresQueueService) ensureSchema(ctx context.Context) error {
	ddl := `
	CREATE TABLE IF NOT EXISTS reminder_jobs (
		id TEXT PRIMARY KEY,
		chore_id TEXT NOT NULL,
		chore_title TEXT NOT NULL,
		due_date TEXT NOT NULL DEFAULT '',
		assignee_email TEXT NOT NULL DEFAULT '',
		assignee_name TEXT NOT NULL DEFAULT '',
		sender_name TEXT NOT NULL DEFAULT '',
		household_name TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'pending',
		attempts INTEGER NOT NULL DEFAULT 0,
		last_error TEXT,
		enqueued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_reminder_jobs_status_enqueued ON reminder_jobs(status, enqueued_at);
	`
	_, err := p.pool.Exec(ctx, ddl)
	return err
}

// IsEnabled returns true if the connection pool is non-nil.
func (p *PostgresQueueService) IsEnabled() bool {
	return p.pool != nil
}

// PublishReminder persists a reminder job into the reminder_jobs PostgreSQL table.
func (p *PostgresQueueService) PublishReminder(ctx context.Context, job ReminderJob) (string, error) {
	if !p.IsEnabled() {
		return "", fmt.Errorf("PostgreSQL/Neon queue is not enabled or connection pool is nil")
	}

	if job.EnqueuedAt.IsZero() {
		job.EnqueuedAt = time.Now().UTC()
	}

	b := make([]byte, 8)
	_, _ = rand.Read(b)
	msgID := fmt.Sprintf("neon-job-%s", hex.EncodeToString(b))

	query := `
	INSERT INTO reminder_jobs (
		id, chore_id, chore_title, due_date, assignee_email, assignee_name, sender_name, household_name, status, enqueued_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'pending', $9, NOW())
	`
	_, err := p.pool.Exec(ctx, query,
		msgID, job.ChoreID, job.ChoreTitle, job.DueDate, job.AssigneeEmail, job.AssigneeName, job.SenderName, job.HouseholdName, job.EnqueuedAt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to enqueue reminder job in Neon/PostgreSQL: %w", err)
	}

	log.Printf("[Neon Queue] Enqueued chore reminder for '%s' (MessageId: %s)", job.ChoreTitle, msgID)
	return msgID, nil
}

// StartReminderWorker launches a background polling consumer loop utilizing FOR UPDATE SKIP LOCKED.
func (p *PostgresQueueService) StartReminderWorker(ctx context.Context, emailService email.Service) {
	if !p.IsEnabled() {
		log.Println("[Neon Queue] Worker not started: database pool is nil")
		return
	}

	go func() {
		log.Println("[Neon Queue] Starting background reminder worker polling Neon PostgreSQL...")
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("[Neon Queue] Stopping background reminder worker...")
				return
			case <-ticker.C:
				p.processNextJob(ctx, emailService)
			}
		}
	}()
}

func (p *PostgresQueueService) processNextJob(ctx context.Context, emailService email.Service) {
	query := `
	UPDATE reminder_jobs
	SET status = 'processing', attempts = attempts + 1, updated_at = NOW()
	WHERE id = (
		SELECT id FROM reminder_jobs
		WHERE status = 'pending'
		ORDER BY enqueued_at ASC
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	)
	RETURNING id, chore_id, chore_title, due_date, assignee_email, assignee_name, sender_name, household_name;
	`

	pollCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var job ReminderJob
	var jobID string
	err := p.pool.QueryRow(pollCtx, query).Scan(
		&jobID, &job.ChoreID, &job.ChoreTitle, &job.DueDate,
		&job.AssigneeEmail, &job.AssigneeName, &job.SenderName, &job.HouseholdName,
	)
	if err != nil {
		// No pending jobs or transient error
		return
	}

	log.Printf("[Neon Queue] Consuming reminder job %s: chore='%s', to='%s'", jobID, job.ChoreTitle, job.AssigneeEmail)

	var dispatchErr error
	if emailService != nil && job.AssigneeEmail != "" {
		dispatchErr = emailService.SendChoreReminder(
			job.AssigneeEmail,
			job.AssigneeName,
			job.SenderName,
			job.ChoreTitle,
			job.DueDate,
			job.HouseholdName,
		)
	}

	updateCtx, updateCancel := context.WithTimeout(ctx, 5*time.Second)
	defer updateCancel()

	if dispatchErr != nil {
		log.Printf("[Neon Queue] Failed to dispatch email for job %s: %v", jobID, dispatchErr)
		_, _ = p.pool.Exec(updateCtx, `UPDATE reminder_jobs SET status = 'failed', last_error = $1, updated_at = NOW() WHERE id = $2`, dispatchErr.Error(), jobID)
	} else {
		log.Printf("[Neon Queue] Email successfully dispatched for chore '%s' (Job: %s)", job.ChoreTitle, jobID)
		_, _ = p.pool.Exec(updateCtx, `UPDATE reminder_jobs SET status = 'completed', updated_at = NOW() WHERE id = $1`, jobID)
	}
}
