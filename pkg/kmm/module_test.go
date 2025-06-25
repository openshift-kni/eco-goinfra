package kmm

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/schemes/kmm/v1beta1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testSchemesV1beta1 = []clients.SchemeAttacher{
		v1beta1.AddToScheme,
	}
	defaultModuleName      = "testmodule"
	defaultModuleNamespace = "testns"
)

func TestNewModuleBuilder(t *testing.T) {
	testCases := []struct {
		name        string
		namespace   string
		expectedErr string
		client      bool
	}{
		{
			name:        defaultModuleName,
			namespace:   defaultModuleNamespace,
			expectedErr: "",
			client:      true,
		},
		{
			name:        defaultModuleName,
			namespace:   defaultModuleNamespace,
			expectedErr: "",
			client:      false,
		},
		{
			name:        defaultModuleName,
			namespace:   "",
			expectedErr: "module 'namespace' cannot be empty",
			client:      true,
		},
		{
			name:        "",
			namespace:   defaultModuleNamespace,
			expectedErr: "module 'name' cannot be empty",
			client:      true,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemesV1beta1})
		}

		testBuilder := NewModuleBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedErr == "" {
			if testCase.client {
				assert.NotNil(t, testBuilder)
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testBuilder.Definition.Namespace)
			} else {
				assert.Nil(t, testBuilder)
			}
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestModulePull(t *testing.T) {
	testCases := []struct {
		name                string
		namespace           string
		expectedError       error
		addToRuntimeObjects bool
		client              bool
	}{
		{
			name:                "test",
			namespace:           "testns",
			expectedError:       nil,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "",
			namespace:           "testns",
			expectedError:       fmt.Errorf("module 'name' cannot be empty"),
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "test",
			namespace:           "",
			expectedError:       fmt.Errorf("module 'namespace' cannot be empty"),
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "test",
			namespace:           "testns",
			expectedError:       fmt.Errorf("module object test does not exist in namespace testns"),
			addToRuntimeObjects: false,
			client:              true,
		},
		{
			name:                "test",
			namespace:           "testns",
			expectedError:       fmt.Errorf("module 'apiClient' cannot be empty"),
			addToRuntimeObjects: true,
			client:              false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testModule := generateModule(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testModule)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemesV1beta1,
			})
		}

		builderResult, err := Pull(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Definition.Name)
			assert.Equal(t, testCase.namespace, builderResult.Definition.Namespace)
		}
	}
}

