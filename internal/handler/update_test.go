package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/botchris/go-pubsub/provider/memory"
	"github.com/scouser-122/go-metrics/internal/config"
	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/repository/db"
	"github.com/scouser-122/go-metrics/internal/service"
	"github.com/stretchr/testify/assert"
)

type want struct {
	code        int
	contentType string
	body        string
}

type request struct {
	method      string
	contentType string
	path        string
	body        string
}

var updateTestsMemStorage = []struct {
	name    string
	request request
	want    want
}{
	{
		name: "positive test gauge",
		request: request{
			method:      http.MethodPost,
			contentType: "text/plain",
			path:        "/update/gauge/Alloc/120.50",
		},
		want: want{
			code:        http.StatusOK,
			contentType: "text/plain",
			body:        "120.5",
		},
	},
	{
		name: "positive test counter",
		request: request{
			method:      http.MethodPost,
			contentType: "text/plain",
			path:        "/update/counter/PollCount/10",
		},
		want: want{
			code:        http.StatusOK,
			contentType: "text/plain",
			body:        "10",
		},
	},
	{
		name: "negative test gauge format",
		request: request{
			method:      http.MethodPost,
			contentType: "text/plain",
			path:        "/update/gauge/Alloc/t23",
		},
		want: want{
			code:        http.StatusBadRequest,
			contentType: "text/plain",
		},
	},
	{
		name: "negative test counter format",
		request: request{
			method:      http.MethodPost,
			contentType: "text/plain",
			path:        "/update/counter/PollCount/t23",
		},
		want: want{
			code:        http.StatusBadRequest,
			contentType: "text/plain",
		},
	},
	{
		name: "negative test incorrect method",
		request: request{
			method:      http.MethodGet,
			contentType: "text/plain",
			path:        "/update/counter/PollCount/10",
		},
		want: want{
			code:        http.StatusMethodNotAllowed,
			contentType: "text/plain",
		},
	},
	{
		name: "negative test missing value",
		request: request{
			method:      http.MethodPost,
			contentType: "text/plain",
			path:        "/update/counter/PollCount",
		},
		want: want{
			code:        http.StatusNotFound,
			contentType: "text/plain",
		},
	},
	{
		name: "negative test missing metric name",
		request: request{
			method:      http.MethodPost,
			contentType: "text/plain",
			path:        "/update/counter",
		},
		want: want{
			code:        http.StatusNotFound,
			contentType: "text/plain",
		},
	},
	{
		name: "negative test incorrect metric type",
		request: request{
			method:      http.MethodPost,
			contentType: "text/plain",
			path:        "/update/histogram/Alloc/10.0",
		},
		want: want{
			code:        http.StatusBadRequest,
			contentType: "text/plain",
		},
	},
}

func TestUpdateHandlerMemStorage(t *testing.T) {
	for _, test := range updateTestsMemStorage {
		t.Run(test.name, func(t *testing.T) {
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

			handlers := InitializeHandlers(metricsService, &cryptoService, nil)

			r := CreateChiRouter(&handlers)

			request := httptest.NewRequest(test.request.method, test.request.path, nil)
			request.Header.Add("Content-Type", test.request.contentType)

			// создаём новый Recorder
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			if res.StatusCode == http.StatusOK && test.want.body != "" {
				bodyBytes, err := io.ReadAll(res.Body)
				assert.Nil(t, err)
				bodyString := string(bodyBytes)
				assert.Equal(t, test.want.body, strings.Replace(bodyString, "\n", "", -1))
			}
			res.Body.Close()
		})
	}
}

var updateJSONTests = []struct {
	name    string
	request request
	want    want
}{
	{
		name: "positive test update metric gauge",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/update",
			body:        `{"id":"LastGC","type":"gauge","value":123.456}`,
		},
		want: want{
			code:        http.StatusOK,
			contentType: "application/json",
			body:        `{"id":"LastGC","type":"gauge","value":123.456}`,
		},
	},
	{
		name: "positive test update metric counter",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/update",
			body:        `{"id":"TotalAlloc","type":"counter","delta":123456}`,
		},
		want: want{
			code:        http.StatusOK,
			contentType: "application/json",
			body:        `{"id":"TotalAlloc","type":"counter","delta":123456}`,
		},
	},
	{
		name: "negative test update bad json",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/update",
			body:        `{"id":"TotalAlloc","type":"coun`,
		},
		want: want{
			code:        http.StatusBadRequest,
			contentType: "application/json",
		},
	},
	{
		name: "negative test update incorrect metric type",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/update",
			body:        `{"id":"LastGC","type":"histogram","value":123.456}`,
		},
		want: want{
			code:        http.StatusBadRequest,
			contentType: "application/json",
		},
	},
	{
		name: "negative test update incorrect gauge metric format",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/update",
			body:        `{"id":"LastGC","type":"gauge","value":"123"}`,
		},
		want: want{
			code:        http.StatusBadRequest,
			contentType: "application/json",
		},
	},
	{
		name: "negative test update incorrect counter metric format",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/update",
			body:        `{"id":"LastGC","type":"counter","delta":123.456}`,
		},
		want: want{
			code:        http.StatusBadRequest,
			contentType: "application/json",
		},
	},
	{
		name: "negative test update counter missing delta",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/update",
			body:        `{"id":"LastGC","type":"counter","value":123.456}`,
		},
		want: want{
			code:        http.StatusBadRequest,
			contentType: "application/json",
		},
	},
	{
		name: "negative test update gauge missing value",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/update",
			body:        `{"id":"LastGC","type":"gauge","delta":123}`,
		},
		want: want{
			code:        http.StatusBadRequest,
			contentType: "application/json",
		},
	},
}

