package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/handler"
	"github.com/scouser-122/go-metrics/internal/repository"
	"github.com/scouser-122/go-metrics/internal/service"
)

func main() {
	serverConfig := config.GetDefaultServerConfig()

	memStorage := repository.MemStorage{}
	memStorage.FillMetrics(&serverConfig)

	service := service.MetricsService{
		Storage: &memStorage,
	}

	metricsHandler := handler.UpdateHandler{
		Service: service,
	}

	listHandler := handler.ListHandler{
		Service: service,
	}
	listHandler.CreateTemplate()

	// mux := http.NewServeMux()
	// mux.HandleFunc("/update", handler.UpdateHandler)
	// mux.HandleFunc("/update/", handler.UpdateHandler)

	// fmt.Println("Starting server on http://localhost:8080")
	// err := http.ListenAndServe(`:8080`, mux)
	// if err != nil {
	// 	panic(err)
	// }

	r := handler.CreateChiRouter(metricsHandler.UpdateHandler, listHandler.ListHandler)

	fmt.Println("Starting server on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
