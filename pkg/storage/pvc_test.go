package storage

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultPVCName      = "persistentvolumeclaim-test"
	defaultPVCNamespace = "persistentvolumeclaim-namespace"
)

func TestPullPersistentVolumeClaim(t *testing.T) {
	generatePVC := func(name, namespace string) *corev1.PersistentVolumeClaim {
		return &corev1.PersistentVolumeClaim{
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
		client              bool
		expectedError       error
	}{
		{
			name:                defaultPVCName,
			namespace:           defaultPVCNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                defaultPVCName,
			namespace:           defaultPVCNamespace,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("persistentVolumeClaim 'apiClient' cannot be empty"),
		},
		{
			name:                defaultPVCName,
			namespace:           defaultPVCNamespace,
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf(
				"persistentVolumeClaim object %s does not exist in namespace %s",
				defaultPVCName, defaultPVCNamespace),
		},
		{
			name:                "",
			namespace:           defaultPVCNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("persistentVolumeClaim 'name' cannot be empty"),
		},
		{
			name:                defaultPVCName,
			namespace:           "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("persistentVolumeClaim 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testPersistentVolumeClaim := generatePVC(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPersistentVolumeClaim)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		persistentVolumeClaimBuilder, err := PullPersistentVolumeClaim(testSettings,
			testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testPersistentVolumeClaim.Name, persistentVolumeClaimBuilder.Definition.Name)
			assert.Equal(t, testPersistentVolumeClaim.Namespace, persistentVolumeClaimBuilder.Definition.Namespace)
		}
	}
}

func TestNewPersistentVolumeClaimBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
		client        bool
	}{
		{
			name:          defaultPVCName,
			namespace:     defaultPVCNamespace,
			expectedError: "",
			client:        true,
		},
		{
			name:          "",
			namespace:     defaultPVCNamespace,
			expectedError: "PVC name is empty",
			client:        true,
		},
		{
			name:          defaultPVCName,
			namespace:     "",
			expectedError: "PVC namespace is empty",
			client:        true,
		},
		{
			name:          defaultPVCName,
			namespace:     defaultPVCNamespace,
			expectedError: "",
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testPVCBuilder := NewPVCBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedError == "" {
			if testCase.client {
				assert.Equal(t, testCase.name, testPVCBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testPVCBuilder.Definition.Namespace)
			} else {
				assert.Nil(t, testPVCBuilder)
			}
		} else {
			assert.Equal(t, testCase.expectedError, testPVCBuilder.errorMsg)
			assert.NotNil(t, testPVCBuilder.Definition)
		}
	}
}

