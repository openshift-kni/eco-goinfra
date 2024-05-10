package clusterlogging

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	clov1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	clusterLoggingGVK = schema.GroupVersionKind{
		Group:   APIGroup,
		Version: APIVersion,
		Kind:    CLOKind,
	}
	defaultClusterLoggingName      = "instance"
	defaultClusterLoggingNamespace = "openshift-logging"
	defaultManagementState         = clov1.ManagementState("")
)

//nolint:funlen
func TestClusterLoggingPull(t *testing.T) {
	generateClusterLogging := func(name, namespace string) *clov1.ClusterLogging {
		return &clov1.ClusterLogging{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: clov1.ClusterLoggingSpec{
				ManagementState: clov1.ManagementStateManaged,
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
			name:                "test",
			namespace:           "openshift-logging",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "openshift-logging",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterLogging 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "test",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterLogging 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "clotest",
			namespace:           "openshift-logging",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("clusterLogging object clotest does not exist in namespace openshift-logging"),
			client:              true,
		},
		{
			name:                "clotest",
			namespace:           "openshift-logging",
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

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testClusterLogging.Name, builderResult.Object.Name)
		}
	}
}

func TestNewBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
	}{
		{
			name:          defaultClusterLoggingName,
			namespace:     defaultClusterLoggingNamespace,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     defaultClusterLoggingNamespace,
			expectedError: "clusterLogging 'name' cannot be empty",
		},
		{
			name:          defaultClusterLoggingName,
			namespace:     "",
			expectedError: "clusterLogging 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{
			GVK: []schema.GroupVersionKind{clusterLoggingGVK},
		})
		testClusterLoggingBuilder := NewBuilder(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, testClusterLoggingBuilder.errorMsg)
		assert.NotNil(t, testClusterLoggingBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testClusterLoggingBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testClusterLoggingBuilder.Definition.Namespace)
		}
	}
}

func TestClusterLoggingExist(t *testing.T) {
	testCases := []struct {
		testClusterLogging *Builder
		expectedStatus     bool
	}{
		{
			testClusterLogging: buildValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
			expectedStatus:     true,
		},
		{
			testClusterLogging: buildInValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
			expectedStatus:     false,
		},
		{
			testClusterLogging: buildValidClusterLoggingBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:     false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testClusterLogging.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestClusterLoggingGet(t *testing.T) {
	testCases := []struct {
		testClusterLogging *Builder
		expectedError      error
	}{
		{
			testClusterLogging: buildValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
			expectedError:      nil,
		},
		{
			testClusterLogging: buildInValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
			expectedError:      fmt.Errorf("clusterloggings.logging.openshift.io \"\" not found"),
		},
		{
			testClusterLogging: buildValidClusterLoggingBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:      fmt.Errorf("clusterloggings.logging.openshift.io \"instance\" not found"),
		},
	}

	for _, testCase := range testCases {
		storageClusterObj, err := testCase.testClusterLogging.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, storageClusterObj, testCase.testClusterLogging.Definition)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestClusterLoggingCreate(t *testing.T) {
	testCases := []struct {
		testClusterLogging *Builder
		expectedError      string
	}{
		{
			testClusterLogging: buildValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
			expectedError:      "",
		},
		{
			testClusterLogging: buildInValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
			expectedError: fmt.Sprintf("ClusterLogging.logging.openshift.io \"\" is invalid: " +
				"metadata.name: Required value: name is required"),
		},
	}

	for _, testCase := range testCases {
		testClusterLoggingBuilder, err := testCase.testClusterLogging.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testClusterLoggingBuilder.Definition, testClusterLoggingBuilder.Object)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestClusterLoggingDelete(t *testing.T) {
	testCases := []struct {
		testClusterLogging *Builder
		expectedError      error
	}{
		{
			testClusterLogging: buildValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
			expectedError:      nil,
		},
		{
			testClusterLogging: buildInValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
			expectedError:      fmt.Errorf("clusterLogging cannot be deleted because it does not exist"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testClusterLogging.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testClusterLogging.Object)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestClusterLoggingUpdate(t *testing.T) {
	testCases := []struct {
		testClusterLogging *Builder
		expectedError      string
		managementState    clov1.ManagementState
	}{
		{
			testClusterLogging: buildValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
			expectedError:      "",
			managementState:    clov1.ManagementStateManaged,
		},
		{
			testClusterLogging: buildInValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
			expectedError: fmt.Sprintf("ClusterLogging.logging.openshift.io \"\" is invalid: " +
				"metadata.name: Required value: name is required"),
			managementState: clov1.ManagementStateManaged,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, defaultManagementState, testCase.testClusterLogging.Definition.Spec.ManagementState)
		assert.Nil(t, nil, testCase.testClusterLogging.Object)
		testCase.testClusterLogging.WithManagementState(testCase.managementState)
		_, err := testCase.testClusterLogging.Update()

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.managementState, testCase.testClusterLogging.Definition.Spec.ManagementState)
		}
	}
}

func TestWithManagementState(t *testing.T) {
	testCases := []struct {
		testManagementState clov1.ManagementState
		expectedError       bool
		expectedErrorText   string
	}{
		{
			testManagementState: clov1.ManagementStateUnmanaged,
			expectedError:       false,
			expectedErrorText:   "",
		},
		{
			testManagementState: clov1.ManagementStateManaged,
			expectedError:       false,
			expectedErrorText:   "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject())

		result := testBuilder.WithManagementState(testCase.testManagementState)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testManagementState, result.Definition.Spec.ManagementState)
		}
	}
}

func TestImageRegistryGetManagementState(t *testing.T) {
	testCases := []struct {
		testImageRegistry *Builder
		expectedError     error
	}{
		{
			testImageRegistry: buildValidClusterLoggingBuilder(buildClusterLoggingClientWithDummyObject()),
			expectedError:     nil,
		},
		{
			testImageRegistry: buildValidClusterLoggingBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:     fmt.Errorf("clusterLogging object does not exist"),
		},
	}

	for _, testCase := range testCases {
		currentManagementState, err := testCase.testImageRegistry.GetManagementState()

		if testCase.expectedError == nil {
			assert.Equal(t, *currentManagementState, testCase.testImageRegistry.Object.Spec.ManagementState)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func buildValidClusterLoggingBuilder(apiClient *clients.Settings) *Builder {
	clusterLoggingBuilder := NewBuilder(
		apiClient, defaultClusterLoggingName, defaultClusterLoggingNamespace)
	clusterLoggingBuilder.Definition.ResourceVersion = "999"
	clusterLoggingBuilder.Definition.Spec.ManagementState = ""

	return clusterLoggingBuilder
}

func buildInValidClusterLoggingBuilder(apiClient *clients.Settings) *Builder {
	clusterLoggingBuilder := NewBuilder(
		apiClient, "", defaultClusterLoggingNamespace)
	clusterLoggingBuilder.Definition.ResourceVersion = "999"
	clusterLoggingBuilder.Definition.Spec.ManagementState = ""

	return clusterLoggingBuilder
}

func buildClusterLoggingClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyClusterLogging(),
		GVK:            []schema.GroupVersionKind{clusterLoggingGVK},
	})
}

func buildDummyClusterLogging() []runtime.Object {
	return append([]runtime.Object{}, &clov1.ClusterLogging{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultClusterLoggingName,
			Namespace: defaultClusterLoggingNamespace,
		},
		Spec: clov1.ClusterLoggingSpec{
			ManagementState: "",
		},
	})
}
