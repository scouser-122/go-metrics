package agent

type AgentConfig struct {
	runtimeMetricNames []string
	ServerAddress      string
	ReportInterval     int
	PollInterval       int
}

func GetDefaultAgentConfig() AgentConfig {
	config := AgentConfig{}
	config.runtimeMetricNames = []string{
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
	config.ServerAddress = "http://localhost:8080"
	config.PollInterval = 2
	config.ReportInterval = 10
	return config
}
