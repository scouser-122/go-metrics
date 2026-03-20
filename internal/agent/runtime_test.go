package agent

import (
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
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
	agent := RuntimeMetricsAgent{
		Config: GetDefaultAgentConfig(),
	}
	agent.Init()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			agent.CollectMetrics()

			respChan := make(chan []models.Metrics)
			agent.reads <- ReadRequest{resp: respChan}
			metrics := <-respChan

			for _, m := range metrics {
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
			agent.Config.ServerAddress = server.URL

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

func TestSendMetricJSON(t *testing.T) {
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
				body:   `{"id":"Alloc","type":"gauge","value":100.20}`,
				err:    nil,
			},
			want: want{
				result:    "100.2",
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
				headers.Add("Content-Type", "application/json")
				headers.Set("Content-Encoding", "gzip")

				rw.WriteHeader(test.response.status)
				if test.response.body != "" {
					gz := gzip.NewWriter(rw)
					defer gz.Close()
					gz.Write([]byte(test.response.body))
				}
			}))
			defer server.Close()

			client := resty.New()
			agent.Config.ServerAddress = server.URL

			result, err := agent.SendMetricJSON(client, &test.metric)
			assert.Equal(t, test.want.result, result)
			if test.want.errNotNil {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
