package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/botchris/go-pubsub/provider/memory"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/scouser-122/go-metrics/internal/config"
	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/repository/db"
	"github.com/scouser-122/go-metrics/internal/service"
	"github.com/stretchr/testify/assert"
)

var listPostgresDBTests = []struct {
	name    string
	request request
	mockDB  db.MockPostgresDBTestData
	want    want
}{
	{
		name: "positive test list metrics",
		request: request{
			method: http.MethodGet,
			path:   "/",
		},
		mockDB: db.MockPostgresDBTestData{
			MockDBCalls: func(tt db.MockPostgresDBTestData) {
				mock := tt.PgxPoolIface
				mock.ExpectPing()
				metrics := []models.Metrics{}
				for i := 0; i < 9; i++ {
					if i%2 == 0 {
						metrics = append(metrics, models.Metrics{
							ID:    fmt.Sprintf("TestCounter_%d", i),
							MType: models.Counter,
							Delta: Ptr(int64(10 * i)),
							Value: nil,
						})
					} else {
						metrics = append(metrics, models.Metrics{
							ID:    fmt.Sprintf("TestGauge_%d", i),
							MType: models.Gauge,
							Value: Ptr(float64(i)),
							Delta: nil,
						})
					}
				}
				rows := mock.NewRows([]string{"id", "type", "delta", "value"})
				for _, m := range metrics {
					rows.AddRow(m.ID, m.MType, m.Delta, m.Value)
				}
				mock.ExpectQuery("SELECT id, type, delta, value FROM metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
				mock.ExpectQuery("SELECT id, type, delta, value FROM metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(mock.NewRows([]string{"id", "type", "delta", "value"}))
			},
		},
		want: want{
			code:        http.StatusOK,
			contentType: "text/html; charset=utf-8",
		},
	},
}

func TestListHandlerDBPostgres(t *testing.T) {
	for _, test := range listPostgresDBTests {
		t.Run(test.name, func(t *testing.T) {
			r := createTestRouterPostgresDB(&test.mockDB)

			var bodyReader io.Reader
			if test.request.body != "" {
				jsonData := []byte(test.request.body)
				bodyReader = bytes.NewBuffer(jsonData)
			}

			request := httptest.NewRequest(test.request.method, test.request.path, bodyReader)
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
				assert.Equal(t, test.want.body, strings.ReplaceAll(bodyString, "\n", ""))
			}
			res.Body.Close()
		})
	}
}

// Benchmark for ListHandler func
func BenchmarkListHandlerDBPostgres(b *testing.B) {
	// prepare config
	serverConfig := config.DefaultServerConfig()
	cryptoService := service.CryptoService{
		ServerConfig: &serverConfig,
	}

	// reset benchmark timer
	b.ResetTimer()

	// run benchmark b.N times
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		// prepare test data
		mock, err := pgxmock.NewPool()
		if err != nil {
			panic(err)
		}
		mockDB := db.MockPostgresDBTestData{
			MockDBCalls: func(tt db.MockPostgresDBTestData) {
				mock := tt.PgxPoolIface
				mock.ExpectPing()
				for i := 0; i < 100; i++ {
					metrics := []models.Metrics{}
					for j := 0; j < 10; j++ {
						if j%2 == 0 {
							metrics = append(metrics, models.Metrics{
								ID:    fmt.Sprintf("TestCounter_%d_%d", i, j),
								MType: models.Counter,
								Delta: Ptr(int64(10*i + j)),
								Value: nil,
							})
						} else {
							metrics = append(metrics, models.Metrics{
								ID:    fmt.Sprintf("TestGauge_%d_%d", i, j),
								MType: models.Gauge,
								Value: Ptr(float64(10*i + j)),
								Delta: nil,
							})
						}
					}
					rows := mock.NewRows([]string{"id", "type", "delta", "value"})
					for _, m := range metrics {
						rows.AddRow(m.ID, m.MType, m.Delta, m.Value)
					}
					mock.ExpectQuery("SELECT id, type, delta, value FROM metrics").
						WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
						WillReturnRows(rows)
					mock.ExpectQuery("SELECT id, type, delta, value FROM metrics").
						WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
						WillReturnRows(mock.NewRows([]string{"id", "type", "delta", "value"}))
				}
			},
		}
		mockDB.PgxPoolIface = mock
		mockDB.MockDBCalls(mockDB)
		db := db.NewPgxMockDB(serverConfig, mock)

		eventBroker := memory.NewBroker()
		brokerContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		auditService := service.NewAuditService(&serverConfig)
		auditService.SubscribeToMetricEvents(eventBroker, brokerContext)

		metricsService := service.NewMetricsService(&serverConfig, &db, eventBroker)
		handlers := InitializeHandlers(metricsService, &cryptoService, nil)

		handlerIndex := slices.IndexFunc(handlers, func(h Handler) bool {
			return h.Method == http.MethodGet && h.URLPathPattern == "/"
		})
		handler := handlers[handlerIndex]

		// create request
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		// create response recorder
		rr := httptest.NewRecorder()

		b.StartTimer()

		// call handler
		handler.HandlerFn(rr, req)

		mock.Close()
	}
}
