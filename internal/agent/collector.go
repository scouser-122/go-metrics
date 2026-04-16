package agent

import (
	"math/rand/v2"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type MetricsCollector struct {
	Config    *AgentConfig
	pollCount int64
}

func NewCollector(config *AgentConfig) MetricsCollector {
	return MetricsCollector{
		Config: config,
	}
}

func (collector *MetricsCollector) CollectMetricsWorker(wg *sync.WaitGroup, dataCh chan<- CollectedData) {
	defer wg.Done()

	ticker := time.NewTicker(time.Duration(collector.Config.PollInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		go collector.CollectRuntimeMetrics(dataCh)
		go collector.CollectGopsutilMetrics(dataCh)
	}
}

func (collector *MetricsCollector) CollectRuntimeMetrics(dataCh chan<- CollectedData) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	runtimeMetricsMap := make(map[string]float64)
	runtimeMetricsMap["Alloc"] = float64(m.Alloc)
	runtimeMetricsMap["BuckHashSys"] = float64(m.BuckHashSys)
	runtimeMetricsMap["Frees"] = float64(m.Frees)
	runtimeMetricsMap["GCCPUFraction"] = float64(m.GCCPUFraction)
	runtimeMetricsMap["GCSys"] = float64(m.GCSys)
	runtimeMetricsMap["HeapAlloc"] = float64(m.HeapAlloc)
	runtimeMetricsMap["HeapIdle"] = float64(m.HeapIdle)
	runtimeMetricsMap["HeapInuse"] = float64(m.HeapInuse)
	runtimeMetricsMap["HeapObjects"] = float64(m.HeapObjects)
	runtimeMetricsMap["HeapReleased"] = float64(m.HeapReleased)
	runtimeMetricsMap["HeapSys"] = float64(m.HeapSys)
	runtimeMetricsMap["LastGC"] = float64(m.LastGC)
	runtimeMetricsMap["Lookups"] = float64(m.Lookups)
	runtimeMetricsMap["MCacheInuse"] = float64(m.MCacheInuse)
	runtimeMetricsMap["MCacheSys"] = float64(m.MCacheSys)
	runtimeMetricsMap["MSpanInuse"] = float64(m.MSpanInuse)
	runtimeMetricsMap["MSpanSys"] = float64(m.MSpanSys)
	runtimeMetricsMap["Mallocs"] = float64(m.Mallocs)
	runtimeMetricsMap["NextGC"] = float64(m.NextGC)
	runtimeMetricsMap["NumForcedGC"] = float64(m.NumForcedGC)
	runtimeMetricsMap["NumGC"] = float64(m.NumGC)
	runtimeMetricsMap["OtherSys"] = float64(m.OtherSys)
	runtimeMetricsMap["PauseTotalNs"] = float64(m.PauseTotalNs)
	runtimeMetricsMap["StackInuse"] = float64(m.StackInuse)
	runtimeMetricsMap["StackSys"] = float64(m.StackSys)
	runtimeMetricsMap["Sys"] = float64(m.Sys)
	runtimeMetricsMap["TotalAlloc"] = float64(m.TotalAlloc)
	runtimeMetricsMap["RandomValue"] = rand.Float64()

	collector.pollCount++

	keys := make([]string, 0, len(runtimeMetricsMap))
	for key := range runtimeMetricsMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	runtimeMetrics := []models.Metrics{}
	for _, key := range keys {
		var value = runtimeMetricsMap[key]
		runtimeMetrics = append(runtimeMetrics, models.Metrics{
			ID:    key,
			MType: models.Gauge,
			Value: &value,
		})
	}

	pollCountMetric := models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: new(int64),
	}
	*pollCountMetric.Delta = collector.pollCount
	runtimeMetrics = append(runtimeMetrics, pollCountMetric)

	if len(dataCh) < cap(dataCh) {
		dataCh <- CollectedData{runtimeMetrics}
		logger.Sugar.Infof("successfully collect %d runtime metrics, poll count: %d", len(runtimeMetrics), collector.pollCount)
	} else {
		logger.Sugar.Errorf("error collecting runtime metrics, channel is full")
	}
}

func (collector *MetricsCollector) CollectGopsutilMetrics(dataCh chan<- CollectedData) {
	gopsutilMetrics := []models.Metrics{}
	vm, err := mem.VirtualMemory()
	if err != nil {
		logger.Sugar.Errorf("Error getting memory: %v\n", err)
	} else {
		totalMem := float64(vm.Total)
		gopsutilMetrics = append(gopsutilMetrics, models.Metrics{
			ID:    "TotalMemory",
			MType: models.Gauge,
			Value: &totalMem,
		})
		freeMem := float64(vm.Free)
		gopsutilMetrics = append(gopsutilMetrics, models.Metrics{
			ID:    "FreeMemory",
			MType: models.Gauge,
			Value: &freeMem,
		})
	}

	percentage, err := cpu.Percent(time.Second, false)
	if err != nil {
		logger.Sugar.Errorf("Error getting CPU: %v\n", err)
	} else {
		cpuUtilization := percentage[0]
		gopsutilMetrics = append(gopsutilMetrics, models.Metrics{
			ID:    "CPUutilization1",
			MType: models.Gauge,
			Value: &cpuUtilization,
		})
	}

	collector.pollCount++
	pollCountMetric := models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: new(int64),
	}
	*pollCountMetric.Delta = collector.pollCount
	gopsutilMetrics = append(gopsutilMetrics, pollCountMetric)

	if len(dataCh) < cap(dataCh) {
		dataCh <- CollectedData{gopsutilMetrics}
		logger.Sugar.Infof("successfully collect %d gopsutil metrics, poll count: %d", len(gopsutilMetrics), collector.pollCount)
	} else {
		logger.Sugar.Errorf("error collecting gopsutil metrics, channel is full")
	}
}
