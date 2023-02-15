package hosted_services

import (
	"fmt"

	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/namespace"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	Context        string
	KubeconfigPath string
	InCluster      bool
}

func NewClient(c *Client) (*kubernetes.Clientset, error) {
	if c.inCluster {
		clientset, err := createKubeClient()
		if err != nil {
			return nil, fmt.Errorf("failed to create kube client: %w", err)
		}
	} else {
		clientset, err := authenticate.CreateClientFromConfigFile() // create kube client
		if err != nil {
			return nil, fmt.Errorf("failed to create kube client: %w", err)
		}
	}
	return clientset, nil
}

func (c *Client) FetchNamespaceDetails() ([]v1.Namespace, error) {

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
