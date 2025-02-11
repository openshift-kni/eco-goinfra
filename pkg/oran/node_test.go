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
	defaultNodeName      = "test-node"
	defaultNodeNamespace = "test-namespace"
)

func TestPullNode(t *testing.T) {
	testCases := []struct {
		name                string
		nsname              string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultNodeName,
			nsname:              defaultNodeNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			nsname:              defaultNodeNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("node 'name' cannot be empty"),
		},
		{
			name:                defaultNodeName,
			nsname:              "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("node 'nsname' cannot be empty"),
		},
		{
			name:                defaultNodeName,
			nsname:              defaultNodeNamespace,
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf(
				"node object %s does not exist in namespace %s", defaultNodeName, defaultNodeNamespace),
		},
		{
			name:                defaultNodeName,
			nsname:              defaultNodeNamespace,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("node 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, buildDummyNode(defaultNodeName, defaultNodeNamespace))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: hardwaremanagementTestSchemes,
			})
		}

		testBuilder, err := PullNode(testSettings, testCase.name, testCase.nsname)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.nsname, testBuilder.Definition.Namespace)
		}
	}
}

func TestNodeGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *NodeBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidNodeTestBuilder(buildTestClientWithDummyNode()),
			expectedError: "",
		},
		{
			testBuilder: buildValidNodeTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Sprintf(
				"nodes.o2ims-hardwaremanagement.oran.openshift.io \"%s\" not found", defaultNodeName),
		},
	}

	for _, testCase := range testCases {
		node, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, node.Name)
			assert.Equal(t, testCase.testBuilder.Definition.Namespace, node.Namespace)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestNodeExists(t *testing.T) {
	testCases := []struct {
		testBuilder *NodeBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidNodeTestBuilder(buildTestClientWithDummyNode()),
			exists:      true,
		},
		{
			testBuilder: buildValidNodeTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

// buildDummyNode returns a Node with the provided name and nsname.
func buildDummyNode(name, nsname string) *hardwaremanagementv1alpha1.Node {
	return &hardwaremanagementv1alpha1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}

// buildTestClientWithDummyNode returns an apiClient with the correct schemes and a Node with default name and
// namespace.
func buildTestClientWithDummyNode() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyNode(defaultNodeName, defaultNodeNamespace),
		},
		SchemeAttachers: hardwaremanagementTestSchemes,
	})
}

// buildValidNodeTestBuilder returns a valid NodeBuilder with all defaults and the provided apiClient.
func buildValidNodeTestBuilder(apiClient *clients.Settings) *NodeBuilder {
	_ = apiClient.AttachScheme(hardwaremanagementv1alpha1.AddToScheme)

	return &NodeBuilder{
		Definition: buildDummyNode(defaultNodeName, defaultNodeNamespace),
		apiClient:  apiClient,
	}
}
