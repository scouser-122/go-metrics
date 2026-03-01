package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
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

func Ptr[T any](v T) *T {
	return &v
}

func TestSendMetric(t *testing.T) {
	type want struct {
		result    string
		errNotNil bool
	}
	type response struct {
		status int
		body   string
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
				body:   "100.20",
				err:    nil,
			},
			want: want{
				result:    "100.20",
				errNotNil: false,
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
				result:    "",
				errNotNil: true,
			},
		},
	}
	agent := RuntimeMetricsAgent{
		Config: GetDefaultAgentConfig(),
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				headers := rw.Header()
				headers.Add("Content-Type", "text/plain")
				rw.WriteHeader(test.response.status)
				if test.response.body != "" {
					rw.Write([]byte(test.response.body))
				}
			}))
			defer server.Close()

			client := resty.New()
			agent.Config.serverAddress = server.URL

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
