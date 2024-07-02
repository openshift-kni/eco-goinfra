package secret

import (
	"fmt"

	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultSecretName      = "test-name"
	defaultSecretNamespace = "test-namespace"
	defaultSecretType      = "test-secretType"
)

func TestSecretPull(t *testing.T) {
	testCases := []struct {
		secretName          string
		secretNamespace     string
		expectedError       bool
		addToRuntimeObjects bool
		client              bool
		expectedErrorText   string
	}{
		{
			secretName:          defaultSecretName,
			secretNamespace:     defaultSecretNamespace,
			expectedError:       false,
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "",
		},
		{
			secretName:          defaultSecretName,
			secretNamespace:     defaultSecretNamespace,
			expectedError:       true,
			addToRuntimeObjects: false,
			client:              true,
			expectedErrorText:   "secret object test-name does not exist in namespace test-namespace",
		},
		{
			secretName:          "",
			secretNamespace:     defaultSecretNamespace,
			expectedError:       true,
			addToRuntimeObjects: false,
			client:              true,
			expectedErrorText:   "secret 'name' cannot be empty",
		},
		{
			secretName:          defaultSecretName,
			secretNamespace:     "",
			expectedError:       true,
			addToRuntimeObjects: false,
			client:              true,
			expectedErrorText:   "secret 'nsname' cannot be empty",
		},
		{
			secretName:          defaultSecretName,
			secretNamespace:     defaultSecretNamespace,
			expectedError:       true,
			addToRuntimeObjects: false,
			client:              false,
			expectedErrorText:   "secret object test-name does not exist in namespace test-namespace",
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testCase.secretName,
					Namespace: testCase.secretNamespace,
				},
			})
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := Pull(testSettings, testCase.secretName, testCase.secretNamespace)

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, builderResult)
		}
	}
}

func TestSecretNewBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		secretType    string
		expectedError string
	}{
		{
			name:          defaultSecretName,
			namespace:     defaultSecretNamespace,
			secretType:    defaultSecretType,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     defaultSecretNamespace,
			secretType:    defaultSecretType,
			expectedError: "secret 'name' cannot be empty",
		},
		{
			name:          defaultSecretName,
			namespace:     "",
			secretType:    defaultSecretType,
			expectedError: "secret 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testSecretBuilder := NewBuilder(
			testSettings,
			testCase.name,
			testCase.namespace,
			corev1.SecretType(testCase.secretType),
		)
		assert.Equal(t, testCase.expectedError, testSecretBuilder.errorMsg)
		assert.NotNil(t, testSecretBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testSecretBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testSecretBuilder.Definition.Namespace)
		}
	}
}

func TestSecretCreate(t *testing.T) {
	testCases := []struct {
		testSecret    *Builder
		expectedError error
	}{
		{
			testSecret:    buildValidSecretBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testSecret:    buildInvalidSecretBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("secret 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		Builder, err := testCase.testSecret.Create()
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.Equal(t, Builder.Definition, Builder.Object)
		}
	}
}

func TestSecretDelete(t *testing.T) {
	testCases := []struct {
		secretExistsAlready bool
		name                string
		namespace           string
	}{
		{
			secretExistsAlready: true,
			name:                defaultSecretName,
			namespace:           defaultSecretNamespace,
		},
		{
			secretExistsAlready: false,
			name:                defaultSecretName,
			namespace:           defaultSecretNamespace,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.secretExistsAlready {
			runtimeObjects = append(runtimeObjects, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testCase.name,
					Namespace: testCase.namespace,
				},
			})
		}

		testBuilder, client := buildTestBuilderWithFakeObjects(runtimeObjects, testCase.name, testCase.namespace)

		err := testBuilder.Delete()
		assert.Nil(t, err)

		// Assert that the object actually does not exist
		_, err = Pull(client, testCase.name, testCase.namespace)
		assert.NotNil(t, err)
	}
}

