package jobs

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type JobStatus string

const (
	StatusQueued     JobStatus = "queued"
	StatusProcessing JobStatus = "processing"
	StatusSucceeded  JobStatus = "succeeded"
	StatusFailed     JobStatus = "failed"
	StatusCanceled   JobStatus = "canceled"
)

type Job struct {
	ID          string         `gorm:"type:uuid;primaryKey"`
	Type        string         `gorm:"type:varchar(100);not null"`
	Payload     datatypes.JSON `gorm:"type:jsonb;not null"`
	Status      JobStatus      `gorm:"type:varchar(30);not null"`
	Priority    int            `gorm:"not null"`
	Attempts    int            `gorm:"not null"`
	MaxAttempts int            `gorm:"not null"`
	RunAt       time.Time      `gorm:"not null"`
	LockedAt    *time.Time
	WorkerID    *string
	LastError   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (j *Job) BeforeCreate(tx *gorm.DB) error {
	if j.ID == "" {
		j.ID = uuid.NewString()
	}

	if j.Status == "" {
		j.Status = StatusQueued
	}

	if j.RunAt.IsZero() {
		j.RunAt = time.Now()
	}

	if j.MaxAttempts == 0 {
		j.MaxAttempts = 3
	}

	return nil
}
