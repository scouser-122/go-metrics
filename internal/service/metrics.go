package service

import (
	"github.com/scouser-122/go-metrics/internal/repository"
)

func SaveCounter(name string, value int64) {
	repository.MemoryStorage.SaveCounter(name, value)
	repository.MemoryStorage.Print()
}

func SaveGauge(name string, value float64) {
	repository.MemoryStorage.SaveGauge(name, value)
	repository.MemoryStorage.Print()
}
