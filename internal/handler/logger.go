package handler

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
	"go.uber.org/zap/zapcore"
)

// RequestLogger is an HTTP middleware that logs request and response details including
// URI, method, status code, response size, and duration.
func RequestLogger(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := uuid.New()
		logger.Log.Sugar().Infof(
			"received request. uri: %s, method: %s, id: %s",
			r.RequestURI,
			r.Method,
			requestID,
		)

		responseData := &models.ResponseData{
			Status: 0,
			Size:   0,
		}
		lw := models.LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData:   responseData,
		}

		if logger.Log.Level() == zapcore.DebugLevel {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "can't read body", http.StatusBadRequest)
				return
			}
			if len(bodyBytes) > 0 {
				logger.Log.Sugar().Debugf(
					"request body. id: %s, body: %d",
					requestID,
					string(bodyBytes),
				)
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		ctx := context.WithValue(r.Context(), models.IPAddressContextKey, getClientIP(r))
		h(&lw, r.WithContext(ctx))

		duration := time.Since(start)

		logger.Log.Sugar().Infof(
			"processed request. id: %s, status: %d, duration: %s",
			requestID,
			responseData.Status,
			duration,
		)
	})
}

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		clientIP := strings.TrimSpace(ips[0])
		if clientIP != "" {
			return clientIP
		}
	}

	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
