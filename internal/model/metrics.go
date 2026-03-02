package models

import (
	"errors"
	"strconv"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func (m *Metrics) GetValueAsString() (string, error) {
	if m.MType == Counter {
		if m.Delta == nil {
			return "", errors.New("absent value for counter")
		}
		return strconv.FormatInt(*m.Delta, 10), nil
	} else if m.MType == Gauge {
		if m.Value == nil {
			return "", errors.New("absent value for gauge")
		}
		return strconv.FormatFloat(*m.Value, 'f', -1, 64), nil
	}
	return "", errors.New("metric type incorrect")
}
