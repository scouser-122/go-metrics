package models

import (
	"errors"
	"fmt"
	"net"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// IncorrectMetricType represents an error for an invalid metric type.
type IncorrectMetricType struct {
	Message string
}

// Error implements the error interface.
func (e IncorrectMetricType) Error() string {
	return e.Message
}

// MetricFormatError represents an error for an invalid metric format.
type MetricFormatError struct {
	Message string
	Err     error
}

// Error implements the error interface.
func (e MetricFormatError) Error() string {
	return fmt.Sprintf("%s %v", e.Message, e.Err)
}

// MetricSaveError represents an error that occurred while saving a metric.
type MetricSaveError struct {
	Message string
	Err     error
}

// Error implements the error interface.
func (e MetricSaveError) Error() string {
	return fmt.Sprintf("%s %v", e.Message, e.Err)
}

// MetricGetError represents an error that occurred while retrieving a metric.
type MetricGetError struct {
	Message string
	Err     error
}

// Error implements the error interface.
func (e MetricGetError) Error() string {
	return fmt.Sprintf("%s %v", e.Message, e.Err)
}

var (
	ErrIncorrectType   = IncorrectMetricType{}
	ErrIncorrectFormat = MetricFormatError{}
	ErrSaveMetric      = MetricSaveError{}
	ErrGetMetric       = MetricGetError{}
)

// ErrorClassification categorizes errors as retryable or non-retryable.
type ErrorClassification int

const (
	// NonRetryable indicates an error that should not be retried.
	NonRetryable ErrorClassification = iota
	// Retryable indicates an error that can be retried.
	Retryable
)

// ClassifyAgentError classifies agent errors to determine if they should be retried.
func ClassifyAgentError(err error) ErrorClassification {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return Retryable
	}
	return NonRetryable
}

// ClassifyPostgreSQLError classifies PostgreSQL errors to determine if they should be retried.
func ClassifyPostgreSQLError(err error) ErrorClassification {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure:
			return Retryable
		}
	}
	var pgConnectErr *pgconn.ConnectError
	if errors.As(err, &pgConnectErr) {
		return Retryable
	}
	return NonRetryable
}
