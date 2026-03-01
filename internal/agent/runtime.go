package agent

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"sort"
	"time"

	models "github.com/scouser-122/go-metrics/internal/model"
)

type RuntimeMetricsAgent struct {
	Config         AgentConfig
	runtimeMetrics RuntimeMetircs
}

func (agent *RuntimeMetricsAgent) FillMetricsModel() {
	agent.runtimeMetrics = RuntimeMetircs{
		Metrics: make(map[string]*models.Metrics),
	}
	for _, v := range agent.Config.runtimeMetricNames {
		agent.runtimeMetrics.Metrics[v] = &models.Metrics{
			ID:    v,
			MType: models.Gauge,
			Value: new(float64),
		}
	}
	agent.runtimeMetrics.Metrics["PollCount"] = &models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: new(int64),
	}
	agent.runtimeMetrics.Metrics["RandomValue"] = &models.Metrics{
		ID:    "RandomValue",
		MType: models.Gauge,
		Value: new(float64),
	}
	agent.runtimeMetrics.SortedKeys = make([]string, 0, len(agent.runtimeMetrics.Metrics))
	for key := range agent.runtimeMetrics.Metrics {
		agent.runtimeMetrics.SortedKeys = append(agent.runtimeMetrics.SortedKeys, key)
	}
	sort.Strings(agent.runtimeMetrics.SortedKeys)
}

func (agent *RuntimeMetricsAgent) CollectMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	*agent.runtimeMetrics.Metrics["Alloc"].Value = float64(m.Alloc)
	*agent.runtimeMetrics.Metrics["TotalAlloc"].Value = float64(m.TotalAlloc)
	*agent.runtimeMetrics.Metrics["BuckHashSys"].Value = float64(m.BuckHashSys)
	*agent.runtimeMetrics.Metrics["Frees"].Value = float64(m.Frees)
	*agent.runtimeMetrics.Metrics["GCCPUFraction"].Value = float64(m.GCCPUFraction)
	*agent.runtimeMetrics.Metrics["GCSys"].Value = float64(m.GCSys)
	*agent.runtimeMetrics.Metrics["HeapAlloc"].Value = float64(m.HeapAlloc)
	*agent.runtimeMetrics.Metrics["HeapIdle"].Value = float64(m.HeapIdle)
	*agent.runtimeMetrics.Metrics["HeapInuse"].Value = float64(m.HeapInuse)
	*agent.runtimeMetrics.Metrics["HeapObjects"].Value = float64(m.HeapObjects)
	*agent.runtimeMetrics.Metrics["HeapReleased"].Value = float64(m.HeapReleased)
	*agent.runtimeMetrics.Metrics["HeapSys"].Value = float64(m.HeapSys)
	*agent.runtimeMetrics.Metrics["LastGC"].Value = float64(m.LastGC)
	*agent.runtimeMetrics.Metrics["Lookups"].Value = float64(m.Lookups)
	*agent.runtimeMetrics.Metrics["MCacheInuse"].Value = float64(m.MCacheInuse)
	*agent.runtimeMetrics.Metrics["MCacheSys"].Value = float64(m.MCacheSys)
	*agent.runtimeMetrics.Metrics["MSpanInuse"].Value = float64(m.MSpanInuse)
	*agent.runtimeMetrics.Metrics["MSpanSys"].Value = float64(m.MSpanSys)
	*agent.runtimeMetrics.Metrics["Mallocs"].Value = float64(m.Mallocs)
	*agent.runtimeMetrics.Metrics["NextGC"].Value = float64(m.NextGC)
	*agent.runtimeMetrics.Metrics["NumForcedGC"].Value = float64(m.NumForcedGC)
	*agent.runtimeMetrics.Metrics["NumGC"].Value = float64(m.NumGC)
	*agent.runtimeMetrics.Metrics["OtherSys"].Value = float64(m.OtherSys)
	*agent.runtimeMetrics.Metrics["PauseTotalNs"].Value = float64(m.PauseTotalNs)
	*agent.runtimeMetrics.Metrics["StackInuse"].Value = float64(m.StackInuse)
	*agent.runtimeMetrics.Metrics["StackSys"].Value = float64(m.StackSys)
	*agent.runtimeMetrics.Metrics["Sys"].Value = float64(m.Sys)
	*agent.runtimeMetrics.Metrics["TotalAlloc"].Value = float64(m.TotalAlloc)
	*agent.runtimeMetrics.Metrics["PollCount"].Delta++
	*agent.runtimeMetrics.Metrics["RandomValue"].Value = rand.Float64()
}

func (agent *RuntimeMetricsAgent) SendMetrics() {
	var client = http.Client{
		Timeout: time.Second * 10,
	}
	for _, key := range agent.runtimeMetrics.SortedKeys {
		metric := agent.runtimeMetrics.Metrics[key]
		_, err := agent.SendMetric(&client, metric)
		if err != nil {
			fmt.Printf("Error sending metric %q: %s\n", metric.ID, err)
			continue
		} else {
			fmt.Printf("Metric %q sent successfully\n", metric.ID)
		}
	}
}

func (agent *RuntimeMetricsAgent) SendMetric(client *http.Client, metric *models.Metrics) (bool, error) {
	var metricValue = ""
	switch metric.MType {
	case models.Counter:
		metricValue = fmt.Sprintf("%d", *metric.Delta)
	case models.Gauge:
		metricValue = fmt.Sprintf("%f", *metric.Value)
	}

	var url = fmt.Sprintf(
		"%s/update/%s/%s/%s",
		agent.Config.serverAddress,
		metric.MType,
		metric.ID,
		metricValue,
	)
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return false, err
	}
	request.Header.Set("Content-Type", "text/plain")
	response, err := client.Do(request)
	if err != nil {
		return false, err
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf("incorrect response status code: %q", response.StatusCode)
	}
	return true, nil
}

func (agent *RuntimeMetricsAgent) CollectAndSendMetricsInLooop() {
	fmt.Println("Start collecting metrics")
	agent.FillMetricsModel()

	ticker := time.NewTicker(agent.Config.pollInterval)
	defer ticker.Stop()

	for range ticker.C {
		agent.CollectMetrics()
		agent.SendMetrics()
	}
}
