package exporter

import (
	"fmt"

	authenticate "github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	namespace "github.com/ministryofjustice/cloud-platform-environments/pkg/namespace"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func NewClient(c Config) (*kubernetes.Clientset, error) {
	clientset := &kubernetes.Clientset{}
	err := error(nil)
	if c.inCluster {
		clientset, err = createKubeClient()
		if err != nil {
			return nil, fmt.Errorf("failed to create in cluster kube client: %w", err)
		}
	} else {
		clientset, err = authenticate.CreateClientFromConfigFile(c.kubeconfigPath, c.context) // create kube client
		if err != nil {
			return nil, fmt.Errorf("failed to create kube client: %w", err)
		}
	}
	return clientset, nil
}

func FetchNamespaceDetails(clientset *kubernetes.Clientset) ([]v1.Namespace, error) {

	// Get the list of namespaces from the cluster which is set in the clientset
	namespaces, err := namespace.GetAllNamespacesFromCluster(clientset)
	if err != nil {
		return nil, fmt.Errorf("failed to GetAllNamespacesFromCluster from cluster: %w", err)
	}
	return namespaces, nil
}

func createKubeClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}
	return clientset, nil
}

// Store the namespaces in the clusterMetrics struct
func UpdateNSDetailsMetrics(namespaces []v1.Namespace, e *Exporter) {
	// get required details of each namespace and store it in namespace map
	for _, ns := range namespaces {
		e.Metrics.namespace_details.With(
			prometheus.Labels{
				"namespace":     ns.Name,
				"application":   ns.Annotations["cloud-platform.justice.gov.uk/application"],
				"business_unit": ns.Annotations["cloud-platform.justice.gov.uk/business-unit"],
				"is_production": ns.Labels["cloud-platform.justice.gov.uk/is-production"],
			}).Set(1)
	}
}
