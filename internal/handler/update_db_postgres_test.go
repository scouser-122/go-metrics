package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/scouser-122/go-metrics/internal/config"
	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/repository/db"
	"github.com/scouser-122/go-metrics/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var updateTestsDBPostgres = []struct {
	name    string
	request request
	mockDB  db.MockPostgresDBTestData
	want    want
}{
	{
		name: "positive test create gauge",
		request: request{
			method:      http.MethodPost,
			contentType: "text/plain",
			path:        "/update/gauge/TestGauge1/120.50",
		},
		mockDB: db.MockPostgresDBTestData{
			MockPool: &db.MockPostgresPool{
				MockMethods: func(tt db.MockPostgresDBTestData) {
					tt.MockPool.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).
						Return(tt.MockRow)
					tt.MockPool.On("Exec", mock.Anything, mock.Anything, mock.Anything).
						Return(tt.MockTag, tt.MockError)
					tt.MockPool.On("Ping", mock.Anything).
						Return(nil)
				},
			},
			MockRow: &db.MockPostgresRow{
				Metric: nil,
			},
			MockTag: pgconn.NewCommandTag("CREATE 1"),
		},
		want: want{
			code:        http.StatusOK,
			contentType: "text/plain",
			body:        "120.5",
		},
	},
	{
		name: "positive test update gauge",
		request: request{
			method:      http.MethodPost,
			contentType: "text/plain",
			path:        "/update/gauge/TestGauge1/120.50",
		},
		mockDB: db.MockPostgresDBTestData{
			MockPool: &db.MockPostgresPool{
				MockMethods: func(tt db.MockPostgresDBTestData) {
					tt.MockPool.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).
						Return(tt.MockRow)
					tt.MockPool.On("Exec", mock.Anything, mock.Anything, mock.Anything).
						Return(tt.MockTag, tt.MockError)
					tt.MockPool.On("Ping", mock.Anything).
						Return(nil)
				},
			},
			MockRow: &db.MockPostgresRow{
				Metric: &models.Metrics{ID: "TestGauge1", MType: models.Gauge, Value: ptr(float64(10.0))},
			},
			MockTag: pgconn.NewCommandTag("UPDATE 1"),
		},
		want: want{
			code:        http.StatusOK,
			contentType: "text/plain",
			body:        "120.5",
		},
	},
	{
		name: "negative test update gauge",
		request: request{
			method:      http.MethodPost,
			contentType: "text/plain",
			path:        "/update/gauge/TestGauge1/120.50",
		},
		mockDB: db.MockPostgresDBTestData{
			MockPool: &db.MockPostgresPool{
				MockMethods: func(tt db.MockPostgresDBTestData) {
					tt.MockPool.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).
						Return(tt.MockRow)
					tt.MockPool.On("Ping", mock.Anything).
						Return(nil)
				},
			},
			MockRow: &db.MockPostgresRow{
				Metric: nil,
				Err:    fmt.Errorf("QueryRow error"),
			},
			MockTag: pgconn.NewCommandTag("UPDATE 1"),
		},
		want: want{
			code:        http.StatusInternalServerError,
			contentType: "text/plain",
			body:        "120.5",
		},
	},
	{
		name: "positive test create counter",
		request: request{
			method:      http.MethodPost,
			contentType: "text/plain",
			path:        "/update/counter/TestCounter1/10",
		},
		mockDB: db.MockPostgresDBTestData{
			MockPool: &db.MockPostgresPool{
				MockMethods: func(tt db.MockPostgresDBTestData) {
					tt.MockPool.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).
						Return(tt.MockRow)
					tt.MockPool.On("Exec", mock.Anything, mock.Anything, mock.Anything).
						Return(tt.MockTag, tt.MockError)
					tt.MockPool.On("Ping", mock.Anything).
						Return(nil)
				},
			},
			MockRow: &db.MockPostgresRow{
				Metric: nil,
			},
			MockTag: pgconn.NewCommandTag("CREATE 1"),
		},
		want: want{
			code:        http.StatusOK,
			contentType: "text/plain",
			body:        "10",
		},
	},
	{
		name: "positive test update counter",
		request: request{
			method:      http.MethodPost,
			contentType: "text/plain",
			path:        "/update/counter/TestCounter1/10",
		},
		mockDB: db.MockPostgresDBTestData{
			MockPool: &db.MockPostgresPool{
				MockMethods: func(tt db.MockPostgresDBTestData) {
					tt.MockPool.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).
						Return(tt.MockRow)
					tt.MockPool.On("Exec", mock.Anything, mock.Anything, mock.Anything).
						Return(tt.MockTag, tt.MockError)
					tt.MockPool.On("Ping", mock.Anything).
						Return(nil)
				},
			},
			MockRow: &db.MockPostgresRow{
				Metric: &models.Metrics{ID: "TestCounter1", MType: models.Counter, Delta: ptr(int64(10))},
			},
			MockTag: pgconn.NewCommandTag("UPDATE 1"),
		},
		want: want{
			code:        http.StatusOK,
			contentType: "text/plain",
			body:        "20",
		},
	},
}

func TestUpdateHandlerDBPostgres(t *testing.T) {
	for _, test := range updateTestsDBPostgres {
		t.Run(test.name, func(t *testing.T) {
			config := config.DefaultServerConfig()
			cryptoService := service.CryptoService{
				ServerConfig: &config,
			}
			test.mockDB.MockPool.MockMethods(test.mockDB)
			db := db.NewMockPostgresDB(config, test.mockDB.MockPool)
			metricsService := service.NewMetricsService(&config, &db)
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
