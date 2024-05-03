package storage

import (
	"fmt"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	ocsoperatorv1 "github.com/red-hat-storage/ocs-operator/api/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

var (
	storageClusterGVK = schema.GroupVersionKind{
		Group:   APIGroup,
		Version: APIVersion,
		Kind:    StorageClusterKind,
	}
	defaultStorageClusterName      = "ocs-storagecluster"
	defaultStorageClusterNamespace = "openshift-storage"
	defaultManageNodes             = false
)

func TestSorageClusterPull(t *testing.T) {
	generateStorageCluster := func(name, namespace string) *ocsoperatorv1.StorageCluster {
		return &ocsoperatorv1.StorageCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
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
			namespace:           "openshift-storage",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "openshift-storage",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("storageCluster 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "test",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("storageCluster 'namespace' cannot be empty"),
			client:              true,
		},
		{
			name:                "ocstest",
			namespace:           "openshift-storage",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("storageCluster object ocstest doesn't exist in namespace openshift-storage"),
			client:              true,
		},
		{
			name:                "ocstest",
			namespace:           "openshift-storage",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("storageCluster 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testStorageCluster := generateStorageCluster(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testStorageCluster)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := PullStorageCluster(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.Equal(t, testStorageCluster.Name, builderResult.Object.Name)
		}
	}
}

func TestStorageClusterExist(t *testing.T) {
	testCases := []struct {
		testStorageCluster *StorageClusterBuilder
		expectedStatus     bool
	}{
		{
			testStorageCluster: buildValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject()),
			expectedStatus:     true,
		},
		{
			testStorageCluster: buildInValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject()),
			expectedStatus:     false,
		},
		{
			testStorageCluster: buildValidStorageClusterBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:     false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testStorageCluster.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestStorageClusterGet(t *testing.T) {
	testCases := []struct {
		testStorageCluster *StorageClusterBuilder
		expectedError      error
	}{
		{
			testStorageCluster: buildValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject()),
			expectedError:      nil,
		},
		{
			testStorageCluster: buildInValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject()),
			expectedError:      fmt.Errorf("storageCluster 'name' cannot be empty"),
		},
		{
			testStorageCluster: buildValidStorageClusterBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:      fmt.Errorf("configs.imageregistry.operator.openshift.io \"cluster\" not found"),
		},
	}

	for _, testCase := range testCases {
		imageRegistryObj, err := testCase.testStorageCluster.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, imageRegistryObj, testCase.testStorageCluster.Definition)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestStorageClusterUpdate(t *testing.T) {
	testCases := []struct {
		testStorageCluster *StorageClusterBuilder
		expectedError      error
		manageNodes        bool
	}{
		{
			testStorageCluster: buildValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject()),
			expectedError:      nil,
			manageNodes:        true,
		},
		{
			testStorageCluster: buildInValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject()),
			expectedError:      fmt.Errorf("storageCluster 'name' cannot be empty"),
			manageNodes:        true,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, defaultManageNodes, testCase.testStorageCluster.Definition.Spec.ManageNodes)
		assert.Nil(t, nil, testCase.testStorageCluster.Object)
		testCase.testStorageCluster.WithManagedNodes(testCase.manageNodes)
		_, err := testCase.testStorageCluster.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.manageNodes, testCase.testStorageCluster.Definition.Spec.ManageNodes)
		}
	}
}

func TestStorageClusterWithManageNodes(t *testing.T) {
	testCases := []struct {
		testManageNodes   bool
		expectedError     bool
		expectedErrorText string
	}{
		{
			testManageNodes:   true,
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testManageNodes:   false,
			expectedError:     false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject())

		result := testBuilder.WithManagedNodes(testCase.testManageNodes)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testManageNodes, result.Definition.Spec.ManageNodes)
		}
	}
}

func TestStorageClusterGetManageNodes(t *testing.T) {
	testCases := []struct {
		testStorageCluster *StorageClusterBuilder
		expectedError      error
	}{
		{
			testStorageCluster: buildValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject()),
			expectedError:      nil,
		},
		{
			testStorageCluster: buildValidStorageClusterBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:      fmt.Errorf("storageCluster object doesn't exist"),
		},
	}

	for _, testCase := range testCases {
		currentStorageClusterManageNodesValue, err := testCase.testStorageCluster.GetManageNodes()

		if testCase.expectedError == nil {
			assert.Equal(t, currentStorageClusterManageNodesValue, testCase.testStorageCluster.Object.Spec.ManageNodes)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func buildValidStorageClusterBuilder(apiClient *clients.Settings) *StorageClusterBuilder {
	return StorageClusterNewBuilder(apiClient, defaultStorageClusterName, defaultStorageClusterNamespace)
}

func buildInValidStorageClusterBuilder(apiClient *clients.Settings) *StorageClusterBuilder {
	return StorageClusterNewBuilder(apiClient, "", defaultStorageClusterNamespace)
}

func buildStorageClusterClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyStorageCluster(),
		GVK:            []schema.GroupVersionKind{storageClusterGVK},
	})
}

func buildDummyStorageCluster() []runtime.Object {
	return append([]runtime.Object{}, &ocsoperatorv1.StorageCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultStorageClusterName,
			Namespace: defaultStorageClusterNamespace,
		},
		Spec: ocsoperatorv1.StorageClusterSpec{
			ManageNodes:        false,
			ManagedResources:   ocsoperatorv1.ManagedResourcesSpec{},
			MonDataDirHostPath: "",
			MultiCloudGateway:  nil,
			StorageDeviceSets:  make([]ocsoperatorv1.StorageDeviceSet, 0),
		},
	})
}
