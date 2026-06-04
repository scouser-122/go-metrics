package agent

import models "github.com/scouser-122/go-metrics/internal/model"

type CollectedData struct {
	metrics []models.Metrics
}

type WriteRequest struct {
	data models.Metrics
}

type ReadRequest struct {
	resp chan []models.Metrics
}
