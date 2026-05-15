package worker

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"
)

func SimulateWorker(ctx context.Context) error {
	select {
	case <-time.After(3 * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}

	n := rand.IntN(100) + 1

	switch {
	case n <= 30:
		return errors.New("hard failure")
	case n <= 35:
		return errors.New("transient failure")
	default:
		return nil
	}
}
