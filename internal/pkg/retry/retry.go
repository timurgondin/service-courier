package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

var ErrMaxAttemptsExceeded = errors.New("max retry attempts exceeded")

type Strategy interface {
	NextDelay(attempt int) time.Duration
}

type RetryConfig struct {
	MaxAttempts int
	Strategy    Strategy
	ShouldRetry func(error) bool
}

type RetryExecutor struct {
	config RetryConfig
}

func NewRetryExecutor(config RetryConfig) *RetryExecutor {
	if config.MaxAttempts == 0 {
		config.MaxAttempts = 3
	}
	if config.Strategy == nil {
		config.Strategy = NewExponentialBackoff(100*time.Millisecond, 5*time.Second, 2.0)
	}
	if config.ShouldRetry == nil {
		config.ShouldRetry = func(err error) bool { return err != nil }
	}
	return &RetryExecutor{config: config}
}

func (r *RetryExecutor) Execute(fn func() error) error {
	var lastErr error

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if !r.config.ShouldRetry(err) {
			return err
		}

		if attempt == r.config.MaxAttempts {
			break
		}

		delay := r.config.Strategy.NextDelay(attempt)
		time.Sleep(delay)
	}

	return fmt.Errorf("%w: %v", ErrMaxAttemptsExceeded, lastErr)
}

func (r *RetryExecutor) ExecuteWithContext(ctx context.Context, fn func(context.Context) error) error {
	var lastErr error

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := fn(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		if !r.config.ShouldRetry(err) {
			return err
		}

		if attempt == r.config.MaxAttempts {
			break
		}

		delay := r.config.Strategy.NextDelay(attempt)
		if !sleepWithContext(ctx, delay) {
			return ctx.Err()
		}
	}

	return fmt.Errorf("%w: %v", ErrMaxAttemptsExceeded, lastErr)
}

func (r *RetryExecutor) ExecuteWithCallback(
	fn func() error,
	onRetry func(attempt int, err error, delay time.Duration),
) error {
	var lastErr error

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if !r.config.ShouldRetry(err) {
			return err
		}

		if attempt == r.config.MaxAttempts {
			break
		}

		delay := r.config.Strategy.NextDelay(attempt)
		if onRetry != nil {
			onRetry(attempt, err, delay)
		}
		time.Sleep(delay)
	}

	return fmt.Errorf("%w: %v", ErrMaxAttemptsExceeded, lastErr)
}

type ExponentialBackoff struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

func NewExponentialBackoff(initial, max time.Duration, multiplier float64) *ExponentialBackoff {
	return &ExponentialBackoff{
		InitialDelay: initial,
		MaxDelay:     max,
		Multiplier:   multiplier,
	}
}

func (e *ExponentialBackoff) NextDelay(attempt int) time.Duration {
	delay := float64(e.InitialDelay) * math.Pow(e.Multiplier, float64(attempt-1))
	if delay > float64(e.MaxDelay) {
		delay = float64(e.MaxDelay)
	}
	return time.Duration(delay)
}

func sleepWithContext(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
