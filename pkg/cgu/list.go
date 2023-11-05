package cgu

import (
	"context"

	"github.com/golang/glog"
	"github.com/openshift-kni/cluster-group-upgrades-operator/api/v1alpha1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListCgusInAllNamespaces returns a cluster-wide cgu inventory.
func ListCgusInAllNamespaces(apiClient *clients.Settings, options metaV1.ListOptions) ([]*CguBuilder, error) {
	glog.V(100).Info("Listing all CGUs in all namespaces")

	policyList := &v1alpha1.ClusterGroupUpgradeList{}

	err := apiClient.Client.List(context.Background(), policyList, &runtimeclient.ListOptions{})

	if err != nil {
		glog.V(100).Infof("Failed to list all CGUs in all namespaces due to %s", err.Error())

		return nil, err
	}

	var cguObjects []*CguBuilder

	for _, policy := range policyList.Items {
		copiedCgu := policy
		cguBuilder := &CguBuilder{
			apiClient:  apiClient,
			Object:     &copiedCgu,
			Definition: &copiedCgu,
		}

		cguObjects = append(cguObjects, cguBuilder)
	}

	return cguObjects, nil
}
