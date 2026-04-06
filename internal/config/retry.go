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

	for attempt := 0; attempt <= config.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		if models.ClassifyAgentError(err) == models.NonRetryable {
			return lastErr
		}

		if attempt != config.MaxAttempts {
			backoff := calculateBackoff(config, attempt)
			logger.Sugar.Infof("received error on attempt %d: %s, will retry after %q", attempt, lastErr, backoff)
			time.Sleep(backoff)
		} else {
			logger.Sugar.Infof("received error on attempt %d: %s", attempt, lastErr)
		}
	}

	logger.Sugar.Infof("max retries (%d) exceeded", config.MaxAttempts)
	return lastErr
}

func DataBaseRequestRetry(ctx context.Context, config RetryConfig, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		if models.ClassifyPostgreSQLError(err) == models.NonRetryable {
			return lastErr
		}

		if attempt != config.MaxAttempts {
			backoff := calculateBackoff(config, attempt)
			logger.Sugar.Infof("received error on attempt %d: %s, will retry after %q", attempt, lastErr, backoff)
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
			}
		} else {
			logger.Sugar.Infof("received error on attempt %d: %s", attempt, lastErr)
		}
	}

	logger.Sugar.Infof("max retries (%d) exceeded", config.MaxAttempts)
	return lastErr
}

func calculateBackoff(config RetryConfig, attempt int) time.Duration {
	backoff := float64(config.InitialBackoff + attempt*config.BackoffMultiplier)
	return time.Duration(backoff * float64(time.Millisecond))
}
