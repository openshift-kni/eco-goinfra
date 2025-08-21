package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const defaultPersistentVolumeName = "persistentvolume-test"

func TestPullPersistentVolume(t *testing.T) {
	testCases := []struct {
		persistentVolumeName string
		addToRuntimeObjects  bool
		client               bool
		expectedErrorText    string
	}{
		{
			persistentVolumeName: defaultPersistentVolumeName,
			addToRuntimeObjects:  true,
			client:               true,
			expectedErrorText:    "",
		},
		{
			persistentVolumeName: defaultPersistentVolumeName,
			addToRuntimeObjects:  true,
			client:               false,
			expectedErrorText:    "persistentVolume 'apiClient' cannot be empty",
		},
		{
			persistentVolumeName: defaultPersistentVolumeName,
			addToRuntimeObjects:  false,
			client:               true,
			expectedErrorText:    fmt.Sprintf("PersistentVolume object %s does not exist", defaultPersistentVolumeName),
		},
		{
			persistentVolumeName: "",
			addToRuntimeObjects:  true,
			client:               true,
			expectedErrorText:    "persistentVolume 'name' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testPersistentVolume := buildDummyPersistentVolume(testCase.persistentVolumeName)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPersistentVolume)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		persistentVolumeBuilder, err := PullPersistentVolume(testSettings, testCase.persistentVolumeName)

		if testCase.expectedErrorText == "" {
			assert.Nil(t, err)
			assert.Equal(t, testPersistentVolume.Name, persistentVolumeBuilder.Definition.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedErrorText)
		}
	}
}

func TestPersistentVolumeExists(t *testing.T) {
	testCases := []struct {
		testBuilder *PVBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidPersistentVolumeTestBuilder(buildTestClientWithDummyPersistentVolume()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidPersistentVolumeTestBuilder(buildTestClientWithDummyPersistentVolume()),
			exists:      false,
		},
		{
			testBuilder: buildValidPersistentVolumeTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestPersistentVolumeDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *PVBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPersistentVolumeTestBuilder(buildTestClientWithDummyPersistentVolume()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidPersistentVolumeTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPersistentVolumeTestBuilder(buildTestClientWithDummyPersistentVolume()),
			expectedError: fmt.Errorf("can not redefine the undefined PersistentVolume"),
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

func TestPersistentVolumeDeleteAndWait(t *testing.T) {
	testCases := []struct {
		testBuilder       *PVBuilder
		expectedErrorText string
	}{
		{
			testBuilder:       buildValidPersistentVolumeTestBuilder(buildTestClientWithDummyPersistentVolume()),
			expectedErrorText: "",
		},
		{
			testBuilder:       buildValidPersistentVolumeTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErrorText: "",
		},
		{
			testBuilder:       buildInvalidPersistentVolumeTestBuilder(buildTestClientWithDummyPersistentVolume()),
			expectedErrorText: "can not redefine the undefined PersistentVolume",
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.DeleteAndWait(time.Second)

		if testCase.expectedErrorText == "" {
			assert.Nil(t, err)
			assert.Nil(t, testCase.testBuilder.Object)
		} else {
			assert.EqualError(t, err, testCase.expectedErrorText)
		}
	}
}

func TestPersistentVolumeWaitUntilDeleted(t *testing.T) {
	testCases := []struct {
		testBuilder   *PVBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPersistentVolumeTestBuilder(buildTestClientWithDummyPersistentVolume()),
			expectedError: context.DeadlineExceeded,
		},
		{
			testBuilder:   buildValidPersistentVolumeTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPersistentVolumeTestBuilder(buildTestClientWithDummyPersistentVolume()),
			expectedError: fmt.Errorf("can not redefine the undefined PersistentVolume"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.WaitUntilDeleted(time.Second)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

// buildDummyPersistentVolume returns a PersistentVolume with the specified name.
func buildDummyPersistentVolume(name string) *corev1.PersistentVolume {
	return &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestClientWithDummyPersistentVolume returns a client with a mock PersistentVolume with the default name.
func buildTestClientWithDummyPersistentVolume() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPersistentVolume(defaultPersistentVolumeName),
		},
	})
}

// buildValidPersistentVolumeTestBuilder returns a valid PVBuilder with the default name and specified client.
func buildValidPersistentVolumeTestBuilder(apiClient *clients.Settings) *PVBuilder {
	return &PVBuilder{
		apiClient:  apiClient,
		Definition: buildDummyPersistentVolume(defaultPersistentVolumeName),
	}
}

// buildInvalidPersistentVolumeTestBuilder returns an invalid PVBuilder with no definition and the specified client.
func buildInvalidPersistentVolumeTestBuilder(apiClient *clients.Settings) *PVBuilder {
	return &PVBuilder{
		apiClient: apiClient,
	}
}
