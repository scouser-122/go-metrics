package models

import "github.com/botchris/go-pubsub"

type MetricEvent interface{}

// MetricsReceivedEvent событие получения метрик
type MetricsReceivedEvent struct {
	MetricEvent `json:"-"`

	// Ts unix timestamp события
	Ts int64 `json:"ts"`

	// Metrics наименование полученных метрик
	Metrics []string `json:"metrics"`

	// IpAddress IP адрес входящего запроса
	IpAddress string `json:"ip_address"`
}

const MetricEventTopic pubsub.Topic = "metricEvents"
