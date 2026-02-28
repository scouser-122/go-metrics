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

var runtimeMetricNames = []string{
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

var runtimeMetics = RuntimeMetircs{
	Metrics: make(map[string]*models.Metrics),
}

var sortedKeys []string

var client = http.Client{
	Timeout: time.Second * 10,
}

const serverAddress = "http://localhost:8080"

const pollInterval = 2

func FillMetricsModel() {
	for _, v := range runtimeMetricNames {
		runtimeMetics.Metrics[v] = &models.Metrics{
			ID:    v,
			MType: models.Gauge,
			Value: new(float64),
		}
	}
	runtimeMetics.Metrics["PollCount"] = &models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: new(int64),
	}
	runtimeMetics.Metrics["RandomValue"] = &models.Metrics{
		ID:    "RandomValue",
		MType: models.Gauge,
		Value: new(float64),
	}
	sortedKeys = make([]string, 0, len(runtimeMetics.Metrics))
	for key := range runtimeMetics.Metrics {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)
}

func CollectAndSendMetricsInLooop() {
	fmt.Println("Start collecting metrics")

	ticker := time.NewTicker(pollInterval * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		collectMetrics()
		sendMetrics()
	}
}

func collectMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	*runtimeMetics.Metrics["Alloc"].Value = float64(m.Alloc)
	*runtimeMetics.Metrics["TotalAlloc"].Value = float64(m.TotalAlloc)
	*runtimeMetics.Metrics["BuckHashSys"].Value = float64(m.BuckHashSys)
	*runtimeMetics.Metrics["Frees"].Value = float64(m.Frees)
	*runtimeMetics.Metrics["GCCPUFraction"].Value = float64(m.GCCPUFraction)
	*runtimeMetics.Metrics["GCSys"].Value = float64(m.GCSys)
	*runtimeMetics.Metrics["HeapAlloc"].Value = float64(m.HeapAlloc)
	*runtimeMetics.Metrics["HeapIdle"].Value = float64(m.HeapIdle)
	*runtimeMetics.Metrics["HeapInuse"].Value = float64(m.HeapInuse)
	*runtimeMetics.Metrics["HeapObjects"].Value = float64(m.HeapObjects)
	*runtimeMetics.Metrics["HeapReleased"].Value = float64(m.HeapReleased)
	*runtimeMetics.Metrics["HeapSys"].Value = float64(m.HeapSys)
	*runtimeMetics.Metrics["LastGC"].Value = float64(m.LastGC)
	*runtimeMetics.Metrics["Lookups"].Value = float64(m.Lookups)
	*runtimeMetics.Metrics["MCacheInuse"].Value = float64(m.MCacheInuse)
	*runtimeMetics.Metrics["MCacheSys"].Value = float64(m.MCacheSys)
	*runtimeMetics.Metrics["MSpanInuse"].Value = float64(m.MSpanInuse)
	*runtimeMetics.Metrics["MSpanSys"].Value = float64(m.MSpanSys)
	*runtimeMetics.Metrics["Mallocs"].Value = float64(m.Mallocs)
	*runtimeMetics.Metrics["NextGC"].Value = float64(m.NextGC)
	*runtimeMetics.Metrics["NumForcedGC"].Value = float64(m.NumForcedGC)
	*runtimeMetics.Metrics["NumGC"].Value = float64(m.NumGC)
	*runtimeMetics.Metrics["OtherSys"].Value = float64(m.OtherSys)
	*runtimeMetics.Metrics["PauseTotalNs"].Value = float64(m.PauseTotalNs)
	*runtimeMetics.Metrics["StackInuse"].Value = float64(m.StackInuse)
	*runtimeMetics.Metrics["StackSys"].Value = float64(m.StackSys)
	*runtimeMetics.Metrics["Sys"].Value = float64(m.Sys)
	*runtimeMetics.Metrics["TotalAlloc"].Value = float64(m.TotalAlloc)
	*runtimeMetics.Metrics["PollCount"].Delta++
	*runtimeMetics.Metrics["RandomValue"].Value = rand.Float64()
}

func sendMetrics() {
	for _, key := range sortedKeys {
		metric := runtimeMetics.Metrics[key]

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
		fmt.Printf("url: %s\n", url)
		request, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			fmt.Printf("Error creating request for metric %q: %s\n", metric.ID, err)
			continue
		}
		request.Header.Set("Content-Type", "text/plain")
		response, err := client.Do(request)
		if err != nil {
			fmt.Printf("Error sending metric %q: %s\n", metric.ID, err)
			continue
		}
		response.Body.Close()
		if response.StatusCode != http.StatusOK {
			fmt.Printf("Error sending metric %q: status %d\n", metric.ID, response.StatusCode)
			continue
		}
		fmt.Printf("Metric %q sent successfully\n", metric.ID)
	}
}
