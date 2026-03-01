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

var agentRuntimeMetricNames = []string{
	"Alloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
	"TotalAlloc",
}

const serverAddress = "http://localhost:8080"

const pollInterval = 2

func CollectAndSendMetricsInLooop() {
	fmt.Println("Start collecting metrics")
	var runtimeMetrics = RuntimeMetircs{
		Metrics: make(map[string]*models.Metrics),
	}

	FillMetricsModel(&runtimeMetrics, agentRuntimeMetricNames)

	ticker := time.NewTicker(pollInterval * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		CollectMetrics(&runtimeMetrics)
		SendMetrics(&runtimeMetrics)
	}
}

func FillMetricsModel(runtimeMetrics *RuntimeMetircs, runtimeMetricNames []string) {
	for _, v := range runtimeMetricNames {
		runtimeMetrics.Metrics[v] = &models.Metrics{
			ID:    v,
			MType: models.Gauge,
			Value: new(float64),
		}
	}
	runtimeMetrics.Metrics["PollCount"] = &models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: new(int64),
	}
	runtimeMetrics.Metrics["RandomValue"] = &models.Metrics{
		ID:    "RandomValue",
		MType: models.Gauge,
		Value: new(float64),
	}
	runtimeMetrics.SortedKeys = make([]string, 0, len(runtimeMetrics.Metrics))
	for key := range runtimeMetrics.Metrics {
		runtimeMetrics.SortedKeys = append(runtimeMetrics.SortedKeys, key)
	}
	sort.Strings(runtimeMetrics.SortedKeys)
}

func CollectMetrics(runtimeMetrics *RuntimeMetircs) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	*runtimeMetrics.Metrics["Alloc"].Value = float64(m.Alloc)
	*runtimeMetrics.Metrics["TotalAlloc"].Value = float64(m.TotalAlloc)
	*runtimeMetrics.Metrics["BuckHashSys"].Value = float64(m.BuckHashSys)
	*runtimeMetrics.Metrics["Frees"].Value = float64(m.Frees)
	*runtimeMetrics.Metrics["GCCPUFraction"].Value = float64(m.GCCPUFraction)
	*runtimeMetrics.Metrics["GCSys"].Value = float64(m.GCSys)
	*runtimeMetrics.Metrics["HeapAlloc"].Value = float64(m.HeapAlloc)
	*runtimeMetrics.Metrics["HeapIdle"].Value = float64(m.HeapIdle)
	*runtimeMetrics.Metrics["HeapInuse"].Value = float64(m.HeapInuse)
	*runtimeMetrics.Metrics["HeapObjects"].Value = float64(m.HeapObjects)
	*runtimeMetrics.Metrics["HeapReleased"].Value = float64(m.HeapReleased)
	*runtimeMetrics.Metrics["HeapSys"].Value = float64(m.HeapSys)
	*runtimeMetrics.Metrics["LastGC"].Value = float64(m.LastGC)
	*runtimeMetrics.Metrics["Lookups"].Value = float64(m.Lookups)
	*runtimeMetrics.Metrics["MCacheInuse"].Value = float64(m.MCacheInuse)
	*runtimeMetrics.Metrics["MCacheSys"].Value = float64(m.MCacheSys)
	*runtimeMetrics.Metrics["MSpanInuse"].Value = float64(m.MSpanInuse)
	*runtimeMetrics.Metrics["MSpanSys"].Value = float64(m.MSpanSys)
	*runtimeMetrics.Metrics["Mallocs"].Value = float64(m.Mallocs)
	*runtimeMetrics.Metrics["NextGC"].Value = float64(m.NextGC)
	*runtimeMetrics.Metrics["NumForcedGC"].Value = float64(m.NumForcedGC)
	*runtimeMetrics.Metrics["NumGC"].Value = float64(m.NumGC)
	*runtimeMetrics.Metrics["OtherSys"].Value = float64(m.OtherSys)
	*runtimeMetrics.Metrics["PauseTotalNs"].Value = float64(m.PauseTotalNs)
	*runtimeMetrics.Metrics["StackInuse"].Value = float64(m.StackInuse)
	*runtimeMetrics.Metrics["StackSys"].Value = float64(m.StackSys)
	*runtimeMetrics.Metrics["Sys"].Value = float64(m.Sys)
	*runtimeMetrics.Metrics["TotalAlloc"].Value = float64(m.TotalAlloc)
	*runtimeMetrics.Metrics["PollCount"].Delta++
	*runtimeMetrics.Metrics["RandomValue"].Value = rand.Float64()
}

func SendMetrics(runtimeMetrics *RuntimeMetircs) {
	var client = http.Client{
		Timeout: time.Second * 10,
	}
	for _, key := range runtimeMetrics.SortedKeys {
		metric := runtimeMetrics.Metrics[key]
		_, err := SendMetric(&client, metric)
		if err != nil {
			fmt.Printf("Error sending metric %q: %s\n", metric.ID, err)
			continue
		} else {
			fmt.Printf("Metric %q sent successfully\n", metric.ID)
		}
	}
}

func SendMetric(client *http.Client, metric *models.Metrics) (bool, error) {
	var metricValue = ""
	switch metric.MType {
	case models.Counter:
		metricValue = fmt.Sprintf("%d", *metric.Delta)
	case models.Gauge:
		metricValue = fmt.Sprintf("%f", *metric.Value)
	}

	var url = fmt.Sprintf(
		"%s/update/%s/%s/%s",
		serverAddress,
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
