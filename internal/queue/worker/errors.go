package worker

import "errors"

var (
	ErrUnknownJobType  = errors.New("unknown job type")
	ErrHandlerNotFound = errors.New("handler not found")
)
