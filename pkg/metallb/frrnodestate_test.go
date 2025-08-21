package metallb

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/frrtypes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultNodeName = "worker-0"
)

func TestFrrNodeStateGet(t *testing.T) {
	var runtimeObjects []runtime.Object
	testCases := []struct {
		testFrrNodeState    *FrrNodeStateBuilder
		addToRuntimeObjects bool
		expectedError       string
		client              bool
	}{
		{
			testFrrNodeState: buildValidFrrNodeStateTestBuilder(buildTestFrrClientWithDummyNode(defaultNodeName)),
			expectedError:    "",
		},
		{
			testFrrNodeState: buildValidFrrNodeStateTestBuilder(clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: frrTestSchemes,
			})),
			expectedError: "frrnodestates.frrk8s.metallb.io \"worker-0\" not found",
		},
	}

	for _, testCase := range testCases {
		frrNodeState, err := testCase.testFrrNodeState.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, frrNodeState.Name, testCase.testFrrNodeState.Definition.Name, frrNodeState.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestFrrNodeStateExist(t *testing.T) {
	testCases := []struct {
		testFrrNodeState *FrrNodeStateBuilder
		exist            bool
	}{
		{
			testFrrNodeState: buildValidFrrNodeStateTestBuilder(buildTestFrrClientWithDummyNode("test-node")),
			exist:            false,
		},
		{
			testFrrNodeState: buildValidFrrNodeStateTestBuilder(buildTestFrrClientWithDummyNode(defaultNodeName)),
			exist:            true,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testFrrNodeState.Exists()
		assert.Equal(t, testCase.exist, exist)
	}
}

func TestPullFrrNodeState(t *testing.T) {
	generateFrrNodeState := func(name string) *frrtypes.FRRNodeState {
		return &frrtypes.FRRNodeState{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Status: frrtypes.FRRNodeStateStatus{},
		}
	}

	testCases := []struct {
		name                string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
		client              bool
	}{
		{
			name:                "test1",
			expectedError:       false,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "",
			expectedError:       true,
			expectedErrorText:   "frrNodeState 'name' cannot be empty",
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "test1",
			expectedError:       true,
			expectedErrorText:   "frrNodeState object test1 does not exist",
			addToRuntimeObjects: false,
			client:              true,
		},
		{
			name:                "test1",
			expectedError:       true,
			expectedErrorText:   "the apiClient cannot be nil",
			addToRuntimeObjects: true,
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testNodeNetState := generateFrrNodeState(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testNodeNetState)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: frrTestSchemes,
			})
		}

		// Test the Pull method
		builderResult, err := PullFrrNodeState(testSettings, testCase.name)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			// Check the error message
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testNodeNetState.Name, builderResult.Object.Name)
		}
	}
}

// buildDummyNode returns a Node with the provided name.
func buildDummyFRRNodeState(name string) *frrtypes.FRRNodeState {
	return &frrtypes.FRRNodeState{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestFrrClientWithDummyNode returns a client with a dummy node.
func buildTestFrrClientWithDummyNode(nodeName string) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{buildDummyFRRNodeState(nodeName)},
		SchemeAttachers: frrTestSchemes,
	})
}

func buildValidFrrNodeStateTestBuilder(apiClient *clients.Settings) *FrrNodeStateBuilder {
	return newFRRNodeStateBuilder(apiClient, defaultNodeName)
}

func newFRRNodeStateBuilder(apiClient *clients.Settings, frr string) *FrrNodeStateBuilder {
	if apiClient == nil {
		return nil
	}

	builder := FrrNodeStateBuilder{
		apiClient:  apiClient.Client,
		Definition: buildDummyFRRNodeState(frr),
	}

	return &builder
}
