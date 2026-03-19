package handler

import (
	"net/http"

	"github.com/scouser-122/go-metrics/internal/service"
)

type Handler struct {
	Method         string
	URLPathPattern string
	HandlerFn      http.HandlerFunc
}

func InitializeHandlers(service *service.MetricsService) []Handler {
	handlers := []Handler{}

	updateHandler := UpdateHandler{
		Service: service,
	}
	handlers = append(handlers, Handler{
		Method:         http.MethodPost,
		URLPathPattern: "/update/{type}/{name}/{value}",
		HandlerFn:      updateHandler.UpdateHandler,
	})
	handlers = append(handlers, Handler{
		Method:         http.MethodPost,
		URLPathPattern: "/update/",
		HandlerFn:      updateHandler.UpdateJSONHandler,
	})
	handlers = append(handlers, Handler{
		Method:         http.MethodPost,
		URLPathPattern: "/update",
		HandlerFn:      updateHandler.UpdateJSONHandler,
	})

	readHandler := ReadHandler{
		Service: service,
	}
	readHandler.CreateTemplate()
	handlers = append(handlers, Handler{
		Method:         http.MethodGet,
		URLPathPattern: "/",
		HandlerFn:      readHandler.ListHandler,
	})
	handlers = append(handlers, Handler{
		Method:         http.MethodGet,
		URLPathPattern: "/value/{type}/{name}",
		HandlerFn:      readHandler.ValueHandler,
	})
	handlers = append(handlers, Handler{
		Method:         http.MethodPost,
		URLPathPattern: "/value",
		HandlerFn:      readHandler.ValueJSONHandler,
	})
	handlers = append(handlers, Handler{
		Method:         http.MethodPost,
		URLPathPattern: "/value/",
		HandlerFn:      readHandler.ValueJSONHandler,
	})

	return handlers
}
