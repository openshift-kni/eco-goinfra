package bmh

import (
	"fmt"
	"testing"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultHFSName      = "hostfirmwaresettings-test"
	defaultHFSNamespace = "test-ns"
)

func TestPullHFS(t *testing.T) {
	testCases := []struct {
		hfsName             string
		hfsNamespace        string
		addToRuntimeObjects bool
		client              bool
		expectedErrorText   string
	}{
		{
			hfsName:             defaultHFSName,
			hfsNamespace:        defaultHFSNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "",
		},
		{
			hfsName:             "",
			hfsNamespace:        defaultHFSNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "hostFirmwareSettings 'name' cannot be empty",
		},
		{
			hfsName:             defaultHFSName,
			hfsNamespace:        "",
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "hostFirmwareSettings 'nsname' cannot be empty",
		},
		{
			hfsName:             defaultHFSName,
			hfsNamespace:        defaultHFSNamespace,
			addToRuntimeObjects: false,
			client:              true,
			expectedErrorText: fmt.Sprintf(
				"hostFirmwareSettings object %s does not exist in namespace %s", defaultHFSName, defaultHFSNamespace),
		},
		{
			hfsName:             defaultHFSName,
			hfsNamespace:        defaultHFSNamespace,
			addToRuntimeObjects: true,
			client:              false,
			expectedErrorText:   "hostFirmwareSettings 'apiClient' cannot be nil",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testHFS := buildDummyHFS(testCase.hfsName, testCase.hfsNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testHFS)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		hfsBuilder, err := PullHFS(testSettings, testCase.hfsName, testCase.hfsNamespace)

		if testCase.expectedErrorText == "" {
			assert.Nil(t, err)
			assert.Equal(t, testHFS.Name, hfsBuilder.Definition.Name)
			assert.Equal(t, testHFS.Namespace, hfsBuilder.Definition.Namespace)
		} else {
			assert.EqualError(t, err, testCase.expectedErrorText)
		}
	}
}

func TestHFSGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *HFSBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidHFSBuilder(buildTestClientWithDummyHFS()),
			expectedError: "",
		},
		{
			testBuilder:   buildValidHFSBuilder(buildTestClientWithHFSScheme()),
			expectedError: "hostfirmwaresettingses.metal3.io \"hostfirmwaresettings-test\" not found",
		},
	}

	for _, testCase := range testCases {
		hfs, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, hfs.Name)
			assert.Equal(t, testCase.testBuilder.Definition.Namespace, hfs.Namespace)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestHFSExists(t *testing.T) {
	testCases := []struct {
		testBuilder *HFSBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidHFSBuilder(buildTestClientWithDummyHFS()),
			exists:      true,
		},
		{
			testBuilder: buildValidHFSBuilder(buildTestClientWithHFSScheme()),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestHFSCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *HFSBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidHFSBuilder(buildTestClientWithHFSScheme()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidHFSBuilder(buildTestClientWithDummyHFS()),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		hfsBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, hfsBuilder.Definition.Name, hfsBuilder.Object.Name)
			assert.Equal(t, hfsBuilder.Definition.Namespace, hfsBuilder.Object.Namespace)
		}
	}
}

func TestHFSDelete(t *testing.T) {
	testCases := []struct {
		testBuilder       *HFSBuilder
		expectedErrorText string
	}{
		{
			testBuilder:       buildValidHFSBuilder(buildTestClientWithDummyHFS()),
			expectedErrorText: "",
		},
		{
			testBuilder:       buildValidHFSBuilder(buildTestClientWithHFSScheme()),
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.Delete()

		if testCase.expectedErrorText == "" {
			assert.Nil(t, err)
			assert.Nil(t, testCase.testBuilder.Object)
		} else {
			assert.EqualError(t, err, testCase.expectedErrorText)
		}
	}
}

// buildDummyHFS returns a HostFirmwareSettings with the provided name and namespace.
func buildDummyHFS(name, namespace string) *bmhv1alpha1.HostFirmwareSettings {
	return &bmhv1alpha1.HostFirmwareSettings{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

// buildTestClientWithDummyHFS returns a client with a mock HostFirmwareSettings.
func buildTestClientWithDummyHFS() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyHFS(defaultHFSName, defaultHFSNamespace),
		},
		SchemeAttachers: testSchemes,
	})
}

// buildTestClientWithHFSScheme returns a client with no objects but the HostFirmwareSettings scheme attached.
func buildTestClientWithHFSScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: testSchemes,
	})
}

// buildValidHFSBuilder returns a valid Builder for testing.
func buildValidHFSBuilder(apiClient *clients.Settings) *HFSBuilder {
	return &HFSBuilder{
		Definition: buildDummyHFS(defaultHFSName, defaultHFSNamespace),
		apiClient:  apiClient,
	}
}
