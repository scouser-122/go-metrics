package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/service"
)

type UpdateHandler struct {
	Service service.MetricsService
}

func (h *UpdateHandler) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("content-type", "text/plain")
	res.WriteHeader(h.processUpdateRequest(req))
}

func (h *UpdateHandler) processUpdateRequest(req *http.Request) int {
	if req.Method != http.MethodPost {
		fmt.Printf("Incorrect request method: %q\n", req.Method)
		return http.StatusNotFound
	}

	contentType := req.Header.Get("Content-Type")
	if contentType != "text/plain" {
		fmt.Printf("Incorrect request content type: %q\n", contentType)
		return http.StatusNotFound
	}

	fullPath := req.URL.Path
	pathSegments := strings.Split(fullPath, "/")

	if len(pathSegments) < 3 {
		fmt.Printf("Metric type not specified\n")
		return http.StatusNotFound
	}

	metricType := pathSegments[2]

	if len(pathSegments) < 4 {
		fmt.Printf("Metric name not specified\n")
		return http.StatusNotFound
	}

	name := pathSegments[3]

	if len(pathSegments) < 5 {
		fmt.Printf("Metric value not specified\n")
		return http.StatusNotFound
	}

	value := pathSegments[4]

	_, err := h.Service.SaveMetric(metricType, name, value)
	if err != nil {
		if errors.As(err, &models.ErrIncorrectType) {
			fmt.Printf("Metric type incorrect: %q\n", err.Error())
			return http.StatusBadRequest
		} else if errors.As(err, &models.ErrIncorrectFormat) {
			fmt.Printf("Metric format incorrect: %q\n", err.Error())
			return http.StatusBadRequest
		} else if errors.As(err, &models.ErrSaveMetric) {
			fmt.Printf("Metric save error: %q\n", err.Error())
			return http.StatusInternalServerError
		}
	}

	fmt.Printf("Metric saved successfully: %q %q %q\n", metricType, name, value)
	return http.StatusOK
}
