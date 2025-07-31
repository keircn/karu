package scraper

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/keircn/karu/internal/config"
)

const (
	DefaultTimeout = 10 * time.Second
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

func executeWithFallback(fn func(string) error) error {
	cfg, _ := config.Load()
	timeout := time.Duration(cfg.RequestTimeout) * time.Second
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	var lastErr error

	for _, source := range sources {
		ctx := context.Background()
		err := withTimeout(ctx, timeout, func() error {
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
