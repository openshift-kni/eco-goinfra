package amdgpu

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	amdgpuv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/amd/gpu-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testSchemes = []clients.SchemeAttacher{
		amdgpuv1.AddToScheme,
	}
	testDeviceConfigName      = "test-deviceconfig"
	testDeviceConfigNamespace = "test-openshift-amd-gpu"
	almExample                = fmt.Sprintf(`
[
	{
		"apiVersion": "amd.com/v1alpha1",
		"kind": "DeviceConfig",
		"metadata": {
			"name": "%s",
			"namespace": "%s"
		},
		"spec": {
			"driver": {
				"enabled": true
			}
		}
	}
]
`, testDeviceConfigName, testDeviceConfigNamespace)
)

func TestAMDGPUNewBuilderFromObjectString(t *testing.T) {
	testCases := []struct {
		almExample    string
		expectedError string
		client        bool
	}{
		{
			almExample:    almExample,
			expectedError: "",
			client:        true,
		},
		{
			almExample:    almExample,
			expectedError: "",
			client:        false,
		},
		{
			almExample:    "{ invalid: data }",
			expectedError: "error initializing DeviceConfig from alm-examples: ",
			client:        true,
		},
		{
			almExample:    "",
			expectedError: "error initializing DeviceConfig from alm-examples: almExample is an empty string",
			client:        true,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings
		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		amdGPUNewBuilder := NewBuilderFromObjectString(testSettings, testCase.almExample)

		if testCase.expectedError == "" {
			if testCase.client {
				assert.Empty(t, amdGPUNewBuilder.errorMsg)
				assert.NotNil(t, amdGPUNewBuilder.Definition)
			} else {
				assert.Nil(t, amdGPUNewBuilder)
			}
		} else {
			assert.Contains(t, amdGPUNewBuilder.errorMsg, testCase.expectedError)
			assert.Nil(t, amdGPUNewBuilder.Definition)
		}
	}
}

func TestAMDGPUPull(t *testing.T) {
	generateDeviceConfig := func(name string, namespace string) *amdgpuv1.DeviceConfig {
		return &amdgpuv1.DeviceConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: amdgpuv1.DeviceConfigSpec{},
		}
	}

	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                testDeviceConfigName,
			namespace:           testDeviceConfigNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           testDeviceConfigNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("DeviceConfig 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                testDeviceConfigName,
			namespace:           testDeviceConfigNamespace,
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf("deviceConfig object %s does not exist in namespace %s",
				testDeviceConfigName, testDeviceConfigNamespace),
			client: true,
		},
		{
			name:                testDeviceConfigName,
			namespace:           testDeviceConfigNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("the apiClient of the Policy is nil"),
			client:              false,
		},
		{
			name:                testDeviceConfigName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("the apiClient of the Policy is nil"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testDeviceConfig := generateDeviceConfig(testCase.name, testDeviceConfigNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testDeviceConfig)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects, SchemeAttachers: testSchemes})
		}

		builderResult, err := Pull(testSettings, testCase.name, testDeviceConfigNamespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
		}
	}
}

func TestAMDGPUGet(t *testing.T) {
	testCases := []struct {
		deviceConfig  *Builder
		expectedError error
	}{
		{
			deviceConfig:  buildValidDeviceConfigBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			deviceConfig:  buildInvalidDeviceConfigBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("can not redefine the undefined DeviceConfig"),
		},
		{
			deviceConfig: buildValidDeviceConfigBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: fmt.Errorf("deviceconfigs.amd.com \"%s\" not found", testDeviceConfigName),
		},
	}

	for _, testCase := range testCases {
		deviceConfig, err := testCase.deviceConfig.Get()
		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}

		if testCase.expectedError == nil {
			assert.Equal(t, deviceConfig.Name, testCase.deviceConfig.Definition.Name)
		}
	}
}

func TestAMDGPUExist(t *testing.T) {
	testCases := []struct {
		deviceConfig   *Builder
		expectedStatus bool
	}{
		{
			deviceConfig:   buildValidDeviceConfigBuilder(buildTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			deviceConfig:   buildInvalidDeviceConfigBuilder(buildTestClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			deviceConfig: buildValidDeviceConfigBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.deviceConfig.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestAMDGPUDelete(t *testing.T) {
	testCases := []struct {
		deviceConfig  *Builder
		expectedError error
	}{
		{
			deviceConfig:  buildValidDeviceConfigBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			deviceConfig: buildValidDeviceConfigBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.deviceConfig.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.deviceConfig.Object)
		}
	}
}

func TestAMDGPUCreate(t *testing.T) {
	testCases := []struct {
		deviceConfig  *Builder
		expectedError error
	}{
		{
			deviceConfig:  buildValidDeviceConfigBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			deviceConfig: buildValidDeviceConfigBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: nil,
		},
		{
			deviceConfig:  buildInvalidDeviceConfigBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("can not redefine the undefined DeviceConfig"),
		},
	}

	for _, testCase := range testCases {
		clusterPolicyBuilder, err := testCase.deviceConfig.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, clusterPolicyBuilder.Definition.Name, clusterPolicyBuilder.Object.Name)
		}
	}
}

func buildInvalidDeviceConfigBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilderFromObjectString(apiClient, "")
}

func buildValidDeviceConfigBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilderFromObjectString(apiClient, almExample)
}

func buildDummyDeviceConfig() []runtime.Object {
	return append([]runtime.Object{}, &amdgpuv1.DeviceConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testDeviceConfigName,
			Namespace: testDeviceConfigNamespace,
		},
		Spec: amdgpuv1.DeviceConfigSpec{},
	})
}

func buildTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyDeviceConfig(),
		SchemeAttachers: testSchemes,
	})
}
