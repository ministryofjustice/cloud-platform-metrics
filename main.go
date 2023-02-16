package main

import (
	"log"
	"ministryofjustice/cloud-platform-metrics/exporter"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/version"
)

var (
	cfg     exporter.Config
	metrics exporter.Metrics
)

const (
	metricsPath = "/metrics"
)

func init() {
	cfg = exporter.Init()

}

func main() {
	promlogConfig := &promlog.Config{}
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("Starting Cloud Platforme Metrics Exporter", version.Info())

	exp := exporter.NewExporter(cfg)

	prometheus.MustRegister(exp)

	go func() {
		for {
			clientset, err := exporter.NewClient(cfg)
			if err != nil {
				level.Error(logger).Log("msg", "failed to create kube client", "err", err)
				os.Exit(1)
			}
			namespaces, err := exporter.FetchNamespaceDetails(clientset)
			if err != nil {
				level.Error(logger).Log("msg", "failed to fetch namespace details", "err", err)
				os.Exit(1)
			}

			// get Cost and Usage data from aws cost explorer api
			awsCostUsageData, err := exporter.FetchAWSCostDetails(namespaces)
			if err != nil {
				level.Error(logger).Log("msg", "failed to fetch aws cost details", "err", err)
				os.Exit(1)
			}

			exporter.UpdateNSDetailsMetrics(namespaces, exp)
			exporter.UpdateAWSCostsMetrics(awsCostUsageData, namespaces, exp)
			time.Sleep(1 * time.Hour)
		}
	}()

	serveMetrics(":8080", "/metrics")
}

func serveMetrics(addr, path string) {
	log.Printf("serveMetrics: addr=%s path=%s", addr, path)
	http.Handle(path, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(addr, nil))
}
