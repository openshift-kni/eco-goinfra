package hive

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListClusterDeploymentsInAllNamespaces(t *testing.T) {
	testCases := []struct {
		clusterDeployment []*ClusterDeploymentBuilder
		listOptions       []client.ListOptions
		expectedError     error
		client            bool
	}{
		{
			clusterDeployment: []*ClusterDeploymentBuilder{buildValidClusterDeploymentBuilder(
				buildClusterDeploymentClientWithDummyObject())},
			expectedError: nil,
			client:        true,
		},
		{
			clusterDeployment: []*ClusterDeploymentBuilder{buildValidClusterDeploymentBuilder(
				buildClusterDeploymentClientWithDummyObject())},
			listOptions: []client.ListOptions{{Continue: "test"}},
			client:      true,
		},
		{
			clusterDeployment: []*ClusterDeploymentBuilder{buildValidClusterDeploymentBuilder(
				buildClusterDeploymentClientWithDummyObject())},
			listOptions:   []client.ListOptions{{Namespace: "test"}, {Continue: "true"}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			clusterDeployment: []*ClusterDeploymentBuilder{buildValidClusterDeploymentBuilder(
				buildClusterDeploymentClientWithDummyObject())},
			expectedError: fmt.Errorf("the apiClient cannot be nil"),
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyClusterDeployment(),
				SchemeAttachers: testSchemes,
			})
		}

		deploymentBuilder, err := ListClusterDeploymentsInAllNamespaces(testSettings, testCase.listOptions...)
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(deploymentBuilder), len(testCase.clusterDeployment))
		}
	}
}
