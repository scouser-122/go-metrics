package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/scouser-122/go-metrics/internal/logger"
)

// CreateChiRouter creates and configures a chi router with the provided handlers.
// It applies GzipMiddleware and RequestLogger to all routes and sets up 404/405 handlers.
func CreateChiRouter(handlers *[]Handler) *chi.Mux {
	r := chi.NewRouter()
	for _, h := range *handlers {
		switch h.Method {
		case http.MethodGet:
			r.Get(h.URLPathPattern, GzipMiddleware(RequestLogger(h.HandlerFn)))
		case http.MethodPost:
			r.Post(h.URLPathPattern, GzipMiddleware(RequestLogger(h.HandlerFn)))
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
	return r
}
