package sriov

import (
	"fmt"
	"testing"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultOperatorConfigNsName = "testnamespace"
)

func TestPullOperatorConfig(t *testing.T) {
	generatePolicy := func(namespace string) *srIovV1.SriovOperatorConfig {
		return &srIovV1.SriovOperatorConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "default",
				Namespace: namespace,
			},
			Spec: srIovV1.SriovOperatorConfigSpec{},
		}
	}

	testCases := []struct {
		operatorConfigNamespace string
		expectedError           bool
		addToRuntimeObjects     bool
		expectedErrorText       string
		client                  bool
	}{
		{
			operatorConfigNamespace: "test-namespace",
			expectedError:           false,
			addToRuntimeObjects:     true,
			client:                  true,
		},
		{
			operatorConfigNamespace: "test-namespace",
			expectedError:           true,
			expectedErrorText:       "SriovOperatorConfig 'apiClient' cannot be empty",
			addToRuntimeObjects:     true,
			client:                  false,
		},
		{
			operatorConfigNamespace: "test-namespace",
			expectedError:           true,
			addToRuntimeObjects:     false,
			expectedErrorText:       "SriovOperatorConfig object default does not exist in namespace test-namespace",
			client:                  true,
		},
		{
			operatorConfigNamespace: "",
			expectedError:           true,
			addToRuntimeObjects:     true,
			expectedErrorText:       "SriovOperatorConfig 'nsname' cannot be empty",
			client:                  true,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testPolicy := generatePolicy(testCase.operatorConfigNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPolicy)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		// Test the Pull method
		builderResult, err := PullOperatorConfig(testSettings, testPolicy.Namespace)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			// Check the error message
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testPolicy.Name, builderResult.Object.Name)
			assert.Equal(t, testPolicy.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestNewOperatorConfigBuilder(t *testing.T) {
	generatePolicyBuilder := NewOperatorConfigBuilder

	testCases := []struct {
		operatorConfigNamespace string
		expectedErrorText       string
	}{
		{
			operatorConfigNamespace: "test-namespace",
			expectedErrorText:       "",
		},
		{
			operatorConfigNamespace: "",
			expectedErrorText:       "SriovOperatorConfig 'nsname' is empty",
		},
	}
	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testPolicyStructure := generatePolicyBuilder(testSettings, testCase.operatorConfigNamespace)
		assert.NotNil(t, testPolicyStructure)
		assert.Equal(t, testCase.expectedErrorText, testPolicyStructure.errorMsg)
	}
}

func TestOperatorConfigCreate(t *testing.T) {
	testCases := []struct {
		testOperatorConfig *OperatorConfigBuilder
		expectedError      error
	}{
		{
			testOperatorConfig: NewOperatorConfigBuilder(
				buildTestClientWithDummyOperatorConfigObject(),
				defaultOperatorConfigNsName),
			expectedError: nil,
		},
		{
			testOperatorConfig: NewOperatorConfigBuilder(buildTestClientWithDummyOperatorConfigObject(), ""),
			expectedError:      fmt.Errorf("SriovOperatorConfig 'nsname' is empty"),
		},
	}

	for _, testCase := range testCases {
		oCBuilder, err := testCase.testOperatorConfig.Create()

		if testCase.expectedError == nil {
			assert.Nil(t, err)
			assert.Equal(t, oCBuilder.Definition.Name, oCBuilder.Object.Name)
		} else {
			assert.Equal(t, testCase.expectedError, err)
		}
	}
}

func TestOperatorConfigExistGet(t *testing.T) {
	testCases := []struct {
		operatorConfig *OperatorConfigBuilder
		expectedError  error
	}{
		{
			operatorConfig: NewOperatorConfigBuilder(
				buildTestClientWithDummyOperatorConfigObject(),
				defaultOperatorConfigNsName),
			expectedError: nil,
		},
		{
			operatorConfig: NewOperatorConfigBuilder(buildTestClientWithDummyOperatorConfigObject(), ""),
			expectedError:  fmt.Errorf("SriovOperatorConfig 'nsname' is empty"),
		},
	}

	for _, testCase := range testCases {
		operatorConfig, err := testCase.operatorConfig.Get()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, operatorConfig.Name, testCase.operatorConfig.Definition.Name)
		}
	}
}

