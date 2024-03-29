package exporter

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Namespace  = "cloud_platform_metrics"
	Deployment = "cloud_platform_metrics"
)

func NewExporter(cfg Config, logger log.Logger) *Exporter {
	return &Exporter{
		Metrics: Metrics{
			namespace_details: prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "", "namespace_details"),
				"Namespace details from cluster",
				[]string{"namespace", "application", "business_unit", "is_production"},
				prometheus.Labels{},
			),
			aws_costs: prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "", "aws_costs"),
				"AWS Costs",
				[]string{"aws_service", "hosted_ns"},
				prometheus.Labels{},
			),
			infrastructure_deployment_details_deployed: prometheus.NewDesc(
				prometheus.BuildFQName(Deployment, "", "infrastructure_deployment_details_deployed"),
				"Successful deployments from infrastructure",
				[]string{"deployed"},
				prometheus.Labels{},
			),
			infrastructure_deployment_details_failed: prometheus.NewDesc(
				prometheus.BuildFQName(Deployment, "", "infrastructure_deployment_details_failed"),
				"Failed deployments from infrastructure",
				[]string{"failed"},
				prometheus.Labels{},
			),
			incidents_mean_time_to_repair: prometheus.NewDesc(
				prometheus.BuildFQName(Deployment, "", "incidents_mean_time_to_repair"),
				"Incidents Mean Time to Repair",
				[]string{"incidents_mean_time_to_repair"},
				prometheus.Labels{},
			),
			incidents_mean_time_to_resolve: prometheus.NewDesc(
				prometheus.BuildFQName(Deployment, "", "incidents_mean_time_to_resolve"),
				"Incidents Mean Time to Resolve",
				[]string{"incidents_mean_time_to_resolve"},
				prometheus.Labels{},
			),
		},
		Config: cfg,
		logger: logger,
	}
}

// Describe implements the prometheus.Collector interface
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.Metrics.namespace_details
	ch <- e.Metrics.aws_costs
	ch <- e.Metrics.infrastructure_deployment_details_deployed
	ch <- e.Metrics.infrastructure_deployment_details_failed
	ch <- e.Metrics.incidents_mean_time_to_repair
	ch <- e.Metrics.incidents_mean_time_to_resolve
}

// Collect implements the prometheus.Collector interface
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	clientset, err := NewClient(e.Config)
	if err != nil {
		level.Error(e.logger).Log("msg", "failed to create kubernetes client", "err", err)
		return
	}
	namespaces, err := FetchNamespaceDetails(clientset)
	if err != nil {
		level.Error(e.logger).Log("msg", "failed to fetch namespace details", "err", err)
		return
	}

	// get Cost and Usage data from aws cost explorer api
	awsCostUsageData, err := FetchAWSCostDetails(namespaces)
	if err != nil {
		level.Error(e.logger).Log("msg", "failed to fetch aws cost details", "err", err)
		return
	}

	deployments, err := FetchCDdeployments()
	if err != nil {
		level.Error(e.logger).Log("msg", "failed to fetch deployment details", "err", err)
		return
	}

	incidentmeantimes, err := FetchIncidentMTTR()
	if err != nil {
		level.Error(e.logger).Log("msg", "failed to fetch incident mean time details", "err", err)
		return
	}

	for _, ns := range namespaces {
		ch <- prometheus.MustNewConstMetric(
			e.Metrics.namespace_details,
			prometheus.GaugeValue,
			1,
			ns.Name,
			ns.Annotations["cloud-platform.justice.gov.uk/application"],
			ns.Annotations["cloud-platform.justice.gov.uk/business-unit"],
			ns.Labels["cloud-platform.justice.gov.uk/is-production"],
		)

	}
	for _, ns := range namespaces {
		services := awsCostUsageData.costPerNamespace[ns.Name]

		for s, val := range services {
			ch <- prometheus.MustNewConstMetric(
				e.Metrics.aws_costs,
				prometheus.GaugeValue,
				val,
				s,
				ns.Name,
			)
		}
	}
	for _, nums := range deployments {
		ch <- prometheus.MustNewConstMetric(
			e.Metrics.infrastructure_deployment_details_deployed,
			prometheus.GaugeValue,
			nums["deployed"],
			"deployed",
		)
		ch <- prometheus.MustNewConstMetric(
			e.Metrics.infrastructure_deployment_details_failed,
			prometheus.GaugeValue,
			nums["failed"],
			"failed",
		)
	}

	for _, inc := range incidentmeantimes {
		ch <- prometheus.MustNewConstMetric(
			e.Metrics.incidents_mean_time_to_repair,
			prometheus.GaugeValue,
			inc["incidents_mean_time_to_repair"],
			"Incidents Mean Time to Repair",
		)
		ch <- prometheus.MustNewConstMetric(
			e.Metrics.incidents_mean_time_to_resolve,
			prometheus.GaugeValue,
			inc["incidents_mean_time_to_resolve"],
			"Incidents Mean_Time to Resolve",
		)
	}
}
