package handler

import (
	"net/http"

	"github.com/scouser-122/go-metrics/internal/repository/db"
	"github.com/scouser-122/go-metrics/internal/service"
)

// Handler represents an HTTP handler configuration with method, path pattern, and handler function.
type Handler struct {
	Method         string
	URLPathPattern string
	HandlerFn      http.HandlerFunc
}

// InitializeHandlers creates and returns a slice of all HTTP handlers for the metrics service.
// It configures both update and read handlers with their respective routes.
func InitializeHandlers(
	metricsService *service.MetricsService,
	cryptoService *service.CryptoService,
	db *db.PostgresDatabase,
) []Handler {
	handlers := []Handler{}

	updateHandler := UpdateHandler{
		MetricsService: metricsService,
		cryptoService:  cryptoService,
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
	handlers = append(handlers, Handler{
		Method:         http.MethodPost,
		URLPathPattern: "/updates/",
		HandlerFn:      updateHandler.UpdateJSONArrayHandler,
	})
	handlers = append(handlers, Handler{
		Method:         http.MethodPost,
		URLPathPattern: "/updates",
		HandlerFn:      updateHandler.UpdateJSONArrayHandler,
	})

	readHandler := ReadHandler{
		Service:  metricsService,
		Database: db,
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
	handlers = append(handlers, Handler{
		Method:         http.MethodGet,
		URLPathPattern: "/ping",
		HandlerFn:      readHandler.PingDB,
	})

	return handlers
}
