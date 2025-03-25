package ocm

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	policiesv1beta1 "open-cluster-management.io/governance-policy-propagator/api/v1beta1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListPolicieSetsInAllNamespaces returns a cluster-wide policySets inventory.
func ListPolicieSetsInAllNamespaces(apiClient *clients.Settings,
	options ...runtimeclient.ListOptions) (
	[]*PolicySetBuilder, error) {
	if apiClient == nil {
		glog.V(100).Info("PolicySets 'apiClient' parameter cannot be nil")

		return nil, fmt.Errorf("failed to list policySets, 'apiClient' parameter is nil")
	}

	err := apiClient.AttachScheme(policiesv1beta1.AddToScheme)
	if err != nil {
		glog.V(100).Info("Failed to add PolicySet scheme to client schemes")

		return nil, err
	}

	logMessage := string("Listing all policySets in all namespaces")
	passedOptions := runtimeclient.ListOptions{}

	if len(options) > 1 {
		glog.V(100).Infof("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	}

	glog.V(100).Infof(logMessage)

	policySetList := new(policiesv1beta1.PolicySetList)
	err = apiClient.List(context.TODO(), policySetList, &passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list all policySets in all namespaces due to %s", err.Error())

		return nil, err
	}

	var policySetObjects []*PolicySetBuilder

	for _, policy := range policySetList.Items {
		copiedPolicySet := policy
		policySetBuilder := &PolicySetBuilder{
			apiClient:  apiClient.Client,
			Object:     &copiedPolicySet,
			Definition: &copiedPolicySet,
		}

		policySetObjects = append(policySetObjects, policySetBuilder)
	}

	return policySetObjects, nil
}
