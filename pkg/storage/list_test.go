package storage

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	defaultLabelSelector = "ceph.rook.io/DeviceSet=ocs-deviceset-0"
	defaultLabel         = map[string]string{"ceph.rook.io/DeviceSet": "ocs-deviceset-0"}
	defaultNamespace     = "persistentvolumeclaim-namespace"
)

func TestListPV(t *testing.T) {
	testCases := []struct {
		pv            []*PVBuilder
		listOptions   []metav1.ListOptions
		expectedError error
		client        bool
	}{
		{
			pv:            []*PVBuilder{buildValidPersistentVolumeTestBuilder(buildTestClientWithDummyPersistentVolume())},
			listOptions:   nil,
			expectedError: nil,
			client:        true,
		},
		{
			pv:            []*PVBuilder{buildValidPersistentVolumeTestBuilder(buildTestClientWithDummyPersistentVolume())},
			listOptions:   []metav1.ListOptions{{LabelSelector: defaultLabelSelector}},
			expectedError: nil,
			client:        true,
		},
		{
			pv:            []*PVBuilder{buildValidPersistentVolumeTestBuilder(buildTestClientWithDummyPersistentVolume())},
			listOptions:   []metav1.ListOptions{{LabelSelector: defaultLabelSelector}, {LabelSelector: defaultLabelSelector}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			pv:            []*PVBuilder{buildValidPersistentVolumeTestBuilder(buildTestClientWithDummyPersistentVolume())},
			listOptions:   []metav1.ListOptions{{LabelSelector: defaultLabelSelector}},
			expectedError: fmt.Errorf("failed to list persistentVolume, 'apiClient' parameter is empty"),
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithDummyPersistentVolume()
		}

		pvBuilders, err := ListPV(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.pv), len(pvBuilders))
		}
	}
}

func TestListPVC(t *testing.T) {
	testCases := []struct {
		testPVC       []*corev1.PersistentVolumeClaim
		namespace     string
		listOptions   []metav1.ListOptions
		expectedError error
		client        bool
	}{
		{
			testPVC: []*corev1.PersistentVolumeClaim{
				buildDummyPVCObject("pvctest1", defaultNamespace),
				buildDummyPVCObject("pvctest2", defaultNamespace),
				buildDummyPVCObject("pvctest3", defaultNamespace),
			},
			namespace:     defaultNamespace,
			listOptions:   nil,
			expectedError: nil,
			client:        true,
		},
		{
			testPVC: []*corev1.PersistentVolumeClaim{
				buildDummyPVCObject("pvctest1", defaultNamespace),
				buildDummyPVCObject("pvctest2", defaultNamespace),
				buildDummyPVCObject("pvctest3", defaultNamespace),
			},
			namespace:     "",
			listOptions:   []metav1.ListOptions{{LabelSelector: defaultLabelSelector}},
			expectedError: fmt.Errorf("PVC namespace can not be empty"),
			client:        true,
		},
		{
			testPVC: []*corev1.PersistentVolumeClaim{
				buildDummyPVCObject("pvctest1", defaultNamespace),
				buildDummyPVCObject("pvctest2", defaultNamespace),
			},
			namespace:     defaultNamespace,
			listOptions:   []metav1.ListOptions{{LabelSelector: defaultLabelSelector}},
			expectedError: nil,
			client:        true,
		},
		{
			testPVC: []*corev1.PersistentVolumeClaim{
				buildDummyPVCObject("", "")},
			namespace:     defaultNamespace,
			listOptions:   []metav1.ListOptions{{LabelSelector: defaultLabelSelector}, {LabelSelector: defaultLabelSelector}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			testPVC: []*corev1.PersistentVolumeClaim{
				buildDummyPVCObject("pvctest1", defaultNamespace),
				buildDummyPVCObject("pvctest2", defaultNamespace),
				buildDummyPVCObject("pvctest3", defaultNamespace),
			},
			namespace:     defaultNamespace,
			listOptions:   []metav1.ListOptions{{LabelSelector: defaultLabelSelector}},
			expectedError: fmt.Errorf("failed to list persistentVolumeClaim, 'apiClient' parameter is empty"),
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var (
			testSettings   *clients.Settings
			runtimeObjects []runtime.Object
		)

		for _, pvc := range testCase.testPVC {
			runtimeObjects = append(runtimeObjects, pvc)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builders, err := ListPVC(testSettings, testCase.namespace, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, len(testCase.testPVC), len(builders))
		}
	}
}

func buildDummyPVCObject(name, namespace string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    defaultLabel,
		},
	}
}
