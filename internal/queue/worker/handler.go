package worker

import (
	"context"

	"github.com/emremy/socialqueue/internal/queue/jobs"
)

type JobHandler func(ctx context.Context, job jobs.Job) error

type HandlerRegistry struct {
	handlers map[string]JobHandler
}

func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[string]JobHandler),
	}
}

func (r *HandlerRegistry) Register(jobType string, handler JobHandler) {
	r.handlers[jobType] = handler
}

func (r *HandlerRegistry) Get(jobType string) (JobHandler, bool) {
	handler, ok := r.handlers[jobType]
	return handler, ok
}
