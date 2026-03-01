package repository

type MetricsStorage interface {
	SaveCounter(string, int64) (int64, error)
	SaveGauge(string, float64) (float64, error)
}
