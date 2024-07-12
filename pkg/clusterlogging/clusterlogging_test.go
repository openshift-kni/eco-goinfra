package clusterlogging

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	clov1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
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
			clusterLogging: buildValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			clusterLogging: buildValidClusterLoggingBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("clusterloggings.logging.openshift.io \"clusterlogging\" not found"),
		},
		{
			clusterLogging: buildInValidClusterLoggingBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("the clusterLogging 'name' cannot be empty"),
		},
		{
			clusterLogging: buildInValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
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
			clusterLogging: buildValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			clusterLogging: buildValidClusterLoggingBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  nil,
		},
		{
			clusterLogging: buildInValidClusterLoggingBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("the clusterLogging 'name' cannot be empty"),
		},
		{
			clusterLogging: buildInValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
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
			clusterLogging: buildValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			clusterLogging: buildValidClusterLoggingBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("clusterLogging cannot be deleted because it does not exist"),
		},
		{
			clusterLogging: buildInValidClusterLoggingBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("the clusterLogging 'name' cannot be empty"),
		},
		{
			clusterLogging: buildInValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
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
			clusterLogging: buildValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			clusterLogging: buildValidClusterLoggingBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},

		{
			clusterLogging: buildInValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
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
			clusterLogging: buildValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			clusterLogging: buildValidClusterLoggingBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("clusterloggings.logging.openshift.io \"clusterlogging\" not found"),
		},
		{
			clusterLogging: buildInValidClusterLoggingBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("the clusterLogging 'name' cannot be empty"),
		},
		{
			clusterLogging: buildInValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
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

func TestClusterLoggingWithCollection(t *testing.T) {
	testCases := []struct {
		testCollection clov1.CollectionSpec
		expectedError  string
	}{
		{
			testCollection: clov1.CollectionSpec{
				Type: "vector",
				CollectorSpec: clov1.CollectorSpec{
					Tolerations: []corev1.Toleration{{
						Key:      "node-role.kubernetes.io/infra",
						Operator: "Exists",
					}, {
						Key:      "node.ocs.openshift.io/storage",
						Operator: "Equal",
						Value:    "true",
						Effect:   "NoSchedule",
					}},
				},
			},
			expectedError: "",
		},
		{
			testCollection: clov1.CollectionSpec{},
			expectedError:  "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject())

		result := testBuilder.WithCollection(testCase.testCollection)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testCollection, *result.Definition.Spec.Collection)
		}
	}
}

func TestClusterLoggingWithManagementState(t *testing.T) {
	testCases := []struct {
		testManagementState clov1.ManagementState
		expectedError       string
	}{
		{
			testManagementState: clov1.ManagementStateManaged,
			expectedError:       "",
		},
		{
			testManagementState: clov1.ManagementStateUnmanaged,
			expectedError:       "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject())

		result := testBuilder.WithManagementState(testCase.testManagementState)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testManagementState, result.Definition.Spec.ManagementState)
		}
	}
}

func TestClusterLoggingWithLogStore(t *testing.T) {
	testCases := []struct {
		testLogStore  clov1.LogStoreSpec
		expectedError string
	}{
		{
			testLogStore: clov1.LogStoreSpec{
				Type:      "lokistack",
				LokiStack: clov1.LokiStackStoreSpec{Name: "logging-loki"},
			},
			expectedError: "",
		},
		{
			testLogStore:  clov1.LogStoreSpec{},
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject())

		result := testBuilder.WithLogStore(testCase.testLogStore)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testLogStore, *result.Definition.Spec.LogStore)
		}
	}
}

func TestClusterLoggingWithVisualization(t *testing.T) {
	testCases := []struct {
		testVisualization clov1.VisualizationSpec
		expectedError     string
	}{
		{
			testVisualization: clov1.VisualizationSpec{
				Type:         "ocp-console",
				NodeSelector: map[string]string{"node-role.kubernetes.io/infra": ""},
				Tolerations: []corev1.Toleration{{
					Key:      "node-role.kubernetes.io/infra",
					Operator: "Exists",
				}},
			},
			expectedError: "",
		},
		{
			testVisualization: clov1.VisualizationSpec{},
			expectedError:     "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject())

		result := testBuilder.WithVisualization(testCase.testVisualization)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testVisualization, *result.Definition.Spec.Visualization)
		}
	}
}

func buildValidClusterLoggingBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultClusterLoggingName, defaultClusterLoggingNsName)
}

func buildInValidClusterLoggingBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "", defaultClusterLoggingNsName)
}

func buildClusterLoggingClientWithDummyObject() *clients.Settings {
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
