package agent

import (
	"sync"

	"github.com/scouser-122/go-metrics/internal/logger"
)

type MetricsAgent struct {
	Config      *AgentConfig
	dataChannel chan CollectedData
	collector   MetricsCollector
	sender      MetricsSender
}

func NewAgent(config *AgentConfig) *MetricsAgent {
	agent := MetricsAgent{
		Config:    config,
		collector: NewCollector(config),
		sender:    NewSender(config),
	}
	return &agent
}

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
