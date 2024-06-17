package multiclusterhub

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	mchv1 "github.com/stolostron/multiclusterhub-operator/api/v1"
)

var (
	testValue string = "test"
)

func TestMultiClusterHubPull(t *testing.T) {
	testCases := []struct {
		expectedError            error
		multiClusterHubName      string
		multiClusterHubNamespace string
		addToRuntimeObjects      bool
	}{
		{
			expectedError:            nil,
			multiClusterHubName:      testValue,
			multiClusterHubNamespace: testValue,
			addToRuntimeObjects:      true,
		},

		{
			expectedError:            fmt.Errorf("multiclusterhub 'name' cannot be empty"),
			multiClusterHubName:      "",
			multiClusterHubNamespace: testValue,
			addToRuntimeObjects:      true,
		},

		{
			expectedError:            fmt.Errorf("multiclusterhub 'namespace' cannot be empty"),
			multiClusterHubName:      testValue,
			multiClusterHubNamespace: "",
			addToRuntimeObjects:      true,
		},

		{
			expectedError:            fmt.Errorf("multiclusterhub 'name' cannot be empty"),
			multiClusterHubName:      "",
			multiClusterHubNamespace: "",
			addToRuntimeObjects:      true,
		},

		{
			expectedError: fmt.Errorf("multiclusterhub object %s in namespace %s does not exist",
				testValue, testValue),
			multiClusterHubName:      testValue,
			multiClusterHubNamespace: testValue,
			addToRuntimeObjects:      false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testMultiClusterHub := generateMultiClusterHub(
			testCase.multiClusterHubName, testCase.multiClusterHubNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testMultiClusterHub)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		// Test the PullMultiClusterHub function
		builderResult, err := PullMultiClusterHub(testSettings,
			testCase.multiClusterHubName, testCase.multiClusterHubNamespace)

		// Check the error
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
		}
	}
}

func TestMultiClusterHubGet(t *testing.T) {
	testCases := []struct {
		expectedError            error
		multiClusterHubName      string
		multiClusterHubNamespace string
		addToRuntimeObjects      bool
	}{
		{
			addToRuntimeObjects: true,
		},

		{
			addToRuntimeObjects: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, &mchv1.MultiClusterHub{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testValue,
					Namespace: testValue,
				},
			})
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		testBuilder, _ := generateMultiClusterHubBuilder(testSettings, testValue, testValue)

		// Test the Get function
		result, err := testBuilder.Get()

		if testCase.addToRuntimeObjects {
			assert.Nil(t, err)
			assert.NotNil(t, result)
		} else {
			assert.NotNil(t, err)
			assert.Nil(t, result)
		}
	}
}

func TestMultiClusterHubCreate(t *testing.T) {
	testCases := []struct {
		addToRuntimeObjects bool
	}{
		{
			addToRuntimeObjects: false,
		},
		{
			addToRuntimeObjects: true,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, &mchv1.MultiClusterHub{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testValue,
					Namespace: testValue,
				},
			})
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		testBuilder, client := generateMultiClusterHubBuilder(testSettings, testValue, testValue)

		// Testing the create function
		_, err := testBuilder.Create()
		assert.Nil(t, err)

		// Assert that the object actually exists
		_, err = PullMultiClusterHub(client, testValue, testValue)
		assert.Nil(t, err)
	}
}

func TestMultiClusterHubUpdate(t *testing.T) {
	testCases := []struct {
		expectedError            error
		addToRuntimeObjects      bool
		multiClusterHubName      string
		multiClusterHubNamespace string
		newImageName             string
	}{
		{
			expectedError:            nil,
			addToRuntimeObjects:      true,
			multiClusterHubName:      testValue,
			multiClusterHubNamespace: testValue,
			newImageName:             "new-image",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testMultiClusterHub := generateMultiClusterHub(
			testCase.multiClusterHubName, testCase.multiClusterHubNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testMultiClusterHub)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		builder, err := PullMultiClusterHub(testSettings,
			testCase.multiClusterHubName, testCase.multiClusterHubNamespace)
		assert.Nil(t, err)

		// Test the Update function
		builder.Definition.Spec.ImagePullSecret = testCase.newImageName
		builderResult, err := builder.Update()

		// Check the error
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.Equal(t, builderResult.Object.Spec.ImagePullSecret, testCase.newImageName)
		}
	}
}

func TestMultiClusterHubDelete(t *testing.T) {
	testCases := []struct {
		expectedError       error
		addToRuntimeObjects bool
	}{
		{
			addToRuntimeObjects: true,
		},
		{
			addToRuntimeObjects: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, &mchv1.MultiClusterHub{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testValue,
					Namespace: testValue,
				},
			})
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		testBuilder, client := generateMultiClusterHubBuilder(testSettings, testValue, testValue)

		// Testing the delete function
		err := testBuilder.Delete()
		assert.Nil(t, err)

		// Assert that the object actually does not exist
		_, err = PullMultiClusterHub(client, testValue, testValue)
		assert.NotNil(t, err)
	}
}

func TestMultiClusterHubExists(t *testing.T) {
	testCases := []struct {
		expectedExists      bool
		addToRuntimeObjects bool
		expectedError       error
	}{
		{
			expectedExists:      true,
			addToRuntimeObjects: true,
			expectedError:       nil,
		},
		{
			expectedExists:      false,
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("MultiClusterHub object test does not exist"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testMultiClusterHub := generateMultiClusterHub(
			testValue, testValue)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testMultiClusterHub)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		builder, err := PullMultiClusterHub(testSettings,
			testValue, testValue)
		if testCase.expectedError == nil {
			assert.Nil(t, err)
		}

		// Test the Exists function
		result := builder.Exists()

		// Check the result
		assert.Equal(t, result, testCase.expectedExists)
	}
}

func TestMultiClusterHubValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		expectedError error
	}{
		{
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: fmt.Errorf("error: received nil MultiClusterHub builder"),
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: fmt.Errorf("can not redefine the undefined MultiClusterHub"),
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: fmt.Errorf("MultiClusterHub builder cannot have nil apiClient"),
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testMultiClusterHub := generateMultiClusterHub(testValue, testValue)

		runtimeObjects = append(runtimeObjects, testMultiClusterHub)

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		testBuilder, err := PullMultiClusterHub(testSettings,
			testValue, testValue)

		if testCase.expectedError == nil {
			assert.Nil(t, err)
		}

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
		if testCase.expectedError != nil {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedError, err)
			assert.False(t, result)
		} else {
			assert.Nil(t, err)
			assert.True(t, result)
		}
	}
}

func generateMultiClusterHub(name, namespace string) *mchv1.MultiClusterHub {
	return &mchv1.MultiClusterHub{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: mchv1.MultiClusterHubSpec{
			ImagePullSecret: "image",
		},
	}
}

func generateMultiClusterHubBuilder(testSettings *clients.Settings, name, namespace string) (*MultiClusterHubBuilder,
	*clients.Settings) {
	return &MultiClusterHubBuilder{
		apiClient: testSettings.Client,
		Definition: &mchv1.MultiClusterHub{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		},
	}, testSettings
}
