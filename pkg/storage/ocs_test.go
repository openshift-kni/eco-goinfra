package storage

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	ocsoperatorv1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/ocs/operatorv1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultStorageClusterName      = "ocs-storagecluster"
	defaultStorageClusterNamespace = "openshift-storage"
	defaultStorageClassName        = "ocs-storagecluster-cephfs"
	defaultVolumeMode              = corev1.PersistentVolumeBlock
	errStorageClusterNotExists     = fmt.Errorf("storageCluster object ocs-storagecluster does not exist in " +
		"namespace openshift-storage")
	ocsTestSchemes = []clients.SchemeAttacher{
		ocsoperatorv1.AddToScheme,
	}
)

//nolint:funlen
func TestSorageClusterPull(t *testing.T) {
	generateStorageCluster := func(name, namespace string) *ocsoperatorv1.StorageCluster {
		return &ocsoperatorv1.StorageCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: ocsoperatorv1.StorageClusterSpec{
				ManagedResources: ocsoperatorv1.ManagedResourcesSpec{
					CephBlockPools: ocsoperatorv1.ManageCephBlockPools{
						ReconcileStrategy: "manage",
					},
					CephFilesystems: ocsoperatorv1.ManageCephFilesystems{
						ReconcileStrategy: "manage",
					},
					CephObjectStoreUsers: ocsoperatorv1.ManageCephObjectStoreUsers{
						ReconcileStrategy: "manage",
					},
					CephObjectStores: ocsoperatorv1.ManageCephObjectStores{
						ReconcileStrategy: "manage",
					},
				},
				MonDataDirHostPath: "/var/lib/rook",
				MultiCloudGateway: &ocsoperatorv1.MultiCloudGatewaySpec{
					ReconcileStrategy: "manage",
				},
				StorageDeviceSets: make([]ocsoperatorv1.StorageDeviceSet, 0),
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
			expectedError:       fmt.Errorf("storageCluster object ocstest does not exist in namespace openshift-storage"),
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
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: ocsTestSchemes,
			})
		}

		builderResult, err := PullStorageCluster(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testStorageCluster.Name, builderResult.Object.Name)
			assert.Equal(t, testStorageCluster.Namespace, builderResult.Object.Namespace)
			assert.Nil(t, err)
		}
	}
}

func TestNewStorageClusterBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
		client        bool
	}{
		{
			name:          defaultStorageClusterName,
			namespace:     defaultStorageClusterNamespace,
			expectedError: "",
			client:        true,
		},
		{
			name:          "",
			namespace:     defaultStorageClusterNamespace,
			expectedError: "storageCluster 'name' cannot be empty",
			client:        true,
		},
		{
			name:          defaultStorageClusterName,
			namespace:     "",
			expectedError: "storageCluster 'nsname' cannot be empty",
			client:        true,
		},
		{
			name:          defaultStorageClusterName,
			namespace:     defaultStorageClusterNamespace,
			expectedError: "",
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testStorageClusterBuilder := NewStorageClusterBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedError == "" {
			if testCase.client {
				assert.Equal(t, testCase.name, testStorageClusterBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testStorageClusterBuilder.Definition.Namespace)
			} else {
				assert.Nil(t, testStorageClusterBuilder)
			}
		} else {
			assert.Equal(t, testCase.expectedError, testStorageClusterBuilder.errorMsg)
			assert.NotNil(t, testStorageClusterBuilder.Definition)
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
			expectedError:      fmt.Errorf("storageclusters.ocs.openshift.io \"ocs-storagecluster\" not found"),
		},
	}

	for _, testCase := range testCases {
		storageClusterObj, err := testCase.testStorageCluster.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, storageClusterObj.Name, testCase.testStorageCluster.Definition.Name)
			assert.Equal(t, storageClusterObj.Namespace, testCase.testStorageCluster.Definition.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestStorageClusterCreate(t *testing.T) {
	testCases := []struct {
		testStorageCluster *StorageClusterBuilder
		expectedError      string
	}{
		{
			testStorageCluster: buildValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject()),
			expectedError:      "",
		},
		{
			testStorageCluster: buildInValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject()),
			expectedError:      "storageCluster 'name' cannot be empty",
		},
		{
			testStorageCluster: buildValidStorageClusterBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:      "",
		},
	}

	for _, testCase := range testCases {
		testStorageClusterBuilder, err := testCase.testStorageCluster.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testStorageClusterBuilder.Definition.Name, testStorageClusterBuilder.Object.Name)
			assert.Equal(t, testStorageClusterBuilder.Definition.Namespace, testStorageClusterBuilder.Object.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestStorageClusterDelete(t *testing.T) {
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
			expectedError:      nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testStorageCluster.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testStorageCluster.Object)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestStorageClusterUpdate(t *testing.T) {
	testCases := []struct {
		testStorageCluster *StorageClusterBuilder
		expectedError      error
		flexibleScaling    bool
	}{
		{
			testStorageCluster: buildValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject()),
			expectedError:      nil,
			flexibleScaling:    true,
		},
		{
			testStorageCluster: buildInValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject()),
			expectedError:      fmt.Errorf("storageCluster 'name' cannot be empty"),
			flexibleScaling:    true,
		},
	}

	for _, testCase := range testCases {
		assert.False(t, testCase.testStorageCluster.Definition.Spec.FlexibleScaling)
		assert.Nil(t, nil, testCase.testStorageCluster.Object)
		testCase.testStorageCluster.WithFlexibleScaling(testCase.flexibleScaling)
		_, err := testCase.testStorageCluster.Update()

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.flexibleScaling, testCase.testStorageCluster.Definition.Spec.FlexibleScaling)
		} else {
			assert.Equal(t, testCase.expectedError, err)
		}
	}
}

