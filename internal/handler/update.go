package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/service"
	"go.uber.org/zap"
)

type UpdateHandler struct {
	Service *service.MetricsService
}

func (h *UpdateHandler) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("content-type", "text/plain")
	h.processUpdateRequest(res, req)
}

func (h *UpdateHandler) UpdateJSONHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("content-type", "application/json")
	h.processUpdateJSONRequest(res, req)
}

func (h *UpdateHandler) processUpdateRequest(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "type")
	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")

	result, err := h.Service.SaveMetric(metricType, name, value)
	if err != nil {
		if errors.As(err, &models.ErrIncorrectType) {
			logger.Sugar.Errorf("metric type incorrect: %q", err.Error())
			res.WriteHeader(http.StatusBadRequest)
			return
		} else if errors.As(err, &models.ErrIncorrectFormat) {
			logger.Sugar.Errorf("metric format incorrect: %q", err.Error())
			res.WriteHeader(http.StatusBadRequest)
			return
		} else if errors.As(err, &models.ErrSaveMetric) {
			logger.Sugar.Errorf("metric save error: %q", err.Error())
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	logger.Sugar.Infof("metric saved successfully: %q %q %s", metricType, name, value)
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(result))
}

func (h *UpdateHandler) processUpdateJSONRequest(res http.ResponseWriter, req *http.Request) {
	var metric models.Metrics
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&metric); err != nil {
		logger.Log.Error("cannot decode request JSON body ", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := h.Service.SaveMetricModel(&metric)
	if err != nil {
		if errors.As(err, &models.ErrIncorrectType) {
			logger.Sugar.Error("metric type incorrect ", zap.Error(err))
			res.WriteHeader(http.StatusBadRequest)
			return
		} else if errors.As(err, &models.ErrIncorrectFormat) {
			logger.Sugar.Errorf("metric format incorrect: %q", err.Error())
			res.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(result); err != nil {
		logger.Log.Error("error encoding response ", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	logger.Sugar.Info("metric saved successfully ", zap.String("metric", result.String()))
	res.WriteHeader(http.StatusOK)
	res.Write(buf.Bytes())
}
