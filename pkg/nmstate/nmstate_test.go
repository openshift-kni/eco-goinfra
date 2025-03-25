package nmstate

import (
	"fmt"
	"testing"

	nmstatev1 "github.com/nmstate/kubernetes-nmstate/api/v1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	v1TestSchemes = []clients.SchemeAttacher{
		nmstatev1.AddToScheme,
	}
	defaultNMStateName = "nmstatename"
)

func TestPullNMState(t *testing.T) {
	generateNMState := func(name string) *nmstatev1.NMState {
		return &nmstatev1.NMState{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: nmstatev1.NMStateSpec{},
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
			expectedErrorText:   "nmState 'name' cannot be empty",
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "test1",
			expectedError:       true,
			expectedErrorText:   "nmState object test1 does not exist",
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

		testNmState := generateNMState(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testNmState)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: v1TestSchemes,
			})
		}

		// Test the Pull method
		builderResult, err := PullNMstate(testSettings, testCase.name)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			// Check the error message
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testNmState.Name, builderResult.Object.Name)
			assert.Equal(t, testNmState.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestNMStateNewBuilder(t *testing.T) {
	testCases := []struct {
		name              string
		expectedErrorText string
		client            bool
	}{
		{
			name:   "test1",
			client: true,
		},
		{
			name:              "",
			expectedErrorText: "NMState 'name' cannot be empty",
			client:            true,
		},
		{
			name:   "test1",
			client: true,
		},
		{
			name:   "test1",
			client: false,
		},
	}
	for _, testCase := range testCases {
		var (
			client *clients.Settings
		)

		if testCase.client {
			client = clients.GetTestClients(clients.TestClientParams{})
		}

		testNMState := NewBuilder(client, testCase.name)
		if testCase.client {
			assert.NotNil(t, testNMState)
		}

		if len(testCase.expectedErrorText) > 0 {
			assert.Equal(t, testCase.expectedErrorText, testNMState.errorMsg)
		}
	}
}

func TestNMStateGet(t *testing.T) {
	testCases := []struct {
		nmStateBuilder *Builder
		expectedError  error
	}{
		{
			nmStateBuilder: buildValidNMStateTestBuilder(buildTestClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			nmStateBuilder: buildInvalidNMStateTestBuilder(buildTestClientWithDummyObject()),
			expectedError:  fmt.Errorf("NMState 'name' cannot be empty"),
		},
		{
			nmStateBuilder: buildValidNMStateTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("nmstates.nmstate.io \"nmstatename\" not found"),
		},
	}

	for _, testCase := range testCases {
		nmState, err := testCase.nmStateBuilder.Get()
		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedError, err)
		}

		if testCase.expectedError == nil {
			assert.Equal(t, nmState.Name, testCase.nmStateBuilder.Definition.Name)
		}
	}
}

func TestNMStateCreate(t *testing.T) {
	testCases := []struct {
		testNMState   *Builder
		expectedError error
	}{
		{
			testNMState:   buildValidNMStateTestBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testNMState:   buildInvalidNMStateTestBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("NMState 'name' cannot be empty"),
		},
		{
			testNMState:   buildValidNMStateTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		nmStateBuilder, err := testCase.testNMState.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, nmStateBuilder.Definition.Name, nmStateBuilder.Object.Name)
		}
	}
}

func TestNMStateDelete(t *testing.T) {
	testCases := []struct {
		testNMState   *Builder
		expectedError error
	}{
		{
			testNMState:   buildValidNMStateTestBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testNMState:   buildInvalidNMStateTestBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("NMState 'name' cannot be empty"),
		},
		{
			testNMState:   buildValidNMStateTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testNMState.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testNMState.Object)
		}
	}
}

func TestNMStateUpdate(t *testing.T) {
	testCases := []struct {
		testNMState   *Builder
		expectedError error
		forceFlag     bool
	}{
		{
			testNMState:   buildValidNMStateTestBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
			forceFlag:     false,
		},
		{
			testNMState:   buildValidNMStateTestBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
			forceFlag:     true,
		},
		{
			testNMState:   buildInvalidNMStateTestBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("NMState 'name' cannot be empty"),
			forceFlag:     true,
		},
	}
	for _, testCase := range testCases {
		assert.Empty(t, testCase.testNMState.Definition.Spec.NodeSelector)
		assert.Nil(t, nil, testCase.testNMState.Object)
		testCase.testNMState.Definition.Spec.NodeSelector = map[string]string{"test": "test"}
		testCase.testNMState.Definition.ResourceVersion = "999"
		nmStateBuilder, err := testCase.testNMState.Update(testCase.forceFlag)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, map[string]string{"test": "test"}, testCase.testNMState.Object.Spec.NodeSelector)
			assert.Equal(t, nmStateBuilder.Definition, nmStateBuilder.Object)
		}
	}
}

func TestNMStateExist(t *testing.T) {
	testCases := []struct {
		testNMState    *Builder
		expectedStatus bool
	}{
		{
			testNMState:    buildValidNMStateTestBuilder(buildTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testNMState:    buildInvalidNMStateTestBuilder(buildTestClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testNMState:    buildValidNMStateTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testNMState.Exists()
		assert.Equal(t, testCase.expectedStatus, exists)
	}
}

// buildValidTestBuilder returns a valid Builder for testing purposes.
func buildValidNMStateTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultNMStateName)
}

// buildValidTestBuilder returns a valid Builder for testing purposes.
func buildInvalidNMStateTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "")
}

func buildTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyNMStateObject(),
		SchemeAttachers: v1TestSchemes,
	})
}

func buildDummyNMStateObject() []runtime.Object {
	return append([]runtime.Object{}, &nmstatev1.NMState{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultNMStateName,
		},
		Spec: nmstatev1.NMStateSpec{},
	})
}
