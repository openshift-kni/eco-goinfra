package cgu

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/cluster-group-upgrades-operator/pkg/api/clustergroupupgrades/v1alpha1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListInAllNamespaces returns a cluster-wide cgu inventory.
func ListInAllNamespaces(apiClient *clients.Settings, options ...client.ListOptions) ([]*CguBuilder, error) {
	logMessage := "Listing CGUS in all namespaces"
	passedOptions := client.ListOptions{}

	if apiClient == nil {
		glog.V(100).Infof("CGUs 'apiClient' parameter can not be empty")

		return nil, fmt.Errorf("failed to list cgu objects, 'apiClient' parameter is empty")
	}

	err := apiClient.AttachScheme(v1alpha1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add cgu v1alpha1 scheme to client schemes")

		return nil, err
	}

	if len(options) > 1 {
		glog.V(100).Infof("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	}

	glog.V(100).Infof(logMessage)

	cguList := &v1alpha1.ClusterGroupUpgradeList{}
	err = apiClient.Client.List(context.TODO(), cguList, &passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list all CGUs in all namespaces due to %s", err.Error())

		return nil, err
	}

	var cguObjects []*CguBuilder

	for _, policy := range cguList.Items {
		copiedCgu := policy
		cguBuilder := &CguBuilder{
			apiClient:  apiClient.Client,
			Object:     &copiedCgu,
			Definition: &copiedCgu,
		}

		cguObjects = append(cguObjects, cguBuilder)
	}

	return cguObjects, nil
}
