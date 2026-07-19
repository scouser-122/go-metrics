package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/botchris/go-pubsub/provider/memory"
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/repository/db"
	"github.com/scouser-122/go-metrics/internal/service"
	"github.com/stretchr/testify/assert"
)

var updateJSONTrustedIPTests = []struct {
	name    string
	request request
	want    want
}{
	{
		name: "positive test trusted IP",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/update",
			headers: map[string]string{
				"X-Real-IP": "192.168.0.1",
			},
			body: `{"id":"LastGC","type":"gauge","value":123.456}`,
		},
		want: want{
			code:        http.StatusOK,
			contentType: "application/json",
		},
	},
	{
		name: "negative test trusted IP untrusted",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/update",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.1",
			},
			body: `{"id":"LastGC","type":"gauge","value":123.456}`,
		},
		want: want{
			code:        http.StatusForbidden,
			contentType: "application/json",
		},
	},
	{
		name: "negative test trusted IP incorrect IP",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/update",
			headers: map[string]string{
				"X-Real-IP": "23452345",
			},
			body: `{"id":"LastGC","type":"gauge","value":123.456}`,
		},
		want: want{
			code:        http.StatusForbidden,
			contentType: "application/json",
		},
	},
	{
		name: "negative test trusted IP absent IP",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/update",
			body:        `{"id":"LastGC","type":"gauge","value":123.456}`,
		},
		want: want{
			code:        http.StatusForbidden,
			contentType: "application/json",
		},
	},
}

func TestUpdateJSONTrustedIP(t *testing.T) {
	for _, test := range updateJSONTrustedIPTests {
		t.Run(test.name, func(t *testing.T) {
			serverConfig := config.DefaultServerConfig()
			serverConfig.TrustedSubnet = "192.168.0.0/24"
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

			r := CreateChiRouterWithHandlers(&handlers, &serverConfig)

			var bodyReader io.Reader
			if test.request.body != "" {
				jsonData := []byte(test.request.body)
				bodyReader = bytes.NewBuffer(jsonData)
			}

			request := httptest.NewRequest(test.request.method, test.request.path, bodyReader)
			request.Header.Add("Content-Type", test.request.contentType)
			for k, v := range test.request.headers {
				request.Header.Add(k, v)
			}

			// создаём новый Recorder
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			res.Body.Close()
		})
	}
}
