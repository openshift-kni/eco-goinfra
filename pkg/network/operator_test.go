package network

import (
	"context"
	"fmt"
	"testing"
	"time"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

var operatorTestSchemes = []clients.SchemeAttacher{
	operatorv1.Install,
}

func TestPullOperator(t *testing.T) {
	testCases := []struct {
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("network.operator object %s does not exist", clusterNetworkName),
		},
		{
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("network.operator 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, buildDummyNetworkOperator())
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: operatorTestSchemes,
			})
		}

		testBuilder, err := PullOperator(testSettings)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, clusterNetworkName, testBuilder.Definition.Name)
		}
	}
}

func TestOperatorGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *OperatorBuilder
		expectedError string
	}{
		{
			testBuilder:   newOperatorBuilder(buildTestClientWithDummyNetworkOperator()),
			expectedError: "",
		},
		{
			testBuilder:   newOperatorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "networks.operator.openshift.io \"cluster\" not found",
		},
	}

	for _, testCase := range testCases {
		networkOperator, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, clusterNetworkName, networkOperator.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestOperatorExists(t *testing.T) {
	testCases := []struct {
		testBuilder *OperatorBuilder
		exists      bool
	}{
		{
			testBuilder: newOperatorBuilder(buildTestClientWithDummyNetworkOperator()),
			exists:      true,
		},
		{
			testBuilder: newOperatorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestOperatorUpdate(t *testing.T) {
	testCases := []struct {
		testBuilder   *OperatorBuilder
		expectedError error
	}{
		{
			testBuilder:   newOperatorBuilder(buildTestClientWithDummyNetworkOperator()),
			expectedError: nil,
		},
		{
			testBuilder:   newOperatorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("network.operator object %s does not exist", clusterNetworkName),
		},
	}

	for _, testCase := range testCases {
		assert.Nil(t, testCase.testBuilder.Definition.Spec.DeployKubeProxy)

		testCase.testBuilder.Definition.ResourceVersion = "999"
		testCase.testBuilder.Definition.Spec.DeployKubeProxy = ptr.To(true)

		testBuilder, err := testCase.testBuilder.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, true, *testBuilder.Definition.Spec.DeployKubeProxy)
		}
	}
}

func TestOperatorWaitUntilInCondition(t *testing.T) {
	testCases := []struct {
		exists        bool
		inCondition   bool
		expectedError error
	}{
		{
			exists:        true,
			inCondition:   true,
			expectedError: nil,
		},
		{
			exists:        false,
			inCondition:   true,
			expectedError: fmt.Errorf("network.operator object %s does not exist", clusterNetworkName),
		},
		{
			exists:        true,
			inCondition:   false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.exists {
			networkOperator := buildDummyNetworkOperator()

			if testCase.inCondition {
				networkOperator.Status.Conditions = []operatorv1.OperatorCondition{{
					Type:   operatorv1.OperatorStatusTypeAvailable,
					Status: operatorv1.ConditionTrue,
				}}
			}

			runtimeObjects = append(runtimeObjects, networkOperator)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: operatorTestSchemes,
		})
		testBuilder := newOperatorBuilder(testSettings)

		err := testBuilder.WaitUntilInCondition(operatorv1.OperatorStatusTypeAvailable, time.Second, operatorv1.ConditionTrue)
		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestOperatorValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		errorMsg      string
		expectedError error
	}{
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			errorMsg:      "",
			expectedError: nil,
		},
		{
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			errorMsg:      "",
			expectedError: fmt.Errorf("error: received nil network.operator builder"),
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			errorMsg:      "",
			expectedError: fmt.Errorf("can not redefine the undefined network.operator"),
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			errorMsg:      "",
			expectedError: fmt.Errorf("network.operator builder cannot have nil apiClient"),
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			errorMsg:      "test error",
			expectedError: fmt.Errorf("test error"),
		},
	}

	for _, testCase := range testCases {
		testBuilder := newOperatorBuilder(clients.GetTestClients(clients.TestClientParams{}))

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		if testCase.errorMsg != "" {
			testBuilder.errorMsg = testCase.errorMsg
		}

		valid, err := testBuilder.validate()

		if testCase.expectedError != nil {
			assert.False(t, valid)
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.True(t, valid)
			assert.Nil(t, err)
		}
	}
}

// buildDummyNetworkOperator builds a dummy network.operator object. It uses the clusterNetworkName.
func buildDummyNetworkOperator() *operatorv1.Network {
	return &operatorv1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterNetworkName,
		},
	}
}

// buildTestClientWithDummyNetworkOperator returns a client with a mock network.operator.
func buildTestClientWithDummyNetworkOperator() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyNetworkOperator(),
		},
		SchemeAttachers: operatorTestSchemes,
	})
}

// newOperatorBuilder returns a new OperatorBuilder for testing. It does not validate the apiClient.
func newOperatorBuilder(apiClient *clients.Settings) *OperatorBuilder {
	return &OperatorBuilder{
		apiClient:  apiClient.Client,
		Definition: buildDummyNetworkOperator(),
	}
}
