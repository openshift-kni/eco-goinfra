package nmstate

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	nmstateV1 "github.com/nmstate/kubernetes-nmstate/api/v1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListPolicy returns a list of NodeNetworkConfigurationPolicy.
func ListPolicy(apiClient *clients.Settings, options ...goclient.ListOptions) ([]*PolicyBuilder, error) {
	glog.V(100).Infof("Listing NodeNetworkConfigurationPolicy")

	passedOptions := goclient.ListOptions{}

	if len(options) == 1 {
		passedOptions = options[0]
	} else if len(options) > 1 {

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	policyList := &nmstateV1.NodeNetworkConfigurationPolicyList{}
	err := apiClient.Client.List(context.Background(), policyList, &passedOptions)

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

// CleanAllNMStatePolicies removes all NodeNetworkConfigurationPolicies.
func CleanAllNMStatePolicies(apiClient *clients.Settings, options ...goclient.ListOptions) error {
	glog.V(100).Infof("Cleaning up NodeNetworkConfigurationPolicies")

	nncpList, err := ListPolicy(apiClient, options...)
	if err != nil {
		glog.V(100).Infof("Failed to list NodeNetworkConfigurationPolicies")

		return err
	}

	for _, nncpPolicy := range nncpList {
		_, err = nncpPolicy.Delete()
		if err != nil {
			glog.V(100).Infof("Failed to delete NodeNetworkConfigurationPolicy: %s", nncpPolicy.Object.Name)

			return err
		}
	}

	return nil
}
