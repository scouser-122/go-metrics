package agent

import (
	"sync"

	"github.com/scouser-122/go-metrics/internal/logger"
)

// MetricsAgent orchestrates metrics collection and sending operations.
type MetricsAgent struct {
	Config      *AgentConfig
	dataChannel chan CollectedData
	collector   MetricsCollector
	sender      MetricsSender
}

// NewAgent creates a new MetricsAgent instance with the provided configuration.
func NewAgent(config *AgentConfig) *MetricsAgent {
	agent := MetricsAgent{
		Config:    config,
		collector: NewCollector(config),
		sender:    NewSender(config),
	}
	agent.sender.LoadPublicKeyIfExists()
	return &agent
}

// CollectAndSendMetricsInLoop starts concurrent workers for collecting and sending metrics.
// This method blocks until all workers complete.
func (agent *MetricsAgent) CollectAndSendMetricsInLoop() {
	logger.Sugar.Info("start collecting metrics")

	dataChannel := make(chan CollectedData, agent.Config.CollectChannelSize)

	var wg sync.WaitGroup

	wg.Add(1)
	go agent.sender.SendMetricsContinuousWorker(&wg, dataChannel)

	wg.Add(1)
	go agent.collector.CollectMetricsWorker(&wg, dataChannel)

	wg.Wait()

	close(dataChannel)

	logger.Sugar.Info("finish collecting metrics")
}
