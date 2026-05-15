package worker

import "time"

type RetryPolicy struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		BaseDelay: 5 * time.Second,
		MaxDelay:  5 * time.Minute,
	}
}

func (p RetryPolicy) NextRunAt(attempts int) time.Time {
	delay := p.Delay(attempts)
	return time.Now().Add(delay)
}

func (p RetryPolicy) Delay(attempts int) time.Duration {
	if attempts < 0 {
		return p.BaseDelay
	}

	delay := p.BaseDelay * time.Duration(1<<(attempts-1))

	if delay > p.MaxDelay {
		return p.MaxDelay
	}

	return delay
}
