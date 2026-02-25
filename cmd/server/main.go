package main

import (
	"net/http"

	"github.com/scouser-122/go-metrics/internal/handler"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update", handler.UpdateHandler)
	mux.HandleFunc("/update/", handler.UpdateHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
