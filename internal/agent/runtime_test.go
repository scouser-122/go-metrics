package agent

import (
	"errors"
	"net/http"
	"testing"

	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestFillMetricsModel(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "fill metrics map",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			agent := RuntimeMetricsAgent{
				Config: AgentConfig{
					runtimeMetricNames: []string{
						"Alloc",
						"BuckHashSys",
						"Frees",
						"TotalAlloc",
					},
				},
			}
			agent.FillMetricsModel()

			assert.Equal(t, 6, len(agent.runtimeMetrics.Metrics))
			assert.Equal(
				t,
				[]string{
					"Alloc",
					"BuckHashSys",
					"Frees",
					"PollCount",
					"RandomValue",
					"TotalAlloc",
				},
				agent.runtimeMetrics.SortedKeys,
			)
			assert.NotNil(t, agent.runtimeMetrics.Metrics["Alloc"])
			assert.Equal(t, float64(0.0), *agent.runtimeMetrics.Metrics["Alloc"].Value)
			assert.NotNil(t, agent.runtimeMetrics.Metrics["BuckHashSys"])
			assert.Equal(t, float64(0.0), *agent.runtimeMetrics.Metrics["BuckHashSys"].Value)
			assert.NotNil(t, agent.runtimeMetrics.Metrics["Frees"])
			assert.Equal(t, float64(0.0), *agent.runtimeMetrics.Metrics["Frees"].Value)
			assert.NotNil(t, agent.runtimeMetrics.Metrics["TotalAlloc"])
			assert.Equal(t, float64(0.0), *agent.runtimeMetrics.Metrics["TotalAlloc"].Value)
			assert.NotNil(t, agent.runtimeMetrics.Metrics["PollCount"])
			assert.Equal(t, int64(0.0), *agent.runtimeMetrics.Metrics["PollCount"].Delta)
			assert.NotNil(t, agent.runtimeMetrics.Metrics["RandomValue"])
			assert.Equal(t, float64(0.0), *agent.runtimeMetrics.Metrics["RandomValue"].Value)
		})
	}
}

func TestCollectMetrics(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "collect runtime metrics",
		},
	}
	agent := RuntimeMetricsAgent{
		Config: GetDefaultAgentConfig(),
	}
	agent.FillMetricsModel()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			agent.CollectMetrics()
			for _, metricName := range agent.Config.runtimeMetricNames {
				assert.True(t, *agent.runtimeMetrics.Metrics[metricName].Value >= 0.0)
			}
			assert.Equal(t, int64(1), *agent.runtimeMetrics.Metrics["PollCount"].Delta)
			assert.True(t, *agent.runtimeMetrics.Metrics["RandomValue"].Value >= 0.0)
			assert.True(t, *agent.runtimeMetrics.Metrics["RandomValue"].Value <= 1.0)
		})
	}
}

// RoundTripFunc is a type that implements http.RoundTripper for testing
type RoundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the http.RoundTripper interface
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// Function to get a mock client
func NewMockClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func Ptr[T any](v T) *T {
	return &v
}

func TestSendMetric(t *testing.T) {
	type want struct {
		result    bool
		errNotNil bool
	}
	type response struct {
		status int
		err    error
	}
	tests := []struct {
		name     string
		metric   models.Metrics
		response response
		want     want
	}{
		{
			name: "send runtime metric",
			metric: models.Metrics{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: Ptr(100.20),
			},
			response: response{
				status: http.StatusOK,
				err:    nil,
			},
			want: want{
				result:    true,
				errNotNil: false,
			},
		},
		{
			name: "send runtime metric url error",
			metric: models.Metrics{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: Ptr(100.20),
			},
			response: response{
				status: http.StatusInternalServerError,
				err:    errors.New("Internal server error"),
			},
			want: want{
				result:    false,
				errNotNil: true,
			},
		},
		{
			name: "send runtime metric incorrect status",
			metric: models.Metrics{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: Ptr(100.20),
			},
			response: response{
				status: http.StatusNotFound,
				err:    nil,
			},
			want: want{
				result:    false,
				errNotNil: true,
			},
		},
	}
	agent := RuntimeMetricsAgent{
		Config: GetDefaultAgentConfig(),
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Mock the RoundTrip function to return a specific response
			mockRoundTripper := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
				header := make(http.Header)
				header.Set("Content-Type", "text/plain")
				return &http.Response{
					StatusCode: test.response.status,
					Body:       nil,
					Header:     header,
				}, test.response.err
			})

			// Create a client with the mock transport
			client := NewMockClient(mockRoundTripper)

			result, err := agent.SendMetric(client, &test.metric)
			assert.Equal(t, test.want.result, result)
			if test.want.errNotNil {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
