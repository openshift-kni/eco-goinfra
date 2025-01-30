package oran

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	hardwaremanagementv1alpha1 "github.com/openshift-kni/oran-o2ims/api/hardwaremanagement/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultNodePoolName      = "test-node-pool"
	defaultNodePoolNamespace = "test-namespace"
)

var hardwaremanagementTestSchemes = []clients.SchemeAttacher{
	hardwaremanagementv1alpha1.AddToScheme,
}

func TestPullNodePool(t *testing.T) {
	testCases := []struct {
		name                string
		nsname              string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultNodePoolName,
			nsname:              defaultNodePoolNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			nsname:              defaultNodePoolNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("nodePool 'name' cannot be empty"),
		},
		{
			name:                defaultNodePoolName,
			nsname:              "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("nodePool 'nsname' cannot be empty"),
		},
		{
			name:                defaultNodePoolName,
			nsname:              defaultNodePoolNamespace,
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf(
				"nodePool object %s does not exist in namespace %s", defaultNodePoolName, defaultNodePoolNamespace),
		},
		{
			name:                defaultNodePoolName,
			nsname:              defaultNodePoolNamespace,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("nodePool 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, buildDummyNodePool(defaultNodePoolName, defaultNodePoolNamespace))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: hardwaremanagementTestSchemes,
			})
		}

		testBuilder, err := PullNodePool(testSettings, testCase.name, testCase.nsname)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.nsname, testBuilder.Definition.Namespace)
		}
	}
}

func TestNodePoolGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *NodePoolBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidNodePoolTestBuilder(buildTestClientWithDummyNodePool()),
			expectedError: "",
		},
		{
			testBuilder: buildValidNodePoolTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Sprintf(
				"nodepools.o2ims-hardwaremanagement.oran.openshift.io \"%s\" not found", defaultNodePoolName),
		},
	}

	for _, testCase := range testCases {
		nodePool, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, nodePool.Name)
			assert.Equal(t, testCase.testBuilder.Definition.Namespace, nodePool.Namespace)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestNodePoolExists(t *testing.T) {
	testCases := []struct {
		testBuilder *NodePoolBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidNodePoolTestBuilder(buildTestClientWithDummyNodePool()),
			exists:      true,
		},
		{
			testBuilder: buildValidNodePoolTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

// buildDummyNodePool returns a NodePool with the provided name and nsname.
func buildDummyNodePool(name, nsname string) *hardwaremanagementv1alpha1.NodePool {
	return &hardwaremanagementv1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}

// buildTestClientWithDummyNodePool returns an apiClient with the correct schemes and a NodePool with default name and
// namespace.
func buildTestClientWithDummyNodePool() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyNodePool(defaultNodePoolName, defaultNodePoolNamespace),
		},
		SchemeAttachers: hardwaremanagementTestSchemes,
	})
}

// buildValidNodePoolTestBuilder returns a valid NodePoolBuilder with all defaults and the provided apiClient.
func buildValidNodePoolTestBuilder(apiClient *clients.Settings) *NodePoolBuilder {
	_ = apiClient.AttachScheme(hardwaremanagementv1alpha1.AddToScheme)

	return &NodePoolBuilder{
		Definition: buildDummyNodePool(defaultNodePoolName, defaultNodePoolNamespace),
		apiClient:  apiClient,
	}
}
