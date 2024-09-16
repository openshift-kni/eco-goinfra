package nodes

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultNodeName         = "test-node"
	defaultNodeLabel        = "node-role.kubernetes.io/control-plane"
	defaultExternalNetworks = `{"ipv4":"10.0.0.0/8","ipv6":"fd00::/8"}`
	defaultExternalIPv4     = "10.0.0.0/8"
)

func TestNodePull(t *testing.T) {
	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultNodeName,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("node 'name' cannot be empty"),
		},
		{
			name:                defaultNodeName,
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("node object %s does not exist", defaultNodeName),
		},
		{
			name:                defaultNodeName,
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
			runtimeObjects = append(runtimeObjects, buildDummyNode(testCase.name))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		testBuilder, err := Pull(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
		}
	}
}

func TestNodeUpdate(t *testing.T) {
	testCases := []struct {
		alreadyExists bool
		expectedError error
	}{
		{
			alreadyExists: true,
			expectedError: nil,
		},
		{
			alreadyExists: false,
			expectedError: fmt.Errorf("node %s object does not exist", defaultNodeName),
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.alreadyExists {
			runtimeObjects = append(runtimeObjects, buildDummyNode(defaultNodeName))
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		testBuilder := buildValidNodeTestBuilder(testSettings)
		assert.False(t, testBuilder.Definition.Spec.Unschedulable)

		testBuilder.Definition.ResourceVersion = "999"
		testBuilder.Definition.Spec.Unschedulable = true

		testBuilder, err := testBuilder.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.True(t, testBuilder.Object.Spec.Unschedulable)
		}
	}
}

func TestNodeExists(t *testing.T) {
	testCases := []struct {
		testBuilder *Builder
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

func TestNodeDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *Builder
		expectedError error
	}{
		{
			testBuilder:   buildValidNodeTestBuilder(buildTestClientWithDummyNode()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidNodeTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestNodeWithNewLabel(t *testing.T) {
	testCases := []struct {
		key           string
		keyExists     bool
		expectedError string
	}{
		{
			key:           defaultNodeLabel,
			keyExists:     false,
			expectedError: "",
		},
		{
			key:           "",
			keyExists:     false,
			expectedError: "error to set empty key to node",
		},
		{
			key:           defaultNodeLabel,
			keyExists:     true,
			expectedError: fmt.Sprintf("cannot overwrite existing node label: %s", defaultNodeLabel),
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidNodeTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

		if testCase.keyExists {
			testBuilder.Definition.Labels = map[string]string{testCase.key: ""}
		}

		testBuilder = testBuilder.WithNewLabel(testCase.key, "")
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)
	}
}

func TestNodeWithOptions(t *testing.T) {
	testCases := []struct {
		testBuilder   *Builder
		options       AdditionalOptions
		expectedError string
	}{
		{
			testBuilder: buildValidNodeTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			options: func(builder *Builder) (*Builder, error) {
				builder.Definition.Spec.Unschedulable = true

				return builder, nil
			},
			expectedError: "",
		},
		{
			testBuilder: buildValidNodeTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			options: func(builder *Builder) (*Builder, error) {
				return builder, fmt.Errorf("error adding additional option")
			},
			expectedError: "error adding additional option",
		},
	}

	for _, testCase := range testCases {
		testBuilder := testCase.testBuilder.WithOptions(testCase.options)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.True(t, testBuilder.Definition.Spec.Unschedulable)
		}
	}
}

func TestNodeRemoveLabel(t *testing.T) {
	testCases := []struct {
		key           string
		expectedError string
	}{
		{
			key:           defaultNodeLabel,
			expectedError: "",
		},
		{
			key:           "",
			expectedError: "error to remove empty key from node",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidNodeTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		testBuilder = testBuilder.RemoveLabel(testCase.key, "")

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.NotContains(t, testBuilder.Definition.Labels, testCase.key)
		}
	}
}

func TestNodeExternalIPv4Network(t *testing.T) {
	testCases := []struct {
		objectNil         bool
		externalAddresses string
		expectedError     error
	}{
		{
			objectNil:         false,
			externalAddresses: defaultExternalNetworks,
			expectedError:     nil,
		},
		{
			objectNil:         true,
			externalAddresses: defaultExternalNetworks,
			expectedError:     fmt.Errorf("cannot collect external networks when node object is nil"),
		},
		{
			objectNil:         false,
			externalAddresses: "",
			expectedError:     fmt.Errorf("node %s does not have external addresses annotation", defaultNodeName),
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidNodeTestBuilder(buildTestClientWithDummyNode())

		if !testCase.objectNil {
			testBuilder.Object = testBuilder.Definition

			if testCase.externalAddresses != "" {
				testBuilder.Object.Annotations = map[string]string{ovnExternalAddresses: testCase.externalAddresses}
			}
		}

		externalIPv4, err := testBuilder.ExternalIPv4Network()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, defaultExternalIPv4, externalIPv4)
		}
	}
}

func TestNodeIsReady(t *testing.T) {
	testCases := []struct {
		exists         bool
		readyCondition bool
		ready          bool
		expectedError  error
	}{
		{
			exists:         true,
			readyCondition: true,
			ready:          true,
			expectedError:  nil,
		},
		{
			exists:         false,
			readyCondition: true,
			ready:          true,
			expectedError:  fmt.Errorf("node object %s does not exist", defaultNodeName),
		},
		{
			exists:         true,
			readyCondition: false,
			ready:          true,
			expectedError:  fmt.Errorf("the Ready condition could not be found for node %s", defaultNodeName),
		},
		{
			exists:         true,
			readyCondition: true,
			ready:          false,
			expectedError:  nil,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.exists {
			if testCase.readyCondition {
				runtimeObjects = append(runtimeObjects, buildDummyNodeWithReadiness(defaultNodeName, testCase.ready))
			} else {
				runtimeObjects = append(runtimeObjects, buildDummyNode(defaultNodeName))
			}
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})
		testBuilder := buildValidNodeTestBuilder(testSettings)

		ready, err := testBuilder.IsReady()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.ready, ready)
		}
	}
}

func TestNodeWaitUntilConditionTrue(t *testing.T) {
	testNodeWaitUntilConditionHelper(t,
		func(testBuilder *Builder, conditionType corev1.NodeConditionType, timeout time.Duration) error {
			return testBuilder.WaitUntilConditionTrue(conditionType, timeout)
		})
}

func TestNodeWaitUntilConditionUnknown(t *testing.T) {
	testNodeWaitUntilConditionHelper(t,
		func(testBuilder *Builder, conditionType corev1.NodeConditionType, timeout time.Duration) error {
			return testBuilder.WaitUntilConditionUnknown(conditionType, timeout)
		})
}

func TestNodeWaitUntilReady(t *testing.T) {
	testNodeWaitUntilConditionHelper(t,
		func(testBuilder *Builder, conditionType corev1.NodeConditionType, timeout time.Duration) error {
			return testBuilder.WaitUntilReady(timeout)
		})
}

func TestNodeWaitUntilNotReady(t *testing.T) {
	testNodeWaitUntilConditionHelper(t,
		func(testBuilder *Builder, conditionType corev1.NodeConditionType, timeout time.Duration) error {
			return testBuilder.WaitUntilNotReady(timeout)
		})
}

func TestNodeValidate(t *testing.T) {
	testCases := []struct {
		builderNil      bool
		definitionNil   bool
		apiClientNil    bool
		builderErrorMsg string
		expectedError   error
	}{
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   nil,
		},
		{
			builderNil:      true,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("error: received nil node builder"),
		},
		{
			builderNil:      false,
			definitionNil:   true,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("can not redefine the undefined node"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    true,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("node builder cannot have nil apiClient"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "test error",
			expectedError:   fmt.Errorf("test error"),
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidNodeTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		if testCase.builderErrorMsg != "" {
			testBuilder.errorMsg = testCase.builderErrorMsg
		}

		valid, err := testBuilder.validate()

		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, testCase.expectedError == nil, valid)
	}
}

func testNodeWaitUntilConditionHelper(
	t *testing.T,
	testedFunc func(testBuilder *Builder, conditionType corev1.NodeConditionType, timeout time.Duration) error) {
	t.Helper()

	testCases := []struct {
		exists        bool
		hasCondition  bool
		conditionTrue bool
		expectedError error
	}{
		{
			exists:        true,
			hasCondition:  true,
			conditionTrue: true,
			expectedError: nil,
		},
		{
			exists:        false,
			hasCondition:  true,
			conditionTrue: true,
			expectedError: fmt.Errorf("node %s object does not exist", defaultNodeName),
		},
		{
			exists:        true,
			hasCondition:  false,
			conditionTrue: true,
			expectedError: fmt.Errorf("the %s condition could not be found for node %s", corev1.NodeReady, defaultNodeName),
		},
		{
			exists:        true,
			hasCondition:  true,
			conditionTrue: false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.exists {
			node := buildDummyNode(defaultNodeName)

			if testCase.hasCondition {
				status := corev1.ConditionUnknown

				if testCase.conditionTrue {
					status = corev1.ConditionTrue
				}

				node = buildDummyNodeWithCondition(defaultNodeName, corev1.NodeReady, status)
			}

			runtimeObjects = append(runtimeObjects, node)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})
		testBuilder := buildValidNodeTestBuilder(testSettings)

		err := testedFunc(testBuilder, corev1.NodeReady, time.Second)
		assert.Equal(t, testCase.expectedError, err)
	}
}

// buildDummyNode returns a Node with the provided name.
func buildDummyNode(name string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildDummyNodeWithCondition returns a Node with one provided condition.
func buildDummyNodeWithCondition(
	name string, conditionType corev1.NodeConditionType, status corev1.ConditionStatus) *corev1.Node {
	node := buildDummyNode(name)

	node.Status.Conditions = []corev1.NodeCondition{{
		Type:   conditionType,
		Status: status,
	}}

	return node
}

// buildTestClientWithDummyNode returns a client with a dummy node.
func buildTestClientWithDummyNode() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{buildDummyNode(defaultNodeName)},
	})
}

func buildValidNodeTestBuilder(apiClient *clients.Settings) *Builder {
	return newNodeBuilder(apiClient, defaultNodeName)
}

// newNodeBuilder creates a new Builder instances for testing purposes.
func newNodeBuilder(apiClient *clients.Settings, name string) *Builder {
	if apiClient == nil {
		return nil
	}

	builder := Builder{
		apiClient:  apiClient.K8sClient,
		Definition: buildDummyNode(name),
	}

	return &builder
}
