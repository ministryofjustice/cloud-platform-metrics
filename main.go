package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/namespace"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/homedir"
)

var (
	kubeconfig  string
	clusterName string
	interval    time.Duration
)

func init() {
	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "absolute path to the kubeconfig file")
	}

	flag.StringVar(&clusterName, "cluster", "arn:aws:eks:eu-west-2:754256621582:cluster/cp-0202-1257", "Kubernetes context specified in kubeconfig")
	flag.DurationVar(&interval, "interval", 10*time.Second, "How often to poll the cluster and aws for data.")
}

func main() {

	// Create new metrics and register them using the custom registry.
	m := NewMetrics()
	// // Add Go module build info.
	// reg.MustRegister(collectors.NewBuildInfoCollector())

	// Periodically record some sample latencies for the three services.
	go func() {
		for {
			namespaces, _ := fetchNamespaceDetails(kubeconfig)
			updateMetrics(namespaces, m)
			time.Sleep(1 * time.Minute)
		}
	}()

	serveMetrics(":8080", "/metrics")

}

type metrics struct {
	namespace_details *prometheus.GaugeVec
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

	// tweak defaults
	// see https://github.com/prometheus/client_golang/blob/v1.11.0/prometheus/registry.go#L61
	defaultRegistry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = defaultRegistry
	prometheus.DefaultGatherer = defaultRegistry

	prometheus.MustRegister(m.namespace_details)

	return m
}

func serveMetrics(addr, path string) {
	log.Printf("serveMetrics: addr=%s path=%s", addr, path)
	//http.Handle(path, promhttp.Handler())
	http.Handle(path, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(addr, nil))
}

func pollForClusterMetrics(m *metrics) error {

	for {
		namespaces, err := fetchNamespaceDetails(kubeconfig)
		if err != nil {
			return err
		}
		m.namespace_details.Reset()
		updateMetrics(namespaces, m)
		time.Sleep(1 * time.Minute)
	}
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

func fetchNamespaceDetails(kubeconfig string) ([]v1.Namespace, error) {

	// Gain access to a Kubernetes cluster using a config file for given cluster context.
	clientset, err := authenticate.CreateClientFromConfigFile(kubeconfig, clusterName)
	if err != nil {
		log.Fatalln(err.Error())
	}
	if err != nil {
		return nil, err
	}

	// Get the list of namespaces from the cluster which is set in the clientset
	namespaces, err := namespace.GetAllNamespacesFromCluster(clientset)
	if err != nil {
		return nil, err
	}
	return namespaces, nil
}
