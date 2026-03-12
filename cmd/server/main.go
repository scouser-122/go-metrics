package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/scouser-122/go-metrics/internal/handler"
	"github.com/scouser-122/go-metrics/internal/repository"
	"github.com/scouser-122/go-metrics/internal/service"
)

func main() {
	memStorage := repository.MemStorage{}

	metricsService := service.MetricsService{
		Storage: &memStorage,
	}

	updateHandler := handler.UpdateHandler{
		Service: metricsService,
	}
	obtainHandler := handler.ReadHandler{
		Service: metricsService,
	}
	obtainHandler.CreateTemplate()

	r := handler.CreateChiRouter(&updateHandler, &obtainHandler)

	parseFlags()
	parseEnvVariables()

	fmt.Printf("Starting server on http://%s\n", flagRunAddr)
	log.Fatal(http.ListenAndServe(flagRunAddr, r))
}
