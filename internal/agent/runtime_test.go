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
			var testRuntimeMetricNames = []string{
				"Alloc",
				"BuckHashSys",
				"Frees",
				"TotalAlloc",
			}
			var runtimeMetrics = RuntimeMetircs{
				Metrics: make(map[string]*models.Metrics),
			}
			FillMetricsModel(&runtimeMetrics, testRuntimeMetricNames)

			assert.Equal(t, 6, len(runtimeMetrics.Metrics))
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
				runtimeMetrics.SortedKeys,
			)
			assert.NotNil(t, runtimeMetrics.Metrics["Alloc"])
			assert.Equal(t, float64(0.0), *runtimeMetrics.Metrics["Alloc"].Value)
			assert.NotNil(t, runtimeMetrics.Metrics["BuckHashSys"])
			assert.Equal(t, float64(0.0), *runtimeMetrics.Metrics["BuckHashSys"].Value)
			assert.NotNil(t, runtimeMetrics.Metrics["Frees"])
			assert.Equal(t, float64(0.0), *runtimeMetrics.Metrics["Frees"].Value)
			assert.NotNil(t, runtimeMetrics.Metrics["TotalAlloc"])
			assert.Equal(t, float64(0.0), *runtimeMetrics.Metrics["TotalAlloc"].Value)
			assert.NotNil(t, runtimeMetrics.Metrics["PollCount"])
			assert.Equal(t, int64(0.0), *runtimeMetrics.Metrics["PollCount"].Delta)
			assert.NotNil(t, runtimeMetrics.Metrics["RandomValue"])
			assert.Equal(t, float64(0.0), *runtimeMetrics.Metrics["RandomValue"].Value)
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
	var testRuntimeMetricNames = []string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
	}
	var runtimeMetrics = RuntimeMetircs{
		Metrics: make(map[string]*models.Metrics),
	}
	FillMetricsModel(&runtimeMetrics, testRuntimeMetricNames)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			CollectMetrics(&runtimeMetrics)
			for _, metricName := range testRuntimeMetricNames {
				assert.True(t, *runtimeMetrics.Metrics[metricName].Value >= 0.0)
			}
			assert.Equal(t, int64(1), *runtimeMetrics.Metrics["PollCount"].Delta)
			assert.True(t, *runtimeMetrics.Metrics["RandomValue"].Value >= 0.0)
			assert.True(t, *runtimeMetrics.Metrics["RandomValue"].Value <= 1.0)
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

			result, err := SendMetric(client, &test.metric)
			assert.Equal(t, test.want.result, result)
			if test.want.errNotNil {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
