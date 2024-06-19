package secret

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultSecretName      = "test-secret-name"
	defaultSecretNamespace = "test-secret-namespace"
	defaultSecretType      = corev1.SecretTypeDockercfg
)

func TestPullSecret(t *testing.T) {
	generateSecret := func(name, namespace string) *corev1.Secret {
		return &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Type: defaultSecretType,
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
			name:                defaultSecretName,
			namespace:           defaultSecretNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           defaultSecretNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("secret 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                defaultSecretName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("secret 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "secret-test",
			namespace:           defaultSecretNamespace,
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf("secret object secret-test does not exist " +
				"in namespace test-secret-namespace"),
			client: true,
		},
		{
			name:                "mon-test",
			namespace:           defaultSecretNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("secret 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testSecret := generateSecret(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testSecret)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := Pull(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testSecret.Name, builderResult.Object.Name)
			assert.Equal(t, testSecret.Namespace, builderResult.Object.Namespace)
			assert.Nil(t, err)
		}
	}
}

//nolint:funlen
func TestNewSecretBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		secretType    corev1.SecretType
		expectedError string
		client        bool
	}{
		{
			name:          defaultSecretName,
			namespace:     defaultSecretNamespace,
			secretType:    corev1.SecretTypeDockercfg,
			expectedError: "",
			client:        true,
		},
		{
			name:          defaultSecretName,
			namespace:     defaultSecretNamespace,
			secretType:    corev1.SecretTypeOpaque,
			expectedError: "",
			client:        true,
		},
		{
			name:          defaultSecretName,
			namespace:     defaultSecretNamespace,
			secretType:    corev1.SecretTypeTLS,
			expectedError: "",
			client:        true,
		},
		{
			name:          defaultSecretName,
			namespace:     defaultSecretNamespace,
			secretType:    corev1.SecretTypeServiceAccountToken,
			expectedError: "",
			client:        true,
		},
		{
			name:          defaultSecretName,
			namespace:     defaultSecretNamespace,
			secretType:    corev1.SecretTypeBasicAuth,
			expectedError: "",
			client:        true,
		},
		{
			name:          defaultSecretName,
			namespace:     defaultSecretNamespace,
			secretType:    corev1.SecretTypeBootstrapToken,
			expectedError: "",
			client:        true,
		},
		{
			name:          defaultSecretName,
			namespace:     defaultSecretNamespace,
			secretType:    corev1.SecretTypeDockerConfigJson,
			expectedError: "",
			client:        true,
		},
		{
			name:          defaultSecretName,
			namespace:     defaultSecretNamespace,
			secretType:    corev1.SecretTypeSSHAuth,
			expectedError: "",
			client:        true,
		},
		{
			name:          "",
			namespace:     defaultSecretNamespace,
			secretType:    defaultSecretType,
			expectedError: "secret 'name' cannot be empty",
			client:        true,
		},
		{
			name:          defaultSecretName,
			namespace:     "",
			secretType:    defaultSecretType,
			expectedError: "secret 'nsname' cannot be empty",
			client:        true,
		},
		{
			name:          defaultSecretName,
			namespace:     defaultSecretNamespace,
			secretType:    defaultSecretType,
			expectedError: "",
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings
		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testSecretBuilder := NewBuilder(testSettings,
			testCase.name,
			testCase.namespace,
			testCase.secretType)

		if testCase.expectedError == "" {
			if testCase.client {
				assert.Equal(t, testCase.name, testSecretBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testSecretBuilder.Definition.Namespace)
				assert.Equal(t, testCase.secretType, testSecretBuilder.Definition.Type)
			} else {
				assert.Nil(t, testSecretBuilder)
			}
		} else {
			assert.Equal(t, testCase.expectedError, testSecretBuilder.errorMsg)
			assert.NotNil(t, testSecretBuilder.Definition)
		}
	}
}

