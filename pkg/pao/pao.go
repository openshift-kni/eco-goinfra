package pao

import (
	"context"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	v2 "github.com/openshift/cluster-node-tuning-operator/pkg/apis/performanceprofile/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
