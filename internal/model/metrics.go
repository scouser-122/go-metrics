package models

import (
	"encoding/json"
	"errors"
	"strconv"
)

//go:generate go run github.com/scouser-122/go-metrics/cmd/reset

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.

// Metrics represents a metric with an ID, type, and value.
// Delta and Value are pointers to distinguish between zero values and unset values.
// swagger:model
// generate:reset
type Metrics struct {
	// The unique identifier for this metric.
	ID string `json:"id" binding:"required"`

	// Metric type.
	MType string `json:"type" enums:"counter,gauge" binding:"required"`

	// Metric delta in case it has counter type.
	Delta *int64 `json:"delta,omitempty"`

	// Metric value in case it has counter gauge.
	Value *float64 `json:"value,omitempty"`

	// Internal data (not exposed in API)
	Hash string `json:"hash,omitempty" swaggerignore:"true"`
}

// GetValueAsString returns the metric value as a string based on its type.
func (m *Metrics) GetValueAsString() (string, error) {
	switch m.MType {
	case Counter:
		if m.Delta == nil {
			return "", errors.New("absent value for counter")
		}
		return strconv.FormatInt(*m.Delta, 10), nil
	case Gauge:
		if m.Value == nil {
			return "", errors.New("absent value for gauge")
		}
		return strconv.FormatFloat(*m.Value, 'f', -1, 64), nil
	}
	return "", errors.New("metric type incorrect")
}

// String returns the metric as a JSON string.
func (m Metrics) String() string {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return err.Error()
	}
	return string(jsonData)
}
