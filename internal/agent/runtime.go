package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
)

type WriteRequest struct {
	data models.Metrics
}

type ReadRequest struct {
	resp chan []models.Metrics
}

type RuntimeMetricsAgent struct {
	Config    AgentConfig
	Metrics   []models.Metrics
	pollCount int64
	writes    chan WriteRequest
	reads     chan ReadRequest
}

func (agent *RuntimeMetricsAgent) Init() {
	agent.writes = make(chan WriteRequest)
	agent.reads = make(chan ReadRequest)
	go monitor(agent.writes, agent.reads)
}

func monitor(writes <-chan WriteRequest, reads <-chan ReadRequest) {
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

func (agent *RuntimeMetricsAgent) CollectMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	runtimeMetrics := make(map[string]float64)
	runtimeMetrics["Alloc"] = float64(m.Alloc)
	runtimeMetrics["BuckHashSys"] = float64(m.BuckHashSys)
	runtimeMetrics["Frees"] = float64(m.Frees)
	runtimeMetrics["GCCPUFraction"] = float64(m.GCCPUFraction)
	runtimeMetrics["GCSys"] = float64(m.GCSys)
	runtimeMetrics["HeapAlloc"] = float64(m.HeapAlloc)
	runtimeMetrics["HeapIdle"] = float64(m.HeapIdle)
	runtimeMetrics["HeapInuse"] = float64(m.HeapInuse)
	runtimeMetrics["HeapObjects"] = float64(m.HeapObjects)
	runtimeMetrics["HeapReleased"] = float64(m.HeapReleased)
	runtimeMetrics["HeapSys"] = float64(m.HeapSys)
	runtimeMetrics["LastGC"] = float64(m.LastGC)
	runtimeMetrics["Lookups"] = float64(m.Lookups)
	runtimeMetrics["MCacheInuse"] = float64(m.MCacheInuse)
	runtimeMetrics["MCacheSys"] = float64(m.MCacheSys)
	runtimeMetrics["MSpanInuse"] = float64(m.MSpanInuse)
	runtimeMetrics["MSpanSys"] = float64(m.MSpanSys)
	runtimeMetrics["Mallocs"] = float64(m.Mallocs)
	runtimeMetrics["NextGC"] = float64(m.NextGC)
	runtimeMetrics["NumForcedGC"] = float64(m.NumForcedGC)
	runtimeMetrics["NumGC"] = float64(m.NumGC)
	runtimeMetrics["OtherSys"] = float64(m.OtherSys)
	runtimeMetrics["PauseTotalNs"] = float64(m.PauseTotalNs)
	runtimeMetrics["StackInuse"] = float64(m.StackInuse)
	runtimeMetrics["StackSys"] = float64(m.StackSys)
	runtimeMetrics["Sys"] = float64(m.Sys)
	runtimeMetrics["TotalAlloc"] = float64(m.TotalAlloc)
	runtimeMetrics["RandomValue"] = rand.Float64()

	agent.pollCount++

	keys := make([]string, 0, len(runtimeMetrics))
	for key := range runtimeMetrics {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		var value = runtimeMetrics[key]
		agent.writes <- WriteRequest{data: models.Metrics{
			ID:    key,
			MType: models.Gauge,
			Value: &value,
		}}
	}

	pollCountMetric := models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: new(int64),
	}
	*pollCountMetric.Delta = agent.pollCount
	agent.writes <- WriteRequest{data: pollCountMetric}

	logger.Sugar.Infof("successfully collect %d metrics, poll count: %d", len(runtimeMetrics), agent.pollCount)
}

func (agent *RuntimeMetricsAgent) SendMetrics() {
	var client = resty.New()

	respChan := make(chan []models.Metrics)
	agent.reads <- ReadRequest{resp: respChan}
	metrics := <-respChan

	// successSentCount := 0
	// for _, metric := range metrics {
	// 	metricValue, err := agent.SendMetricJSON(client, &metric)
	// 	if err != nil {
	// 		logger.Sugar.Errorf("error sending metric %q: %s", metric.ID, err)
	// 		continue
	// 	} else {
	// 		logger.Sugar.Infof("metric %q sent successfully, value: %s", metric.ID, metricValue)
	// 		successSentCount++
	// 	}
	// }
	// logger.Sugar.Infof("successfully sent %d metrics", successSentCount)

	logger.Sugar.Info("start sending metrics")
	err := config.AgentRetry(
		agent.Config.RetryConfig,
		func() error {
			response, err := agent.SendMetricsJSON(client, metrics)
			if err != nil {
				return err
			} else {
				logger.Sugar.Info(response)
			}
			return nil
		},
	)
	if err != nil {
		logger.Sugar.Errorf("error sending metrics: %s", err)
	}
}

func (agent *RuntimeMetricsAgent) SendMetric(client *resty.Client, metric *models.Metrics) (string, error) {
	var metricValue = ""
	switch metric.MType {
	case models.Counter:
		metricValue = fmt.Sprintf("%d", *metric.Delta)
	case models.Gauge:
		metricValue = fmt.Sprintf("%f", *metric.Value)
	}

	var url = fmt.Sprintf(
		"%s/update/%s/%s/%s",
		agent.Config.ServerAddress,
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

func (agent *RuntimeMetricsAgent) SendMetricJSON(client *resty.Client, metric *models.Metrics) (string, error) {
	var url = fmt.Sprintf("%s/update", agent.Config.ServerAddress)
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
	if agent.Config.HMACKey != "" {
		h := hmac.New(sha256.New, []byte(agent.Config.HMACKey))
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

func (agent *RuntimeMetricsAgent) SendMetricsJSON(client *resty.Client, metrics []models.Metrics) (string, error) {
	var url = fmt.Sprintf("%s/updates", agent.Config.ServerAddress)
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
	if agent.Config.HMACKey != "" {
		h := hmac.New(sha256.New, []byte(agent.Config.HMACKey))
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

func (agent *RuntimeMetricsAgent) CollectAndSendMetricsInLoop() {
	logger.Sugar.Info("start collecting metrics")
	agent.Init()

	var wg sync.WaitGroup

	wg.Add(1)
	go agent.CollectMetricsWorker(&wg)

	wg.Add(1)
	go agent.SendMetricsWorker(&wg)

	wg.Wait()

	logger.Sugar.Info("finish collecting metrics")
}

func (agent *RuntimeMetricsAgent) CollectMetricsWorker(wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(time.Duration(agent.Config.PollInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		agent.CollectMetrics()
	}
}

func (agent *RuntimeMetricsAgent) SendMetricsWorker(wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(time.Duration(agent.Config.ReportInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		agent.SendMetrics()
	}
}
