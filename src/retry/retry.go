package retry

import (
	"context"
	"log/slog"
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

func Do[T any](ctx context.Context, name string, backoff Backoff, f func() (T, error)) (T, error) {
	slog.Info("started retry", slog.String("name", name))
	for attempt := 0; ; attempt++ {
		res, err := f()
		if err == nil {
			slog.Info("completed retry", slog.String("name", name), slog.Int("attempt", attempt))
			return res, nil
		}
		if err, ok := err.(break_); ok {
			slog.Info("break retry", slog.String("name", name), slog.Int("attempt", attempt))
			return res, err.error
		}

		slog.Info(
			"retry got err",
			slog.String("name", name),
			slog.Int("attempt", attempt),
			slog.Any("err", err),
		)
		delay, ok := backoff.GetDelay(attempt)
		if !ok {
			slog.Error("stopped retry", slog.String("name", name), slog.Int("attempt", attempt))
			return res, err
		}

		slog.Info(
			"retry sleep",
			slog.String("name", name),
			slog.Int("attempt", attempt),
			slog.String("delay", delay.String()),
		)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return res, ctx.Err()
		}
	}
}