func TestPersistentVolumeClaimExists(t *testing.T) {
	testCases := []struct {
		testBuilder *PVCBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidPVCTestBuilder(buildTestClientWithDummyPVC()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidPVCTestBuilder(buildTestClientWithDummyPVC()),
			exists:      false,
		},
		{
			testBuilder: buildValidPVCTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestStorageSystemCreate(t *testing.T) {
	testCases := []struct {
		testPVC       *PVCBuilder
		expectedError error
	}{
		{
			testPVC:       buildValidPVCTestBuilder(buildTestClientWithDummyPVC()),
			expectedError: nil,
		},
		{
			testPVC:       buildInvalidPVCTestBuilder(buildTestClientWithDummyPVC()),
			expectedError: fmt.Errorf("PVC name is empty"),
		},
		{
			testPVC: buildValidPVCTestBuilder(
				clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testPVCBuilder, err := testCase.testPVC.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testPVCBuilder.Definition.Name, testPVCBuilder.Object.Name)
			assert.Equal(t, testPVCBuilder.Definition.Namespace, testPVCBuilder.Object.Namespace)
		}
	}
}

func TestPersistentVolumeClaimDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *PVCBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPVCTestBuilder(buildTestClientWithDummyPVC()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidPVCTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPVCTestBuilder(buildTestClientWithDummyPVC()),
			expectedError: fmt.Errorf("PVC name is empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, err)
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

func TestPersistentVolumeClaimDeleteAndWait(t *testing.T) {
	testCases := []struct {
		testBuilder   *PVCBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPVCTestBuilder(buildTestClientWithDummyPVC()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidPVCTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPVCTestBuilder(buildTestClientWithDummyPVC()),
			expectedError: fmt.Errorf("PVC name is empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.DeleteAndWait(time.Second)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

func TestPersistentVolumeClaimWithVolumeMode(t *testing.T) {
	testCases := []struct {
		testPVC       string
		expectedError string
	}{
		{
			testPVC:       "Block",
			expectedError: "",
		},
		{
			testPVC:       "Filesystem",
			expectedError: "",
		},
		{
			testPVC: "Disk",
			expectedError: "unsupported VolumeMode \"Disk\" requested for persistentvolumeclaim-test " +
				"PersistentVolumeClaim in persistentvolumeclaim-test namespace",
		},
		{
			testPVC: "",
			expectedError: "empty volumeMode requested for the PersistentVolumeClaim persistentvolumeclaim-test " +
				"in persistentvolumeclaim-namespace namespace",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPVCTestBuilder(buildTestClientWithDummyPVC())

		result, err := testBuilder.WithVolumeMode(testCase.testPVC)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Nil(t, err)
			assert.Equal(t, corev1.PersistentVolumeMode(testCase.testPVC), *result.Definition.Spec.VolumeMode)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestPersistentVolumeClaimWithStorageClass(t *testing.T) {
	testCases := []struct {
		testStorageClass string
		expectedError    string
	}{
		{
			testStorageClass: "test-storage-class",
			expectedError:    "",
		},
		{
			testStorageClass: "",
			expectedError:    "empty storageClass requested for the PersistentVolumeClaim persistentvolumeclaim-test",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPVCTestBuilder(buildTestClientWithDummyPVC())

		result, err := testBuilder.WithStorageClass(testCase.testStorageClass)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Nil(t, err)
			assert.Equal(t, testCase.testStorageClass, *result.Definition.Spec.StorageClassName)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestPersistentVolumeClaimWithPVCCapacity(t *testing.T) {
	testCases := []struct {
		testCapacity  string
		expectedError string
	}{
		{
			testCapacity:  "5Gi",
			expectedError: "",
		},
		{
			testCapacity:  "",
			expectedError: "capacity of the PersistentVolumeClaim is empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPVCTestBuilder(buildTestClientWithDummyPVC())

		result, err := testBuilder.WithPVCCapacity(testCase.testCapacity)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Nil(t, err)

			capacityMap := make(map[corev1.ResourceName]resource.Quantity)
			capacityMap[corev1.ResourceStorage] = resource.MustParse(testCase.testCapacity)
			assert.Equal(t, corev1.VolumeResourceRequirements{Requests: capacityMap},
				result.Definition.Spec.Resources)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestPersistentVolumeClaimWithPVCAccessMode(t *testing.T) {
	testCases := []struct {
		testAccessMode string
		expectedError  string
	}{
		{
			testAccessMode: "ReadWriteOnce",
			expectedError:  "",
		},
		{
			testAccessMode: "",
			expectedError:  "Empty accessMode for PVC requested",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPVCTestBuilder(buildTestClientWithDummyPVC())

		result, err := testBuilder.WithPVCAccessMode(testCase.testAccessMode)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Nil(t, err)

			assert.Equal(t, corev1.PersistentVolumeAccessMode(testCase.testAccessMode),
				result.Definition.Spec.AccessModes[0])
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func buildValidPVCTestBuilder(apiClient *clients.Settings) *PVCBuilder {
	pvcBuilder := NewPVCBuilder(
		apiClient, defaultPVCName, defaultPVCNamespace)

	return pvcBuilder
}

func buildInvalidPVCTestBuilder(apiClient *clients.Settings) *PVCBuilder {
	pvcBuilder := NewPVCBuilder(
		apiClient, "", defaultPVCNamespace)

	return pvcBuilder
}

func buildTestClientWithDummyPVC() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyPersistentVolumeClaim(),
	})
}

func buildDummyPersistentVolumeClaim() []runtime.Object {
	return append([]runtime.Object{}, &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultPVCName,
			Namespace: defaultPVCNamespace,
		},
	})
}
