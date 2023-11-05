package ocm

import (
	"context"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	policiesv1 "open-cluster-management.io/governance-policy-propagator/api/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListPoliciesInAllNamespaces returns a cluster-wide policy inventory.
func ListPoliciesInAllNamespaces(apiClient *clients.Settings, options metaV1.ListOptions) ([]*PolicyBuilder, error) {
	glog.V(100).Info("Listing all policies in all namespaces")

	policyList := &policiesv1.PolicyList{}

	err := apiClient.Client.List(context.Background(), policyList, &runtimeclient.ListOptions{})

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
