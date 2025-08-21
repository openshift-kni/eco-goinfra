package serviceaccount

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestNewBuilder(t *testing.T) {
	testCases := []struct {
		name        string
		namespace   string
		expectedSA  *corev1.ServiceAccount
		expectedErr string
	}{
		{
			name:      "test-sa",
			namespace: "test-ns",
			expectedSA: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sa",
					Namespace: "test-ns",
				},
			},
			expectedErr: "",
		},
		{
			name:        "",
			namespace:   "test-ns",
			expectedSA:  nil,
			expectedErr: "serviceaccount 'name' cannot be empty",
		},
		{
			name:        "test-sa",
			namespace:   "",
			expectedSA:  nil,
			expectedErr: "serviceaccount 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})

		testBuilder := NewBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedErr != "" {
			assert.NotNil(t, testBuilder)
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		} else {
			assert.NotNil(t, testBuilder)
			assert.Equal(t, testCase.expectedSA.Name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.expectedSA.Namespace, testBuilder.Definition.Namespace)
		}
	}
}

func TestServiceAccountPull(t *testing.T) {
	generateServiceAccount := func(name, namespace string) *corev1.ServiceAccount {
		return &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		saName              string
		saNamespace         string
		expectedErrorStr    string
		addToRuntimeObjects bool
	}{
		{
			saName:              "test-sa",
			saNamespace:         "test-ns",
			expectedErrorStr:    "",
			addToRuntimeObjects: true,
		},
		{
			saName:              "test-sa",
			saNamespace:         "test-ns",
			expectedErrorStr:    "serviceaccount object test-sa does not exist in namespace test-ns",
			addToRuntimeObjects: false,
		},
		{
			saName:              "",
			saNamespace:         "test-ns",
			expectedErrorStr:    "serviceaccount 'name' cannot be empty",
			addToRuntimeObjects: false,
		},
		{
			saName:              "test-sa",
			saNamespace:         "",
			expectedErrorStr:    "serviceaccount 'namespace' cannot be empty",
			addToRuntimeObjects: false,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testSA := generateServiceAccount(testCase.saName, testCase.saNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testSA)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		builderResult, err := Pull(testSettings, testCase.saName, testCase.saNamespace)

		if testCase.expectedErrorStr != "" {
			assert.Equal(t, testCase.expectedErrorStr, err.Error())
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testSA.Name, builderResult.Object.Name)
			assert.Equal(t, testSA.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestServiceAccountCreate(t *testing.T) {
	generateServiceAccount := func(name, namespace string) *corev1.ServiceAccount {
		return &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		saExistsAlready bool
	}{
		{
			saExistsAlready: false,
		},
		{
			saExistsAlready: true,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		testSA := generateServiceAccount("test-sa", "test-ns")

		if testCase.saExistsAlready {
			runtimeObjects = append(runtimeObjects, testSA)
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects, testSA.Name, testSA.Namespace)
		result, err := testBuilder.Create()
		assert.Nil(t, err)
		assert.Equal(t, testBuilder.Object.Name, result.Object.Name)
	}
}

func TestServiceAccountDelete(t *testing.T) {
	generateServiceAccount := func(name, namespace string) *corev1.ServiceAccount {
		return &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		saExistsAlready bool
	}{
		{
			saExistsAlready: false,
		},
		{
			saExistsAlready: true,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		testSA := generateServiceAccount("test-sa", "test-ns")

		if testCase.saExistsAlready {
			runtimeObjects = append(runtimeObjects, testSA)
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects, testSA.Name, testSA.Namespace)
		err := testBuilder.Delete()
		assert.Nil(t, err)
	}
}

func TestServiceAccountWithOptions(t *testing.T) {
	testBuilder := buildValidTestBuilder()

	testBuilder.WithOptions(func(builder *Builder) (*Builder, error) {
		return builder, nil
	})

	assert.Equal(t, "", testBuilder.errorMsg)
}

func TestServiceAccountExists(t *testing.T) {
	generateServiceAccount := func(name, namespace string) *corev1.ServiceAccount {
		return &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		saExistsAlready bool
	}{
		{
			saExistsAlready: false,
		},
		{
			saExistsAlready: true,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		testSA := generateServiceAccount("test-sa", "test-ns")

		if testCase.saExistsAlready {
			runtimeObjects = append(runtimeObjects, testSA)
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects, testSA.Name, testSA.Namespace)
		result := testBuilder.Exists()
		assert.Equal(t, testCase.saExistsAlready, result)
	}
}

func TestServiceAccountValidate(t *testing.T) {
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
			expectedError: "error: received nil ServiceAccount builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined ServiceAccount",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "ServiceAccount builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestBuilder()

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

func buildTestBuilderWithFakeObjects(objects []runtime.Object, name, namespace string) *Builder {
	fakeClient := k8sfake.NewSimpleClientset(objects...)

	return NewBuilder(&clients.Settings{
		K8sClient:       fakeClient,
		CoreV1Interface: fakeClient.CoreV1(),
		AppsV1Interface: fakeClient.AppsV1(),
	}, name, namespace)
}

func buildValidTestBuilder() *Builder {
	fakeClient := k8sfake.NewSimpleClientset()

	return NewBuilder(&clients.Settings{
		K8sClient:       fakeClient,
		AppsV1Interface: fakeClient.AppsV1(),
		CoreV1Interface: fakeClient.CoreV1(),
	}, "test-sa", "test-ns")
}
