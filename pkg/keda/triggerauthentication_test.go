package keda

import (
	"fmt"
	"testing"

	kedav2v1alpha1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultTriggerAuthName      = "keda-trigger-auth-prometheus"
	defaultTriggerAuthNamespace = "test-appspace"
)

func TestPullTriggerAuthentication(t *testing.T) {
	generateTriggerAuth := func(name, namespace string) *kedav2v1alpha1.TriggerAuthentication {
		return &kedav2v1alpha1.TriggerAuthentication{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
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
			name:                defaultTriggerAuthName,
			namespace:           defaultTriggerAuthNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           defaultTriggerAuthNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("triggerAuthentication 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                defaultTriggerAuthName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("triggerAuthentication 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "triggerauthtest",
			namespace:           defaultTriggerAuthNamespace,
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf("triggerAuthentication object triggerauthtest does not exist " +
				"in namespace test-appspace"),
			client: true,
		},
		{
			name:                "triggerauthtest",
			namespace:           defaultTriggerAuthNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("triggerAuthentication 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testTriggerAuth := generateTriggerAuth(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testTriggerAuth)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := Pull(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testTriggerAuth.Name, builderResult.Object.Name)
			assert.Equal(t, testTriggerAuth.Namespace, builderResult.Object.Namespace)
			assert.Nil(t, err)
		}
	}
}

func TestNewTriggerAuthenticationBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
	}{
		{
			name:          defaultTriggerAuthName,
			namespace:     defaultTriggerAuthNamespace,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     defaultTriggerAuthNamespace,
			expectedError: "triggerAuthentication 'name' cannot be empty",
		},
		{
			name:          defaultTriggerAuthName,
			namespace:     "",
			expectedError: "triggerAuthentication 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testTriggerAuthBuilder := NewTriggerAuthenticationBuilder(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, testTriggerAuthBuilder.errorMsg)
		assert.NotNil(t, testTriggerAuthBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testTriggerAuthBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testTriggerAuthBuilder.Definition.Namespace)
		}
	}
}

