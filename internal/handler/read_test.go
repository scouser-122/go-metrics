package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/scouser-122/go-metrics/internal/config"
	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/repository"
	"github.com/scouser-122/go-metrics/internal/repository/db"
	"github.com/scouser-122/go-metrics/internal/service"
	"github.com/stretchr/testify/assert"
)

var listTests = []struct {
	name    string
	request request
	want    want
}{
	{
		name: "positive test list metrics",
		request: request{
			method: http.MethodGet,
			path:   "/",
		},
		want: want{
			code:        http.StatusOK,
			contentType: "text/html; charset=utf-8",
		},
	},
}

func TestListHandler(t *testing.T) {
	config := config.DefaultServerConfig()
	cryptoService := service.CryptoService{
		ServerConfig: &config,
	}
	metricsService := service.MetricsService{}
	metricsService.Initialize(&config, &db.PostgresDatabase{})
	handlers := InitializeHandlers(&metricsService, &cryptoService, nil)
	for _, test := range listTests {
		t.Run(test.name, func(t *testing.T) {
			r := CreateChiRouter(&handlers)

			request := httptest.NewRequest(test.request.method, test.request.path, nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			contentType := strings.Join(res.Header.Values("Content-Type"), "; ")
			assert.Equal(t, test.want.contentType, contentType)
			if res.StatusCode == http.StatusOK {
				bodyBytes, err := io.ReadAll(res.Body)
				assert.Nil(t, err)
				bodyString := string(bodyBytes)
				assert.True(t, len(bodyString) > 0)
			}
			res.Body.Close()
		})
	}
}

var valueTests = []struct {
	name    string
	request request
	want    want
}{
	{
		name: "positive test get counter value",
		request: request{
			method: http.MethodGet,
			path:   "/value/counter/PollCount",
		},
		want: want{
			code:        http.StatusOK,
			contentType: "text/plain",
			body:        "0",
		},
	},
	{
		name: "positive test get gauge value",
		request: request{
			method: http.MethodGet,
			path:   "/value/gauge/Alloc",
		},
		want: want{
			code:        http.StatusOK,
			contentType: "text/plain",
			body:        "0.00",
		},
	},
	{
		name: "negative test get metric value",
		request: request{
			method: http.MethodGet,
			path:   "/value/counter/SomeCounter",
		},
		want: want{
			code:        http.StatusNotFound,
			contentType: "text/plain",
		},
	},
}

func TestValueHandler(t *testing.T) {
	config := config.DefaultServerConfig()
	memStorage := repository.MemStorage{}
	memStorage.Metrics = append(memStorage.Metrics, models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: new(int64),
	})
	memStorage.Metrics = append(memStorage.Metrics, models.Metrics{
		ID:    "Alloc",
		MType: models.Gauge,
		Value: new(float64),
	})

	metricsService := service.MetricsService{
		Storage: &memStorage,
	}
	cryptoService := service.CryptoService{
		ServerConfig: &config,
	}
	handlers := InitializeHandlers(&metricsService, &cryptoService, nil)
	for _, test := range valueTests {
		t.Run(test.name, func(t *testing.T) {
			r := CreateChiRouter(&handlers)

			request := httptest.NewRequest(test.request.method, test.request.path, nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			contentType := strings.Join(res.Header.Values("Content-Type"), "; ")
			assert.Equal(t, test.want.contentType, contentType)
			if res.StatusCode == http.StatusOK {
				bodyBytes, err := io.ReadAll(res.Body)
				assert.Nil(t, err)
				bodyString := string(bodyBytes)
				assert.True(t, len(bodyString) > 0)
			}
			res.Body.Close()
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}

var valueJSONTests = []struct {
	name    string
	request request
	metrics []models.Metrics
	want    want
}{
	{
		name: "positive test value metric gauge",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/value",
			body:        `{"id":"Alloc","type":"gauge"}`,
		},
		metrics: []models.Metrics{
			{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: ptr(float64(123.456)),
			},
		},
		want: want{
			code:        http.StatusOK,
			contentType: "application/json",
			body:        `{"id":"Alloc","type":"gauge","value":123.456}`,
		},
	},
	{
		name: "positive test value metric counter",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/value",
			body:        `{"id":"Alloc","type":"counter"}`,
		},
		metrics: []models.Metrics{
			{
				ID:    "Alloc",
				MType: models.Counter,
				Delta: ptr(int64(10)),
			},
		},
		want: want{
			code:        http.StatusOK,
			contentType: "application/json",
			body:        `{"id":"Alloc","type":"counter","delta":10}`,
		},
	},
	{
		name: "negative test bad json",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/value",
			body:        `{"id":"Alloc","ty`,
		},
		want: want{
			code:        http.StatusBadRequest,
			contentType: "application/json",
		},
	},
	{
		name: "positive test value metric not found",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/value",
			body:        `{"id":"Alloc","type":"gauge"}`,
		},
		want: want{
			code:        http.StatusNotFound,
			contentType: "application/json",
		},
	},
	{
		name: "negative test value incorrect metric type",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/value",
			body:        `{"id":"Alloc","type":"histogram"}`,
		},
		want: want{
			code:        http.StatusBadRequest,
			contentType: "application/json",
		},
	},
}

func TestValueJSONHandler(t *testing.T) {
	for _, test := range valueJSONTests {
		t.Run(test.name, func(t *testing.T) {
			config := config.DefaultServerConfig()
			metricsService := service.MetricsService{}
			cryptoService := service.CryptoService{
				ServerConfig: &config,
			}
			metricsService.Initialize(&config, &db.PostgresDatabase{})
			metricsService.Storage.SaveMetrics(context.Background(), test.metrics)
			handlers := InitializeHandlers(&metricsService, &cryptoService, nil)

			r := CreateChiRouter(&handlers)

			var bodyReader io.Reader
			if test.request.body != "" {
				jsonData := []byte(test.request.body)
				bodyReader = bytes.NewBuffer(jsonData)
			}

			request := httptest.NewRequest(test.request.method, test.request.path, bodyReader)
			// создаём новый Recorder
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			contentType := strings.Join(res.Header.Values("Content-Type"), "; ")
			assert.Equal(t, test.want.contentType, contentType)
			if res.StatusCode == http.StatusOK {
				bodyBytes, err := io.ReadAll(res.Body)
				assert.Nil(t, err)
				bodyString := string(bodyBytes)
				assert.True(t, len(bodyString) > 0)
			}
			res.Body.Close()
		})
	}
}
