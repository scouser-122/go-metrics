package models

import "github.com/botchris/go-pubsub"

// MetricsReceivedEvent metric received event.
type MetricsReceivedEvent struct {
	// TS event's unix timestamp.
	TS int64 `json:"ts"`

	// Metrics names of received metrics.
	Metrics []string `json:"metrics"`

	// IPAddress incoming request IP address.
	IPAddress string `json:"ip_address"`
}

const MetricEventTopic pubsub.Topic = "metricEvents"
