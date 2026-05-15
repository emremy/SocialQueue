package jobs

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateJob(ctx context.Context, job *Job) error {
	return s.db.WithContext(ctx).Create(job).Error
}

func (s *Store) GetByID(ctx context.Context, id string) (*Job, error) {
	var job Job

	err := s.db.WithContext(ctx).Where("id = ?", id).First(&job).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrJobNotFound
		}
		return nil, err
	}

	return &job, nil
}

func (s *Store) ClaimJobs(ctx context.Context, workerID string, limit int) ([]Job, error) {
	var jobs []Job

	err := s.db.WithContext(ctx).Raw(`
		WITH picked AS (
			SELECT id
			FROM jobs
			WHERE status = ?
			  AND run_at <= NOW()
			ORDER BY priority DESC, created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT ?
		)
		UPDATE jobs
		SET status = ?,
		    locked_at = NOW(),
		    worker_id = ?,
		    attempts = attempts + 1,
		    updated_at = NOW()
		WHERE id IN (SELECT id FROM picked)
		RETURNING *
	`, StatusQueued, limit, StatusProcessing, workerID).Scan(&jobs).Error

	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func (s *Store) MarkJobSucceeded(ctx context.Context, jobID string) error {
	return s.db.WithContext(ctx).
		Model(&Job{}).
		Where("id = ?", jobID).
		Updates(map[string]any{
			"status":     StatusSucceeded,
			"locked_at":  nil,
			"worker_id":  nil,
			"last_error": nil,
			"updated_at": time.Now(),
		}).
		Error
}

func (s *Store) MarkJobFailed(
	ctx context.Context,
	jobID string,
	errMessage string,
	nextRunAt *time.Time,
) error {
	updates := map[string]any{
		"last_error": errMessage,
		"locked_at":  nil,
		"worker_id":  nil,
		"updated_at": time.Now(),
	}

	if nextRunAt != nil {
		updates["status"] = StatusQueued
		updates["run_at"] = *nextRunAt
	} else {
		updates["status"] = StatusFailed
	}

	return s.db.WithContext(ctx).
		Model(&Job{}).
		Where("id = ?", jobID).
		Updates(updates).
		Error
}

func (s *Store) RecoverStaleJobs(ctx context.Context, timeout time.Duration) (int64, error) {
	result := s.db.WithContext(ctx).
		Model(&Job{}).
		Where("status = ?", StatusProcessing).
		Where("locked_at < NOW() - (? * INTERVAL '1 millisecond')", timeout.Milliseconds()).
		Updates(map[string]any{
			"status":     StatusQueued,
			"locked_at":  nil,
			"worker_id":  nil,
			"last_error": "job recovered from stale processing state",
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (s *Store) StartAttempt(ctx context.Context, job Job, workerID string) (*JobAttempt, error) {
	attempt := &JobAttempt{
		JobID:         job.ID,
		AttemptNumber: job.Attempts,
		WorkerID:      workerID,
		Status:        AttemptRunning,
	}

	if err := s.db.Create(attempt).Error; err != nil {
		return nil, err
	}

	return attempt, nil
}

func (s *Store) FinishAttempt(
	ctx context.Context,
	attemptID string,
	status AttemptStatus,
	errMessage *string,
	startedAt time.Time,
) error {
	finishedAt := time.Now()
	durationMs := int(finishedAt.Sub(startedAt).Milliseconds())

	return s.db.WithContext(ctx).
		Model(&JobAttempt{}).
		Where("id = ?", attemptID).
		Updates(map[string]any{
			"status":      status,
			"error":       errMessage,
			"finished_at": finishedAt,
			"duration_ms": durationMs,
		}).
		Error
}
