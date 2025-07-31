package bmh

import (
	"fmt"
	"testing"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultHFCName      = "hostfirmwarecomponents-test"
	defaultHFCNamespace = "test-ns"
)

func TestPullHFC(t *testing.T) {
	testCases := []struct {
		hfcName             string
		hfcNamespace        string
		addToRuntimeObjects bool
		client              bool
		expectedErrorText   string
	}{
		{
			hfcName:             defaultHFCName,
			hfcNamespace:        defaultHFCNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "",
		},
		{
			hfcName:             "",
			hfcNamespace:        defaultHFCNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "hostFirmwareComponents 'name' cannot be empty",
		},
		{
			hfcName:             defaultHFCName,
			hfcNamespace:        "",
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "hostFirmwareComponents 'nsname' cannot be empty",
		},
		{
			hfcName:             defaultHFCName,
			hfcNamespace:        defaultHFCNamespace,
			addToRuntimeObjects: false,
			client:              true,
			expectedErrorText: fmt.Sprintf(
				"hostFirmwareComponents object %s does not exist in namespace %s", defaultHFCName, defaultHFCNamespace),
		},
		{
			hfcName:             defaultHFCName,
			hfcNamespace:        defaultHFCNamespace,
			addToRuntimeObjects: true,
			client:              false,
			expectedErrorText:   "hostFirmwareComponents 'apiClient' cannot be nil",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testHFC := buildDummyHFC(testCase.hfcName, testCase.hfcNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testHFC)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		hfcBuilder, err := PullHFC(testSettings, testCase.hfcName, testCase.hfcNamespace)

		if testCase.expectedErrorText == "" {
			assert.Nil(t, err)
			assert.Equal(t, testHFC.Name, hfcBuilder.Definition.Name)
			assert.Equal(t, testHFC.Namespace, hfcBuilder.Definition.Namespace)
		} else {
			assert.EqualError(t, err, testCase.expectedErrorText)
		}
	}
}

func TestHFCGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *HFCBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidHFCBuilder(buildTestClientWithDummyHFC()),
			expectedError: "",
		},
		{
			testBuilder:   buildValidHFCBuilder(buildTestClientWithHFCScheme()),
			expectedError: "hostfirmwarecomponentses.metal3.io \"hostfirmwarecomponents-test\" not found",
		},
	}

	for _, testCase := range testCases {
		hfc, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, hfc.Name)
			assert.Equal(t, testCase.testBuilder.Definition.Namespace, hfc.Namespace)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestHFCExists(t *testing.T) {
	testCases := []struct {
		testBuilder *HFCBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidHFCBuilder(buildTestClientWithDummyHFC()),
			exists:      true,
		},
		{
			testBuilder: buildValidHFCBuilder(buildTestClientWithHFCScheme()),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

// buildDummyHFC returns a HostFirmwareComponents with the provided name and namespace.
func buildDummyHFC(name, namespace string) *bmhv1alpha1.HostFirmwareComponents {
	return &bmhv1alpha1.HostFirmwareComponents{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

// buildTestClientWithDummyHFC returns a client with a mock HostFirmwareComponents.
func buildTestClientWithDummyHFC() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyHFC(defaultHFCName, defaultHFCNamespace),
		},
		SchemeAttachers: testSchemes,
	})
}

// buildTestClientWithHFCScheme returns a client with no objects but the HostFirmwareComponents scheme attached.
func buildTestClientWithHFCScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: testSchemes,
	})
}

// buildValidHFCBuilder returns a valid Builder for testing.
func buildValidHFCBuilder(apiClient *clients.Settings) *HFCBuilder {
	return &HFCBuilder{
		Definition: buildDummyHFC(defaultHFCName, defaultHFCNamespace),
		apiClient:  apiClient.Client,
	}
}
