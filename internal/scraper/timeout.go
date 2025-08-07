package scraper

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/keircn/karu/internal/config"
)

const (
	DefaultTimeout    = 10 * time.Second
	MaxRetryAttempts  = 3
	BaseRetryDelay    = 1 * time.Second
	MaxRetryDelay     = 30 * time.Second
	BackoffMultiplier = 2.0
)

type Source struct {
	Name string
	URL  string
}

var sources = []Source{
	{"primary", "https://api.allanime.day/api"},
	{"fallback1", "https://api.allanime.day/api"},
	{"fallback2", "https://api.allanime.day/api"},
}

type TimeoutError struct {
	Source  string
	Timeout time.Duration
}

func (e TimeoutError) Error() string {
	return fmt.Sprintf("timeout after %v for source %s", e.Timeout, e.Source)
}

type RetryError struct {
	Attempts int
	LastErr  error
}

func (e RetryError) Error() string {
	return fmt.Sprintf("operation failed after %d attempts: %v", e.Attempts, e.LastErr)
}

func withTimeout(ctx context.Context, timeout time.Duration, fn func() error) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func executeWithRetry(ctx context.Context, timeout time.Duration, fn func() error) error {
	var lastErr error

	for attempt := 1; attempt <= MaxRetryAttempts; attempt++ {
		err := withTimeout(ctx, timeout, fn)

		if err == nil {
			return nil
		}

		lastErr = err

		if attempt == MaxRetryAttempts {
			break
		}

		if shouldRetry(err) {
			delay := calculateBackoffDelay(attempt)
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		} else {
			break
		}
	}

	return RetryError{Attempts: MaxRetryAttempts, LastErr: lastErr}
}

func shouldRetry(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	errStr := err.Error()
	retryableErrors := []string{
		"connection reset",
		"connection refused",
		"temporary failure",
		"timeout",
		"network is unreachable",
		"no route to host",
		"i/o timeout",
	}

	for _, retryable := range retryableErrors {
		if contains(errStr, retryable) {
			return true
		}
	}

	return false
}

func calculateBackoffDelay(attempt int) time.Duration {
	delay := time.Duration(float64(BaseRetryDelay) * math.Pow(BackoffMultiplier, float64(attempt-1)))
	if delay > MaxRetryDelay {
		delay = MaxRetryDelay
	}
	return delay
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func executeWithFallback(fn func(string) error) error {
	cfg, _ := config.Load()
	timeout := time.Duration(cfg.RequestTimeout) * time.Second
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	ctx := context.Background()
	var lastErr error

	for _, source := range sources {
		err := executeWithRetry(ctx, timeout, func() error {
			return fn(source.URL)
		})

		if err == nil {
			return nil
		}

		if errors.Is(err, context.DeadlineExceeded) {
			lastErr = TimeoutError{Source: source.Name, Timeout: timeout}
		} else {
			lastErr = err
		}
	}

	return lastErr
}
