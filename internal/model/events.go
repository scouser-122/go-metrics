package models

import "github.com/botchris/go-pubsub"

type MetricEvent interface{}

// MetricsReceivedEvent событие получения метрик
type MetricsReceivedEvent struct {
	MetricEvent `json:"-"`

	// TS unix timestamp события
	TS int64 `json:"ts"`

	// Metrics наименование полученных метрик
	Metrics []string `json:"metrics"`

	// IPAddress IP адрес входящего запроса
	IPAddress string `json:"ip_address"`
}

const MetricEventTopic pubsub.Topic = "metricEvents"
