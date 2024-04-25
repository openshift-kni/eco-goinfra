package ingress

import (
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestIngressPull(t *testing.T) {
	testCases := []struct {
		ingressName         string
		ingressNamespace    string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			expectedError:       false,
			addToRuntimeObjects: true,
			expectedErrorText:   "",
			ingressName:         "test",
			ingressNamespace:    "test",
		},
		{
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "ingresscontroller object test not found in namespace test",
			ingressName:         "test",
			ingressNamespace:    "test",
		},
		{
			expectedError:       true,
			addToRuntimeObjects: true,
			expectedErrorText:   "ingresscontroller object test not found in namespace test",
			ingressName:         "",
			ingressNamespace:    "test",
		},
		{
			expectedError:       true,
			addToRuntimeObjects: true,
			expectedErrorText:   "ingresscontroller object test not found in namespace test",
			ingressName:         "test",
			ingressNamespace:    "",
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, &operatorv1.IngressController{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testCase.ingressName,
					Namespace: testCase.ingressNamespace,
				},
			})
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		builderResult, err := Pull(testSettings, "test", "test")

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

// func TestIngressUpdate(t *testing.T) {
// 	testCases := []struct {
// 		ingressExistsAlready bool
// 		name                 string
// 		namespace            string
// 	}{
// 		{
// 			ingressExistsAlready: true,
// 			name:                 "test",
// 			namespace:            "test",
// 		},
// 		{
// 			ingressExistsAlready: false,
// 			name:                 "test",
// 			namespace:            "test",
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		var runtimeObjects []runtime.Object

// 		if testCase.ingressExistsAlready {
// 			runtimeObjects = append(runtimeObjects, &operatorv1.IngressController{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      testCase.name,
// 					Namespace: testCase.namespace,
// 				},
// 			})
// 		}

// 		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects, testCase.name, testCase.namespace)

// 		testBuilder.Definition.CreationTimestamp = metav1.Time{}
// 		testBuilder.Definition.ResourceVersion = ""

// 		// Updating an ingress controller that already exists leads to failure
// 		// because it cannot be modified in place.
// 		_, err := testBuilder.Update()
// 		assert.Nil(t, err)
// 	}
// }

func TestIngressCreate(t *testing.T) {
	testCases := []struct {
		ingressExistsAlready bool
		name                 string
		namespace            string
	}{
		{
			ingressExistsAlready: true,
			name:                 "test",
			namespace:            "test",
		},
		{
			ingressExistsAlready: false,
			name:                 "test",
			namespace:            "test",
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.ingressExistsAlready {
			runtimeObjects = append(runtimeObjects, &operatorv1.IngressController{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testCase.name,
					Namespace: testCase.namespace,
				},
			})
		}

		testBuilder, client := buildTestBuilderWithFakeObjects(runtimeObjects, testCase.name, testCase.namespace)

		_, err := testBuilder.Create()
		assert.Nil(t, err)

		// Assert that the object actually exists
		_, err = Pull(client, testCase.name, testCase.namespace)
		assert.Nil(t, err)
	}
}

func TestIngressDelete(t *testing.T) {
	testCases := []struct {
		ingressExistsAlready bool
		name                 string
		namespace            string
	}{
		{
			ingressExistsAlready: true,
			name:                 "test",
			namespace:            "test",
		},
		{
			ingressExistsAlready: false,
			name:                 "test",
			namespace:            "test",
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.ingressExistsAlready {
			runtimeObjects = append(runtimeObjects, &operatorv1.IngressController{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testCase.name,
					Namespace: testCase.namespace,
				},
			})
		}

		testBuilder, client := buildTestBuilderWithFakeObjects(runtimeObjects, testCase.name, testCase.namespace)

		err := testBuilder.Delete()
		assert.Nil(t, err)

		// Assert that the object actually doesn't exist
		_, err = Pull(client, testCase.name, testCase.namespace)
		assert.NotNil(t, err)
	}
}

func TestIngressValidate(t *testing.T) {
	testCases := []struct {
		builderNil      bool
		definitionNil   bool
		readerClientNil bool
		writerClientNil bool
		expectedError   string
	}{
		{
			builderNil:      true,
			definitionNil:   false,
			readerClientNil: false,
			writerClientNil: false,
			expectedError:   "error: received nil IngressController builder",
		},
		{
			builderNil:      false,
			definitionNil:   true,
			readerClientNil: false,
			writerClientNil: false,
			expectedError:   "can not redefine the undefined IngressController",
		},
		{
			builderNil:      false,
			definitionNil:   false,
			readerClientNil: true,
			writerClientNil: false,
			expectedError:   "IngressController builder cannot have nil apiClient",
		},
		{
			builderNil:      false,
			definitionNil:   false,
			readerClientNil: false,
			writerClientNil: false,
			expectedError:   "",
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

		if testCase.readerClientNil {
			testBuilder.readerClient = nil
		}

		if testCase.writerClientNil {
			testBuilder.writerClient = nil
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

func buildTestBuilderWithFakeObjects(runtimeObjects []runtime.Object,
	name, namespace string) (*Builder, *clients.Settings) {
	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: runtimeObjects,
	})

	return &Builder{
		readerClient: testSettings.Client,
		writerClient: testSettings.Client,
		Definition: &operatorv1.IngressController{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		},
	}, testSettings
}
