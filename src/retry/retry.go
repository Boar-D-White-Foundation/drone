package retry

import (
	"context"
	"time"
)

type break_ struct {
	error
}

func Break(err error) error {
	return break_{err}
}

type Backoff interface {
	GetDelay(attempt int) (time.Duration, bool)
}

type LinearBackoff struct {
	Delay       time.Duration
	MaxAttempts int
}

func (b LinearBackoff) GetDelay(attempt int) (time.Duration, bool) {
	if attempt >= b.MaxAttempts {
		return 0, false
	}
	return b.Delay, true
}

func Do[T any](ctx context.Context, backoff Backoff, f func() (T, error)) (T, error) {
	for attempt := 0; ; attempt++ {
		res, err := f()
		if err == nil {
			return res, nil
		}
		if err, ok := err.(break_); ok {
			return res, err.error
		}

		delay, ok := backoff.GetDelay(attempt)
		if !ok {
			return res, err
		}

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return res, ctx.Err()
		}
	}
}
