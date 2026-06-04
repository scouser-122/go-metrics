package models

// MetricsReceivedEvent событие получения метрик
type MetricsReceivedEvent struct {
	// Ts unix timestamp события
	Ts int64 `json:"ts"`

	// Metrics наименование полученных метрик
	Metrics []string `json:"metrics"`

	// IpAddress IP адрес входящего запроса
	IpAddress string `json:"ip_address"`
}
