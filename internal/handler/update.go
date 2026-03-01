package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/service"
)

type UpdateHandler struct {
	Service service.MetricsService
}

func (h *UpdateHandler) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("content-type", "text/plain")
	h.processUpdateRequest(res, req)
}

func (h *UpdateHandler) processUpdateRequest(res http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	if contentType != "text/plain" {
		fmt.Printf("Incorrect request content type: %q\n", contentType)
		res.WriteHeader(http.StatusNotFound)
		return
	}

	metricType := chi.URLParam(req, "type")
	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")

	result, err := h.Service.SaveMetric(metricType, name, value)
	if err != nil {
		if errors.As(err, &models.ErrIncorrectType) {
			fmt.Printf("Metric type incorrect: %q\n", err.Error())
			res.WriteHeader(http.StatusBadRequest)
			return
		} else if errors.As(err, &models.ErrIncorrectFormat) {
			fmt.Printf("Metric format incorrect: %q\n", err.Error())
			res.WriteHeader(http.StatusBadRequest)
			return
		} else if errors.As(err, &models.ErrSaveMetric) {
			if err.Error() == "Unknown metric" {
				fmt.Printf("Unknown metric: %q\n", err.Error())
				res.WriteHeader(http.StatusNotFound)
			} else {
				fmt.Printf("Metric save error: %q\n", err.Error())
				res.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}

	fmt.Printf("Metric saved successfully: %q %q %q\n", metricType, name, value)
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(result))
}