func TestUpdateJSONHandler(t *testing.T) {
	for _, test := range updateJSONTests {
		t.Run(test.name, func(t *testing.T) {
			serverConfig := config.DefaultServerConfig()
			serverConfig.HMACKey = "secret_key"
			cryptoService := service.CryptoService{
				ServerConfig: &serverConfig,
			}

			eventBroker := memory.NewBroker()
			brokerContext, cancel := context.WithCancel(context.Background())
			defer cancel()
			auditService := service.NewAuditService(&serverConfig)
			auditService.SubscribeToMetricEvents(eventBroker, brokerContext)

			metricsService := service.NewMetricsService(&serverConfig, &db.PostgresDatabase{}, eventBroker)
			handlers := InitializeHandlers(metricsService, &cryptoService, nil)

			r := CreateChiRouter(&handlers)

			var bodyReader io.Reader
			var bodyHash string
			if test.request.body != "" {
				jsonData := []byte(test.request.body)
				bodyReader = bytes.NewBuffer(jsonData)
				bodyHash = cryptoService.CalculateHash(jsonData)
			}

			request := httptest.NewRequest(test.request.method, test.request.path, bodyReader)
			request.Header.Add("Content-Type", test.request.contentType)
			if bodyHash != "" {
				request.Header.Add("HashSHA256", bodyHash)
			}

			// создаём новый Recorder
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			if res.StatusCode == http.StatusOK && test.want.body != "" {
				bodyBytes, err := io.ReadAll(res.Body)
				assert.Nil(t, err)
				bodyString := string(bodyBytes)
				assert.Equal(t, test.want.body, strings.Replace(bodyString, "\n", "", -1))
			}
			res.Body.Close()
		})
	}
}

// Benchmark for UpdateJSONArrayHandler func
func BenchmarkUpdateJSONArrayHandler(b *testing.B) {
	// prepare handler
	serverConfig := config.DefaultServerConfig()
	serverConfig.HMACKey = "secret_key"
	cryptoService := service.CryptoService{
		ServerConfig: &serverConfig,
	}

	eventBroker := memory.NewBroker()
	brokerContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	auditService := service.NewAuditService(&serverConfig)
	auditService.SubscribeToMetricEvents(eventBroker, brokerContext)

	metricsService := service.NewMetricsService(&serverConfig, &db.PostgresDatabase{}, eventBroker)
	handlers := InitializeHandlers(metricsService, &cryptoService, nil)

	handlerIndex := slices.IndexFunc(handlers, func(h Handler) bool {
		return h.Method == http.MethodPost && h.URLPathPattern == "/updates"
	})
	handler := handlers[handlerIndex]

	// prepare request data
	metricsSize := 500
	requestData := make([]models.Metrics, 0, metricsSize)
	for i := 0; i < metricsSize/2; i += 2 {
		requestData = append(requestData, models.Metrics{
			ID:    fmt.Sprintf("TestCounter_%d", i),
			MType: models.Counter,
			Delta: Ptr(int64(i * 10)),
		})
		requestData = append(requestData, models.Metrics{
			ID:    fmt.Sprintf("TestGauge_%d", i+1),
			MType: models.Counter,
			Value: Ptr(float64((i + 1) * 10)),
		})
	}

	body, _ := json.Marshal(requestData)

	// reset benchmark timer
	b.ResetTimer()

	// run benchmark b.N times
	for i := 0; i < b.N; i++ {
		// create request
		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// create response recorder
		rr := httptest.NewRecorder()

		// call handler
		handler.HandlerFn(rr, req)
	}
}