func TestStorageClusterWithFlexibleScaling(t *testing.T) {
	testCases := []struct {
		testFlexibleScaling bool
		expectedError       string
	}{
		{
			testFlexibleScaling: true,
			expectedError:       "",
		},
		{
			testFlexibleScaling: false,
			expectedError:       "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject())

		result := testBuilder.WithFlexibleScaling(testCase.testFlexibleScaling)

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testFlexibleScaling, result.Definition.Spec.FlexibleScaling)
		}
	}
}

func TestStorageClusterWithManagedResources(t *testing.T) {
	testCases := []struct {
		testManagedResources ocsoperatorv1.ManagedResourcesSpec
		expectedError        string
	}{
		{
			testManagedResources: ocsoperatorv1.ManagedResourcesSpec{
				CephBlockPools: ocsoperatorv1.ManageCephBlockPools{
					ReconcileStrategy: "manage",
				},
				CephFilesystems: ocsoperatorv1.ManageCephFilesystems{
					ReconcileStrategy: "manage",
				},
				CephObjectStoreUsers: ocsoperatorv1.ManageCephObjectStoreUsers{
					ReconcileStrategy: "manage",
				},
				CephObjectStores: ocsoperatorv1.ManageCephObjectStores{
					ReconcileStrategy: "manage",
				},
			},
			expectedError: "",
		},
		{
			testManagedResources: ocsoperatorv1.ManagedResourcesSpec{},
			expectedError:        "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject())

		result := testBuilder.WithManagedResources(testCase.testManagedResources)

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testManagedResources, result.Definition.Spec.ManagedResources)
		}
	}
}

