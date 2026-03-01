package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/repository"
	"github.com/scouser-122/go-metrics/internal/service"
	"github.com/stretchr/testify/assert"
)

var listTests = []struct {
	name    string
	request request
	want    want
}{
	{
		name: "positive test",
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
	for _, test := range listTests {
		t.Run(test.name, func(t *testing.T) {
			serverConfig := config.GetDefaultServerConfig()

			memStorage := repository.MemStorage{}
			memStorage.FillMetrics(&serverConfig)

			metricsService := service.MetricsService{
				Storage: &memStorage,
			}
			updateHandler := UpdateHandler{
				Service: metricsService,
			}
			listHandler := ListHandler{
				Service: metricsService,
			}
			listHandler.CreateTemplate()

			r := CreateChiRouter(updateHandler.UpdateHandler, listHandler.ListHandler)

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
