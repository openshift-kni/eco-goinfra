package lso

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	lsov1 "github.com/openshift/local-storage-operator/api/v1"
	lsov1alpha1 "github.com/openshift/local-storage-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultLocalVolumeSetName      = "ocs-deviceset"
	defaultLocalVolumeSetNamespace = "test-lvsspace"
)

func TestPullLocalVolumeSet(t *testing.T) {
	generateLocalVolumeSet := func(name, namespace string) *lsov1alpha1.LocalVolumeSet {
		return &lsov1alpha1.LocalVolumeSet{
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
			name:                defaultLocalVolumeSetName,
			namespace:           defaultLocalVolumeSetNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           defaultLocalVolumeSetNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("localVolumeSet 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                defaultLocalVolumeSetName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("localVolumeSet 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "lvstest",
			namespace:           defaultLocalVolumeSetNamespace,
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf("localVolumeSet object lvstest does not exist " +
				"in namespace test-lvsspace"),
			client: true,
		},
		{
			name:                "lvstest",
			namespace:           defaultLocalVolumeSetNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("localVolumeSet 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testLocalVolumeSet := generateLocalVolumeSet(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testLocalVolumeSet)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := PullLocalVolumeSet(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testLocalVolumeSet.Name, builderResult.Object.Name)
			assert.Equal(t, testLocalVolumeSet.Namespace, builderResult.Object.Namespace)
			assert.Nil(t, err)
		}
	}
}

func TestNewLocalVolumeSetBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
		client        bool
	}{
		{
			name:          defaultLocalVolumeSetName,
			namespace:     defaultLocalVolumeSetNamespace,
			expectedError: "",
			client:        true,
		},
		{
			name:          "",
			namespace:     defaultLocalVolumeSetNamespace,
			expectedError: "localVolumeSet 'name' cannot be empty",
			client:        true,
		},
		{
			name:          defaultLocalVolumeSetName,
			namespace:     "",
			expectedError: "localVolumeSet 'nsname' cannot be empty",
			client:        true,
		},
		{
			name:          defaultLocalVolumeSetName,
			namespace:     defaultLocalVolumeSetNamespace,
			expectedError: "",
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings
		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testLocalVolumeSet := NewLocalVolumeSetBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedError == "" {
			if testCase.client {
				assert.Equal(t, testCase.name, testLocalVolumeSet.Definition.Name)
				assert.Equal(t, testCase.namespace, testLocalVolumeSet.Definition.Namespace)
			} else {
				assert.Nil(t, testLocalVolumeSet)
			}
		} else {
			assert.Equal(t, testCase.expectedError, testLocalVolumeSet.errorMsg)
			assert.NotNil(t, testLocalVolumeSet.Definition)
		}
	}
}

func TestLocalVolumeSetExists(t *testing.T) {
	testCases := []struct {
		testLocalVolumeSet *LocalVolumeSetBuilder
		expectedStatus     bool
	}{
		{
			testLocalVolumeSet: buildValidLocalVolumeSetObjectBuilder(buildLocalVolumeSetClientWithDummyObject()),
			expectedStatus:     true,
		},
		{
			testLocalVolumeSet: buildInValidLocalVolumeSetObjectBuilder(buildLocalVolumeSetClientWithDummyObject()),
			expectedStatus:     false,
		},
		{
			testLocalVolumeSet: buildValidLocalVolumeSetObjectBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:     false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testLocalVolumeSet.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestLocalVolumeSetGet(t *testing.T) {
	testCases := []struct {
		testLocalVolumeSet *LocalVolumeSetBuilder
		expectedError      error
	}{
		{
			testLocalVolumeSet: buildValidLocalVolumeSetObjectBuilder(buildLocalVolumeSetClientWithDummyObject()),
			expectedError:      nil,
		},
		{
			testLocalVolumeSet: buildInValidLocalVolumeSetObjectBuilder(buildLocalVolumeSetClientWithDummyObject()),
			expectedError:      fmt.Errorf("localVolumeSet 'name' cannot be empty"),
		},
		{
			testLocalVolumeSet: buildValidLocalVolumeSetObjectBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:      fmt.Errorf("localvolumesets.local.storage.openshift.io \"ocs-deviceset\" not found"),
		},
	}

	for _, testCase := range testCases {
		localVolumeSetObj, err := testCase.testLocalVolumeSet.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, localVolumeSetObj.Name, testCase.testLocalVolumeSet.Definition.Name)
			assert.Equal(t, localVolumeSetObj.Namespace, testCase.testLocalVolumeSet.Definition.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestLocalVolumeSetCreate(t *testing.T) {
	testCases := []struct {
		testLocalVolumeSet *LocalVolumeSetBuilder
		expectedError      string
	}{
		{
			testLocalVolumeSet: buildValidLocalVolumeSetObjectBuilder(buildLocalVolumeSetClientWithDummyObject()),
			expectedError:      "",
		},
		{
			testLocalVolumeSet: buildInValidLocalVolumeSetObjectBuilder(buildLocalVolumeSetClientWithDummyObject()),
			expectedError:      "localVolumeSet 'name' cannot be empty",
		},
		{
			testLocalVolumeSet: buildValidLocalVolumeSetObjectBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:      "",
		},
	}

	for _, testCase := range testCases {
		testLocalVolumeSetBuilder, err := testCase.testLocalVolumeSet.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testLocalVolumeSetBuilder.Definition.Name, testLocalVolumeSetBuilder.Object.Name)
			assert.Equal(t, testLocalVolumeSetBuilder.Definition.Namespace, testLocalVolumeSetBuilder.Object.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestLocalVolumeSetDelete(t *testing.T) {
	testCases := []struct {
		testLocalVolumeSet *LocalVolumeSetBuilder
		expectedError      error
	}{
		{
			testLocalVolumeSet: buildValidLocalVolumeSetObjectBuilder(buildLocalVolumeSetClientWithDummyObject()),
			expectedError:      nil,
		},
		{
			testLocalVolumeSet: buildValidLocalVolumeSetObjectBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:      nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testLocalVolumeSet.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testLocalVolumeSet.Object)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestLocalVolumeSetUpdate(t *testing.T) {
	testCases := []struct {
		testLocalVolumeSet   *LocalVolumeSetBuilder
		testStorageClassName string
		expectedError        string
	}{
		{
			testLocalVolumeSet:   buildValidLocalVolumeSetObjectBuilder(buildLocalVolumeSetClientWithDummyObject()),
			testStorageClassName: "test-storage-class",
			expectedError:        "",
		},
		{
			testLocalVolumeSet:   buildInValidLocalVolumeSetObjectBuilder(buildLocalVolumeSetClientWithDummyObject()),
			testStorageClassName: "",
			expectedError:        "localVolumeSet 'name' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, "", testCase.testLocalVolumeSet.Definition.Spec.StorageClassName)
		assert.Nil(t, nil, testCase.testLocalVolumeSet.Object)
		testCase.testLocalVolumeSet.WithStorageClassName(testCase.testStorageClassName)
		_, err := testCase.testLocalVolumeSet.Update()

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.testStorageClassName, testCase.testLocalVolumeSet.Definition.Spec.StorageClassName)
		}
	}
}

func TestLocalVolumeSetWithTolerations(t *testing.T) {
	testCases := []struct {
		testTolerations   []corev1.Toleration
		expectedErrorText string
	}{
		{
			testTolerations: []corev1.Toleration{{
				Key:      "node.ocs.openshift.io/storage",
				Operator: "Equal",
				Value:    "true",
				Effect:   "NoSchedule",
			}},
			expectedErrorText: "",
		},
		{
			testTolerations:   []corev1.Toleration{},
			expectedErrorText: "'tolerations' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLocalVolumeSetObjectBuilder(buildLocalVolumeSetClientWithDummyObject())

		result := testBuilder.WithTolerations(testCase.testTolerations)
		assert.Equal(t, testCase.expectedErrorText, result.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testTolerations, result.Definition.Spec.Tolerations)
		}
	}
}

func TestLocalVolumeSetWithNodeSelector(t *testing.T) {
	testCases := []struct {
		testNodeSelector corev1.NodeSelector
		expectedError    string
	}{
		{
			testNodeSelector: corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "cluster.ocs.openshift.io/openshift-storage",
						Operator: "In",
						Values:   []string{""},
					}}},
				},
			},
			expectedError: "",
		},
		{
			testNodeSelector: corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "cluster.ocs.openshift.io/openshift-storage",
						Operator: "Exists",
					}}},
				},
			},
			expectedError: "",
		},
		{
			testNodeSelector: corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "cluster.ocs.openshift.io/openshift-storage",
						Operator: "Exists",
					}, {
						Key:      "machineconfiguration.openshift.io/role",
						Operator: "In",
						Values:   []string{"customcnf", "worker"},
					}}},
				},
			},
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLocalVolumeSetObjectBuilder(buildLocalVolumeSetClientWithDummyObject())

		result := testBuilder.WithNodeSelector(testCase.testNodeSelector)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testNodeSelector, *result.Definition.Spec.NodeSelector)
		}
	}
}

