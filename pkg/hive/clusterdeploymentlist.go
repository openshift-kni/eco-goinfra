package hive

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	hiveV1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/hive/api/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListClusterDeploymentsInAllNamespaces returns a cluster-wide clusterdeployment inventory.
func ListClusterDeploymentsInAllNamespaces(
	apiClient *clients.Settings,
	options ...goclient.ListOptions) ([]*ClusterDeploymentBuilder, error) {
	passedOptions := goclient.ListOptions{}
	logMessage := "Listing all clusterdeployments"

	if len(options) > 1 {
		glog.V(100).Infof("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	}

	glog.V(100).Infof(logMessage)

	clusterDeployments := new(hiveV1.ClusterDeploymentList)
	err := apiClient.List(context.TODO(), clusterDeployments, &passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list all clusterDeployments due to %s", err.Error())

		return nil, err
	}

	var clusterDeploymentObjects []*ClusterDeploymentBuilder

	for _, clusterDeployment := range clusterDeployments.Items {
		copiedClusterDeployment := clusterDeployment
		clusterDeploymentBuilder := &ClusterDeploymentBuilder{
			apiClient:  apiClient,
			Object:     &copiedClusterDeployment,
			Definition: &copiedClusterDeployment,
		}

		clusterDeploymentObjects = append(clusterDeploymentObjects, clusterDeploymentBuilder)
	}

	return clusterDeploymentObjects, nil
}
