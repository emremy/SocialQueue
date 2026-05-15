package jobs

import (
	"context"
	"time"

	"gorm.io/datatypes"
)

type CreateJobInput struct {
	Type        string
	Payload     datatypes.JSON
	Priority    int
	MaxAttempts int
	RunAt       *time.Time
}

type Service struct {
	store     *Store
	publisher *Publisher
}

func NewService(store *Store, publisher *Publisher) *Service {
	return &Service{store: store, publisher: publisher}
}

func (s *Service) CreateJob(ctx context.Context, input CreateJobInput) (*Job, error) {
	job := &Job{
		Type:        input.Type,
		Payload:     input.Payload,
		Priority:    input.Priority,
		MaxAttempts: input.MaxAttempts,
	}

	if input.RunAt != nil {
		job.RunAt = *input.RunAt
	}

	if err := s.store.CreateJob(ctx, job); err != nil {
		return nil, err
	}

	if s.publisher != nil {
		if err := s.publisher.PublishJobCreated(ctx); err != nil {
			return nil, err
		}
	}

	return job, nil
}

func (s *Service) GetJobByID(ctx context.Context, id string) (*Job, error) {
	return s.store.GetByID(ctx, id)
}

func (s *Service) ClaimJobs(ctx context.Context, workerID string, limit int) ([]Job, error) {
	return s.store.ClaimJobs(ctx, workerID, limit)
}

func (s *Service) MarkJobSucceeded(ctx context.Context, jobID string) error {
	return s.store.MarkJobSucceeded(ctx, jobID)
}

func (s *Service) MarkJobFailed(
	ctx context.Context,
	jobID string,
	errMessage string,
	nextRunAt *time.Time,
) error {
	return s.store.MarkJobFailed(ctx, jobID, errMessage, nextRunAt)
}

func (s *Service) RecoverStaleJobs(ctx context.Context, timeout time.Duration) (int64, error) {
	return s.store.RecoverStaleJobs(ctx, timeout)
}

func (s *Service) StartAttempt(
	ctx context.Context,
	job Job,
	workerID string,
) (*JobAttempt, error) {
	return s.store.StartAttempt(ctx, job, workerID)
}

func (s *Service) FinishAttempt(
	ctx context.Context,
	attemptID string,
	status AttemptStatus,
	errMessage *string,
	startedAt time.Time,
) error {
	return s.store.FinishAttempt(
		ctx,
		attemptID,
		status,
		errMessage,
		startedAt,
	)
}
