package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/repository/db"
	"github.com/stretchr/testify/assert"
)

var updatePostgresDBTests = []struct {
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
			MockDBCalls: func(tt db.MockPostgresDBTestData) {
				mock := tt.PgxPoolIface
				mock.ExpectPing()
				mock.ExpectQuery("SELECT id, type FROM metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(pgx.ErrNoRows)
				mock.ExpectExec("INSERT INTO metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("CREATE", 1))
			},
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
			MockDBCalls: func(tt db.MockPostgresDBTestData) {
				mock := tt.PgxPoolIface
				mock.ExpectPing()
				mock.ExpectQuery("SELECT id, type FROM metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(mock.NewRows([]string{"id", "type"}).
						AddRow("TestGauge1", models.Gauge))
				mock.ExpectExec("UPDATE metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
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
			MockDBCalls: func(tt db.MockPostgresDBTestData) {
				mock := tt.PgxPoolIface
				mock.ExpectPing()
				mock.ExpectQuery("SELECT id, type FROM metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(fmt.Errorf("DB query error"))
			},
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
			MockDBCalls: func(tt db.MockPostgresDBTestData) {
				mock := tt.PgxPoolIface
				mock.ExpectPing()
				mock.ExpectQuery("SELECT id, type, delta FROM metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(pgx.ErrNoRows)
				mock.ExpectExec("INSERT INTO metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("CREATE", 1))
			},
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
			MockDBCalls: func(tt db.MockPostgresDBTestData) {
				mock := tt.PgxPoolIface
				mock.ExpectPing()
				mock.ExpectQuery("SELECT id, type, delta FROM metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(mock.NewRows([]string{"id", "type", "delta"}).
						AddRow("TestCounter1", models.Counter, Ptr(int64(10))))
				mock.ExpectExec("UPDATE metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		want: want{
			code:        http.StatusOK,
			contentType: "text/plain",
			body:        "20",
		},
	},
}

func TestUpdateHandlerDBPostgres(t *testing.T) {
	for _, test := range updatePostgresDBTests {
		t.Run(test.name, func(t *testing.T) {
			r := createTestRouterPostgresDB(&test.mockDB)

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

var updateJSONPostgresDBTests = []struct {
	name    string
	request request
	mockDB  db.MockPostgresDBTestData
	want    want
}{
	{
		name: "positive test create metric gauge",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/update",
			body:        `{"id":"LastGC","type":"gauge","value":123.456}`,
		},
		mockDB: db.MockPostgresDBTestData{
			MockDBCalls: func(tt db.MockPostgresDBTestData) {
				mock := tt.PgxPoolIface
				mock.ExpectPing()
				mock.ExpectQuery("SELECT delta, value FROM metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(pgx.ErrNoRows)
				mock.ExpectExec("INSERT INTO metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("CREATE", 1))
			},
		},
		want: want{
			code:        http.StatusOK,
			contentType: "application/json",
			body:        `{"id":"LastGC","type":"gauge","value":123.456}`,
		},
	},
	{
		name: "positive test update metric gauge",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/update",
			body:        `{"id":"LastGC","type":"gauge","value":123.456}`,
		},
		mockDB: db.MockPostgresDBTestData{
			MockDBCalls: func(tt db.MockPostgresDBTestData) {
				mock := tt.PgxPoolIface
				mock.ExpectPing()
				mock.ExpectQuery("SELECT delta, value FROM metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(mock.NewRows([]string{"delta", "value"}).
						AddRow(nil, Ptr(float64(10.0))))
				mock.ExpectExec("UPDATE metrics SET value").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		want: want{
			code:        http.StatusOK,
			contentType: "application/json",
			body:        `{"id":"LastGC","type":"gauge","value":123.456}`,
		},
	},
	{
		name: "positive test create metric counter",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/update",
			body:        `{"id":"TotalAlloc","type":"counter","delta":123456}`,
		},

		mockDB: db.MockPostgresDBTestData{
			MockDBCalls: func(tt db.MockPostgresDBTestData) {
				mock := tt.PgxPoolIface
				mock.ExpectPing()
				mock.ExpectQuery("SELECT delta, value FROM metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(pgx.ErrNoRows)
				mock.ExpectExec("INSERT INTO metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("CREATE", 1))
			},
		},
		want: want{
			code:        http.StatusOK,
			contentType: "application/json",
			body:        `{"id":"TotalAlloc","type":"counter","delta":123456}`,
		},
	},
	{
		name: "positive test update metric counter",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/update",
			body:        `{"id":"TotalAlloc","type":"counter","delta":10}`,
		},

		mockDB: db.MockPostgresDBTestData{
			MockDBCalls: func(tt db.MockPostgresDBTestData) {
				mock := tt.PgxPoolIface
				mock.ExpectPing()
				mock.ExpectQuery("SELECT delta, value FROM metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(mock.NewRows([]string{"delta", "value"}).
						AddRow(Ptr(int64(10)), nil))
				mock.ExpectExec("UPDATE metrics SET delta").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		want: want{
			code:        http.StatusOK,
			contentType: "application/json",
			body:        `{"id":"TotalAlloc","type":"counter","delta":20}`,
		},
	},
}

func TestUpdateJSONHandlerDBPostgres(t *testing.T) {
	for _, test := range updateJSONPostgresDBTests {
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
				assert.Equal(t, test.want.body, strings.Replace(bodyString, "\n", "", -1))
			}
			res.Body.Close()
		})
	}
}

var updateJSONArrayPostgresDBTests = []struct {
	name    string
	request request
	mockDB  db.MockPostgresDBTestData
	want    want
}{
	{
		name: "positive test create metrics",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/updates",
			body:        `[{"id":"TotalAlloc","type":"counter","delta":10},{"id":"LastGC","type":"gauge","value":123.456}]`,
		},
		mockDB: db.MockPostgresDBTestData{
			MockDBCalls: func(tt db.MockPostgresDBTestData) {
				mock := tt.PgxPoolIface
				mock.ExpectPing()
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT delta, value FROM metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(pgx.ErrNoRows)
				mock.ExpectExec("INSERT INTO metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("CREATE", 1))
				mock.ExpectQuery("SELECT delta, value FROM metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(pgx.ErrNoRows)
				mock.ExpectExec("INSERT INTO metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("CREATE", 1))
				mock.ExpectCommit()
			},
		},
		want: want{
			code:        http.StatusOK,
			contentType: "application/json",
			body:        `{"status":"ok","message":"successfully saved 2 metrics"}`,
		},
	},
	{
		name: "positive test update metrics",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/updates",
			body:        `[{"id":"TotalAlloc","type":"counter","delta":10},{"id":"LastGC","type":"gauge","value":123.456}]`,
		},
		mockDB: db.MockPostgresDBTestData{
			MockDBCalls: func(tt db.MockPostgresDBTestData) {
				mock := tt.PgxPoolIface
				mock.ExpectPing()
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT delta, value FROM metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(mock.NewRows([]string{"delta", "value"}).
						AddRow(Ptr(int64(10)), nil))
				mock.ExpectExec("UPDATE metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectQuery("SELECT delta, value FROM metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(mock.NewRows([]string{"delta", "value"}).
						AddRow(nil, Ptr(float64(10.0))))
				mock.ExpectExec("UPDATE metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectCommit()
			},
		},
		want: want{
			code:        http.StatusOK,
			contentType: "application/json",
			body:        `{"status":"ok","message":"successfully saved 2 metrics"}`,
		},
	},
	{
		name: "negative test create metrics",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/updates",
			body:        `[{"id":"TotalAlloc","type":"counter","delta":10},{"id":"LastGC","type":"gauge","value":123.456}]`,
		},
		mockDB: db.MockPostgresDBTestData{
			MockDBCalls: func(tt db.MockPostgresDBTestData) {
				mock := tt.PgxPoolIface
				mock.ExpectPing()
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT delta, value FROM metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(pgx.ErrNoRows)
				mock.ExpectExec("INSERT INTO metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(fmt.Errorf("DB query error"))
			},
		},
		want: want{
			code:        http.StatusInternalServerError,
			contentType: "application/json",
		},
	},
	{
		name: "negative test update metrics",
		request: request{
			method:      http.MethodPost,
			contentType: "application/json",
			path:        "/updates",
			body:        `[{"id":"TotalAlloc","type":"counter","delta":10},{"id":"LastGC","type":"gauge","value":123.456}]`,
		},
		mockDB: db.MockPostgresDBTestData{
			MockDBCalls: func(tt db.MockPostgresDBTestData) {
				mock := tt.PgxPoolIface
				mock.ExpectPing()
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT delta, value FROM metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(mock.NewRows([]string{"delta", "value"}).
						AddRow(Ptr(int64(10)), nil))
				mock.ExpectExec("INSERT INTO metrics").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(fmt.Errorf("DB query error"))
			},
		},
		want: want{
			code:        http.StatusInternalServerError,
			contentType: "application/json",
		},
	},
}

func TestUpdateJSONArrayHandlerDBPostgres(t *testing.T) {
	for _, test := range updateJSONArrayPostgresDBTests {
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
