package agent

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
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
	runtimeMetrics["TotalAlloc"] = float64(m.TotalAlloc)
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

	fmt.Printf("Successfully collect %d metrics, poll count: %d\n", len(runtimeMetrics), agent.pollCount)
}

func (agent *RuntimeMetricsAgent) SendMetrics() {
	var client = resty.New()

	respChan := make(chan []models.Metrics)
	agent.reads <- ReadRequest{resp: respChan}
	metrics := <-respChan

	for _, metric := range metrics {
		metricValue, err := agent.SendMetric(client, &metric)
		if err != nil {
			fmt.Printf("Error sending metric %q: %s\n", metric.ID, err)
			continue
		} else {
			fmt.Printf("Metric %q sent successfully, value: %s\n", metric.ID, metricValue)
		}
	}
	fmt.Printf("Successfully sent %d metrics\n", len(metrics))

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
	if metric.MType == models.Counter {
		fmt.Printf("Counter %s, value: %s, url: %s\n", metric.ID, metricValue, url)
	}
	if err != nil {
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("incorrect response status code: %q", resp.StatusCode())
	}
	return string(resp.Body()), nil
}

func (agent *RuntimeMetricsAgent) CollectAndSendMetricsInLooop() {
	fmt.Println("Start collecting metrics")
	agent.Init()

	var wg sync.WaitGroup

	wg.Add(1)
	go agent.CollectMetricsWorker(&wg)

	wg.Add(1)
	go agent.SendMetricsWorker(&wg)

	wg.Wait()

	fmt.Println("Finish collecting metrics")
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
