package ocm

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	policiesv1 "open-cluster-management.io/governance-policy-propagator/api/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListPoliciesInAllNamespaces returns a cluster-wide policy inventory.
func ListPoliciesInAllNamespaces(apiClient *clients.Settings,
	options ...runtimeclient.ListOptions) (
	[]*PolicyBuilder, error) {
	logMessage := string("Listing all policies in all namespaces")
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

	policyList := &policiesv1.PolicyList{}

	err := apiClient.Client.List(context.TODO(), policyList, &passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list all policies in all namespaces due to %s", err.Error())

		return nil, err
	}

	var policyObjects []*PolicyBuilder

	for _, policy := range policyList.Items {
		copiedPolicy := policy
		policyBuilder := &PolicyBuilder{
			apiClient:  apiClient,
			Object:     &copiedPolicy,
			Definition: &copiedPolicy,
		}

		policyObjects = append(policyObjects, policyBuilder)
	}

	return policyObjects, nil
}
