package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func CreateChiRouter(updateHandler *UpdateHandler, obtainHandler *ReadHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", updateHandler.UpdateHandler)
	r.Get("/", obtainHandler.ListHandler)
	r.Get("/value/{type}/{name}", obtainHandler.GetHandler)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(404)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(404)
	})
	return r
}
