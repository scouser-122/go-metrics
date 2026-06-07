package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/botchris/go-pubsub"
	"github.com/go-resty/resty/v2"
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
)

// AuditService service to log metrics events
type AuditService struct {
	serverConfig *config.ServerConfig
}

// NewAuditService creates AuditService instance
func NewAuditService(serverConfig *config.ServerConfig) AuditService {
	return AuditService{
		serverConfig: serverConfig,
	}
}

// SubscribeToMetricEvents subscribes to metric receieve events, in case of error - just log
func (service *AuditService) SubscribeToMetricEvents(eventBroker pubsub.Broker, ctx context.Context) {
	if service.serverConfig.AuditFile != "" {
		handler := pubsub.NewHandler(func(ctx context.Context, topic pubsub.Topic, event models.MetricEvent) error {
			err := service.LogMetricEventToFile(event)
			if err != nil {
				logger.Sugar.Errorf("metrics event save to file error: %w", err)
			}
			return nil
		})
		sub, err := eventBroker.Subscribe(ctx, models.MetricEventTopic, handler)
		if err != nil {
			logger.Sugar.Errorf("subscribe to metric events: %w", err)
		} else {
			go func() {
				<-ctx.Done()
				if err := sub.Unsubscribe(); err != nil {
					logger.Sugar.Errorf("error unsubscribing: %v", err)
				}
			}()
			logger.Sugar.Info("file-system audit start listen to metrics event")
		}
	}
	if service.serverConfig.AuditURL != "" {
		handler := pubsub.NewHandler(func(ctx context.Context, topic pubsub.Topic, event models.MetricEvent) error {
			err := service.LogMetricEventToRemoteService(event)
			if err != nil {
				logger.Sugar.Errorf("metrics event send to remote service error: %w", err)
			}
			return nil
		})
		sub, err := eventBroker.Subscribe(ctx, models.MetricEventTopic, handler)
		if err != nil {
			logger.Sugar.Errorf("subscribe to metric events: %w", err)
		} else {
			go func() {
				<-ctx.Done()
				if err := sub.Unsubscribe(); err != nil {
					logger.Sugar.Errorf("error unsubscribing: %v", err)
				}
			}()
			logger.Sugar.Info("remote service audit start listen to metrics event")
		}
	}
}

// LogMetricEventToFile writes audit event to local file
func (service *AuditService) LogMetricEventToFile(event models.MetricEvent) error {
	file, err := os.OpenFile(service.serverConfig.AuditFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(event); err != nil {
		return err
	}
	logger.Sugar.Infof("metrics event successfully saved in file")
	return nil
}

// LogMetricEventToRemoteService sends audit event to remote service
func (service *AuditService) LogMetricEventToRemoteService(event models.MetricEvent) error {
	var client = resty.New()
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(event)
	if err != nil {
		return err
	}
	request := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(buf.Bytes())
	resp, err := request.Post(service.serverConfig.AuditURL)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("incorrect response status code: %d", resp.StatusCode())
	}
	logger.Sugar.Infof("metrics event successfully sent to remote service")
	return nil
}
