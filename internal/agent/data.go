package agent

import models "github.com/scouser-122/go-metrics/internal/model"

// CollectedData represents a batch of collected metrics.
type CollectedData struct {
	metrics []models.Metrics
}

// WriteRequest represents a request to write a metric to internal storage.
type WriteRequest struct {
	data models.Metrics
}

// ReadRequest represents a request to read all stored metrics.
type ReadRequest struct {
	resp chan []models.Metrics
}
