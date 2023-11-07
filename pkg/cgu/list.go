package cgu

import (
	"context"

	"github.com/golang/glog"
	"github.com/openshift-kni/cluster-group-upgrades-operator/api/v1alpha1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListInAllNamespaces returns a cluster-wide cgu inventory.
func ListInAllNamespaces(apiClient *clients.Settings, options runtimeclient.ListOptions) ([]*CguBuilder, error) {
	glog.V(100).Info("Listing all CGUs in all namespaces")

	cguList := &v1alpha1.ClusterGroupUpgradeList{}

	err := apiClient.Client.List(context.Background(), cguList, &options)

	if err != nil {
		glog.V(100).Infof("Failed to list all CGUs in all namespaces due to %s", err.Error())

		return nil, err
	}

	var cguObjects []*CguBuilder

	for _, policy := range cguList.Items {
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
