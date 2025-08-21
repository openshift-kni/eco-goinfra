package clusterlogging

import (
	"fmt"
	"testing"

	clov1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultClusterLoggingName   = "clusterlogging"
	defaultClusterLoggingNsName = "test-namespace"
)

func TestClusterLoggingPull(t *testing.T) {
	generateClusterLogging := func(name, namespace string) *clov1.ClusterLogging {
		return &clov1.ClusterLogging{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: clov1.ClusterLoggingSpec{},
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
			name:                "clusterlogging",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterLogging 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "clusterlogging",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterLogging 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "clusterlogging",
			namespace:           "test-namespace",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("clusterLogging object clusterlogging does not exist in namespace test-namespace"),
			client:              true,
		},
		{
			name:                "clusterlogging",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterLogging 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testClusterLogging := generateClusterLogging(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testClusterLogging)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := Pull(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

func TestClusterLoggingNewBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
	}{
		{
			name:          "metallbio",
			namespace:     "test-namespace",
			expectedError: "",
		},
		{
			name:          "",
			namespace:     "test-namespace",
			expectedError: "the clusterLogging 'name' cannot be empty",
		},
		{
			name:          "metallbio",
			namespace:     "",
			expectedError: "the clusterLogging 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testClusterLogging := NewBuilder(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, testClusterLogging.errorMsg)
		assert.NotNil(t, testClusterLogging.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testClusterLogging.Definition.Name)
			assert.Equal(t, testCase.namespace, testClusterLogging.Definition.Namespace)
		}
	}
}

func TestClusterLoggingGet(t *testing.T) {
	testCases := []struct {
		clusterLogging *Builder
		expectedError  error
	}{
		{
			clusterLogging: buildValidClusterLogging(buildClusterLoggingTestClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			clusterLogging: buildValidClusterLogging(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("clusterloggings.logging.openshift.io \"clusterlogging\" not found"),
		},
		{
			clusterLogging: buildInValidClusterLogging(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("the clusterLogging 'name' cannot be empty"),
		},
		{
			clusterLogging: buildInValidClusterLogging(buildClusterLoggingTestClientWithDummyObject()),
			expectedError:  fmt.Errorf("the clusterLogging 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		clusterLogging, err := testCase.clusterLogging.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, clusterLogging.Name, testCase.clusterLogging.Definition.Name)
			assert.Equal(t, clusterLogging.Namespace, testCase.clusterLogging.Definition.Namespace)
		} else {
			assert.Equal(t, err.Error(), testCase.expectedError.Error())
		}
	}
}

func TestClusterLoggingCreate(t *testing.T) {
	testCases := []struct {
		clusterLogging *Builder
		expectedError  error
	}{
		{
			clusterLogging: buildValidClusterLogging(buildClusterLoggingTestClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			clusterLogging: buildValidClusterLogging(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  nil,
		},
		{
			clusterLogging: buildInValidClusterLogging(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("the clusterLogging 'name' cannot be empty"),
		},
		{
			clusterLogging: buildInValidClusterLogging(buildClusterLoggingTestClientWithDummyObject()),
			expectedError:  fmt.Errorf("the clusterLogging 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		clusterLogging, err := testCase.clusterLogging.Create()

		if testCase.expectedError == nil {
			assert.Equal(t, clusterLogging.Definition.Name, clusterLogging.Object.Name)
			assert.Equal(t, clusterLogging.Definition.Namespace, clusterLogging.Object.Namespace)
		} else {
			assert.Equal(t, err.Error(), testCase.expectedError.Error())
		}
	}
}

func TestClusterLoggingDelete(t *testing.T) {
	testCases := []struct {
		clusterLogging *Builder
		expectedError  error
	}{
		{
			clusterLogging: buildValidClusterLogging(buildClusterLoggingTestClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			clusterLogging: buildValidClusterLogging(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("clusterLogging cannot be deleted because it does not exist"),
		},
		{
			clusterLogging: buildInValidClusterLogging(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("the clusterLogging 'name' cannot be empty"),
		},
		{
			clusterLogging: buildInValidClusterLogging(buildClusterLoggingTestClientWithDummyObject()),
			expectedError:  fmt.Errorf("the clusterLogging 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.clusterLogging.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.clusterLogging.Object)
		}
	}
}

func TestClusterLoggingExist(t *testing.T) {
	testCases := []struct {
		clusterLogging *Builder
		expectedStatus bool
	}{
		{
			clusterLogging: buildValidClusterLogging(buildClusterLoggingTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			clusterLogging: buildValidClusterLogging(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},

		{
			clusterLogging: buildInValidClusterLogging(buildClusterLoggingTestClientWithDummyObject()),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.clusterLogging.Exists()
		assert.Equal(t, exist, testCase.expectedStatus)
	}
}

func TestClusterLoggingUpdate(t *testing.T) {
	testCases := []struct {
		clusterLogging *Builder
		expectedError  error
	}{
		{
			clusterLogging: buildValidClusterLogging(buildClusterLoggingTestClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			clusterLogging: buildValidClusterLogging(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("clusterloggings.logging.openshift.io \"clusterlogging\" not found"),
		},
		{
			clusterLogging: buildInValidClusterLogging(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("the clusterLogging 'name' cannot be empty"),
		},
		{
			clusterLogging: buildInValidClusterLogging(buildClusterLoggingTestClientWithDummyObject()),
			expectedError:  fmt.Errorf("the clusterLogging 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		assert.Empty(t, testCase.clusterLogging.Definition.Spec.ManagementState)
		testCase.clusterLogging.Definition.Spec.ManagementState = "test"
		testCase.clusterLogging.Definition.ResourceVersion = "999"
		assert.Nil(t, nil, testCase.clusterLogging.Object)

		_, err := testCase.clusterLogging.Update(false)
		if testCase.expectedError != nil {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func buildValidClusterLogging(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultClusterLoggingName, defaultClusterLoggingNsName)
}

func buildInValidClusterLogging(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "", defaultClusterLoggingNsName)
}

func buildClusterLoggingTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyClusterLogging(),
	})
}

func buildDummyClusterLogging() []runtime.Object {
	return append([]runtime.Object{}, &clov1.ClusterLogging{
		Spec: clov1.ClusterLoggingSpec{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultClusterLoggingName,
			Namespace: defaultClusterLoggingNsName,
		},
	})
}
