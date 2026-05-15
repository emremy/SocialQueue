package jobs

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttemptStatus string

const (
	AttemptRunning   AttemptStatus = "running"
	AttemptSucceeded AttemptStatus = "succeeded"
	AttemptFailed    AttemptStatus = "failed"
)

type JobAttempt struct {
	ID            string        `gorm:"type:uuid;primaryKey"`
	JobID         string        `gorm:"type:uuid;not null"`
	AttemptNumber int           `gorm:"not null"`
	WorkerID      string        `gorm:"type:varchar(100);not null"`
	Status        AttemptStatus `gorm:"type:varchar(30);not null"`
	Error         *string       `gorm:"type:text"`
	StartedAt     time.Time     `gorm:"not null"`
	FinishedAt    *time.Time
	DurationMs    *int
	CreatedAt     time.Time
}

func (a *JobAttempt) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}

	if a.Status == "" {
		a.Status = AttemptRunning
	}

	if a.StartedAt.IsZero() {
		a.StartedAt = time.Now()
	}

	return nil
}
