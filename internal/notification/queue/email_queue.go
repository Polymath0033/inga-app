package queue

import (
	"context"
	"sync"
	"time"

	"example.com/inga/internal/logger"
	"example.com/inga/internal/notification/emails"
	"go.uber.org/zap"
)

// EmailQueue handles asynchronous email sending
type EmailQueue struct {
	sender chan emailJob
	wg     sync.WaitGroup
	mu     sync.Mutex
	stop   chan struct{}
}

type emailJob struct {
	ctx context.Context
	msg emails.Message
}

// NewEmailQueue creates a new email queue with specified worker count
func NewEmailQueue(emailer emails.Sender, workers int) *EmailQueue {
	if workers <= 0 {
		workers = 1
	}

	q := &EmailQueue{
		sender: make(chan emailJob, 100), // buffered channel
		stop:   make(chan struct{}),
	}

	// Start worker goroutines only if emailer is not nil
	if emailer != nil {
		for i := 0; i < workers; i++ {
			q.wg.Add(1)
			go q.worker(emailer)
		}
	}

	return q
}

// SendTemplate queues an email template to be sent asynchronously
func (q *EmailQueue) SendTemplate(ctx context.Context, msg emails.Message) error {
	// If no workers are running (emailer was nil), silently succeed
	// This prevents blocking the request when email is disabled
	select {
	case q.sender <- emailJob{ctx: ctx, msg: msg}:
		return nil
	case <-q.stop:
		return ErrQueueClosed
	default:
		// Channel is full, return error to indicate backpressure
		return ErrQueueFull
	}
}

// worker processes emails from the queue
func (q *EmailQueue) worker(emailer emails.Sender) {
	defer q.wg.Done()

	for {
		select {
		case job := <-q.sender:
			// Retry logic with exponential backoff
			maxAttempts := 3
			backoff := 1 * time.Second
			var lastErr error
			for attempt := 1; attempt <= maxAttempts; attempt++ {
				sendCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				lastErr = emailer.SendTemplate(sendCtx, job.msg)
				cancel()
				if lastErr == nil {
					break
				}
				logger.Error(job.ctx, "failed to send email from queue, retrying",
					zap.Error(lastErr), zap.Strings("to", job.msg.To), zap.Int("attempt", attempt))
				time.Sleep(backoff)
				backoff *= 2
			}
			if lastErr != nil {
				logger.Error(job.ctx, "failed to send email from queue",
					zap.Error(lastErr), zap.Strings("to", job.msg.To))
			}
		case <-q.stop:
			return
		}
	}
}

// Close gracefully shuts down the queue and waits for pending emails
func (q *EmailQueue) Close(ctx context.Context) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	close(q.stop)

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
