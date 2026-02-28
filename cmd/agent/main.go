package main

import runtime "github.com/scouser-122/go-metrics/internal/agent"

func main() {
	runtime.FillMetricsModel()
	runtime.CollectAndSendMetricsInLooop()
}
