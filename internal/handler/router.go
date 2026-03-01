package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func CreateChiRouter(updateHandler http.HandlerFunc, listHandler http.HandlerFunc) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", updateHandler)
	r.Get("/", listHandler)
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