func TestSecretExists(t *testing.T) {
	testCases := []struct {
		secretExistsAlready bool
		name                string
		namespace           string
	}{
		{
			secretExistsAlready: true,
			name:                defaultSecretName,
			namespace:           defaultSecretNamespace,
		},
		{
			secretExistsAlready: false,
			name:                defaultSecretName,
			namespace:           defaultSecretNamespace,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.secretExistsAlready {
			runtimeObjects = append(runtimeObjects, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testCase.name,
					Namespace: testCase.namespace,
				},
			})
		}

		testBuilder, _ := buildTestBuilderWithFakeObjects(runtimeObjects, testCase.name, testCase.namespace)

		result := testBuilder.Exists()
		if testCase.secretExistsAlready {
			assert.True(t, result)
		} else {
			assert.False(t, result)
		}
	}
}
func TestSecretWithOptions(t *testing.T) {
	testSettings := buildTestClientWithDummyObject()
	testBuilder := buildValidSecretBuilder(testSettings).WithOptions(
		func(builder *Builder) (*Builder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidSecretBuilder(testSettings).WithOptions(
		func(builder *Builder) (*Builder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestSecretWithData(t *testing.T) {
	testCases := []struct {
		key         string
		value       string
		expectedErr string
	}{
		{
			key:         "key",
			value:       "value",
			expectedErr: "",
		},
		{
			key:         "",
			value:       "",
			expectedErr: "'data' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object
		testBuilder, _ := buildTestBuilderWithFakeObjects(runtimeObjects, defaultSecretName, defaultSecretNamespace)

		// Convert the test case data to map[string][]byte
		data := make(map[string][]byte)
		if testCase.key != "" && testCase.value != "" {
			data[testCase.key] = []byte(testCase.value)
		}

		if testCase.expectedErr == "" {
			testBuilder.WithData(data)

			assert.Equal(t, []byte(testCase.value), testBuilder.Definition.Data[testCase.key])
		} else {
			testBuilder.WithData(map[string][]byte{})

			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestSecretWithAnnotations(t *testing.T) {
	testCases := []struct {
		testAnnotations   map[string]string
		expectedErrorText string
	}{
		{
			testAnnotations:   map[string]string{"openshift.io/internal-registry-auth-token.binding": "bound"},
			expectedErrorText: "",
		},
		{
			testAnnotations:   map[string]string{"openshift.io/internal-registry-auth-token.service-account": "default"},
			expectedErrorText: "",
		},
		{
			testAnnotations:   map[string]string{},
			expectedErrorText: "'annotations' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidSecretBuilder(buildTestClientWithDummyObject())

		testBuilder.WithAnnotations(testCase.testAnnotations)

		assert.Equal(t, testCase.expectedErrorText, testBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.testAnnotations, testBuilder.Definition.Annotations)
		}
	}
}

func TestSecretUpdate(t *testing.T) {
	generateTestSecret := func() *corev1.Secret {
		return &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      defaultSecretName,
				Namespace: defaultSecretNamespace,
			},
		}
	}

	testCases := []struct {
		secretExistsAlready bool
		Name                string
		Namespace           string
	}{
		{
			secretExistsAlready: false,
			Name:                "nameBeforeUpdate",
			Namespace:           "namespaceBeforeUpdate",
		},
		{
			secretExistsAlready: true,
			Name:                defaultSecretName,
			Namespace:           defaultSecretNamespace,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.secretExistsAlready {
			runtimeObjects = append(runtimeObjects, generateTestSecret())
		}

		testBuilder, _ := buildTestBuilderWithFakeObjects(runtimeObjects, testCase.Name, testCase.Namespace)

		// Assert the secret before the update
		assert.NotNil(t, testBuilder.Definition)

		assert.Equal(t, testCase.Name, testBuilder.Definition.Name)
		assert.Equal(t, testCase.Namespace, testBuilder.Definition.Namespace)

		// Perform the update
		result, err := testBuilder.Update()

		// Assert the result
		assert.NotNil(t, testBuilder.Definition)

		if !testCase.secretExistsAlready {
			assert.NotNil(t, err)
			assert.Nil(t, result.Object)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testBuilder.Definition.Name, result.Definition.Name)
			assert.Equal(t, testBuilder.Definition.Namespace, result.Definition.Namespace)
		}
	}
}

func buildTestBuilderWithFakeObjects(runtimeObjects []runtime.Object,
	name, namespace string) (*Builder, *clients.Settings) {
	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: runtimeObjects,
	})

	return &Builder{
		apiClient: testSettings,
		Definition: &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		},
	}, testSettings
}

func buildValidSecretBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(
		apiClient,
		defaultSecretName,
		defaultSecretNamespace,
		corev1.SecretType(""),
	)
}

func buildInvalidSecretBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(
		apiClient,
		"",
		defaultSecretNamespace,
		corev1.SecretType(defaultSecretType),
	)
}

func buildTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildSecretWithDummyObject(),
	})
}

func buildSecretWithDummyObject() []runtime.Object {
	return append([]runtime.Object{}, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultSecretName,
			Namespace: defaultSecretNamespace,
		},
	})
}

func TestSecretValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		expectedError string
	}{
		{
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "error: received nil Secret builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined Secret",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "Secret builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder, _ := buildTestBuilderWithFakeObjects(nil, "test", "test")

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		result, err := testBuilder.validate()
		if testCase.expectedError != "" {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedError, err.Error())
			assert.False(t, result)
		} else {
			assert.Nil(t, err)
			assert.True(t, result)
		}
	}
}
