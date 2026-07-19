package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
)

// CreateChiRouterWithHandlers creates and configures a chi router with the provided handlers.
// It applies GzipMiddleware and RequestLogger to all routes and sets up 404/405 handlers.
func CreateChiRouterWithHandlers(handlers *[]Handler, config *config.ServerConfig) *chi.Mux {
	r := chi.NewRouter()
	AddHandlersForRouter(r, handlers, config)
	return r
}

// AddHandlersForRouter  configures a chi router with the provided handlers.
// It applies GzipMiddleware and RequestLogger to all routes and sets up 404/405 handlers.
func AddHandlersForRouter(r *chi.Mux, handlers *[]Handler, config *config.ServerConfig) {
	if config.TrustedSubnet != "" {
		logger.Sugar.Infof("trusted subnet: %s", config.TrustedSubnet)
	}
	for _, h := range *handlers {
		switch h.Method {
		case http.MethodGet:
			r.Get(
				h.URLPathPattern,
				TrustedMiddleware(
					GzipMiddleware(
						RequestLogger(h.HandlerFn),
					),
					config,
				),
			)
		case http.MethodPost:
			r.Post(
				h.URLPathPattern,
				TrustedMiddleware(
					GzipMiddleware(
						RequestLogger(h.HandlerFn),
					),
					config,
				),
			)
		default:
			logger.Log.Sugar().Errorf("Provided unsupported request handler method: %q", h)
		}
	}
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(405)
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(404)
	})
}
