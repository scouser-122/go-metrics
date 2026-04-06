package models

import (
	"errors"
	"fmt"
	"net"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type IncorrectMetricType struct {
	Message string
}

func (e IncorrectMetricType) Error() string {
	return e.Message
}

type MetricFormatError struct {
	Message string
	Err     error
}

func (e MetricFormatError) Error() string {
	return fmt.Sprintf("%s %v", e.Message, e.Err)
}

type MetricSaveError struct {
	Message string
	Err     error
}

func (e MetricSaveError) Error() string {
	return fmt.Sprintf("%s %v", e.Message, e.Err)
}

type MetricGetError struct {
	Message string
	Err     error
}

func (e MetricGetError) Error() string {
	return fmt.Sprintf("%s %v", e.Message, e.Err)
}

var (
	ErrIncorrectType   = IncorrectMetricType{}
	ErrIncorrectFormat = MetricFormatError{}
	ErrSaveMetric      = MetricSaveError{}
	ErrGetMetric       = MetricGetError{}
)

type ErrorClassification int

const (
	NonRetryable ErrorClassification = iota
	Retryable
)

func ClassifyAgentError(err error) ErrorClassification {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return Retryable
	}
	return NonRetryable
}

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
