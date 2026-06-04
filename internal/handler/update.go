package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/service"
	"go.uber.org/zap"
)

type UpdateHandler struct {
	MetricsService *service.MetricsService
	cryptoService  *service.CryptoService
}

func (h *UpdateHandler) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("content-type", "text/plain")
	h.processUpdateRequest(res, req)
}

func (h *UpdateHandler) UpdateJSONHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("content-type", "application/json")
	h.processUpdateJSONRequest(res, req)
}

func (h *UpdateHandler) UpdateJSONArrayHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("content-type", "application/json")
	h.processUpdateJSONArrayRequest(res, req)
}

func (h *UpdateHandler) processUpdateRequest(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "type")
	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")

	result, err := h.MetricsService.SaveMetric(req.Context(), metricType, name, value)
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
	bodyBuf, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Log.Error("cannot read request body ", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	hashHeader := req.Header.Get("HashSHA256")
	if h.cryptoService.KeyPresent() && hashHeader != "" {
		bodyHash := h.cryptoService.CalculateHash(bodyBuf)
		if hashHeader != bodyHash {
			logger.Sugar.Errorf("HashSHA256 header %q doesn't match body hash %q", hashHeader, bodyHash)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	var metric models.Metrics
	if err := json.Unmarshal(bodyBuf, &metric); err != nil {
		logger.Log.Error("cannot decode request JSON body ", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := h.MetricsService.SaveMetricModel(req.Context(), &metric)
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
	if h.cryptoService.KeyPresent() {
		bodyHash := h.cryptoService.CalculateHash(buf.Bytes())
		res.Header().Set("HashSHA256", bodyHash)
	}
	logger.Sugar.Info("metric saved successfully ", zap.String("metric", result.String()))
	res.WriteHeader(http.StatusOK)
	res.Write(buf.Bytes())
}

func (h *UpdateHandler) processUpdateJSONArrayRequest(res http.ResponseWriter, req *http.Request) {
	bodyBuf, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Log.Error("cannot read request body ", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	hashHeader := req.Header.Get("HashSHA256")
	if h.cryptoService.KeyPresent() && hashHeader != "" {
		bodyHash := h.cryptoService.CalculateHash(bodyBuf)
		if hashHeader != bodyHash {
			logger.Sugar.Errorf("HashSHA256 header %q doesn't match body hash %q", hashHeader, bodyHash)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	var metrics []models.Metrics
	if err := json.Unmarshal(bodyBuf, &metrics); err != nil {
		logger.Log.Error("cannot decode request JSON body ", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	count, err := h.MetricsService.SaveMetricsModel(req.Context(), metrics)
	if err != nil {
		if errors.As(err, &models.ErrIncorrectType) {
			logger.Sugar.Error("metrics type incorrect ", zap.Error(err))
			res.WriteHeader(http.StatusBadRequest)
			return
		} else if errors.As(err, &models.ErrIncorrectFormat) {
			logger.Sugar.Errorf("metrics format incorrect: %q", err.Error())
			res.WriteHeader(http.StatusBadRequest)
			return
		} else if errors.As(err, &models.ErrSaveMetric) {
			logger.Sugar.Errorf("metrics save failed: %q", err.Error())
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	response := models.ResponsePayload{
		Status:  "ok",
		Message: fmt.Sprintf("successfully saved %d metrics", count),
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(response); err != nil {
		logger.Log.Error("error encoding response ", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	if h.cryptoService.KeyPresent() {
		bodyHash := h.cryptoService.CalculateHash(buf.Bytes())
		res.Header().Set("HashSHA256", bodyHash)
	}
	logger.Sugar.Info("metrics saved successfully ", zap.Int64("count", count))
	res.WriteHeader(http.StatusOK)
	res.Write(buf.Bytes())
}
