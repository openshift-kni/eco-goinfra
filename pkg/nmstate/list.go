package nmstate

import (
	"context"

	"github.com/golang/glog"
	nmstateV1 "github.com/nmstate/kubernetes-nmstate/api/v1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	assistedv1beta1 "github.com/openshift/assisted-service/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListPolicy returns a list of NodeNetworkConfigurationPolicy.
func ListPolicy(apiClient *clients.Settings) ([]*PolicyBuilder, error) {
	glog.V(100).Infof("Listing NodeNetworkConfigurationPolicy")

	policyList := &nmstateV1.NodeNetworkConfigurationPolicyList{}
	err := apiClient.Client.List(context.Background(), policyList)

	if err != nil {
		glog.V(100).Infof("Failed to list NodeNetworkConfigurationPolicy due to %s", err.Error())

		return nil, err
	}

	var networkConfigurationPolicyObjects []*PolicyBuilder

	for _, policy := range policyList.Items {
		copiedPolicy := policy
		policyBuilder := &PolicyBuilder{
			apiClient:  apiClient,
			Definition: &copiedPolicy,
			Object:     &copiedPolicy}

		networkConfigurationPolicyObjects = append(networkConfigurationPolicyObjects, policyBuilder)
	}

	return networkConfigurationPolicyObjects, nil
}

// ListNmStateConfig returns a NMStateConfig list.
func ListNmStateConfig(apiClient *clients.Settings) ([]*NmStateConfigBuilder, error) {
	nmStateConfigList := &assistedv1beta1.NMStateConfigList{}

	err := apiClient.List(context.Background(), nmStateConfigList, &client.ListOptions{})
	if err != nil {
		glog.V(100).Infof("Failed to list nmStateConfig due to %s", err.Error())

		return nil, err
	}

	var nmstateConfigObjects []*NmStateConfigBuilder

	for _, nmStateConfigObj := range nmStateConfigList.Items {
		nmStateConf := nmStateConfigObj
		nmStateConfBuilder := &NmStateConfigBuilder{
			apiClient:  apiClient,
			Definition: &nmStateConf,
			Object:     &nmStateConf,
		}

		nmstateConfigObjects = append(nmstateConfigObjects, nmStateConfBuilder)
	}

	return nmstateConfigObjects, err
}
