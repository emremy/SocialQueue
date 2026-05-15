package worker

import (
	"context"

	"github.com/emremy/socialqueue/internal/queue/jobs"
)

type Executor struct {
	registry *HandlerRegistry
}

func NewExecutor() *Executor {
	registry := NewHandlerRegistry()

	executor := &Executor{
		registry: registry,
	}
	executor.registerDefaults()
	return executor
}

func (e *Executor) Execute(ctx context.Context, job jobs.Job) error {
	handler, ok := e.registry.Get(job.Type)
	if !ok {
		return ErrHandlerNotFound
	}

	return handler(ctx, job)
}

func (e *Executor) Register(jobType string, handler JobHandler) {
	e.registry.Register(jobType, handler)
}

func (e *Executor) registerDefaults() {
	e.Register("simulate", func(ctx context.Context, job jobs.Job) error {
		return SimulateWorker(ctx)
	})
}
