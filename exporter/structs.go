package exporter

import (
	"flag"

	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shurcooL/githubv4"
)

type Metrics struct {
	namespace_details           *prometheus.Desc
	aws_costs                   *prometheus.Desc
	deployment_details_deployed *prometheus.Desc
	deployment_details_failed   *prometheus.Desc
}

type Config struct {
	context        string
	kubeconfigPath string
	inCluster      bool
	region         string
	scrapeInterval time.Duration
}

type Exporter struct {
	logger  log.Logger
	Metrics Metrics
	Config  Config
}

type nodes struct {
	PullRequest struct {
		Title githubv4.String
		Url   githubv4.String
	} `graphql:"... on PullRequest"`
}

type date struct {
	first      time.Time
	last       time.Time
	monthIndex string
}

type infraPRs struct {
	deployed, failed int
}

func Init() Config {

	var (
		ctx            = flag.String("context", "live.cloud-platform.service.justice.gov.uk", "Kubernetes context specified in kubeconfig")
		kubeconfigPath = flag.String("kubeconfig", os.Getenv("KUBECONFIG"), "Name of kubeconfig file in S3 bucket")
		inCluster      = flag.Bool("in-cluster", false, "Use in-cluster config")
		interval       = flag.Duration("interval", 10*time.Second, "How often to poll the cluster and aws for data.")
		region         = flag.String("region", os.Getenv("AWS_REGION"), "AWS Region")
	)

	config := Config{
		context:        *ctx,
		kubeconfigPath: *kubeconfigPath,
		inCluster:      *inCluster,
		region:         *region,
		scrapeInterval: *interval,
	}
	return config
}