func TestLocalVolumeSetWithStorageClassName(t *testing.T) {
	testCases := []struct {
		testStorageClassName string
		expectedError        string
	}{
		{
			testStorageClassName: "test-storage-class",
			expectedError:        "",
		},
		{
			testStorageClassName: "",
			expectedError:        "'storageClassName' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLocalVolumeSetObjectBuilder(buildLocalVolumeSetClientWithDummyObject())

		result := testBuilder.WithStorageClassName(testCase.testStorageClassName)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testStorageClassName, result.Definition.Spec.StorageClassName)
		}
	}
}

func TestLocalVolumeSetWithVolumeMode(t *testing.T) {
	testCases := []struct {
		testVolumeMode lsov1.PersistentVolumeMode
		expectedError  string
	}{
		{
			testVolumeMode: lsov1.PersistentVolumeBlock,
			expectedError:  "",
		},
		{
			testVolumeMode: lsov1.PersistentVolumeFilesystem,
			expectedError:  "",
		},
		{
			testVolumeMode: "",
			expectedError:  "'volumeMode' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLocalVolumeSetObjectBuilder(buildLocalVolumeSetClientWithDummyObject())

		result := testBuilder.WithVolumeMode(testCase.testVolumeMode)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testVolumeMode, result.Definition.Spec.VolumeMode)
		}
	}
}