func TestModuleGet(t *testing.T) {
	testCases := []struct {
		testModule    *ModuleBuilder
		expectedError error
	}{
		{
			testModule:    buildValidTestModule(buildModuleTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testModule:    buildInValidTestModule(buildModuleTestClientWithDummyObject()),
			expectedError: fmt.Errorf("module 'namespace' cannot be empty"),
		},
		{
			testModule:    buildValidTestModule(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("modules.kmm.sigs.x-k8s.io \"testmodule\" not found"),
		},
	}

	for _, testCase := range testCases {
		module, err := testCase.testModule.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, module.Name, testCase.testModule.Definition.Name)
			assert.Equal(t, module.Namespace, testCase.testModule.Definition.Namespace)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestModuleExists(t *testing.T) {
	testCases := []struct {
		testModule     *ModuleBuilder
		expectedStatus bool
	}{
		{
			testModule:     buildValidTestModule(buildModuleTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testModule:     buildInValidTestModule(buildModuleTestClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testModule:     buildValidTestModule(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testModule.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestModuleCreate(t *testing.T) {
	testCases := []struct {
		testModule    *ModuleBuilder
		expectedError string
	}{
		{
			testModule:    buildValidTestModule(buildModuleTestClientWithDummyObject()),
			expectedError: "",
		},
		{
			testModule:    buildInValidTestModule(buildModuleTestClientWithDummyObject()),
			expectedError: "module 'namespace' cannot be empty",
		},
		{
			testModule:    buildValidTestModule(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testModuleBuilder, err := testCase.testModule.Create()
		if testCase.expectedError == "" {
			assert.Equal(t, testModuleBuilder.Definition.Name, testModuleBuilder.Object.Name)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestModuleDelete(t *testing.T) {
	testCases := []struct {
		testModule    *ModuleBuilder
		expectedError error
	}{
		{
			testModule:    buildValidTestModule(buildModuleTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testModule:    buildInValidTestModule(buildModuleTestClientWithDummyObject()),
			expectedError: fmt.Errorf("module 'namespace' cannot be empty"),
		},
		{
			testModule:    buildValidTestModule(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testModule.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testModule.Object)
		}
	}
}

func TestModuleUpdate(t *testing.T) {
	testCases := []struct {
		testModule    *ModuleBuilder
		expectedError error
	}{
		{
			testModule:    buildValidTestModule(buildModuleTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testModule:    buildInValidTestModule(buildModuleTestClientWithDummyObject()),
			expectedError: fmt.Errorf("module 'namespace' cannot be empty"),
		},
		{
			testModule:    buildValidTestModule(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("modules.kmm.sigs.x-k8s.io \"testmodule\" not found"),
		},
	}

	for _, testCase := range testCases {
		assert.Empty(t, testCase.testModule.Definition.Spec.Selector)
		testCase.testModule.Definition.ResourceVersion = "999"
		testCase.testModule.Definition.Spec.Selector = map[string]string{"test": "test"}
		_, err := testCase.testModule.Update()

		if errors.IsNotFound(err) {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedError, err)
		}

		if testCase.expectedError == nil {
			assert.Equal(t, map[string]string{"test": "test"}, testCase.testModule.Object.Spec.Selector)
		}
	}
}

func TestModuleWithNodeSelector(t *testing.T) {
	testCases := []struct {
		nodeSelector map[string]string
		expectedErr  string
	}{
		{
			nodeSelector: map[string]string{"test": "test"},
			expectedErr:  "",
		},
		{
			nodeSelector: map[string]string{},
			expectedErr:  "Module 'nodeSelector' cannot be empty map",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithNodeSelector(testCase.nodeSelector)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.Selector, testCase.nodeSelector)
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestModuleWithLoadServiceAccount(t *testing.T) {
	testCases := []struct {
		loadServiceAccount string
		expectedErr        string
	}{
		{
			loadServiceAccount: "test",
			expectedErr:        "",
		},
		{
			loadServiceAccount: "",
			expectedErr:        "can not redefine module with empty ServiceAccount",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithLoadServiceAccount(testCase.loadServiceAccount)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.ModuleLoader.ServiceAccountName, testCase.loadServiceAccount)
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestModuleWithDevicePluginServiceAccount(t *testing.T) {
	testCases := []struct {
		devicePluginServiceAccount string
		expectedErr                string
	}{
		{
			devicePluginServiceAccount: "test",
			expectedErr:                "",
		},
		{
			devicePluginServiceAccount: "",
			expectedErr:                "can not redefine module with empty ServiceAccount",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithDevicePluginServiceAccount(testCase.devicePluginServiceAccount)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.DevicePlugin.ServiceAccountName, testCase.devicePluginServiceAccount)
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestModuleWithImageRepoSecret(t *testing.T) {
	testCases := []struct {
		repoSecret  string
		expectedErr string
	}{
		{
			repoSecret:  "test",
			expectedErr: "",
		},
		{
			repoSecret:  "",
			expectedErr: "can not redefine module with empty imageRepoSecret",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithImageRepoSecret(testCase.repoSecret)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.ImageRepoSecret.Name, testCase.repoSecret)
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestModuleWithDevicePluginVolume(t *testing.T) {
	testCases := []struct {
		pluginVolumeName string
		configMapName    string
		expectedErr      string
	}{
		{
			pluginVolumeName: "test",
			configMapName:    "test",
			expectedErr:      "",
		},
		{
			pluginVolumeName: "",
			configMapName:    "test",
			expectedErr:      "cannot redefine with empty volume 'name'",
		},
		{
			pluginVolumeName: "test",
			configMapName:    "",
			expectedErr:      "cannot redefine with empty 'configMapName'",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithDevicePluginVolume(testCase.pluginVolumeName, testCase.configMapName)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.DevicePlugin.Volumes[0].Name, testCase.pluginVolumeName)
			assert.Equal(t, testBuilder.Definition.Spec.DevicePlugin.Volumes[0].ConfigMap.Name, testCase.configMapName)
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestModuleWithModuleLoaderContainer(t *testing.T) {
	testCases := []struct {
		containerSpec *v1beta1.ModuleLoaderContainerSpec
		expectedErr   string
	}{
		{
			containerSpec: &v1beta1.ModuleLoaderContainerSpec{ContainerImage: "test"},
			expectedErr:   "",
		},
		{
			containerSpec: nil,
			expectedErr:   "invalid 'container' argument can not be nil",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithModuleLoaderContainer(testCase.containerSpec)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.ModuleLoader.Container, *testCase.containerSpec)
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestModuleWithDevicePluginContainer(t *testing.T) {
	testCases := []struct {
		containerSpec *v1beta1.DevicePluginContainerSpec
		expectedErr   string
	}{
		{
			containerSpec: &v1beta1.DevicePluginContainerSpec{Image: "test"},
			expectedErr:   "",
		},
		{
			containerSpec: nil,
			expectedErr:   "invalid 'container' argument can not be nil",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithDevicePluginContainer(testCase.containerSpec)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.DevicePlugin.Container, *testCase.containerSpec)
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestModuleWithOptions(t *testing.T) {
	testSettings := buildModuleTestClientWithDummyObject()
	testBuilder := buildValidTestModule(testSettings).WithOptions(
		func(builder *ModuleBuilder) (*ModuleBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidTestModule(testSettings).WithOptions(
		func(builder *ModuleBuilder) (*ModuleBuilder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestModuleWithToleration(t *testing.T) {
	var tenSeconds = int64(10)

	testCases := []struct {
		key         string
		operator    string
		value       string
		effect      string
		seconds     *int64
		expectedErr string
	}{
		{
			key:         "",
			operator:    "test",
			value:       "test",
			effect:      "test",
			seconds:     &tenSeconds,
			expectedErr: "cannot redefine with empty 'key' value",
		},
		{
			key:         "test",
			operator:    "",
			value:       "test",
			effect:      "test",
			seconds:     &tenSeconds,
			expectedErr: "cannot redefine with empty 'operator' value",
		},
		{
			key:         "test",
			operator:    "test",
			value:       "test",
			effect:      "",
			seconds:     &tenSeconds,
			expectedErr: "cannot redefine with empty 'effect' value",
		},
		{
			key:         "testkey",
			operator:    "Equals",
			value:       "testvalue",
			effect:      "NoSchedule",
			expectedErr: "",
		},
		{
			key:         "testkey",
			operator:    "Equals",
			value:       "testvalue",
			effect:      "NoSchedule",
			expectedErr: "",
		},
	}
	for _, testCase := range testCases {
		testSettings := buildModuleTestClientWithDummyObject()
		testBuilder := buildValidTestModule(testSettings).WithToleration(
			testCase.key, testCase.operator, testCase.value, testCase.effect, testCase.seconds)

		if testCase.expectedErr == "" {
			assert.Equal(t, testCase.key, testBuilder.Definition.Spec.Tolerations[0].Key)
			assert.Equal(t, corev1.TaintEffect(testCase.effect), testBuilder.Definition.Spec.Tolerations[0].Effect)
			assert.Equal(t, corev1.TolerationOperator(testCase.operator), testBuilder.Definition.Spec.Tolerations[0].Operator)
			assert.Equal(t, testCase.value, testBuilder.Definition.Spec.Tolerations[0].Value)
			assert.Equal(t, testCase.seconds, testBuilder.Definition.Spec.Tolerations[0].TolerationSeconds)
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestModuleBuildModuleSpec(t *testing.T) {
	testCases := []struct {
		testModule    *ModuleBuilder
		expectedError string
	}{
		{
			testModule:    buildValidTestModule(buildModuleTestClientWithDummyObject()),
			expectedError: "",
		},
		{
			testModule:    buildInValidTestModule(buildModuleTestClientWithDummyObject()),
			expectedError: "module 'namespace' cannot be empty",
		},
	}
	for _, testCase := range testCases {
		moduleSpec, err := testCase.testModule.BuildModuleSpec()
		if testCase.expectedError == "" {
			assert.NotEmpty(t, moduleSpec)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func buildValidTestModule(apiClient *clients.Settings) *ModuleBuilder {
	moduleBuilder := NewModuleBuilder(apiClient, defaultModuleName, defaultModuleNamespace)
	moduleBuilder.Definition.Spec.ModuleLoader = &v1beta1.ModuleLoaderSpec{ServiceAccountName: "Test"}

	return moduleBuilder
}

func buildInValidTestModule(apiClient *clients.Settings) *ModuleBuilder {
	moduleBuilder := NewModuleBuilder(apiClient, defaultModuleName, "")

	return moduleBuilder
}

func buildModuleTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyModule(),
		SchemeAttachers: testSchemesV1beta1,
	})
}

func buildDummyModule() []runtime.Object {
	return append([]runtime.Object{}, &v1beta1.Module{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultModuleName,
			Namespace: defaultModuleNamespace,
		},

		Spec: v1beta1.ModuleSpec{},
	})
}

func generateModule(name, nsname string) *v1beta1.Module {
	return &v1beta1.Module{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}
