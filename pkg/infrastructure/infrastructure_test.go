package infrastructure

import (
	"fmt"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var testSchemes = []clients.SchemeAttacher{
	configv1.Install,
}

func TestPullInfrastructure(t *testing.T) {
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
			expectedError:       fmt.Errorf("infrastructure object %s does not exist", infrastructureName),
		},
		{
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("infrastructure 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, buildDummyInfrastructure())
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
			assert.Equal(t, infrastructureName, testBuilder.Definition.Name)
		}
	}
}

func TestInfrastructureGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *Builder
		expectedError string
	}{
		{
			testBuilder:   newInfrastructureBuilder(buildTestClientWithDummyInfrastructure()),
			expectedError: "",
		},
		{
			testBuilder:   newInfrastructureBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Sprintf("infrastructures.config.openshift.io \"%s\" not found", infrastructureName),
		},
	}

	for _, testCase := range testCases {
		infrastructure, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, infrastructureName, infrastructure.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestInfrastructureExists(t *testing.T) {
	testCases := []struct {
		testBuilder *Builder
		exists      bool
	}{
		{
			testBuilder: newInfrastructureBuilder(buildTestClientWithDummyInfrastructure()),
			exists:      true,
		},
		{
			testBuilder: newInfrastructureBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

// buildDummyInfrastructure returns a new Infrastructure with the infrastructureName.
func buildDummyInfrastructure() *configv1.Infrastructure {
	return &configv1.Infrastructure{
		ObjectMeta: metav1.ObjectMeta{
			Name: infrastructureName,
		},
	}
}

// buildTestClientWithDummyInfrastructure returns a new client with a mock Infrastructure object.
func buildTestClientWithDummyInfrastructure() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{buildDummyInfrastructure()},
		SchemeAttachers: testSchemes,
	})
}

// newInfrastructureBuilder returns a Builder for an Infrastructure with the provided client.
func newInfrastructureBuilder(apiClient *clients.Settings) *Builder {
	return &Builder{
		apiClient:  apiClient.Client,
		Definition: buildDummyInfrastructure(),
	}
}
