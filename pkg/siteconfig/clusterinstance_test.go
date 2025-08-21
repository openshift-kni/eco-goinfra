package siteconfig

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	siteconfigv1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/siteconfig/v1alpha1"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
)

const (
	testClusterInstance = "test-cluster-instance"
)

var testSchemes = []clients.SchemeAttacher{
	siteconfigv1alpha1.AddToScheme,
}

func TestClusterInstancePull(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		client        bool
		exists        bool
		expectedError error
	}{
		{
			name:          testClusterInstance,
			namespace:     testClusterInstance,
			client:        true,
			exists:        true,
			expectedError: nil,
		},
		{
			name:          "",
			namespace:     testClusterInstance,
			client:        true,
			exists:        true,
			expectedError: fmt.Errorf("clusterinstance 'name' cannot be empty"),
		},
		{
			name:          testClusterInstance,
			namespace:     "",
			client:        true,
			exists:        true,
			expectedError: fmt.Errorf("clusterinstance 'nsname' cannot be empty"),
		},
		{
			name:          testClusterInstance,
			namespace:     testClusterInstance,
			client:        false,
			exists:        true,
			expectedError: fmt.Errorf("apiClient cannot be nil"),
		},
		{
			name:      testClusterInstance,
			namespace: testClusterInstance,
			client:    true,
			exists:    false,
			expectedError: fmt.Errorf(
				"clusterinstance object %s does not exist in namespace %s",
				testClusterInstance, testClusterInstance),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testClient     *clients.Settings
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateClusterInstance())
		}

		if testCase.client {
			testClient = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := PullClusterInstance(testClient, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testBuilder.Definition.Namespace)
		}
	}
}

func TestClusterInstanceGet(t *testing.T) {
	testCases := []struct {
		exists bool
	}{
		{
			exists: true,
		},
		{
			exists: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateClusterInstance())
		}

		testBuilder := generateClusterInstanceBuilderWithFakeObjects(runtimeObjects)

		clusterinstance, err := testBuilder.Get()
		if testCase.exists {
			assert.Nil(t, err)
			assert.NotNil(t, clusterinstance)
		} else {
			assert.NotNil(t, err)
			assert.Nil(t, clusterinstance)
		}
	}
}

func TestClusterInstanceCreate(t *testing.T) {
	testCases := []struct {
		exists bool
	}{
		{
			exists: true,
		},
		{
			exists: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateClusterInstance())
		}

		testBuilder := generateClusterInstanceBuilderWithFakeObjects(runtimeObjects)

		result, err := testBuilder.Create()
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testClusterInstance, result.Definition.Name)
		assert.Equal(t, testClusterInstance, result.Definition.Namespace)
	}
}

func TestClusterInstanceDelete(t *testing.T) {
	testCases := []struct {
		exists        bool
		expectedError error
	}{
		{
			exists:        true,
			expectedError: nil,
		},
		{
			exists:        false,
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateClusterInstance())
		}

		testBuilder := generateClusterInstanceBuilderWithFakeObjects(runtimeObjects)

		err := testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testBuilder.Object)
		}
	}
}
func TestClusterInstanceExists(t *testing.T) {
	testCases := []struct {
		exists bool
	}{
		{
			exists: true,
		},
		{
			exists: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateClusterInstance())
		}

		testBuilder := generateClusterInstanceBuilderWithFakeObjects(runtimeObjects)

		assert.Equal(t, testCase.exists, testBuilder.Exists())
	}
}

func TestClusterInstanceValidate(t *testing.T) {
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
			expectedError: "error: received nil ClusterInstance builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined ClusterInstance",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "ClusterInstance builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateClusterInstanceBuilderWithFakeObjects([]runtime.Object{})

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

func generateClusterInstanceBuilderWithFakeObjects(objects []runtime.Object) *ClusterInstanceBuilder {
	return &ClusterInstanceBuilder{
		apiClient: clients.GetTestClients(
			clients.TestClientParams{K8sMockObjects: objects, SchemeAttachers: testSchemes}).Client,
		Definition: generateClusterInstance(),
	}
}

func generateClusterInstance() *siteconfigv1alpha1.ClusterInstance {
	return &siteconfigv1alpha1.ClusterInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testClusterInstance,
			Namespace: testClusterInstance,
		},
	}
}