func TestStorageClusterGetManagedResources(t *testing.T) {
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
			expectedError:      errStorageClusterNotExists,
		},
	}

	for _, testCase := range testCases {
		currentStorageClusterManagedResourcesValue, err := testCase.testStorageCluster.GetManagedResources()

		if testCase.expectedError == nil {
			assert.Equal(t, *currentStorageClusterManagedResourcesValue,
				testCase.testStorageCluster.Object.Spec.ManagedResources)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestStorageClusterWithMonDataDirHostPath(t *testing.T) {
	testCases := []struct {
		testMonDataDirHostPath string
		expectedError          string
	}{
		{
			testMonDataDirHostPath: "/var/lib/rook",
			expectedError:          "",
		},
		{
			testMonDataDirHostPath: "",
			expectedError:          "the expectedMonDataDirHostPath can not be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject())

		result := testBuilder.WithMonDataDirHostPath(testCase.testMonDataDirHostPath)

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testMonDataDirHostPath, result.Definition.Spec.MonDataDirHostPath)
		}
	}
}

func TestStorageClusterGetMonDataDirHostPath(t *testing.T) {
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
			expectedError:      errStorageClusterNotExists,
		},
	}

	for _, testCase := range testCases {
		currentStorageClusterMonDataDirHostPathValue, err := testCase.testStorageCluster.GetMonDataDirHostPath()

		if testCase.expectedError == nil {
			assert.Equal(t, currentStorageClusterMonDataDirHostPathValue,
				testCase.testStorageCluster.Object.Spec.MonDataDirHostPath)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestStorageClusterWithMultiCloudGateway(t *testing.T) {
	testCases := []struct {
		testMultiCloudGateway ocsoperatorv1.MultiCloudGatewaySpec
		expectedError         string
	}{
		{
			testMultiCloudGateway: ocsoperatorv1.MultiCloudGatewaySpec{
				ReconcileStrategy: "manage",
			},
			expectedError: "",
		},
		{
			testMultiCloudGateway: ocsoperatorv1.MultiCloudGatewaySpec{},
			expectedError:         "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject())

		result := testBuilder.WithMultiCloudGateway(testCase.testMultiCloudGateway)

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, &testCase.testMultiCloudGateway, result.Definition.Spec.MultiCloudGateway)
		}
	}
}

func TestStorageClusterGetMultiCloudGateway(t *testing.T) {
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
			expectedError:      errStorageClusterNotExists,
		},
	}

	for _, testCase := range testCases {
		currentStorageClusterMultiCloudGatewayValue, err := testCase.testStorageCluster.GetMultiCloudGateway()

		if testCase.expectedError == nil {
			assert.Equal(t, currentStorageClusterMultiCloudGatewayValue,
				testCase.testStorageCluster.Object.Spec.MultiCloudGateway)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestStorageClusterWithStorageDeviceSets(t *testing.T) {
	resourceListMap := make(map[corev1.ResourceName]resource.Quantity)
	resourceListMap[corev1.ResourceStorage] = resource.MustParse("1")

	testCases := []struct {
		testStorageDeviceSets ocsoperatorv1.StorageDeviceSet
		expectedError         string
	}{
		{
			testStorageDeviceSets: ocsoperatorv1.StorageDeviceSet{
				Count:    3,
				Replica:  1,
				Portable: false,
				Name:     "local-block",
				DataPVCTemplate: corev1.PersistentVolumeClaim{
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.VolumeResourceRequirements{
							Requests: resourceListMap,
						},
						StorageClassName: &defaultStorageClassName,
						VolumeMode:       &defaultVolumeMode,
					},
				},
			},
			expectedError: "",
		},
		{
			testStorageDeviceSets: ocsoperatorv1.StorageDeviceSet{},
			expectedError:         "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject())

		result := testBuilder.WithStorageDeviceSet(testCase.testStorageDeviceSets)

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, []ocsoperatorv1.StorageDeviceSet{testCase.testStorageDeviceSets},
				result.Definition.Spec.StorageDeviceSets)
		}
	}
}

func TestStorageClusterGetStorageDeviceSets(t *testing.T) {
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
			expectedError:      errStorageClusterNotExists,
		},
	}

	for _, testCase := range testCases {
		currentStorageClusterStorageDeviceSets, err := testCase.testStorageCluster.GetStorageDeviceSets()

		if testCase.expectedError == nil {
			assert.Equal(t, currentStorageClusterStorageDeviceSets,
				testCase.testStorageCluster.Object.Spec.StorageDeviceSets)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestStorageClusterWithAnnotations(t *testing.T) {
	testCases := []struct {
		testAnnotations   map[string]string
		expectedErrorText string
	}{
		{
			testAnnotations:   map[string]string{"uninstall.ocs.openshift.io/cleanup-policy": "delete"},
			expectedErrorText: "",
		},
		{
			testAnnotations: map[string]string{"uninstall.ocs.openshift.io/cleanup-policy": "delete",
				"uninstall.ocs.openshift.io/mode": "graceful"},
			expectedErrorText: "",
		},
		{
			testAnnotations:   map[string]string{},
			expectedErrorText: "'annotations' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidStorageClusterBuilder(buildStorageClusterClientWithDummyObject())

		result := testBuilder.WithAnnotations(testCase.testAnnotations)

		if testCase.expectedErrorText != "" {
			assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testAnnotations, result.Definition.Annotations)
		}
	}
}

func buildValidStorageClusterBuilder(apiClient *clients.Settings) *StorageClusterBuilder {
	storageClusterBuilder := NewStorageClusterBuilder(
		apiClient, defaultStorageClusterName, defaultStorageClusterNamespace)

	return storageClusterBuilder
}

func buildInValidStorageClusterBuilder(apiClient *clients.Settings) *StorageClusterBuilder {
	storageClusterBuilder := NewStorageClusterBuilder(
		apiClient, "", defaultStorageClusterNamespace)

	return storageClusterBuilder
}

func buildStorageClusterClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyStorageCluster(),
		SchemeAttachers: ocsTestSchemes,
	})
}

func buildDummyStorageCluster() []runtime.Object {
	return append([]runtime.Object{}, &ocsoperatorv1.StorageCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultStorageClusterName,
			Namespace: defaultStorageClusterNamespace,
		},
		Spec: ocsoperatorv1.StorageClusterSpec{
			ManagedResources:   ocsoperatorv1.ManagedResourcesSpec{},
			MonDataDirHostPath: "",
			MultiCloudGateway:  nil,
			StorageDeviceSets:  make([]ocsoperatorv1.StorageDeviceSet, 0),
		},
	})
}
