package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/service"
)

type UpdateHandler struct {
	Service *service.MetricsService
}

func (h *UpdateHandler) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("content-type", "text/plain")
	h.processUpdateRequest(res, req)
}

func (h *UpdateHandler) processUpdateRequest(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "type")
	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")

	result, err := h.Service.SaveMetric(metricType, name, value)
	if err != nil {
		if errors.As(err, &models.ErrIncorrectType) {
			logger.Sugar.Errorf("Metric type incorrect: %q", err.Error())
			res.WriteHeader(http.StatusBadRequest)
			return
		} else if errors.As(err, &models.ErrIncorrectFormat) {
			logger.Sugar.Errorf("Metric format incorrect: %q", err.Error())
			res.WriteHeader(http.StatusBadRequest)
			return
		} else if errors.As(err, &models.ErrSaveMetric) {
			logger.Sugar.Errorf("Metric save error: %q", err.Error())
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	logger.Sugar.Infof("Metric saved successfully: %q %q %s", metricType, name, value)
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(result))
}
