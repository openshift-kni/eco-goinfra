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
			addToRuntimeObjects:     false,
			expectedErrorText:       "SriovOperatorConfig object default doesn't exist in namespace test-namespace",
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
			testSettings = clients.GetTestClients(runtimeObjects)
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
		testSettings := clients.GetTestClients([]runtime.Object{})
		testPolicyStructure := generatePolicyBuilder(testSettings, testCase.operatorConfigNamespace)
		assert.NotNil(t, testPolicyStructure)
		assert.Equal(t, testPolicyStructure.errorMsg, testCase.expectedErrorText)
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
			assert.Equal(t, oCBuilder.Definition, oCBuilder.Object)
		} else {
			assert.Equal(t, err, testCase.expectedError)
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
		assert.Equal(t, &testCase.enableInjector, operatorConfigBuilder.Definition.Spec.EnableInjector)
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
		assert.Equal(t, &testCase.webhook, operatorConfigBuilder.Definition.Spec.EnableOperatorWebhook)
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
		assert.Equal(t, &testCase.webhook, operatorConfigBuilder.Definition.Spec.EnableOperatorWebhook)

		if testCase.webhook {
			testCase.webhook = false
		} else {
			testCase.webhook = true
		}

		operatorConfigBuilder, err = operatorConfigBuilder.WithOperatorWebhook(testCase.webhook).Update()
		assert.Equal(t, nil, err)
		assert.Equal(t, &testCase.webhook, testCase.testOperatorConfig.Object.Spec.EnableOperatorWebhook)
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
	return clients.GetTestClients(buildDummySrIovOperatorConfigObject())
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
