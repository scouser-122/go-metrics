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
	h.processUpdateRequest(res, req)
}

func (h *UpdateHandler) processUpdateRequest(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		fmt.Printf("Incorrect request method: %q\n", req.Method)
		res.WriteHeader(http.StatusNotFound)
		return
	}

	contentType := req.Header.Get("Content-Type")
	if contentType != "text/plain" {
		fmt.Printf("Incorrect request content type: %q\n", contentType)
		res.WriteHeader(http.StatusNotFound)
		return
	}

	fullPath := req.URL.Path
	pathSegments := strings.Split(fullPath, "/")

	if len(pathSegments) < 3 {
		fmt.Printf("Metric type not specified\n")
		res.WriteHeader(http.StatusNotFound)
		return
	}

	metricType := pathSegments[2]

	if len(pathSegments) < 4 {
		fmt.Printf("Metric name not specified\n")
		res.WriteHeader(http.StatusNotFound)
		return
	}

	name := pathSegments[3]

	if len(pathSegments) < 5 {
		fmt.Printf("Metric value not specified\n")
		res.WriteHeader(http.StatusNotFound)
		return
	}

	value := pathSegments[4]

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
			fmt.Printf("Metric save error: %q\n", err.Error())
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	fmt.Printf("Metric saved successfully: %q %q %q\n", metricType, name, value)
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(result))
}
