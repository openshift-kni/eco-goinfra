package oran

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	pluginsv1alpha1 "github.com/openshift-kni/oran-o2ims/api/hardwaremanagement/plugins/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultAllocatedNodeName      = "test-node"
	defaultAllocatedNodeNamespace = "test-namespace"
)

var pluginsTestSchemes = []clients.SchemeAttacher{
	pluginsv1alpha1.AddToScheme,
}

func TestPullAllocatedNode(t *testing.T) {
	testCases := []struct {
		name                string
		nsname              string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultAllocatedNodeName,
			nsname:              defaultAllocatedNodeNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			nsname:              defaultAllocatedNodeNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("allocatedNode 'name' cannot be empty"),
		},
		{
			name:                defaultAllocatedNodeName,
			nsname:              "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("allocatedNode 'nsname' cannot be empty"),
		},
		{
			name:                defaultAllocatedNodeName,
			nsname:              defaultAllocatedNodeNamespace,
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf(
				"allocatedNode object %s does not exist in namespace %s", defaultAllocatedNodeName, defaultAllocatedNodeNamespace),
		},
		{
			name:                defaultAllocatedNodeName,
			nsname:              defaultAllocatedNodeNamespace,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("allocatedNode 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects,
				buildDummyAllocatedNode(defaultAllocatedNodeName, defaultAllocatedNodeNamespace))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: pluginsTestSchemes,
			})
		}

		testBuilder, err := PullAllocatedNode(testSettings, testCase.name, testCase.nsname)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.nsname, testBuilder.Definition.Namespace)
		}
	}
}

func TestAllocatedNodeGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *AllocatedNodeBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidAllocatedNodeTestBuilder(buildTestClientWithDummyAllocatedNode()),
			expectedError: "",
		},
		{
			testBuilder: buildValidAllocatedNodeTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Sprintf(
				"allocatednodes.plugins.clcm.openshift.io \"%s\" not found", defaultAllocatedNodeName),
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

func TestAllocatedNodeExists(t *testing.T) {
	testCases := []struct {
		testBuilder *AllocatedNodeBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidAllocatedNodeTestBuilder(buildTestClientWithDummyAllocatedNode()),
			exists:      true,
		},
		{
			testBuilder: buildValidAllocatedNodeTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

// buildDummyAllocatedNode returns an AllocatedNode with the provided name and nsname.
func buildDummyAllocatedNode(name, nsname string) *pluginsv1alpha1.AllocatedNode {
	return &pluginsv1alpha1.AllocatedNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}

// buildTestClientWithDummyAllocatedNode returns an apiClient with the correct schemes and an AllocatedNode with default
// name and namespace.
func buildTestClientWithDummyAllocatedNode() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyAllocatedNode(defaultAllocatedNodeName, defaultAllocatedNodeNamespace),
		},
		SchemeAttachers: pluginsTestSchemes,
	})
}

// buildValidAllocatedNodeTestBuilder returns a valid AllocatedNodeBuilder with all defaults and the provided apiClient.
func buildValidAllocatedNodeTestBuilder(apiClient *clients.Settings) *AllocatedNodeBuilder {
	_ = apiClient.AttachScheme(pluginsv1alpha1.AddToScheme)

	return &AllocatedNodeBuilder{
		Definition: buildDummyAllocatedNode(defaultAllocatedNodeName, defaultAllocatedNodeNamespace),
		apiClient:  apiClient,
	}
}
