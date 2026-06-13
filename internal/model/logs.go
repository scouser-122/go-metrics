package models

import "net/http"

type (
	// ResponseData contains HTTP response metadata.
	ResponseData struct {
		Status int
		Size   int
	}

	// LoggingResponseWriter wraps http.ResponseWriter to capture response data.
	LoggingResponseWriter struct {
		http.ResponseWriter
		ResponseData *ResponseData
	}
)

// Write captures the response size and writes the data.
func (r *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.ResponseData.Size += size
	return size, err
}

// WriteHeader captures the response status code.
func (r *LoggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.ResponseData.Status = statusCode
}

type contextKey string

const IPAddressContextKey contextKey = "clientIpAddress"
