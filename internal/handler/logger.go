package handler

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
)

func RequestLogger(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestId := uuid.New()
		logger.Log.Sugar().Infof(
			"Received request. uri: %s, method: %s, id: %s",
			r.RequestURI,
			r.Method,
			requestId,
		)

		responseData := &models.ResponseData{
			Status: 0,
			Size:   0,
		}
		lw := models.LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData:   responseData,
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
		if len(bodyBytes) > 0 {
			logger.Log.Sugar().Debugf(
				"Request body. id: %s, body: %d",
				requestId,
				string(bodyBytes),
			)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		h(&lw, r)

		duration := time.Since(start)

		logger.Log.Sugar().Infof(
			"Processed request. id: %s, status: %d, duration: %s",
			requestId,
			responseData.Status,
			duration,
		)
	})
}