func TestOperatorConfigExist(t *testing.T) {
	testCases := []struct {
		testOperatorConfig *OperatorConfigBuilder
		expectedStatus     bool
	}{
		{
			testOperatorConfig: NewOperatorConfigBuilder(
				buildTestClientWithDummyOperatorConfigObject(),
				defaultOperatorConfigNsName),
			expectedStatus: true,
		},
		{
			testOperatorConfig: NewOperatorConfigBuilder(buildTestClientWithDummyOperatorConfigObject(), ""),
			expectedStatus:     false,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.expectedStatus, testCase.testOperatorConfig.Exists())
	}
}

func TestOperatorConfigWithInjector(t *testing.T) {
	testCases := []struct {
		enableInjector bool
	}{
		{
			enableInjector: true,
		},
		{
			enableInjector: false,
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyPolicyObject()
		operatorConfigBuilder := NewOperatorConfigBuilder(testSettings, "testnamespace").
			WithInjector(testCase.enableInjector)
		assert.Equal(t, operatorConfigBuilder.errorMsg, "")
		assert.Equal(t, testCase.enableInjector, operatorConfigBuilder.Definition.Spec.EnableInjector)
	}
}

func TestOperatorConfigWithOperatorWebhook(t *testing.T) {
	testCases := []struct {
		webhook bool
	}{
		{
			webhook: true,
		},
		{
			webhook: false,
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyPolicyObject()
		operatorConfigBuilder := NewOperatorConfigBuilder(testSettings, "testnamespace").
			WithOperatorWebhook(testCase.webhook)
		assert.Equal(t, operatorConfigBuilder.errorMsg, "")
		assert.Equal(t, testCase.webhook, operatorConfigBuilder.Definition.Spec.EnableOperatorWebhook)
	}
}

func TestOperatorConfigWithConfigDaemonNodeSelector(t *testing.T) {
	testCases := []struct {
		configDaemonNodeSelector map[string]string
		expectedErrorText        string
	}{
		{
			configDaemonNodeSelector: map[string]string{"test-node-selector-key": "test-node-selector-value"},
			expectedErrorText:        "",
		},
		{
			configDaemonNodeSelector: map[string]string{"test-node-selector-key": ""},
			expectedErrorText:        "",
		},
		{
			configDaemonNodeSelector: map[string]string{"": "test-node-selector-value"},
			expectedErrorText:        "can not apply configDaemonNodeSelector with an empty selectorKey value",
		},
		{
			configDaemonNodeSelector: map[string]string{},
			expectedErrorText:        "can not apply empty configDaemonNodeSelector",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyPolicyObject()
		operatorConfigBuilder := NewOperatorConfigBuilder(testSettings, "testnamespace").
			WithConfigDaemonNodeSelector(testCase.configDaemonNodeSelector)

		assert.Equal(t, testCase.expectedErrorText, operatorConfigBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.configDaemonNodeSelector,
				operatorConfigBuilder.Definition.Spec.ConfigDaemonNodeSelector)
		}
	}
}

func TestOperatorConfigWithDisablePlugins(t *testing.T) {
	testCases := []struct {
		plugins           []string
		expectedErrorText string
	}{
		{
			plugins:           []string{"mellanox"},
			expectedErrorText: "",
		},
		{
			plugins:           []string{"test"},
			expectedErrorText: "invalid plugin parameter",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyPolicyObject()
		operatorConfigBuilder := NewOperatorConfigBuilder(testSettings, "testnamespace").
			WithDisablePlugins(testCase.plugins)

		assert.Equal(t, testCase.expectedErrorText, operatorConfigBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			var pluginSlice srIovV1.PluginNameSlice
			for _, plugin := range testCase.plugins {
				pluginSlice = append(pluginSlice, srIovV1.PluginNameValue(plugin))
			}

			assert.Equal(t, pluginSlice,
				operatorConfigBuilder.Definition.Spec.DisablePlugins)
		}
	}
}

func TestOperatorConfigRemoveDisablePlugins(t *testing.T) {
	testCases := []struct {
		testOperatorConfig *OperatorConfigBuilder
	}{
		{
			testOperatorConfig: NewOperatorConfigBuilder(
				buildTestClientWithDummyOperatorConfigObject(),
				defaultOperatorConfigNsName).WithDisablePlugins([]string{"mellanox"}),
		},
	}
	for _, testCase := range testCases {
		operatorConfigBuilder, _ := testCase.testOperatorConfig.Create()

		operatorConfigBuilder.RemoveDisablePlugins()
		assert.Nil(t, operatorConfigBuilder.Definition.Spec.DisablePlugins)
	}
}

func TestOperatorConfigUpdate(t *testing.T) {
	testCases := []struct {
		testOperatorConfig *OperatorConfigBuilder
		webhook            bool
	}{
		{
			testOperatorConfig: NewOperatorConfigBuilder(
				buildTestClientWithDummyOperatorConfigObject(),
				defaultOperatorConfigNsName),
		},
		{
			testOperatorConfig: NewOperatorConfigBuilder(
				buildTestClientWithDummyOperatorConfigObject(),
				defaultOperatorConfigNsName),
			webhook: true,
		},
	}
	for _, testCase := range testCases {
		operatorConfigBuilder, err := testCase.testOperatorConfig.WithOperatorWebhook(testCase.webhook).Create()
		assert.Nil(t, err)
		assert.Equal(t, testCase.webhook, operatorConfigBuilder.Definition.Spec.EnableOperatorWebhook)

		if testCase.webhook {
			testCase.webhook = false
		} else {
			testCase.webhook = true
		}

		operatorConfigBuilder.Definition.ObjectMeta.ResourceVersion = "999"
		operatorConfigBuilder, err = operatorConfigBuilder.WithOperatorWebhook(testCase.webhook).Update()
		assert.Equal(t, nil, err)
		assert.Equal(t, testCase.webhook, testCase.testOperatorConfig.Object.Spec.EnableOperatorWebhook)
		assert.Equal(t, operatorConfigBuilder.Definition, operatorConfigBuilder.Object)
	}
}

func TestOperatorConfigDelete(t *testing.T) {
	testCases := []struct {
		testOperatorConfig *OperatorConfigBuilder
		expectedError      error
	}{
		{
			testOperatorConfig: NewOperatorConfigBuilder(
				buildTestClientWithDummyOperatorConfigObject(),
				defaultOperatorConfigNsName),
		},
		{
			testOperatorConfig: NewOperatorConfigBuilder(
				buildTestClientWithDummyOperatorConfigObject(),
				""),
			expectedError: fmt.Errorf("SriovOperatorConfig 'nsname' is empty"),
		},
	}
	for _, testCase := range testCases {
		operatorConfigBuilder, err := testCase.testOperatorConfig.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, operatorConfigBuilder.Object)
		}

		operatorConfigBuilder, err = operatorConfigBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)
		assert.Nil(t, operatorConfigBuilder.Object)
	}
}

func buildTestClientWithDummyOperatorConfigObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummySrIovOperatorConfigObject(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummySrIovOperatorConfigObject() []runtime.Object {
	return append([]runtime.Object{}, &srIovV1.SriovOperatorConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: defaultOperatorConfigNsName,
		},
		Spec: srIovV1.SriovOperatorConfigSpec{},
	})
}