func TestSecretExists(t *testing.T) {
	testCases := []struct {
		testSecret     *Builder
		expectedStatus bool
	}{
		{
			testSecret:     buildValidSecretBuilder(buildSecretClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testSecret:     buildInValidSecretBuilder(buildSecretClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testSecret:     buildValidSecretBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testSecret.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestSecretGet(t *testing.T) {
	testCases := []struct {
		testSecret    *Builder
		expectedError error
	}{
		{
			testSecret:    buildValidSecretBuilder(buildSecretClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testSecret:    buildInValidSecretBuilder(buildSecretClientWithDummyObject()),
			expectedError: fmt.Errorf("secret 'name' cannot be empty"),
		},
		{
			testSecret:    buildValidSecretBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("secrets \"test-secret-name\" not found"),
		},
	}

	for _, testCase := range testCases {
		secretObj, err := testCase.testSecret.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, secretObj.Name, testCase.testSecret.Definition.Name)
			assert.Equal(t, secretObj.Namespace, testCase.testSecret.Definition.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestSecretCreate(t *testing.T) {
	testCases := []struct {
		testSecret    *Builder
		expectedError string
	}{
		{
			testSecret:    buildValidSecretBuilder(buildSecretClientWithDummyObject()),
			expectedError: "",
		},
		{
			testSecret:    buildInValidSecretBuilder(buildSecretClientWithDummyObject()),
			expectedError: "secret 'name' cannot be empty",
		},
		{
			testSecret:    buildValidSecretBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		secretObj, err := testCase.testSecret.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, secretObj.Definition.Name, secretObj.Object.Name)
			assert.Equal(t, secretObj.Definition.Namespace, secretObj.Object.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestSecretDelete(t *testing.T) {
	testCases := []struct {
		testSecret    *Builder
		expectedError error
	}{
		{
			testSecret:    buildValidSecretBuilder(buildSecretClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testSecret:    buildValidSecretBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testSecret.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testSecret.Object)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestSecretUpdate(t *testing.T) {
	testCases := []struct {
		testSecret     *Builder
		expectedError  string
		testAnnotation map[string]string
	}{
		{
			testSecret: buildValidSecretBuilder(buildSecretClientWithDummyObject()),
			testAnnotation: map[string]string{"first-test-annotation-key": "first-test-annotation-value",
				"second-test-annotation-key": ""},
			expectedError: "",
		},
		{
			testSecret:     buildValidSecretBuilder(buildSecretClientWithDummyObject()),
			testAnnotation: map[string]string{},
			expectedError:  "'annotations' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, map[string]string(nil), testCase.testSecret.Definition.Annotations)
		testCase.testSecret.WithAnnotations(testCase.testAnnotation)
		_, err := testCase.testSecret.Update()

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.testAnnotation,
				testCase.testSecret.Definition.Annotations)
		}
	}
}

func TestSecretWithData(t *testing.T) {
	testCases := []struct {
		testData          map[string][]byte
		expectedErrorText string
	}{
		{
			testData:          map[string][]byte{"ssh-privatekey": []byte("MIIEpQIBAAKCAQEAulqb/Y")},
			expectedErrorText: "",
		},
		{
			testData:          map[string][]byte{".dockerconfig": []byte("bZ2dnZ2dnZ2dnZ2cgYXV0aCBrZXlzCg==")},
			expectedErrorText: "",
		},
		{
			testData: map[string][]byte{"user-name": []byte("bZ2dnZMMMTAKKLTrZXlzCg=="),
				"password": []byte("MIIEpQIBAAKCAQEAulqbY")},
			expectedErrorText: "",
		},
		{
			testData:          map[string][]byte{},
			expectedErrorText: "'data' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidSecretBuilder(buildSecretClientWithDummyObject())

		result := testBuilder.WithData(testCase.testData)

		if testCase.expectedErrorText != "" {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testData, result.Definition.Data)
		}
	}
}

func TestSecretWithAnnotations(t *testing.T) {
	testCases := []struct {
		testAnnotation map[string]string
		expectedErrMsg string
	}{
		{
			testAnnotation: map[string]string{"test-annotation-key": "test-annotation-value"},
			expectedErrMsg: "",
		},
		{
			testAnnotation: map[string]string{"test-annotation1-key": "test-annotation1-value",
				"test-annotation2-key": "test-annotation2-value",
			},
			expectedErrMsg: "",
		},
		{
			testAnnotation: map[string]string{"test-annotation-key": ""},
			expectedErrMsg: "",
		},
		{
			testAnnotation: map[string]string{"": "test-annotation-value"},
			expectedErrMsg: "can not apply an annotations with an empty key",
		},
		{
			testAnnotation: map[string]string{},
			expectedErrMsg: "'annotations' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidSecretBuilder(buildSecretClientWithDummyObject())

		testBuilder.WithAnnotations(testCase.testAnnotation)

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.testAnnotation, testBuilder.Definition.Annotations)
		}
	}
}

func buildValidSecretBuilder(apiClient *clients.Settings) *Builder {
	serviceMonitorBuilder := NewBuilder(
		apiClient, defaultSecretName, defaultSecretNamespace, defaultSecretType)

	return serviceMonitorBuilder
}

func buildInValidSecretBuilder(apiClient *clients.Settings) *Builder {
	serviceMonitorBuilder := NewBuilder(
		apiClient, "", defaultSecretNamespace, defaultSecretType)

	return serviceMonitorBuilder
}

func buildSecretClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummySecret(),
	})
}

func buildDummySecret() []runtime.Object {
	return append([]runtime.Object{}, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultSecretName,
			Namespace: defaultSecretNamespace,
		},
	})
}
