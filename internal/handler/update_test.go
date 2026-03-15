package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scouser-122/go-metrics/internal/repository"
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
}

var updateTests = []struct {
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
		name: "negative test missing metric type",
		request: request{
			method:      http.MethodPost,
			contentType: "text/plain",
			path:        "/update",
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

func TestUpdateHandler(t *testing.T) {
	for _, test := range updateTests {
		t.Run(test.name, func(t *testing.T) {
			memStorage := repository.MemStorage{}

			metricsService := service.MetricsService{
				Storage: &memStorage,
			}
			handlers := InitializeHandlers(&metricsService)

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
				assert.Equal(t, test.want.body, bodyString)
			}
			res.Body.Close()
		})
	}
}
