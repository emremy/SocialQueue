package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/emremy/socialqueue/internal/queue/jobs"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type WorkerConfig struct {
	WorkerID              string
	Concurrency           int
	PollInterval          time.Duration
	Logger                *slog.Logger
	BatchSize             int
	StaleTimeout          time.Duration
	StaleRecoveryInterval time.Duration
	Redis                 *redis.Client
}

type Worker struct {
	config      WorkerConfig
	jobService  *jobs.Service
	executor    *Executor
	retryPolicy RetryPolicy
}

func New(config WorkerConfig, jobService *jobs.Service, executor *Executor) *Worker {
	if config.WorkerID == "" {
		config.WorkerID = "worker-" + uuid.NewString()
	}

	if config.PollInterval == 0 {
		config.PollInterval = 2 * time.Second
	}

	if config.BatchSize == 0 {
		config.BatchSize = 5
	}

	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	if config.StaleTimeout == 0 {
		config.StaleTimeout = 30 * time.Second
	}

	if config.StaleRecoveryInterval == 0 {
		config.StaleRecoveryInterval = 10 * time.Second
	}

	if executor == nil {
		executor = NewExecutor()
	}

	return &Worker{
		config:      config,
		jobService:  jobService,
		executor:    executor,
		retryPolicy: DefaultRetryPolicy(),
	}

}

func (w *Worker) Run(ctx context.Context) {
	w.config.Logger.Info("worker started", "worker_id", w.config.WorkerID)

	staleTicker := time.NewTicker(w.config.StaleRecoveryInterval)
	defer staleTicker.Stop()

	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()

	pubsub := w.config.Redis.Subscribe(ctx, jobs.JobCreatedChannel)
	defer pubsub.Close()

	messages := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			w.config.Logger.Info("worker stopped", "worker_id", w.config.WorkerID)
			return

		case <-ticker.C:
			w.processBatch(ctx)
		case <-staleTicker.C:
			w.recoverStaleJobs(ctx)
		case <-messages:
			w.config.Logger.Info("received redis wake-up signal")

			w.processBatch(ctx)
		}

	}
}

func (w *Worker) recoverStaleJobs(ctx context.Context) {
	recovered, err := w.jobService.RecoverStaleJobs(ctx, w.config.StaleTimeout)
	if err != nil {
		w.config.Logger.Error("failed to recover stale jobs", "error", err)
	} else if recovered > 0 {
		w.config.Logger.Warn("stale jobs recovered", "count", recovered)
	}
}

func (w *Worker) processBatch(ctx context.Context) {
	claimedJobs, err := w.jobService.ClaimJobs(ctx, w.config.WorkerID, w.config.BatchSize)
	if err != nil {
		w.config.Logger.Error("failed to claim jobs", "error", err)
		return
	}

	if len(claimedJobs) == 0 {
		return
	}

	w.config.Logger.Info("jobs claimed", "count", len(claimedJobs))

	concurrency := w.config.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	sem := make(chan struct{}, concurrency)

	var wg sync.WaitGroup

	for _, job := range claimedJobs {
		select {
		case <-ctx.Done():
			w.config.Logger.Info("batch processing stopped", "reason", ctx.Err())
			return

		case sem <- struct{}{}:
		}

		wg.Add(1)

		go func(job jobs.Job) {
			defer wg.Done()
			defer func() {
				<-sem
			}()

			w.processJob(ctx, job)
		}(job)
	}

	wg.Wait()

}

func (w *Worker) processJob(ctx context.Context, job jobs.Job) {
	w.config.Logger.Info(
		"job processing started",
		"id", job.ID,
		"type", job.Type,
		"attempts", job.Attempts,
	)

	attempt, err := w.jobService.StartAttempt(ctx, job, w.config.WorkerID)
	if err != nil {
		w.config.Logger.Error(
			"failed to start job attempt",
			"id", job.ID,
			"error", err,
		)
		return
	}

	err = w.executor.Execute(ctx, job)

	if err != nil {
		errMessage := err.Error()

		_ = w.jobService.FinishAttempt(
			ctx,
			attempt.ID,
			jobs.AttemptFailed,
			&errMessage,
			attempt.StartedAt,
		)
		var nextRunAt *time.Time

		if job.Attempts < job.MaxAttempts {
			t := w.retryPolicy.NextRunAt(job.Attempts)
			nextRunAt = &t
		}

		if markErr := w.jobService.MarkJobFailed(ctx, job.ID, err.Error(), nextRunAt); markErr != nil {
			w.config.Logger.Error(
				"failed to update failed job",
				"id", job.ID,
				"error", markErr,
			)
			return
		}

		w.config.Logger.Warn(
			"job failed",
			"id", job.ID,
			"error", err.Error(),
			"will_retry", nextRunAt != nil,
		)

		return
	}
	if err := w.jobService.FinishAttempt(
		ctx,
		attempt.ID,
		jobs.AttemptSucceeded,
		nil,
		attempt.StartedAt,
	); err != nil {
		w.config.Logger.Error(
			"failed to finish job attempt",
			"id", job.ID,
			"attempt_id", attempt.ID,
			"error", err,
		)
		return
	}

	if err := w.jobService.MarkJobSucceeded(ctx, job.ID); err != nil {
		w.config.Logger.Error("failed to mark job as succeeded", "id", job.ID, "error", err)
		return
	}

	w.config.Logger.Info("job succeeded", "id", job.ID)
}
