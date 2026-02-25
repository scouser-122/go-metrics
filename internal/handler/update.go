package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/service"
)

func UpdateHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("content-type", "text/plain")
	res.WriteHeader(processUpdateRequest(req))
}

func processUpdateRequest(req *http.Request) int {
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
	if metricType != models.Counter && metricType != models.Gauge {
		fmt.Printf("Metric type incorrect: %q\n", metricType)
		return http.StatusBadRequest
	}

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

	switch metricType {
	case models.Counter:
		counterValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			fmt.Printf("Metric counter incorrect format: %v\n", err)
			return http.StatusBadRequest
		}
		service.SaveCounter(name, counterValue)
	case models.Gauge:
		gaugeValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			fmt.Printf("Metric gauge incorrect format: %v\n", err)
			return http.StatusBadRequest
		}
		service.SaveGauge(name, gaugeValue)
	}

	fmt.Printf("Metric saved successfully: %q %q %q\n", metricType, name, value)
	return http.StatusOK
}
