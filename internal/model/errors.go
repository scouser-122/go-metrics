package models

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

var (
	ErrIncorrectType   = IncorrectMetricType{}
	ErrIncorrectFormat = MetricFormatError{}
	ErrSaveMetric      = MetricSaveError{}
)
