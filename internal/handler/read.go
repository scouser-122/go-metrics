package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/repository/db"
	"github.com/scouser-122/go-metrics/internal/service"
	"go.uber.org/zap"
)

// ReadHandler handles all read operations for metrics including listing and retrieving metric values.
type ReadHandler struct {
	Service  *service.MetricsService
	Database *db.PostgresDatabase
	tmpl     *template.Template
}

// PageData represents the data structure for rendering the metrics HTML page.
type PageData struct {
	Title       string
	Subtitle    string
	Metrics     []models.Metrics
	LastUpdated string
}

// CreateTemplate initializes and parses the HTML template for displaying metrics.
func (h *ReadHandler) CreateTemplate() {
	var err error
	h.tmpl, err = template.New("page").Funcs(template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("Jan 02, 2006 15:04")
		},
	}).Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: system-ui, sans-serif;
            max-width: 800px;
            margin: 40px auto;
            line-height: 1.6;
            padding: 0 20px;
            background: #0f0f15;
            color: #e0e0ff;
        }
        h1 { color: #a5b4fc; margin-bottom: 0.4em; }
        .subtitle { color: #94a3b8; font-style: italic; margin-bottom: 2rem; }
        .item {
            background: #1e1e2e;
            padding: 1.2rem;
            margin: 1rem 0;
            border-radius: 8px;
            border-left: 4px solid #6366f1;
        }
        .item-header {
            display: flex;
            justify-content: space-between;
            align-items: baseline;
            margin-bottom: 0.6rem;
            color: #c7d2fe;
        }
        .meta { font-size: 0.9em; color: #64748b; }
        footer {
            margin-top: 4rem;
            text-align: center;
            color: #475569;
            font-size: 0.95em;
        }
    </style>
</head>
<body>
    <h1>{{.Title}}</h1>
    <p class="subtitle">{{.Subtitle}}</p>

    {{if .Metrics}}
        {{range .Metrics}}
        <div class="item">
            <div class="item-header">
                <strong>{{.ID}} [{{.MType}}]</strong>
                <span class="meta">{{.GetValueAsString}}</span>
            </div>
        </div>
        {{end}}
    {{else}}
        <p style="color: #94a3b8; font-style: italic;">No items yet...</p>
    {{end}}

    <footer>
        Last updated: {{.LastUpdated}} | Served via Go + Chi
    </footer>
</body>
</html>
	`)
	if err != nil {
		log.Fatalf("Template parsing failed: %v", err)
	}
}

// ListHandler handles GET / requests and returns an HTML page with all metrics.
// @Tags Read
// @Summary Metrics list as html page
// @ID ListHandler
// @Accept  text/plain
// @Produce text/html; charset=utf-8
// @Success 200 {string} string  "Html page with metrics list"
// @Failure 500 {string} string  "Internal error"
// @Router / [get]
func (h *ReadHandler) ListHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.processListRequest(res, req)
}

// ValueHandler handles GET /value/{type}/{name} requests and returns a metric value as plain text.
// @Tags Read
// @Summary Get metric value by type and name
// @ID ValueHandler
// @Accept  text/plain
// @Produce text/plain
// @Param type path string true "Metric type" Enums(counter, gauge) default(counter)
// @Param name path string true "Metric name"
// @Success 200 {string} string  "Metric value"
// @Success 400 {string} string  "Metric type incorrect"
// @Success 404 {string} string  "Not found"
// @Failure 500 {string} string  "Internal error"
// @Router /value/{type}/{name} [get]
func (h *ReadHandler) ValueHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain")
	h.processValueRequest(res, req)
}

// ValueJSONHandler handles POST /value requests and returns a metric value as JSON.
// @Tags Read
// @Summary Get metric value by specified params
// @ID ValueJSONHandler
// @Accept  application/json
// @Produce application/json
// @Param metric body models.Metrics true "Metric params"
// @Success 200 {object} models.Metrics "Metric data"
// @Success 400 {string} string  "Metric type incorrect"
// @Success 404 {string} string  "Not found"
// @Failure 500 {string} string  "Internal error"
// @Router /value/ [post]
func (h *ReadHandler) ValueJSONHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	h.processValueJSONRequest(res, req)
}

// PingDB handles GET /ping requests and checks database connectivity.
// @Tags Read
// @Summary Ping DB connection
// @ID PingDB
// @Accept  text/plain
// @Produce text/plain
// @Success 200 {string} string "ok"
// @Failure 500 {string} string  "Internal error"
// @Router /ping [get]
func (h *ReadHandler) PingDB(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain")
	var status int
	var result string
	if err := h.Database.Ping(req.Context()); err != nil {
		status = http.StatusInternalServerError
		result = "error"
		logger.Sugar.Errorf("DB ping failed: %s", err)
	} else {
		status = http.StatusOK
		result = "ok"
	}
	res.WriteHeader(status)
	res.Write([]byte(result))
}

func (h *ReadHandler) processListRequest(res http.ResponseWriter, req *http.Request) {
	metrics := h.Service.Storage.GetAllMetrics(req.Context())
	data := PageData{
		Title:       "Metrics",
		Subtitle:    "For each metric specified it's type and current value",
		LastUpdated: time.Now().Format("2006-01-02 15:04 MST"),
	}
	data.Metrics = metrics
	var sb strings.Builder
	if err := h.tmpl.Execute(&sb, data); err != nil {
		http.Error(res, "Template rendering error", http.StatusInternalServerError)
		return
	}
	result := sb.String()
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(result))
	logger.Sugar.Infof("metrics list sent successfully")
}

func (h *ReadHandler) processValueRequest(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "type")
	name := chi.URLParam(req, "name")

	result, err := h.Service.GetValue(req.Context(), metricType, name)
	if err != nil {
		if errors.As(err, &models.ErrIncorrectType) {
			logger.Sugar.Errorf("metric type incorrect: %q", err.Error())
			res.WriteHeader(http.StatusBadRequest)
		} else if errors.As(err, &models.ErrGetMetric) {
			logger.Sugar.Errorf("can't get metric value: %q", err.Error())
			res.WriteHeader(http.StatusNotFound)
		}
		return
	}

	logger.Sugar.Infof("metric value obtained successfully: %s [%s] %q", metricType, name, result)
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(result))
}

func (h *ReadHandler) processValueJSONRequest(res http.ResponseWriter, req *http.Request) {
	var metric models.Metrics
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&metric); err != nil {
		logger.Log.Error("cannot decode request JSON body ", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := h.Service.ReadMetric(req.Context(), &metric)
	if err != nil {
		if errors.As(err, &models.ErrIncorrectType) {
			logger.Sugar.Errorf("metric type incorrect: %q", err.Error())
			res.WriteHeader(http.StatusBadRequest)
		} else if errors.As(err, &models.ErrGetMetric) {
			logger.Sugar.Errorf("can't get metric value: %q", err.Error())
			res.WriteHeader(http.StatusNotFound)
		}
		return
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(result); err != nil {
		logger.Log.Error("error encoding response ", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	logger.Sugar.Infof("metric value obtained successfully ", zap.String("metric", result.String()))
	res.WriteHeader(http.StatusOK)
	res.Write(buf.Bytes())
}
