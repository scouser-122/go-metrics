package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/botchris/go-pubsub/provider/memory"
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
	serverConfig := config.DefaultServerConfig()
	cryptoService := service.CryptoService{
		ServerConfig: &serverConfig,
	}
	metricsService := service.NewMetricsService(&serverConfig, &db.PostgresDatabase{}, nil)
	handlers := InitializeHandlers(metricsService, &cryptoService, nil)
	for _, test := range listTests {
		t.Run(test.name, func(t *testing.T) {
			r := CreateChiRouterWithHandlers(&handlers, &serverConfig)

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

func ExampleReadHandler_ListHandler() {
	// prepare handler
	serverConfig := config.DefaultServerConfig()
	cryptoService := service.CryptoService{
		ServerConfig: &serverConfig,
	}

	eventBroker := memory.NewBroker()
	brokerContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	auditService := service.NewAuditService(&serverConfig)
	auditService.SubscribeToMetricEvents(eventBroker, brokerContext)

	metricsService := service.NewMetricsService(&serverConfig, &db.PostgresDatabase{}, eventBroker)
	metricsService.SaveMetricsModel(context.Background(), generateTestMetrics(10))

	handlers := InitializeHandlers(metricsService, &cryptoService, nil)
	r := CreateChiRouterWithHandlers(&handlers, &serverConfig)

	// call handler
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, request)
	res := w.Result()

	fmt.Println(res.StatusCode)
	res.Body.Close()

	// Output:
	// 200
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
	serverConfig := config.DefaultServerConfig()
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
		ServerConfig: &serverConfig,
	}
	handlers := InitializeHandlers(&metricsService, &cryptoService, nil)
	for _, test := range valueTests {
		t.Run(test.name, func(t *testing.T) {
			r := CreateChiRouterWithHandlers(&handlers, &serverConfig)

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

func ExampleReadHandler_ValueHandler() {
	// prepare handler
	serverConfig := config.DefaultServerConfig()
	cryptoService := service.CryptoService{
		ServerConfig: &serverConfig,
	}

	eventBroker := memory.NewBroker()
	brokerContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	auditService := service.NewAuditService(&serverConfig)
	auditService.SubscribeToMetricEvents(eventBroker, brokerContext)

	metricsService := service.NewMetricsService(&serverConfig, &db.PostgresDatabase{}, eventBroker)
	metricsService.SaveMetricsModel(context.Background(), []models.Metrics{
		{ID: "TestCounter", MType: models.Counter, Delta: Ptr(int64(10.0))},
	})

	handlers := InitializeHandlers(metricsService, &cryptoService, nil)
	r := CreateChiRouterWithHandlers(&handlers, &serverConfig)

	// call handler
	request := httptest.NewRequest(http.MethodGet, "/value/counter/TestCounter", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, request)
	res := w.Result()

	fmt.Println(res.StatusCode)
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bodyBytes))
	res.Body.Close()

	// Output:
	// 200
	// 10
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
				Value: Ptr(float64(123.456)),
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
				Delta: Ptr(int64(10)),
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
			serverConfig := config.DefaultServerConfig()
			metricsService := service.NewMetricsService(&serverConfig, &db.PostgresDatabase{}, nil)
			cryptoService := service.CryptoService{
				ServerConfig: &serverConfig,
			}
			metricsService.Storage.SaveMetrics(context.Background(), test.metrics)
			handlers := InitializeHandlers(metricsService, &cryptoService, nil)

			r := CreateChiRouterWithHandlers(&handlers, &serverConfig)

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

func ExampleReadHandler_ValueJSONHandler() {
	// prepare handler
	serverConfig := config.DefaultServerConfig()
	cryptoService := service.CryptoService{
		ServerConfig: &serverConfig,
	}

	eventBroker := memory.NewBroker()
	brokerContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	auditService := service.NewAuditService(&serverConfig)
	auditService.SubscribeToMetricEvents(eventBroker, brokerContext)

	metricsService := service.NewMetricsService(&serverConfig, &db.PostgresDatabase{}, eventBroker)
	metricsService.SaveMetricsModel(context.Background(), []models.Metrics{
		{ID: "TestCounter", MType: models.Counter, Delta: Ptr(int64(10.0))},
	})

	handlers := InitializeHandlers(metricsService, &cryptoService, nil)
	r := CreateChiRouterWithHandlers(&handlers, &serverConfig)

	// call handler
	body := `{"id":"TestCounter","type":"counter"}`
	request := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader([]byte(body)))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, request)
	res := w.Result()

	fmt.Println(res.StatusCode)
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bodyBytes))
	res.Body.Close()

	// Output:
	// 200
	// {"id":"TestCounter","type":"counter","delta":10}
}
