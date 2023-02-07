package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ministryofjustice/cloud-platform-environments/pkg/namespace"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	interval time.Duration
)

type metrics struct {
	namespace_details *prometheus.GaugeVec
}

func init() {
	flag.DurationVar(&interval, "interval", 10*time.Second, "How often to poll the cluster and aws for data.")
}

func main() {

	// Create new metrics and register them using the custom registry.
	m := NewMetrics()

	go func() {
		for {
			namespaces, _ := fetchNamespaceDetails()
			updateMetrics(namespaces, m)
			time.Sleep(1 * time.Minute)
		}
	}()

	serveMetrics(":8080", "/metrics")

}

func NewMetrics() *metrics {

	m := &metrics{
		namespace_details: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "namespace_details",
			Help: "Namespace details from cluster",
		},
			[]string{"namespace", "application", "business_unit", "is_production"},
		),
	}

	prometheus.MustRegister(m.namespace_details)

	return m
}

func serveMetrics(addr, path string) {
	log.Printf("serveMetrics: addr=%s path=%s", addr, path)
	http.Handle(path, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(addr, nil))
}

// Store the namespaces in the clusterMetrics struct
func updateMetrics(namespaces []v1.Namespace, m *metrics) {

	// get required details of each namespace and store it in namespace map
	for _, ns := range namespaces {
		log.Printf("namespace: %s", ns.Name)
		m.namespace_details.With(
			prometheus.Labels{
				"namespace":     ns.Name,
				"application":   ns.Annotations["cloud-platform.justice.gov.uk/application"],
				"business_unit": ns.Annotations["cloud-platform.justice.gov.uk/business-unit"],
				"is_production": ns.Labels["cloud-platform.justice.gov.uk/is-production"],
			}).Set(1)
	}
}

func fetchNamespaceDetails() ([]v1.Namespace, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Get the list of namespaces from the cluster which is set in the clientset
	namespaces, err := namespace.GetAllNamespacesFromCluster(clientset)
	if err != nil {
		return nil, fmt.Errorf("failed to GetAllNamespacesFromCluster from cluster: %w", err)
	}
	return namespaces, nil
}
