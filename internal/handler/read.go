package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/service"
	"go.uber.org/zap"
)

type ReadHandler struct {
	Service *service.MetricsService
	tmpl    *template.Template
}

type MetricItem struct {
	ID    string
	Type  string
	Value string
}

type PageData struct {
	Title       string
	Subtitle    string
	Metrics     []MetricItem
	LastUpdated string
}

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
                <strong>{{.ID}} [{{.Type}}]</strong>
                <span class="meta">{{.Value}}</span>
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

func (h *ReadHandler) ListHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.processListRequest(res, req)
}

func (h *ReadHandler) ValueHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain")
	h.processValueRequest(res, req)
}

func (h *ReadHandler) ValueJSONHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	h.processValueJSONRequest(res, req)
}

func (h *ReadHandler) processListRequest(res http.ResponseWriter, req *http.Request) {
	data := PageData{
		Title:       "Metrics",
		Subtitle:    "For each metric specified it's type and current value",
		Metrics:     []MetricItem{},
		LastUpdated: time.Now().Format("2006-01-02 15:04 MST"),
	}
	for _, m := range h.Service.Storage.GetAllMetrics() {
		value, err := m.GetValueAsString()
		if err != nil {
			logger.Sugar.Errorf("can't get value for metric %s, type %s, err: %q", m.ID, m.MType, err)
			continue
		}
		data.Metrics = append(data.Metrics, MetricItem{
			ID:    m.ID,
			Type:  m.MType,
			Value: value,
		})
	}
	if err := h.tmpl.Execute(res, data); err != nil {
		http.Error(res, "Template rendering error", http.StatusInternalServerError)
		return
	}
	logger.Sugar.Infof("metrics list sent successfully")
}

func (h *ReadHandler) processValueRequest(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "type")
	name := chi.URLParam(req, "name")

	result, err := h.Service.GetValue(metricType, name)
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
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, err := h.Service.ReadMetric(&metric)
	if err != nil {
		if errors.As(err, &models.ErrGetMetric) {
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
