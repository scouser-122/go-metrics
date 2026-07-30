package handler

import (
	"context"
	"fmt"
	"testing"

	"github.com/botchris/go-pubsub/provider/memory"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/scouser-122/go-metrics/internal/agent"
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
	pb "github.com/scouser-122/go-metrics/internal/proto"
	"github.com/scouser-122/go-metrics/internal/repository/db"
	"github.com/scouser-122/go-metrics/internal/service"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type grpcRequest struct {
	metrics []models.Metrics
	localIP string
}

type grpcWant struct {
	errCode codes.Code
}

var updateGrpcMetricsTests = []struct {
	name    string
	request grpcRequest
	mockDB  db.MockPostgresDBTestData
	want    grpcWant
}{
	{
		name: "positive test update metrics",
		request: grpcRequest{
			metrics: []models.Metrics{
				{
					ID:    "TotalAlloc",
					MType: models.Counter,
					Delta: Ptr(int64(10)),
				},
				{
					ID:    "LastGC",
					MType: models.Gauge,
					Value: Ptr(float64(123.456)),
				},
			},
			localIP: "192.168.0.1",
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
		want: grpcWant{
			errCode: codes.OK,
		},
	},
	{
		name: "negative test update metrics",
		request: grpcRequest{
			metrics: []models.Metrics{
				{
					ID:    "TotalAlloc",
					MType: models.Counter,
					Delta: Ptr(int64(10)),
				},
				{
					ID:    "LastGC",
					MType: models.Gauge,
					Value: Ptr(float64(123.456)),
				},
			},
			localIP: "192.168.0.1",
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
		want: grpcWant{
			errCode: codes.Internal,
		},
	},

	{
		name: "negative test wrong IP",
		request: grpcRequest{
			metrics: []models.Metrics{
				{
					ID:    "TotalAlloc",
					MType: models.Counter,
					Delta: Ptr(int64(10)),
				},
				{
					ID:    "LastGC",
					MType: models.Gauge,
					Value: Ptr(float64(123.456)),
				},
			},
			localIP: "192.168.10.1",
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
		want: grpcWant{
			errCode: codes.PermissionDenied,
		},
	},
}

func TestGrpcUpdateMetrics(t *testing.T) {
	for _, test := range updateGrpcMetricsTests {
		t.Run(test.name, func(t *testing.T) {
			serverConfig := config.DefaultServerConfig()
			serverConfig.GrpcPort = "45150"
			serverConfig.TrustedSubnet = "192.168.0.1/24"
			if err := logger.Initialize(serverConfig.LogLevel, serverConfig.Environment); err != nil {
				panic(err)
			}
			cryptoService := service.CryptoService{
				ServerConfig: &serverConfig,
			}
			cryptoService.LoadPrivateKeyIfExists()
			mock, err := pgxmock.NewPool()
			if err != nil {
				panic(err)
			}
			defer mock.Close()
			mockDB := &test.mockDB
			mockDB.PgxPoolIface = mock
			mockDB.MockDBCalls(*mockDB)
			db := db.NewPgxMockDB(&serverConfig, mock)

			eventBroker := memory.NewBroker()
			brokerContext, cancel := context.WithCancel(context.Background())
			defer cancel()
			auditService := service.NewAuditService(&serverConfig)
			auditService.SubscribeToMetricEvents(eventBroker, brokerContext)

			metricsService := service.NewMetricsService(&serverConfig, &db, eventBroker)

			grpcServer := NewGrpcServer(&serverConfig)
			go func() {
				grpcServer.Start(metricsService)
			}()
			defer grpcServer.Stop()

			conn, err := grpc.NewClient(":45150", grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Fatalf("Failed to dial: %v", err)
			}
			defer conn.Close()

			client := pb.NewMetricsClient(conn)

			ctx := context.Background()
			md := metadata.New(map[string]string{"X-Real-IP": test.request.localIP})
			ctx = metadata.NewOutgoingContext(ctx, md)
			_, err = client.UpdateMetrics(ctx, pb.UpdateMetricsRequest_builder{
				Metrics: agent.MetricsToPbMetrics(test.request.metrics),
			}.Build())

			var errCode codes.Code
			if err != nil {
				logger.Sugar.Errorf("error sending metrics: %s", err)
				grpcStatus, _ := status.FromError(err)
				errCode = grpcStatus.Code()
			} else {
				errCode = codes.OK
			}
			assert.Equal(t, test.want.errCode, errCode)
		})
	}
}
