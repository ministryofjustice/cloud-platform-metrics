package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

func NewExporter(cfg Config) *Exporter {
	return &Exporter{
		Metrics: Metrics{
			namespace_details: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "namespace_details",
				Help: "Namespace details from cluster",
			},
				[]string{"namespace", "application", "business_unit", "is_production"},
			),
			aws_costs: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "aws_costs",
				Help: "AWS Costs",
			},
				[]string{"aws_service", "hosted_ns"},
			),
		},
		Config: cfg,
	}
}

// Describe implements the prometheus.Collector interface
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.Metrics.namespace_details.Describe(ch)
	e.Metrics.aws_costs.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.Metrics.namespace_details.Collect(ch)
	e.Metrics.aws_costs.Collect(ch)
}