func TestTriggerAuthenticationExists(t *testing.T) {
	testCases := []struct {
		testTriggerAuth *TriggerAuthenticationBuilder
		expectedStatus  bool
	}{
		{
			testTriggerAuth: buildValidTriggerAuthBuilder(buildTriggerAuthClientWithDummyObject()),
			expectedStatus:  true,
		},
		{
			testTriggerAuth: buildInValidTriggerAuthBuilder(buildTriggerAuthClientWithDummyObject()),
			expectedStatus:  false,
		},
		{
			testTriggerAuth: buildValidTriggerAuthBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:  false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testTriggerAuth.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestTriggerAuthenticationGet(t *testing.T) {
	testCases := []struct {
		testTriggerAuth *TriggerAuthenticationBuilder
		expectedError   error
	}{
		{
			testTriggerAuth: buildValidTriggerAuthBuilder(buildTriggerAuthClientWithDummyObject()),
			expectedError:   nil,
		},
		{
			testTriggerAuth: buildInValidTriggerAuthBuilder(buildTriggerAuthClientWithDummyObject()),
			expectedError:   fmt.Errorf("triggerauthentications.keda.sh \"\" not found"),
		},
		{
			testTriggerAuth: buildValidTriggerAuthBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:   fmt.Errorf("triggerauthentications.keda.sh \"keda-trigger-auth-prometheus\" not found"),
		},
	}

	for _, testCase := range testCases {
		triggerAuthObj, err := testCase.testTriggerAuth.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, triggerAuthObj.Name, testCase.testTriggerAuth.Definition.Name)
			assert.Equal(t, triggerAuthObj.Namespace, testCase.testTriggerAuth.Definition.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestTriggerAuthenticationCreate(t *testing.T) {
	testCases := []struct {
		TriggerAuth   *TriggerAuthenticationBuilder
		expectedError string
	}{
		{
			TriggerAuth:   buildValidTriggerAuthBuilder(buildTriggerAuthClientWithDummyObject()),
			expectedError: "",
		},
		{
			TriggerAuth:   buildInValidTriggerAuthBuilder(buildTriggerAuthClientWithDummyObject()),
			expectedError: " \"\" is invalid: metadata.name: Required value: name is required",
		},
		{
			TriggerAuth:   buildValidTriggerAuthBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testTriggerAuthBuilder, err := testCase.TriggerAuth.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testTriggerAuthBuilder.Definition.Name, testTriggerAuthBuilder.Object.Name)
			assert.Equal(t, testTriggerAuthBuilder.Definition.Namespace, testTriggerAuthBuilder.Object.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestTriggerAuthenticationDelete(t *testing.T) {
	testCases := []struct {
		testTriggerAuth *TriggerAuthenticationBuilder
		expectedError   error
	}{
		{
			testTriggerAuth: buildValidTriggerAuthBuilder(buildTriggerAuthClientWithDummyObject()),
			expectedError:   nil,
		},
		{
			testTriggerAuth: buildInValidTriggerAuthBuilder(buildTriggerAuthClientWithDummyObject()),
			expectedError:   nil,
		},
		{
			testTriggerAuth: buildValidTriggerAuthBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:   nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testTriggerAuth.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testTriggerAuth.Object)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestTriggerAuthenticationUpdate(t *testing.T) {
	testCases := []struct {
		testTriggerAuth     *TriggerAuthenticationBuilder
		expectedError       string
		testSecretTargetRef []kedav2v1alpha1.AuthSecretTargetRef
	}{
		{
			testTriggerAuth: buildValidTriggerAuthBuilder(buildTriggerAuthClientWithDummyObject()),
			expectedError:   "",
			testSecretTargetRef: []kedav2v1alpha1.AuthSecretTargetRef{{
				Name: "token-name",
				Key:  "token",
			},
				{
					Name: "cert-name",
					Key:  "ca.crt",
				}},
		},
		{
			testTriggerAuth:     buildInValidTriggerAuthBuilder(buildTriggerAuthClientWithDummyObject()),
			expectedError:       " \"\" is invalid: metadata.name: Required value: name is required",
			testSecretTargetRef: []kedav2v1alpha1.AuthSecretTargetRef{},
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, []kedav2v1alpha1.AuthSecretTargetRef(nil), testCase.testTriggerAuth.Definition.Spec.SecretTargetRef)
		assert.Nil(t, nil, testCase.testTriggerAuth.Object)
		testCase.testTriggerAuth.WithSecretTargetRef(testCase.testSecretTargetRef)
		_, err := testCase.testTriggerAuth.Update()

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.testSecretTargetRef, testCase.testTriggerAuth.Definition.Spec.SecretTargetRef)
		}
	}
}

func TestTriggerAuthenticationWithSecretTargetRef(t *testing.T) {
	testCases := []struct {
		testSecretTargetRef []kedav2v1alpha1.AuthSecretTargetRef
		expectedError       bool
		expectedErrorText   string
	}{
		{
			testSecretTargetRef: []kedav2v1alpha1.AuthSecretTargetRef{{
				Name: "token-name",
				Key:  "token",
			}},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testSecretTargetRef: []kedav2v1alpha1.AuthSecretTargetRef{{
				Name: "cert-name",
				Key:  "ca.crt",
			}},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testSecretTargetRef: []kedav2v1alpha1.AuthSecretTargetRef{{
				Name: "token-name",
				Key:  "token",
			},
				{
					Name: "cert-name",
					Key:  "ca.crt",
				}},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testSecretTargetRef: []kedav2v1alpha1.AuthSecretTargetRef{},
			expectedError:       true,
			expectedErrorText:   "'secretTargetRef' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTriggerAuthBuilder(buildTriggerAuthClientWithDummyObject())

		result := testBuilder.WithSecretTargetRef(testCase.testSecretTargetRef)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testSecretTargetRef, result.Definition.Spec.SecretTargetRef)
		}
	}
}

func buildValidTriggerAuthBuilder(apiClient *clients.Settings) *TriggerAuthenticationBuilder {
	triggerAuthBuilder := NewTriggerAuthenticationBuilder(
		apiClient, defaultTriggerAuthName, defaultTriggerAuthNamespace)

	return triggerAuthBuilder
}

func buildInValidTriggerAuthBuilder(apiClient *clients.Settings) *TriggerAuthenticationBuilder {
	triggerAuthBuilder := NewTriggerAuthenticationBuilder(
		apiClient, "", defaultTriggerAuthNamespace)

	return triggerAuthBuilder
}

func buildTriggerAuthClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyTriggerAuthentication(),
	})
}

func buildDummyTriggerAuthentication() []runtime.Object {
	return append([]runtime.Object{}, &kedav2v1alpha1.TriggerAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultTriggerAuthName,
			Namespace: defaultTriggerAuthNamespace,
		},
	})
}
