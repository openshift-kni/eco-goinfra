package dns

import (
	"fmt"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testSchemes = []clients.SchemeAttacher{
		configv1.Install,
	}
)

func TestPullDNS(t *testing.T) {
	testCases := []struct {
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("dns object %s does not exist", clusterDNSName),
		},
		{
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("dns 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testDNS := buildDummyDNS()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testDNS)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := Pull(testSettings)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, clusterDNSName, testBuilder.Definition.Name)
		}
	}
}

func TestDNSGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *Builder
		expectedError string
	}{
		{
			testBuilder:   newDNSBuilder(buildTestClientWithDummyDNS()),
			expectedError: "",
		},
		{
			testBuilder:   newDNSBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "dnses.config.openshift.io \"cluster\" not found",
		},
	}

	for _, testCase := range testCases {
		dnsObject, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, dnsObject.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestDNSExists(t *testing.T) {
	testCases := []struct {
		testBuilder *Builder
		exists      bool
	}{
		{
			testBuilder: newDNSBuilder(buildTestClientWithDummyDNS()),
			exists:      true,
		},
		{
			testBuilder: newDNSBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestDNSUpdate(t *testing.T) {
	testCases := []struct {
		testBuilder   *Builder
		expectedError error
	}{
		{
			testBuilder:   newDNSBuilder(buildTestClientWithDummyDNS()),
			expectedError: nil,
		},
		{
			testBuilder:   newDNSBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("dns object %s does not exist", clusterDNSName),
		},
	}

	for _, testCase := range testCases {
		assert.Empty(t, testCase.testBuilder.Definition.Spec.Platform.Type)

		testCase.testBuilder.Definition.Spec.Platform.Type = configv1.BareMetalPlatformType

		testBuilder, err := testCase.testBuilder.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, configv1.BareMetalPlatformType, testBuilder.Object.Spec.Platform.Type)
		}
	}
}

// buildDummyDNS returns a DNS with the clusterDNSName.
func buildDummyDNS() *configv1.DNS {
	return &configv1.DNS{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterDNSName,
		},
	}
}

func buildTestClientWithDummyDNS() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{buildDummyDNS()},
		SchemeAttachers: testSchemes,
	})
}

// newDNSBuilder returns a Builder with the provided apiClient and the clusterDNSName.
func newDNSBuilder(apiClient *clients.Settings) *Builder {
	return &Builder{
		apiClient:  apiClient.Client,
		Definition: buildDummyDNS(),
	}
}
