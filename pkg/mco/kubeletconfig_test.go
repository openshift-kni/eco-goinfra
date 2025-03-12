package mco

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	mcv1 "github.com/openshift/api/machineconfiguration/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"k8s.io/utils/ptr"
)

const (
	defaultKubeletConfigName    = "test-kubeletconfig"
	defaultMCPoolSelectorKey    = "test-pool-selector"
	defaultMCPoolSelectorValue  = ""
	defaultSystemReservedCPU    = "100m"
	defaultSystemReservedMemory = "100Gi"
)

func TestNewKubeletConfigBuilder(t *testing.T) {
	testCases := []struct {
		name              string
		client            bool
		expectedErrorText string
	}{
		{
			name:              defaultKubeletConfigName,
			client:            true,
			expectedErrorText: "",
		},
		{
			name:              "",
			client:            true,
			expectedErrorText: "kubeletconfig 'name' cannot be empty",
		},
		{
			name:              defaultKubeletConfigName,
			client:            false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewKubeletConfigBuilder(testSettings, testCase.name)

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

func TestPullKubeletConfig(t *testing.T) {
	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultKubeletConfigName,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("kubeletconfig 'name' cannot be empty"),
		},
		{
			name:                defaultKubeletConfigName,
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("kubeletconfig object %s does not exist", defaultKubeletConfigName),
		},
		{
			name:                defaultKubeletConfigName,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("kubeletconfig 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testKubeletConfig := buildDummyKubeletConfig(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testKubeletConfig)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := PullKubeletConfig(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testKubeletConfig.Name, testBuilder.Definition.Name)
		}
	}
}

func TestKubeletConfigGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *KubeletConfigBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidKubeletConfigBuilder(buildTestClientWithDummyKubeletConfig()),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidKubeletConfigBuilder(buildTestClientWithDummyKubeletConfig()),
			expectedError: "kubeletconfig 'name' cannot be empty",
		},
		{
			testBuilder:   buildValidKubeletConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "kubeletconfigs.machineconfiguration.openshift.io \"test-kubeletconfig\" not found",
		},
	}

	for _, testCase := range testCases {
		kubeletConfig, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, kubeletConfig.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestKubeletConfigCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *KubeletConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidKubeletConfigBuilder(buildTestClientWithDummyKubeletConfig()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidKubeletConfigBuilder(buildTestClientWithDummyKubeletConfig()),
			expectedError: fmt.Errorf("kubeletconfig 'name' cannot be empty"),
		},
		{
			testBuilder:   buildValidKubeletConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
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

func TestKubeletConfigDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *KubeletConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidKubeletConfigBuilder(buildTestClientWithDummyKubeletConfig()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidKubeletConfigBuilder(buildTestClientWithDummyKubeletConfig()),
			expectedError: fmt.Errorf("kubeletconfig 'name' cannot be empty"),
		},
		{
			testBuilder:   buildValidKubeletConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
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

func TestKubeletConfigExists(t *testing.T) {
	testCases := []struct {
		testBuilder *KubeletConfigBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidKubeletConfigBuilder(buildTestClientWithDummyKubeletConfig()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidKubeletConfigBuilder(buildTestClientWithDummyKubeletConfig()),
			exists:      false,
		},
		{
			testBuilder: buildValidKubeletConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestKubeletConfigWithMCPoolSelector(t *testing.T) {
	testCases := []struct {
		testBuilder   *KubeletConfigBuilder
		key           string
		expectedError string
	}{
		{
			testBuilder:   buildValidKubeletConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			key:           defaultMCPoolSelectorKey,
			expectedError: "",
		},
		{
			testBuilder:   buildValidKubeletConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			key:           "",
			expectedError: "'key' cannot be empty",
		},
		{
			testBuilder:   buildInvalidKubeletConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			key:           defaultMCPoolSelectorKey,
			expectedError: "kubeletconfig 'name' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := testCase.testBuilder.WithMCPoolSelector(testCase.key, defaultMCPoolSelectorValue)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(
				t, testBuilder.Definition.Spec.MachineConfigPoolSelector.MatchLabels[testCase.key], defaultMCPoolSelectorValue)
		}
	}
}

func TestKubeletConfigWithSystemReserved(t *testing.T) {
	testCases := []struct {
		testBuilder   *KubeletConfigBuilder
		cpu           string
		memory        string
		expectedError string
	}{
		{
			testBuilder:   buildValidKubeletConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			cpu:           defaultSystemReservedCPU,
			memory:        defaultSystemReservedMemory,
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidKubeletConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			cpu:           defaultSystemReservedCPU,
			memory:        defaultSystemReservedMemory,
			expectedError: "kubeletconfig 'name' cannot be empty",
		},
		{
			testBuilder:   buildValidKubeletConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			cpu:           "",
			memory:        defaultSystemReservedMemory,
			expectedError: "'cpu' cannot be empty",
		},
		{
			testBuilder:   buildValidKubeletConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			cpu:           defaultSystemReservedCPU,
			memory:        "",
			expectedError: "'memory' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := testCase.testBuilder.WithSystemReserved(testCase.cpu, testCase.memory)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			systemReserved, ok := testBuilder.Definition.Spec.KubeletConfig.Object.(*kubeletconfigv1beta1.KubeletConfiguration)
			assert.True(t, ok)

			assert.Equal(t, testCase.cpu, systemReserved.SystemReserved["cpu"])
			assert.Equal(t, testCase.memory, systemReserved.SystemReserved["memory"])
		}
	}
}

func TestKubeletConfigWithOptions(t *testing.T) {
	testCases := []struct {
		testBuilder   *KubeletConfigBuilder
		options       AdditionalOptions
		expectedError string
	}{
		{
			testBuilder: buildValidKubeletConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			options: func(builder *KubeletConfigBuilder) (*KubeletConfigBuilder, error) {
				builder.Definition.Spec.AutoSizingReserved = ptr.To(true)

				return builder, nil
			},
			expectedError: "",
		},
		{
			testBuilder: buildInvalidKubeletConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			options: func(builder *KubeletConfigBuilder) (*KubeletConfigBuilder, error) {
				return builder, nil
			},
			expectedError: "kubeletconfig 'name' cannot be empty",
		},
		{
			testBuilder: buildValidKubeletConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			options: func(builder *KubeletConfigBuilder) (*KubeletConfigBuilder, error) {
				return builder, fmt.Errorf("error adding additional option")
			},
			expectedError: "error adding additional option",
		},
	}

	for _, testCase := range testCases {
		testBuilder := testCase.testBuilder.WithOptions(testCase.options)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.True(t, *testBuilder.Definition.Spec.AutoSizingReserved)
		}
	}
}

// buildDummyKubeletConfig returns a KubeletConfig with the provided name.
func buildDummyKubeletConfig(name string) *mcv1.KubeletConfig {
	return &mcv1.KubeletConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestClientWithDummyKubeletConfig returns a client with a mock KubeletConfig.
func buildTestClientWithDummyKubeletConfig() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyKubeletConfig(defaultKubeletConfigName),
		},
		SchemeAttachers: testSchemes,
	})
}

// buildValidKubeletConfigBuilder returns a valid KubeletConfigBuilder for testing.
func buildValidKubeletConfigBuilder(apiClient *clients.Settings) *KubeletConfigBuilder {
	return NewKubeletConfigBuilder(apiClient, defaultKubeletConfigName)
}

// buildInvalidKubeletConfigBuilder returns a invalid KubeletConfigBuilder for testing.
func buildInvalidKubeletConfigBuilder(apiClient *clients.Settings) *KubeletConfigBuilder {
	return NewKubeletConfigBuilder(apiClient, "")
}
