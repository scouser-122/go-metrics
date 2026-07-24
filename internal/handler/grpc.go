package handler

import (
	"context"
	"fmt"
	"net"
	"os"

	models "github.com/scouser-122/go-metrics/internal/model"
	pb "github.com/scouser-122/go-metrics/internal/proto"
	"github.com/scouser-122/go-metrics/internal/service"

	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServer

	metricsService *service.MetricsService
}

func NewMetricsServer(metricsService *service.MetricsService) *MetricsServer {
	return &MetricsServer{
		metricsService: metricsService,
	}
}

func (s *MetricsServer) UpdateMetrics(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	var response pb.UpdateMetricsResponse

	inMetrics := in.GetMetrics()
	metrics := []models.Metrics{}
	for _, v := range inMetrics {
		metrics = append(metrics, pbMetricToMetric(v))
	}
	_, err := s.metricsService.SaveMetricsModel(ctx, metrics)
	if err != nil {
		return nil, status.Errorf(codes.Internal, `Error saving metrics: %s`, err)
	}

	return &response, nil
}

func pbMetricToMetric(pbMetric *pb.Metric) models.Metrics {
	var metricType string
	var delta *int64
	var value *float64
	switch pbMetric.GetType() {
	case pb.Metric_GAUGE:
		metricType = models.Gauge
		value = Ptr(float64(pbMetric.GetValue()))
	case pb.Metric_COUNTER:
		metricType = models.Counter
		delta = Ptr(int64(pbMetric.GetDelta()))
	}
	return models.Metrics{
		ID:    pbMetric.GetId(),
		MType: metricType,
		Delta: delta,
		Value: value,
	}
}

type GrpcServer struct {
	config *config.ServerConfig
	server *grpc.Server
}

func NewGrpcServer(config *config.ServerConfig) *GrpcServer {
	return &GrpcServer{
		config: config,
	}
}

func (g *GrpcServer) Start(metricsService *service.MetricsService) {
	if g.config.GrpcPort == "" {
		logger.Sugar.Info("gRPC port not provided")
		return
	}

	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", g.config.GrpcPort))
	if err != nil {
		logger.Sugar.Error("gRPC listener initialization error", "error", err)
		os.Exit(1)
	}
	g.server = grpc.NewServer()

	// Регистрируем сервис
	pb.RegisterMetricsServer(g.server, NewMetricsServer(metricsService))

	logger.Sugar.Infof("gRPC server started on port: %s", g.config.GrpcPort)
	// Получение запроса gRpc
	if err := g.server.Serve(listen); err != nil {
		logger.Sugar.Error("gRPC server work error", "error", err)
		os.Exit(1)
	}
}
