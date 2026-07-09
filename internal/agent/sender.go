package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
)

// MetricsSender handles sending collected metrics to the metrics server.
type MetricsSender struct {
	Config *AgentConfig
	writes chan WriteRequest
	reads  chan ReadRequest
	pubKey *rsa.PublicKey
}

// NewSender creates a new MetricsSender instance with the provided configuration.
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

// SendMetricsTikerWorker runs a worker that sends metrics at regular intervals using a ticker.
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

// SendMetricsContinuousWorker runs workers that continuously send metrics as they are collected.
// It respects the rate limit and report interval configurations.
func (sender *MetricsSender) SendMetricsContinuousWorker(
	wg *sync.WaitGroup,
	dataCh <-chan CollectedData,
	stopCh chan struct{},
) {
	interval := time.Duration(sender.Config.ReportInterval) * time.Second
	for w := 1; w <= sender.Config.RequestRateLimit; w++ {
		wg.Add(1)
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
				select {
				case <-stopCh:
					logger.Sugar.Infof("stop sending metrics in worker %d", w)
					wg.Done()
					return
				default:
				}
			}
		}()
	}
}

// SendMetrics sends a batch of metrics to the server with retry logic.
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

// SendMetric sends a single metric using the plain text format via URL path parameters.
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

// SendMetricJSON sends a single metric using JSON format with gzip compression and optional HMAC signing.
func (sender *MetricsSender) SendMetricJSON(client *resty.Client, metric *models.Metrics) (string, error) {
	var url = fmt.Sprintf("%s/update", sender.Config.ServerAddress)
	var savedMetric models.Metrics
	jsonData, err := json.Marshal(*metric)
	if err != nil {
		return "", err
	}
	bodyEncrypted := false
	if sender.pubKey != nil {
		jsonData, err = sender.encryptBytes(jsonData)
		if err != nil {
			return "", err
		}
		bodyEncrypted = true
	}
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	if _, err = gzw.Write(jsonData); err != nil {
		return "", err
	}
	if err = gzw.Close(); err != nil {
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
	if bodyEncrypted {
		request = request.SetHeader("X-Body-Encrypted", "true")
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

// SendMetricsJSON sends multiple metrics using JSON format with gzip compression and optional HMAC signing.
func (sender *MetricsSender) SendMetricsJSON(client *resty.Client, metrics []models.Metrics) (string, error) {
	var url = fmt.Sprintf("%s/updates", sender.Config.ServerAddress)
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return "", err
	}
	bodyEncrypted := false
	if sender.pubKey != nil {
		jsonData, err = sender.encryptBytes(jsonData)
		if err != nil {
			return "", err
		}
		bodyEncrypted = true
	}
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	if _, err = gzw.Write(jsonData); err != nil {
		return "", err
	}
	if err = gzw.Close(); err != nil {
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
	if bodyEncrypted {
		request = request.SetHeader("X-Body-Encrypted", "true")
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

// LoadPublicKeyIfExists loads public key if it exists in FS,
// if loaded - will be used to encode requests to server
func (sender *MetricsSender) LoadPublicKeyIfExists() {
	if sender.Config.CryptoKey == "" {
		logger.Sugar.Info("crypto key path not specified")
		return
	}
	lastSlash := strings.LastIndex(sender.Config.CryptoKey, "/")
	if lastSlash < 0 {
		logger.Sugar.Errorf("crypto key path not correct: %s", sender.Config.CryptoKey)
		return
	}
	dirPath := sender.Config.CryptoKey[:lastSlash]

	root, err := os.OpenRoot(dirPath)
	if err != nil {
		logger.Sugar.Error(err)
		return
	}
	defer root.Close()

	fileName := sender.Config.CryptoKey[lastSlash+1:]
	file, err := root.Open(fileName)
	if err != nil {
		logger.Sugar.Error(err)
		return
	}
	defer file.Close()

	publicKeyBytes, err := io.ReadAll(file)
	if err != nil {
		logger.Sugar.Error(err)
		return
	}

	block, _ := pem.Decode(publicKeyBytes)
	if block == nil {
		logger.Sugar.Errorf("failed to decode public key PEM block")
		return
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		logger.Sugar.Errorf("failed to parse public key: %w", err)
		return
	}

	// Type assert to RSA public key
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		logger.Sugar.Errorf("not an RSA public key")
		return
	}

	sender.pubKey = rsaPub
	logger.Sugar.Info("successfully loaded public key")
}

func (sender *MetricsSender) encryptBytes(message []byte) ([]byte, error) {
	// Calculate maximum message size per chunk
	// For RSA OAEP with SHA-256: keySize/8 - 2*hashSize - 2
	hash := sha256.New()
	maxChunkSize := sender.pubKey.Size() - 2*hash.Size() - 2

	// If the message is small enough, encrypt directly
	if len(message) <= maxChunkSize {
		return rsa.EncryptOAEP(sha256.New(), rand.Reader, sender.pubKey, message, nil)
	}

	// Split the message into chunks
	var encryptedData []byte
	for start := 0; start < len(message); start += maxChunkSize {
		end := start + maxChunkSize
		if end > len(message) {
			end = len(message)
		}

		chunk := message[start:end]
		encryptedChunk, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, sender.pubKey, chunk, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt chunk: %w", err)
		}

		// Append the encrypted chunk to the result
		encryptedData = append(encryptedData, encryptedChunk...)
	}

	return encryptedData, nil
}
