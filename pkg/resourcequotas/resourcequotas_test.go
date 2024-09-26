package resourcequotas

import (
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

//nolint:funlen
func TestResourceQuotaPull(t *testing.T) {
	generateResourceQuota := func(name, namespace string) *corev1.ResourceQuota {
		return &corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		testName            string
		testNamespace       string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
		apiClientNil        bool
	}{
		{
			testName:            "test1",
			testNamespace:       "testNamespace",
			expectedError:       false,
			addToRuntimeObjects: true,
			apiClientNil:        false,
		},
		{
			testName:            "test2",
			testNamespace:       "testNamespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "resource quota test2 does not exist in namespace testNamespace",
			apiClientNil:        false,
		},
		{
			testName:            "",
			testNamespace:       "testNamespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "resource quota 'name' cannot be empty",
			apiClientNil:        false,
		},
		{
			testName:            "test3",
			testNamespace:       "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "resource quota 'namespace' cannot be empty",
			apiClientNil:        false,
		},
		{
			testName:            "test4",
			testNamespace:       "testNamespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "apiClient is nil",
			apiClientNil:        true,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testResourceQuota := generateResourceQuota(testCase.testName, testCase.testNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testResourceQuota)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		if testCase.apiClientNil {
			testSettings = nil
		}

		builderResult, err := Pull(testSettings, testCase.testName, testCase.testNamespace)

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testResourceQuota.Name, builderResult.Object.Name)
			assert.Equal(t, testResourceQuota.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestResourceQuotaCreate(t *testing.T) {
	generateResourceQuota := func(name, namespace string) *corev1.ResourceQuota {
		return &corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		rqExistsAlready bool
	}{
		{
			rqExistsAlready: true,
		},
		{
			rqExistsAlready: false,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		testRQ := generateResourceQuota("testRQ", "testNamespace")

		if testCase.rqExistsAlready {
			runtimeObjects = append(runtimeObjects, testRQ)
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects, testRQ.Name, testRQ.Namespace)
		builderResult, err := testBuilder.Create()

		assert.Nil(t, err)
		assert.NotNil(t, testBuilder.Object)
		assert.Equal(t, testRQ.Name, builderResult.Object.Name)
		assert.Equal(t, testRQ.Namespace, builderResult.Object.Namespace)
	}
}

func TestResourceQuotaDelete(t *testing.T) {
	generateResourceQuota := func(name, namespace string) *corev1.ResourceQuota {
		return &corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		rqExistsAlready bool
	}{
		{
			rqExistsAlready: true,
		},
		{
			rqExistsAlready: false,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		testRQ := generateResourceQuota("testRQ", "testNamespace")

		if testCase.rqExistsAlready {
			runtimeObjects = append(runtimeObjects, testRQ)
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects, testRQ.Name, testRQ.Namespace)
		err := testBuilder.Delete()

		assert.Nil(t, err)
		assert.Nil(t, testBuilder.Object)
	}
}

func TestValidate(t *testing.T) {
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
			expectedError: "error: received nil ResourceQuota builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined ResourceQuota",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "ResourceQuota builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildTestBuilderWithFakeObjects(nil, "testRQ", "testNamespace")

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

func buildTestBuilderWithFakeObjects(runtimeObjects []runtime.Object, name, namespace string) *Builder {
	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: runtimeObjects,
	})

	testBuilder := &Builder{
		apiClient: testSettings.CoreV1Interface,
		Definition: &corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		},
	}

	return testBuilder
}
