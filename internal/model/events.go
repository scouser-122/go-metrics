package models

import "github.com/botchris/go-pubsub"

//go:generate go run github.com/scouser-122/go-metrics/cmd/reset

// MetricsReceivedEvent metric received event.
// generate:reset
type MetricsReceivedEvent struct {
	// TS event's unix timestamp.
	TS int64 `json:"ts"`

	// Metrics names of received metrics.
	Metrics []string `json:"metrics"`

	// IPAddress incoming request IP address.
	IPAddress string `json:"ip_address"`
}

const MetricEventTopic pubsub.Topic = "metricEvents"
