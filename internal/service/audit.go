package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/botchris/go-pubsub"
	"github.com/go-resty/resty/v2"
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
)

// AuditService service to log metrics events.
type AuditService struct {
	serverConfig *config.ServerConfig
}

// NewAuditService creates AuditService instance.
func NewAuditService(serverConfig *config.ServerConfig) AuditService {
	return AuditService{
		serverConfig: serverConfig,
	}
}

// SubscribeToMetricEvents subscribes to metric receieve events, in case of error - just log.
func (service *AuditService) SubscribeToMetricEvents(eventBroker pubsub.Broker, ctx context.Context) {
	if service.serverConfig.AuditFile != "" {
		fileSaveEvents := make(chan models.MetricsReceivedEvent, 50)
		handler := pubsub.NewHandler(func(ctx context.Context, topic pubsub.Topic, event models.MetricsReceivedEvent) error {
			fileSaveEvents <- event
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
			go logMetricEventsToFile(fileSaveEvents, service.serverConfig.AuditFile)
		}
	}
	if service.serverConfig.AuditURL != "" {
		serviceSendEvents := make(chan models.MetricsReceivedEvent, 50)
		handler := pubsub.NewHandler(func(ctx context.Context, topic pubsub.Topic, event models.MetricsReceivedEvent) error {
			serviceSendEvents <- event
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
			go logMetricEventToRemoteService(serviceSendEvents, service.serverConfig.AuditURL)
		}
	}
}

func logMetricEventsToFile(events chan models.MetricsReceivedEvent, auditFilePath string) {
	for event := range events {
		file, err := os.OpenFile(auditFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			logger.Sugar.Errorf("metrics event save to file error: %w", err)
			continue
		}
		defer file.Close()
		encoder := json.NewEncoder(file)
		if err := encoder.Encode(event); err != nil {
			logger.Sugar.Errorf("metrics event save to file error: %w", err)
			continue
		}
		timestamp := time.UnixMilli(event.TS).Format("2006-01-02 15:04:05.000")
		logger.Sugar.Infof("metrics event with TS %s successfully saved in file", timestamp)
	}
}

func logMetricEventToRemoteService(events chan models.MetricsReceivedEvent, auditURL string) {
	for event := range events {
		var client = resty.New()
		buf := new(bytes.Buffer)
		encoder := json.NewEncoder(buf)
		err := encoder.Encode(event)
		if err != nil {
			logger.Sugar.Errorf("metrics event send to remote service error: %w", err)
			continue
		}
		request := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(buf.Bytes())
		resp, err := request.Post(auditURL)
		if err != nil {
			logger.Sugar.Errorf("metrics event send to remote service error: %w", err)
			continue
		}
		if resp.StatusCode() != http.StatusOK {
			err := fmt.Errorf("incorrect response status code: %d", resp.StatusCode())
			logger.Sugar.Errorf("metrics event send to remote service error: %w", err)
			continue
		}
		timestamp := time.UnixMilli(event.TS).Format("2006-01-02 15:04:05.000")
		logger.Sugar.Infof("metrics event with TS %s successfully sent to remote service", timestamp)
	}

}
