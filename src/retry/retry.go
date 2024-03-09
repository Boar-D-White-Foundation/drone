package retry

import (
	"context"
	"errors"
	"time"
)

type break_ struct {
	error
}

func (break_) Is(err error) bool {
	_, ok := err.(break_)
	return ok
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
		if errors.Is(err, break_{}) {
			return res, err
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
