package models

import (
	"errors"
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
}

func (e MetricFormatError) Error() string {
	return e.Message
}

type MetricSaveError struct {
	Message string
}

func (e MetricSaveError) Error() string {
	return e.Message
}

type MetricGetError struct {
	Message string
}

func (e MetricGetError) Error() string {
	return e.Message
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
