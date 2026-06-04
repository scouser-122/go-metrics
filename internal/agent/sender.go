package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
)

type MetricsSender struct {
	Config *AgentConfig
	writes chan WriteRequest
	reads  chan ReadRequest
}

func NewSender(config *AgentConfig) MetricsSender {
	return MetricsSender{
		Config: config,
		writes: make(chan WriteRequest),
		reads:  make(chan ReadRequest),
	}
}

func (sender *MetricsSender) monitor(writes <-chan WriteRequest, reads <-chan ReadRequest) {
	slice := []models.Metrics{}

	for {
		select {
		case req := <-writes:
			slice = append(slice, req.data)
		case req := <-reads:
			sliceCopy := make([]models.Metrics, len(slice))
			copy(sliceCopy, slice)
			slice = make([]models.Metrics, 0)
			req.resp <- sliceCopy
		}
	}
}

func (sender *MetricsSender) writer(dataCh <-chan CollectedData) {
	for data := range dataCh {
		for _, m := range data.metrics {
			sender.writes <- WriteRequest{data: m}
		}
	}
}

func (sender *MetricsSender) SendMetricsTikerWorker(wg *sync.WaitGroup, dataCh <-chan CollectedData) {
	defer wg.Done()

	go sender.monitor(sender.writes, sender.reads)
	go sender.writer(dataCh)

	ticker := time.NewTicker(time.Duration(sender.Config.ReportInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		logger.Sugar.Info("start sending metrics on ticker")
		respChan := make(chan []models.Metrics)
		sender.reads <- ReadRequest{resp: respChan}
		metrics := <-respChan
		sender.SendMetrics(metrics)
	}
}

func (sender *MetricsSender) SendMetricsContinuousWorker(wg *sync.WaitGroup, dataCh <-chan CollectedData) {
	defer wg.Done()
	interval := time.Duration(sender.Config.ReportInterval) * time.Second
	for w := 1; w <= sender.Config.RequestRateLimit; w++ {
		go func() {
			prevTime := time.Now()
			time.Sleep(interval)
			for data := range dataCh {
				logger.Sugar.Infof("start sending metrics in worker %d", w)
				sender.SendMetrics(data.metrics)
				if len(dataCh) == 0 {
					timeDiff := time.Since(prevTime)
					if timeDiff < interval {
						time.Sleep(interval - timeDiff)
					}
				}
				prevTime = time.Now()
			}
		}()
	}
}

func (sender *MetricsSender) SendMetrics(metrics []models.Metrics) {
	var client = resty.New()

	err := config.AgentRetry(
		sender.Config.RetryConfig,
		func() error {
			response, err := sender.SendMetricsJSON(client, metrics)
			if err != nil {
				return err
			} else {
				pollCount := int64(0)
				for _, m := range metrics {
					if m.ID == "PollCount" && *m.Delta > int64(pollCount) {
						pollCount = *m.Delta
					}
				}
				logger.Sugar.Infof("%s, poll count: %d", response, pollCount)
			}
			return nil
		},
	)
	if err != nil {
		logger.Sugar.Errorf("error sending metrics: %s", err)
	}
}

func (sender *MetricsSender) SendMetric(client *resty.Client, metric *models.Metrics) (string, error) {
	var metricValue = ""
	switch metric.MType {
	case models.Counter:
		metricValue = fmt.Sprintf("%d", *metric.Delta)
	case models.Gauge:
		metricValue = fmt.Sprintf("%f", *metric.Value)
	}

	var url = fmt.Sprintf(
		"%s/update/%s/%s/%s",
		sender.Config.ServerAddress,
		metric.MType,
		metric.ID,
		metricValue,
	)
	resp, err := client.R().SetHeader("Content-Type", "text/plain").Post(url)
	if err != nil {
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("incorrect response status code: %q", resp.StatusCode())
	}
	return string(resp.Body()), nil
}

func (sender *MetricsSender) SendMetricJSON(client *resty.Client, metric *models.Metrics) (string, error) {
	var url = fmt.Sprintf("%s/update", sender.Config.ServerAddress)
	var savedMetric models.Metrics
	jsonData, err := json.Marshal(*metric)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	if _, err := gzw.Write(jsonData); err != nil {
		return "", err
	}
	if err := gzw.Close(); err != nil {
		return "", err
	}
	request := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(&buf).
		SetResult(&savedMetric)
	if sender.Config.HMACKey != "" {
		h := hmac.New(sha256.New, []byte(sender.Config.HMACKey))
		h.Write(buf.Bytes())
		hash := h.Sum(nil)
		request = request.SetHeader("HashSHA256", hex.EncodeToString(hash))
	}
	resp, err := request.Post(url)
	if err != nil {
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("incorrect response status code: %d", resp.StatusCode())
	}
	return savedMetric.GetValueAsString()
}

func (sender *MetricsSender) SendMetricsJSON(client *resty.Client, metrics []models.Metrics) (string, error) {
	var url = fmt.Sprintf("%s/updates", sender.Config.ServerAddress)
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	if _, err := gzw.Write(jsonData); err != nil {
		return "", err
	}
	if err := gzw.Close(); err != nil {
		return "", err
	}
	response := models.ResponsePayload{}
	request := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(&buf).
		SetResult(&response)
	if sender.Config.HMACKey != "" {
		h := hmac.New(sha256.New, []byte(sender.Config.HMACKey))
		h.Write(jsonData)
		hash := h.Sum(nil)
		request = request.SetHeader("HashSHA256", hex.EncodeToString(hash))
	}
	resp, err := request.Post(url)
	if err != nil {
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("incorrect response status code: %d", resp.StatusCode())
	}
	return response.Message, nil
}
