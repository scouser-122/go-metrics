package config

import (
	"context"
	"fmt"
	"time"

	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
)

type RetryConfig struct {
	MaxAttempts       int
	InitialBackoff    int
	BackoffMultiplier int
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    1000,
		BackoffMultiplier: 2000,
	}
}

func AgentRetry(config RetryConfig, operation func() error) error {
	var lastErr error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		if models.ClassifyAgentError(err) == models.NonRetryable {
			return lastErr
		}

		backoff := calculateBackoff(config, attempt)

		logger.Sugar.Infof("received error %q on attempt %d, will retry after %q", lastErr, attempt, backoff)
		time.Sleep(backoff)
	}

	logger.Sugar.Errorf("max retries (%d) exceeded, error: %q", config.MaxAttempts, lastErr)
	return lastErr
}

func DataBaseRequestRetry(ctx context.Context, config RetryConfig, operation func() error) error {
	var lastErr error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		if models.ClassifyPostgreSQLError(err) == models.NonRetryable {
			return lastErr
		}

		backoff := calculateBackoff(config, attempt)

		logger.Sugar.Infof("received error %s on attempt %d, will retry after %q", lastErr, attempt, backoff)
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
		case <-time.After(backoff):
		}
	}

	logger.Sugar.Errorf("max retries (%d) exceeded, error: %q", config.MaxAttempts, lastErr)
	return lastErr
}

func calculateBackoff(config RetryConfig, attempt int) time.Duration {
	backoff := float64(config.InitialBackoff + (attempt-1)*config.BackoffMultiplier)
	return time.Duration(backoff * float64(time.Millisecond))
}
