package repository

type MetricsStorage interface {
	SaveCounter(string, int64) (bool, error)
	SaveGauge(string, float64) (bool, error)
}
