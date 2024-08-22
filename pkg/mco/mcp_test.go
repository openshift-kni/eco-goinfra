package mco

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	mcv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const defaultMCPName = "test-machine-config-pool"

var defaultMCPCondition = mcv1.MachineConfigPoolCondition{
	Type:   mcv1.MachineConfigPoolBuildSuccess,
	Status: corev1.ConditionTrue,
}

func TestNewMCPBuilder(t *testing.T) {
	testCases := []struct {
		name              string
		client            bool
		expectedErrorText string
	}{
		{
			name:              defaultMCPName,
			client:            true,
			expectedErrorText: "",
		},
		{
			name:              "",
			client:            true,
			expectedErrorText: "machineconfigpool 'name' cannot be empty",
		},
		{
			name:              defaultMCPName,
			client:            false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewMCPBuilder(testSettings, testCase.name)

		if testCase.client {
			assert.Equal(t, testCase.expectedErrorText, testBuilder.errorMsg)

			if testCase.expectedErrorText == "" {
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			}
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestPullMachineConfigPool(t *testing.T) {
	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultMCPName,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("machineconfigpool 'name' cannot be empty"),
		},
		{
			name:                defaultMCPName,
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("machineconfigpool object %s does not exist", defaultMCPName),
		},
		{
			name:                defaultMCPName,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("machineconfigpool 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testMC := buildDummyMCP(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testMC)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := Pull(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testMC.Name, testBuilder.Definition.Name)
		}
	}
}

func TestMachineConfigPoolGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *MCPBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidMCPTestBuilder(buildTestClientWithDummyMCP()),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidMCPTestBuilder(buildTestClientWithDummyMCP()),
			expectedError: "machineconfigpool 'name' cannot be empty",
		},
		{
			testBuilder:   buildValidMCPTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "machineconfigpools.machineconfiguration.openshift.io \"test-machine-config-pool\" not found",
		},
	}

	for _, testCase := range testCases {
		machineConfig, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, machineConfig.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestMachineConfigPoolCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *MCPBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidMCPTestBuilder(buildTestClientWithDummyMCP()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidMCPTestBuilder(buildTestClientWithDummyMCP()),
			expectedError: fmt.Errorf("machineconfigpool 'name' cannot be empty"),
		},
		{
			testBuilder:   buildValidMCPTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testBuilder.Definition.Name, testBuilder.Object.Name)
		}
	}
}

func TestMachineConfigPoolDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *MCPBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidMCPTestBuilder(buildTestClientWithDummyMCP()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidMCPTestBuilder(buildTestClientWithDummyMCP()),
			expectedError: fmt.Errorf("machineconfigpool 'name' cannot be empty"),
		},
		{
			testBuilder:   buildValidMCPTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

func TestMachineConfigPoolExists(t *testing.T) {
	testCases := []struct {
		testBuilder *MCPBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidMCPTestBuilder(buildTestClientWithDummyMCP()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidMCPTestBuilder(buildTestClientWithDummyMCP()),
			exists:      false,
		},
		{
			testBuilder: buildValidMCPTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestMachineConfigPoolWithMCSelector(t *testing.T) {
	testCases := []struct {
		mcSelector    map[string]string
		expectedError string
	}{
		{
			mcSelector:    map[string]string{"test": "test"},
			expectedError: "",
		},
		{
			mcSelector:    nil,
			expectedError: "machineConfigSelector 'MatchLabels' field cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidMCPTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		testBuilder = testBuilder.WithMcSelector(testCase.mcSelector)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.mcSelector, testBuilder.Definition.Spec.MachineConfigSelector.MatchLabels)
		}
	}
}

func TestMachineConfigPoolWithOptions(t *testing.T) {
	testCases := []struct {
		testBuilder   *MCPBuilder
		options       MCPAdditionalOptions
		expectedError string
	}{
		{
			testBuilder: buildValidMCPTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			options: func(builder *MCPBuilder) (*MCPBuilder, error) {
				builder.Definition.Spec.Paused = true

				return builder, nil
			},
			expectedError: "",
		},
		{
			testBuilder: buildInvalidMCPTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			options: func(builder *MCPBuilder) (*MCPBuilder, error) {
				return builder, nil
			},
			expectedError: "machineconfigpool 'name' cannot be empty",
		},
		{
			testBuilder: buildValidMCPTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			options: func(builder *MCPBuilder) (*MCPBuilder, error) {
				return builder, fmt.Errorf("error adding additional option")
			},
			expectedError: "error adding additional option",
		},
	}

	for _, testCase := range testCases {
		testBuilder := testCase.testBuilder.WithOptions(testCase.options)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.True(t, testBuilder.Definition.Spec.Paused)
		}
	}
}

func TestMachineConfigPoolIsInCondition(t *testing.T) {
	testCases := []struct {
		exists        bool
		valid         bool
		hasCondition  bool
		isInCondition bool
	}{
		{
			exists:        true,
			valid:         true,
			hasCondition:  true,
			isInCondition: true,
		},
		{
			exists:        true,
			valid:         false,
			hasCondition:  true,
			isInCondition: false,
		},
		{
			exists:        false,
			valid:         true,
			hasCondition:  true,
			isInCondition: false,
		},
		{
			exists:        true,
			valid:         true,
			hasCondition:  false,
			isInCondition: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testBuilder    *MCPBuilder
		)

		if testCase.exists {
			mcp := buildDummyMCP(defaultMCPName)

			if testCase.hasCondition {
				mcp.Status.Conditions = []mcv1.MachineConfigPoolCondition{defaultMCPCondition}
			}

			runtimeObjects = append(runtimeObjects, mcp)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: testSchemes,
		})

		if testCase.valid {
			testBuilder = buildValidMCPTestBuilder(testSettings)
		} else {
			testBuilder = buildInvalidMCPTestBuilder(testSettings)
		}

		isInCondition := testBuilder.IsInCondition(defaultMCPCondition.Type)
		assert.Equal(t, testCase.isInCondition, isInCondition)
	}
}

// buildDummyMCP returns a MachineConfigPool with the provided name.
func buildDummyMCP(name string) *mcv1.MachineConfigPool {
	return &mcv1.MachineConfigPool{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestClientWithDummyMCP returns a client with a dummy MachineConfigPool.
func buildTestClientWithDummyMCP() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyMCP(defaultMCPName),
		},
		SchemeAttachers: testSchemes,
	})
}

// buildValidMCPTestBuilder returns a valid MCPBuilder for testing.
func buildValidMCPTestBuilder(apiClient *clients.Settings) *MCPBuilder {
	return NewMCPBuilder(apiClient, defaultMCPName)
}

// buildInvalidMCPTestBuilder returns a valid MCPBuilder for testing.
func buildInvalidMCPTestBuilder(apiClient *clients.Settings) *MCPBuilder {
	return NewMCPBuilder(apiClient, "")
}
