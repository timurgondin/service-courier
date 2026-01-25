package retry_test

import (
	"errors"
	"testing"
	"time"

	"service-courier/internal/pkg/retry"
)

type zeroStrategy struct{}

func (zeroStrategy) NextDelay(int) time.Duration { return 0 }

func TestExecuteWithCallback_RetriesUntilSuccess(t *testing.T) {
	var attempts int
	var onRetryCalls int

	exec := retry.NewRetryExecutor(retry.RetryConfig{
		MaxAttempts: 3,
		Strategy:    zeroStrategy{},
		ShouldRetry: func(err error) bool { return err != nil },
	})

	err := exec.ExecuteWithCallback(
		func() error {
			attempts++
			if attempts < 2 {
				return errors.New("temporary")
			}
			return nil
		},
		func(attempt int, err error, delay time.Duration) {
			onRetryCalls++
		},
	)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	if onRetryCalls != 1 {
		t.Fatalf("expected 1 onRetry call, got %d", onRetryCalls)
	}
}

func TestExecute_NoRetryWhenShouldRetryFalse(t *testing.T) {
	sentinel := errors.New("permanent")
	var attempts int

	exec := retry.NewRetryExecutor(retry.RetryConfig{
		MaxAttempts: 3,
		Strategy:    zeroStrategy{},
		ShouldRetry: func(err error) bool { return false },
	})

	err := exec.Execute(func() error {
		attempts++
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}
