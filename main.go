package main

import (
	"ministryofjustice/cloud-platform-metrics/exporter"
	"net/http"
	"os"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
)

var (
	cfg     exporter.Config
	metrics exporter.Metrics
)

const (
	metricsPath = "/metrics"
	port        = ":8080"
)

func init() {
	cfg = exporter.Init()

}

func main() {
	promlogConfig := &promlog.Config{}
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("Starting Cloud Platforme Metrics Exporter")

	exp := exporter.NewExporter(cfg, logger)

	prometheus.MustRegister(exp)

	level.Info(logger).Log("serveMetrics: addr=%s path=%s", port, metricsPath)
	http.Handle(metricsPath, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Cloud Platform Metrics Exporter</title></head>
             <body>
             <h1>Cloud Platform Metrics Exporter</h1>
             <p><a href='` + metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	if err := http.ListenAndServe(port, nil); err != nil {
		level.Error(logger).Log("msg", "failed to serve metrics", "err", err)
		os.Exit(1)
	}

}
