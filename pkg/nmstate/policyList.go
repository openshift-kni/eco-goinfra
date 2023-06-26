package nmstate

import (
	"context"
	"fmt"

	"github.com/golang/glog"

	nmstateV1 "github.com/nmstate/kubernetes-nmstate/api/v1"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
)

// PolicyBuilderList provides struct for NodeNetworkConfigurationPolicyList object
// which contains connection to cluster and NodeNetworkConfigurationPolicyList definitions.
type PolicyBuilderList struct {
	// Dynamically discovered NodeNetworkConfigurationPolicyList object.
	Objects *nmstateV1.NodeNetworkConfigurationPolicyList
	// apiClient opens api connection to the cluster.
	apiClient *clients.Settings
	// errorMsg used in discovery function before sending api request to cluster.
	errorMsg string
}

// NewPolicyBuilderList creates new instance of NodeNetworkConfigurationPolicyList.
func NewPolicyBuilderList(apiClient *clients.Settings) *PolicyBuilderList {
	glog.V(100).Infof(
		"Initializing new PolicyBuilderList structure")

	builder := &PolicyBuilderList{
		apiClient: apiClient,
		Objects:   &nmstateV1.NodeNetworkConfigurationPolicyList{},
	}

	return builder
}

// Discover method gets the NodeNetworkConfigurationPolicyList items and stores them in the PolicyBuilderList struct.
func (builder *PolicyBuilderList) Discover() error {
	if builder.errorMsg != "" {
		return fmt.Errorf(builder.errorMsg)
	}

	glog.V(100).Infof("Getting the NodeNetworkConfigurationPolicyList object)")

	err := builder.apiClient.List(context.TODO(), builder.Objects)

	return err
}
