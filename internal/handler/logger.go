package handler

import (
	"net/http"
	"time"

	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
)

func RequestLogger(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &models.ResponseData{
			Status: 0,
			Size:   0,
		}
		lw := models.LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData:   responseData,
		}
		h(&lw, r)

		duration := time.Since(start)

		logger.Log.Sugar().Infof(
			"Processed request. uri: %s, method: %s, status: %d, duration: %s",
			r.RequestURI,
			r.Method,
			responseData.Status,
			duration,
		)
	})
}
