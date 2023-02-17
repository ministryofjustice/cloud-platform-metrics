package exporter

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	ceTypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	v1 "k8s.io/api/core/v1"
)

// costs is a map which has the namespace name as key and the value a map
// of resource names as key and costs as value
type costs struct {
	costPerNamespace map[string]map[string]float64
}

const SHARED_COSTS string = "SHARED_COSTS"

// Annual cost of the Cloud Platform team is £1,260,000
// This is the monthly cost in USD
const MONTHLY_TEAM_COST = 136_000

const DAYS_TOGET_DATA int = 30

func FetchAWSCostDetails(namespaces []v1.Namespace) (costs, error) {

	//create a new costs object
	c := costs{
		costPerNamespace: map[string]map[string]float64{},
	}

	// get Cost and Usage data from aws cost explorer api
	awsCostUsageData, err := getAwsCostAndUsageData()
	if err != nil {
		return c, fmt.Errorf("failed to getAwsCostAndUsageData: %w", err)
	}

	// create the resources map for namespaces which are listed in the cluster
	// This is needed later to update shared costs for namespaces which doesnot have any aws resources
	for _, ns := range namespaces {
		resources := make(map[string]float64)
		c.costPerNamespace[ns.Name] = resources
	}

	// update the costs per namespace in a map for all aws resources from CostUsage data
	err = c.updatecostsByNamespace(awsCostUsageData)
	if err != nil {
		return c, fmt.Errorf("failed to updatecostsByNamespace : %w", err)
	}

	// add shared aws resources costs i.e resources which doesnot have namespace tags but global
	// resources to the CP account e.g ec2 instances, elasticsearch
	c.addSharedCosts()

	// add shared CP team costs
	c.addSharedTeamCosts()
	return c, nil
}

// getAwsCostAndUsageData get the data from aws cost explorer api and build a slice of [date,resourcename,namespacename,cost]
func getAwsCostAndUsageData() ([][]string, error) {

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to LoadDefaultConfig: %w", err)
	}
	svc := costexplorer.NewFromConfig(cfg)
	now, monthBefore := timeNow(DAYS_TOGET_DATA)

	param := &costexplorer.GetCostAndUsageInput{
		Granularity: ceTypes.GranularityMonthly,
		TimePeriod: &ceTypes.DateInterval{
			Start: aws.String(monthBefore),
			End:   aws.String(now),
		},
		Metrics: []string{"BlendedCost"},
		GroupBy: []ceTypes.GroupDefinition{
			{
				Type: ceTypes.GroupDefinitionTypeDimension,
				Key:  aws.String("SERVICE"),
			},
			{
				Type: ceTypes.GroupDefinitionTypeTag,
				Key:  aws.String("namespace"),
			},
		},
	}

	GetCostAndUsageOutput, err := svc.GetCostAndUsage(context.TODO(), param)
	if err != nil {
		return nil, fmt.Errorf("failed to GetCostAndUsage: %w", err)
	}

	var resultsCosts [][]string
	for _, results := range GetCostAndUsageOutput.ResultsByTime {

		for _, groups := range results.Groups {
			for _, metrics := range groups.Metrics {
				tag_value := strings.Split(groups.Keys[1], "$")
				if tag_value[1] == "" {
					tag_value[1] = SHARED_COSTS
				}
				info := []string{groups.Keys[0], tag_value[1], *metrics.Amount}

				resultsCosts = append(resultsCosts, info)

			}
		}
	}
	return resultsCosts, nil
}

// updatecostsByNamespace get the aws CostUsageData and update the costPerNamespace
// with resources and map per namespace
func (c *costs) updatecostsByNamespace(awsCostUsageData [][]string) error {

	for _, col := range awsCostUsageData {
		cost, err := strconv.ParseFloat(col[2], 64)
		if err != nil {
			return fmt.Errorf("failed to convert cost to float: %w", err)
		}

		c.addResource(col[1], col[0], cost)

	}
	return nil
}

// addSharedCosts get the value of shared costs for each namespace, delete the shared_costs key and
// and assign the shared_costs per namespace
func (c *costs) addSharedCosts() error {

	costsPerNs := c.getSharedCosts()
	delete(c.costPerNamespace, SHARED_COSTS)
	c.addSharedPerNamespace(costsPerNs)
	return nil

}

// getSharedCosts calculates the shared costs by adding
// all the costs of global resources needed for the Platform and
// divide it by number of namespaces in the cluster
func (c *costs) getSharedCosts() float64 {
	nKeys := len(c.costPerNamespace)

	sharedCosts := c.costPerNamespace[SHARED_COSTS]
	var totalCost float64
	for _, v := range sharedCosts {
		totalCost += v
	}
	// calculate per namespace cost taking away the shared_costs key
	perNsSharedCosts := totalCost / float64(nKeys-1)
	return math.Round(perNsSharedCosts*100) / 100
}

// addSharedPerNamespace get the shared cost and assign the shared_costs per namespace
func (c *costs) addSharedPerNamespace(costsPerNs float64) {
	for _, v := range c.costPerNamespace {
		v["Shared AWS Costs"] = costsPerNs
	}

}

// add shared team costs per namespace
func (c *costs) addSharedTeamCosts() error {

	nKeys := len(c.costPerNamespace)
	perNsSharedCPCosts := MONTHLY_TEAM_COST / float64(nKeys)
	roundedCPCost := math.Round(perNsSharedCPCosts*100) / 100

	for _, v := range c.costPerNamespace {
		v["Shared CP Team Costs"] = roundedCPCost
	}
	return nil

}

func (c *costs) addResource(ns, resource string, cost float64) {
	resources := c.costPerNamespace[ns]

	if resources == nil {
		resources = make(map[string]float64)
		c.costPerNamespace[ns] = resources
		resources[resource] = cost
	} else {

		curCost := c.hasResource(ns, resource)
		if curCost == 0 {
			resources[resource] = curCost
		}
		curCost = cost + curCost
		resources[resource] = math.Round(curCost*100) / 100
	}

}

// hasResource get the namespace name and resource name and checks if it has value in costPerNamespace
func (c *costs) hasResource(ns, resource string) float64 {
	return c.costPerNamespace[ns][resource]
}

// timeNow will take the number of days as input and return the current month and the month past 30 days
func timeNow(x int) (string, string) {
	dt := time.Now()
	now := dt.Format("2006-01-02")
	month := dt.AddDate(0, 0, -x).Format("2006-01-02")
	return now, month
}
