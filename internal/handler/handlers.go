package handler

import (
	"net/http"

	"github.com/scouser-122/go-metrics/internal/service"
)

type Handler struct {
	Name           string
	Method         string
	UrlPathPattern string
	HandlerFn      http.HandlerFunc
}

func InitializeHandlers(service *service.MetricsService) []Handler {
	handlers := []Handler{}

	updateHandler := UpdateHandler{
		Service: service,
	}
	handlers = append(handlers, Handler{
		Name:           "metric update",
		Method:         http.MethodPost,
		UrlPathPattern: "/update/{type}/{name}/{value}",
		HandlerFn:      updateHandler.UpdateHandler,
	})

	readHandler := ReadHandler{
		Service: service,
	}
	readHandler.CreateTemplate()
	handlers = append(handlers, Handler{
		Name:           "metrics list",
		Method:         http.MethodGet,
		UrlPathPattern: "/",
		HandlerFn:      readHandler.ListHandler,
	})
	handlers = append(handlers, Handler{
		Name:           "metrics list",
		Method:         http.MethodGet,
		UrlPathPattern: "/value/{type}/{name}",
		HandlerFn:      readHandler.GetHandler,
	})

	return handlers
}