func TestLocalVolumeSetWithFSType(t *testing.T) {
	testCases := []struct {
		testFSType    string
		expectedError string
	}{
		{
			testFSType:    "ext4",
			expectedError: "",
		},
		{
			testFSType:    "",
			expectedError: "'fstype' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLocalVolumeSetObjectBuilder(buildLocalVolumeSetClientWithDummyObject())

		result := testBuilder.WithFSType(testCase.testFSType)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testFSType, result.Definition.Spec.FSType)
		}
	}
}

func TestLocalVolumeSetWithMaxDeviceCount(t *testing.T) {
	testCases := []struct {
		testMaxDeviceCount int32
		expectedError      string
	}{
		{
			testMaxDeviceCount: int32(42),
			expectedError:      "",
		},
		{
			testMaxDeviceCount: int32(0),
			expectedError:      "'maxDeviceCount' argument cannot be equal zero",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLocalVolumeSetObjectBuilder(buildLocalVolumeSetClientWithDummyObject())

		result := testBuilder.WithMaxDeviceCount(testCase.testMaxDeviceCount)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testMaxDeviceCount, *result.Definition.Spec.MaxDeviceCount)
		}
	}
}

func TestLocalVolumeSetWithDeviceInclusionSpec(t *testing.T) {
	testCases := []struct {
		testDeviceInclusionSpec lsov1alpha1.DeviceInclusionSpec
		expectedError           string
	}{
		{
			testDeviceInclusionSpec: lsov1alpha1.DeviceInclusionSpec{
				DeviceTypes:                []lsov1alpha1.DeviceType{lsov1alpha1.RawDisk},
				DeviceMechanicalProperties: []lsov1alpha1.DeviceMechanicalProperty{lsov1alpha1.NonRotational},
			},
			expectedError: "",
		},
		{
			testDeviceInclusionSpec: lsov1alpha1.DeviceInclusionSpec{
				DeviceTypes:                []lsov1alpha1.DeviceType{lsov1alpha1.Partition},
				DeviceMechanicalProperties: []lsov1alpha1.DeviceMechanicalProperty{lsov1alpha1.Rotational},
			},
			expectedError: "",
		},
		{
			testDeviceInclusionSpec: lsov1alpha1.DeviceInclusionSpec{
				DeviceTypes:                []lsov1alpha1.DeviceType{lsov1alpha1.Loop},
				DeviceMechanicalProperties: []lsov1alpha1.DeviceMechanicalProperty{lsov1alpha1.Rotational},
			},
			expectedError: "",
		},
		{
			testDeviceInclusionSpec: lsov1alpha1.DeviceInclusionSpec{
				DeviceTypes:                []lsov1alpha1.DeviceType{lsov1alpha1.MultiPath},
				DeviceMechanicalProperties: []lsov1alpha1.DeviceMechanicalProperty{lsov1alpha1.NonRotational},
			},
			expectedError: "",
		},
		{
			testDeviceInclusionSpec: lsov1alpha1.DeviceInclusionSpec{
				DeviceTypes: []lsov1alpha1.DeviceType{lsov1alpha1.RawDisk,
					lsov1alpha1.MultiPath},
				DeviceMechanicalProperties: []lsov1alpha1.DeviceMechanicalProperty{lsov1alpha1.Rotational},
			},
			expectedError: "",
		},
		{
			testDeviceInclusionSpec: lsov1alpha1.DeviceInclusionSpec{
				DeviceTypes: []lsov1alpha1.DeviceType{lsov1alpha1.RawDisk},
			},
			expectedError: "",
		},
		{
			testDeviceInclusionSpec: lsov1alpha1.DeviceInclusionSpec{
				DeviceMechanicalProperties: []lsov1alpha1.DeviceMechanicalProperty{lsov1alpha1.NonRotational},
			},
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLocalVolumeSetObjectBuilder(buildLocalVolumeSetClientWithDummyObject())

		result := testBuilder.WithDeviceInclusionSpec(testCase.testDeviceInclusionSpec)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testDeviceInclusionSpec, *result.Definition.Spec.DeviceInclusionSpec)
		}
	}
}

func buildValidLocalVolumeSetObjectBuilder(apiClient *clients.Settings) *LocalVolumeSetBuilder {
	localVolumeSet := NewLocalVolumeSetBuilder(
		apiClient, defaultLocalVolumeSetName, defaultLocalVolumeSetNamespace)

	return localVolumeSet
}

func buildInValidLocalVolumeSetObjectBuilder(apiClient *clients.Settings) *LocalVolumeSetBuilder {
	localVolumeSet := NewLocalVolumeSetBuilder(
		apiClient, "", defaultLocalVolumeSetNamespace)

	return localVolumeSet
}

func buildLocalVolumeSetClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyLocalVolumeSet(),
	})
}

func buildDummyLocalVolumeSet() []runtime.Object {
	return append([]runtime.Object{}, &lsov1alpha1.LocalVolumeSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultLocalVolumeSetName,
			Namespace: defaultLocalVolumeSetNamespace,
		},
	})
}
