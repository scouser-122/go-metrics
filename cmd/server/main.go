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

	fmt.Println("Starting server on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
