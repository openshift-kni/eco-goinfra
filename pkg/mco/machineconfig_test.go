package mco

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	mcv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const defaultMachineConfigName = "test-machine-config"

var testSchemes = []clients.SchemeAttacher{
	mcv1.Install,
}

func TestNewMachineConfigBuilder(t *testing.T) {
	testCases := []struct {
		name              string
		client            bool
		expectedErrorText string
	}{
		{
			name:              defaultMachineConfigName,
			client:            true,
			expectedErrorText: "",
		},
		{
			name:              "",
			client:            true,
			expectedErrorText: "machineconfig 'name' cannot be empty",
		},
		{
			name:              defaultMachineConfigName,
			client:            false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewMCBuilder(testSettings, testCase.name)

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

func TestPullMachineConfig(t *testing.T) {
	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultMachineConfigName,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("machineconfig 'name' cannot be empty"),
		},
		{
			name:                defaultMachineConfigName,
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("machineconfig object %s does not exist", defaultMachineConfigName),
		},
		{
			name:                defaultMachineConfigName,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("machineconfig 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testMC := buildDummyMachineConfig(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testMC)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := PullMachineConfig(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testMC.Name, testBuilder.Definition.Name)
		}
	}
}

func TestMachineConfigGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *MCBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidMachineConfigTestBuilder(buildTestClientWithDummyMachineConfig()),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidMachineConfigTestBuilder(buildTestClientWithDummyMachineConfig()),
			expectedError: "machineconfig 'name' cannot be empty",
		},
		{
			testBuilder:   buildValidMachineConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "machineconfigs.machineconfiguration.openshift.io \"test-machine-config\" not found",
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

func TestMachineConfigCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *MCBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidMachineConfigTestBuilder(buildTestClientWithDummyMachineConfig()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidMachineConfigTestBuilder(buildTestClientWithDummyMachineConfig()),
			expectedError: fmt.Errorf("machineconfig 'name' cannot be empty"),
		},
		{
			testBuilder:   buildValidMachineConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
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

func TestMachineConfigUpdate(t *testing.T) {
	testCases := []struct {
		testBuilder   *MCBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidMachineConfigTestBuilder(buildTestClientWithDummyMachineConfig()),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidMachineConfigTestBuilder(buildTestClientWithDummyMachineConfig()),
			expectedError: "machineconfig 'name' cannot be empty",
		},
		{
			testBuilder:   buildValidMachineConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "machineconfigs.machineconfiguration.openshift.io \"test-machine-config\" not found",
		},
	}

	for _, testCase := range testCases {
		assert.False(t, testCase.testBuilder.Definition.Spec.FIPS)

		testCase.testBuilder.Definition.ResourceVersion = "999"
		testCase.testBuilder.Definition.Spec.FIPS = true

		_, err := testCase.testBuilder.Update()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.True(t, testCase.testBuilder.Object.Spec.FIPS)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestMachineConfigDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *MCBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidMachineConfigTestBuilder(buildTestClientWithDummyMachineConfig()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidMachineConfigTestBuilder(buildTestClientWithDummyMachineConfig()),
			expectedError: fmt.Errorf("machineconfig 'name' cannot be empty"),
		},
		{
			testBuilder:   buildValidMachineConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
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

func TestMachineConfigExists(t *testing.T) {
	testCases := []struct {
		testBuilder *MCBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidMachineConfigTestBuilder(buildTestClientWithDummyMachineConfig()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidMachineConfigTestBuilder(buildTestClientWithDummyMachineConfig()),
			exists:      false,
		},
		{
			testBuilder: buildValidMachineConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestMachineConfigWithLabel(t *testing.T) {
	testCases := []struct {
		key           string
		expectedError string
	}{
		{
			key:           "test",
			expectedError: "",
		},
		{
			key:           "",
			expectedError: "'key' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidMachineConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		testBuilder = testBuilder.WithLabel(testCase.key, "test")

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, "test", testBuilder.Definition.Labels[testCase.key])
		}
	}
}

func TestMachineConfigWithOptions(t *testing.T) {
	testCases := []struct {
		testBuilder   *MCBuilder
		options       MCAdditionalOptions
		expectedError string
	}{
		{
			testBuilder: buildValidMachineConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			options: func(builder *MCBuilder) (*MCBuilder, error) {
				builder.Definition.Spec.FIPS = true

				return builder, nil
			},
			expectedError: "",
		},
		{
			testBuilder: buildInvalidMachineConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			options: func(builder *MCBuilder) (*MCBuilder, error) {
				return builder, nil
			},
			expectedError: "machineconfig 'name' cannot be empty",
		},
		{
			testBuilder: buildValidMachineConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			options: func(builder *MCBuilder) (*MCBuilder, error) {
				return builder, fmt.Errorf("error adding additional option")
			},
			expectedError: "error adding additional option",
		},
	}

	for _, testCase := range testCases {
		testBuilder := testCase.testBuilder.WithOptions(testCase.options)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.True(t, testBuilder.Definition.Spec.FIPS)
		}
	}
}

func TestMachineConfigWithKernelArguments(t *testing.T) {
	testCases := []struct {
		kernelArgs    []string
		expectedError string
	}{
		{
			kernelArgs:    []string{"test"},
			expectedError: "",
		},
		{
			kernelArgs:    nil,
			expectedError: "'kernelArgs' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidMachineConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		testBuilder = testBuilder.WithKernelArguments(testCase.kernelArgs)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.kernelArgs, testBuilder.Definition.Spec.KernelArguments)
		}
	}
}

func TestMachineConfigWithExtensions(t *testing.T) {
	testCases := []struct {
		extensions    []string
		expectedError string
	}{
		{
			extensions:    []string{"test"},
			expectedError: "",
		},
		{
			extensions:    nil,
			expectedError: "'extensions' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidMachineConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		testBuilder = testBuilder.WithExtensions(testCase.extensions)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.extensions, testBuilder.Definition.Spec.Extensions)
		}
	}
}

func TestMachineConfigWithRawConfig(t *testing.T) {
	testCases := []struct {
		rawConfig     []byte
		expectedError string
	}{
		{
			rawConfig:     []byte{0, 0, 0},
			expectedError: "",
		},
		{
			rawConfig:     []byte{},
			expectedError: "'Config.Raw' cannot be empty",
		},
	}
	for _, testCase := range testCases {
		testBuilder := buildValidMachineConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		testBuilder = testBuilder.WithRawConfig(testCase.rawConfig)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.rawConfig, testBuilder.Definition.Spec.Config.Raw)
		}
	}
}

func TestMachineConfigWithFIPS(t *testing.T) {
	testBuilder := buildValidMachineConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
	testBuilder = testBuilder.WithFIPS(true)

	assert.True(t, testBuilder.Definition.Spec.FIPS)
}

func TestMachineConfigWithKernelType(t *testing.T) {
	testCases := []struct {
		kernelType    string
		expectedError string
	}{
		{
			kernelType:    "test",
			expectedError: "",
		},
		{
			kernelType:    "",
			expectedError: "'kernelType' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidMachineConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		testBuilder = testBuilder.WithKernelType(testCase.kernelType)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.kernelType, testBuilder.Definition.Spec.KernelType)
		}
	}
}

// buildDummyMachineConfig returns a MachineConfig with the provided name.
func buildDummyMachineConfig(name string) *mcv1.MachineConfig {
	return &mcv1.MachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestClientWithDummyMachineConfig returns a client with a dummy MachineConfig.
func buildTestClientWithDummyMachineConfig() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyMachineConfig(defaultMachineConfigName),
		},
		SchemeAttachers: testSchemes,
	})
}

// buildValidMachineConfigTestBuilder returns a valid MCBuilder for testing.
func buildValidMachineConfigTestBuilder(apiClient *clients.Settings) *MCBuilder {
	return NewMCBuilder(apiClient, defaultMachineConfigName)
}

// buildInvalidMachineConfigTestBuilder returns a valid MCBuilder for testing.
func buildInvalidMachineConfigTestBuilder(apiClient *clients.Settings) *MCBuilder {
	return NewMCBuilder(apiClient, "")
}
