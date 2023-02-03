package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/cluster"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/namespace"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/homedir"
)

var (
	kubeconfig  string
	clusterName string
	awsRegion   string
	interval    time.Duration
)

type metrics struct {
	namespace_details *prometheus.GaugeVec
}

func init() {
	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "absolute path to the kubeconfig file")
	}

	flag.StringVar(&clusterName, "cluster", "cp-0202-1257", "cluster name to Authenticate to")
	flag.StringVar(&awsRegion, "region", "eu-west-2", "AWS region to Authenticate to")
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

	creds, err := getCredentials(awsRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to auth to cluster: %w", err)
	}

	clientset, err := cluster.AuthToCluster(clusterName, creds.Eks, kubeconfig, creds.Profile)
	if err != nil {
		return nil, fmt.Errorf("failed to auth to cluster: %w", err)
	}

	// Get the list of namespaces from the cluster which is set in the clientset
	namespaces, err := namespace.GetAllNamespacesFromCluster(clientset)
	if err != nil {
		return nil, fmt.Errorf("failed to GetAllNamespacesFromCluster from cluster: %w", err)
	}
	return namespaces, nil
}

func getCredentials(awsRegion string) (*client.AwsCredentials, error) {
	creds, err := client.NewAwsCreds(awsRegion)
	if err != nil {
		return nil, err
	}

	return creds, nil
}
