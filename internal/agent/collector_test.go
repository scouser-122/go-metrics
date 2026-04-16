package agent

import (
	"testing"

	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestCollectMetrics(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "collect runtime metrics",
		},
	}
	config := GetDefaultAgentConfig()
	collector := NewCollector(&config)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			dataChannel := make(chan CollectedData, 1)

			collector.CollectRuntimeMetrics(dataChannel)

			data := <-dataChannel

			for _, m := range data.metrics {
				if m.MType == models.Gauge {
					assert.True(t, *m.Value >= 0.0)
				}
				if m.ID == "PollCount" {
					assert.Equal(t, int64(1), *m.Delta)
				}
				if m.ID == "RandomValue" {
					assert.True(t, *m.Value >= 0.0)
					assert.True(t, *m.Value <= 1.0)
				}
			}

			close(dataChannel)
		})
	}
}
