package pao

import (
	"context"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/mco"
	v2 "github.com/openshift/cluster-node-tuning-operator/pkg/apis/performanceprofile/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

var apiClient *clients.Settings

// CreatePerformanceProfile creates performance profile.
func CreatePerformanceProfile(performanceProfileName string, mcpPoolName string) error {
	isolatedCPUSet := v2.CPUSet("8-15")
	reservedCPUSet := v2.CPUSet("0-7")
	hugepageSize := v2.HugePageSize("1G")
	performanceProfile := &v2.PerformanceProfile{
		ObjectMeta: v1.ObjectMeta{
			Name: performanceProfileName,
		},
		Spec: v2.PerformanceProfileSpec{
			CPU: &v2.CPU{
				Isolated: &isolatedCPUSet,
				Reserved: &reservedCPUSet,
			},
			HugePages: &v2.HugePages{
				DefaultHugePagesSize: &hugepageSize,
				Pages: []v2.HugePage{
					{
						Count: 10,
						Size:  hugepageSize,
					},
				},
			},
			NodeSelector: map[string]string{
				mcpPoolName: "",
			},
		},
	}

	return apiClient.Client.Create(context.TODO(), performanceProfile)
}

// CleanAllPerformanceProfile removes all PerformanceProfile from cluster.
func CleanAllPerformanceProfile(cnfNodeLabel string, snoTimeoutMultiplier time.Duration) error {
	performanceProfileList := &v2.PerformanceProfileList{}
	err := apiClient.Client.List(context.TODO(), performanceProfileList)

	if err != nil {
		return err
	}

	if len(performanceProfileList.Items) > 0 {
		for _, performanceProfile := range performanceProfileList.Items {
			err := apiClient.Client.Delete(
				context.TODO(),
				&performanceProfile)
			if err != nil {
				return err
			}
		}
		mcp := mco.NewMCPListBuilder(apiClient)
		err = mcp.WaitToBeStableFor()
		if err != nil {
			return err
		}
	}

	return nil
}

//err = mco.WaitToBeStableFor(cnfNodeLabel, snoTimeoutMultiplier)
