package network

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var configTestSchemes = []clients.SchemeAttacher{
	configv1.Install,
}

func TestPullConfig(t *testing.T) {
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
			expectedError:       fmt.Errorf("network.config object %s does not exist", clusterNetworkName),
		},
		{
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("network.config 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, buildDummyNetwork())
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: configTestSchemes,
			})
		}

		testbBuilder, err := PullConfig(testSettings)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, clusterNetworkName, testbBuilder.Definition.Name)
		}
	}
}

func TestConfigGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *ConfigBuilder
		expectedError string
	}{
		{
			testBuilder:   newConfigBuilder(buildTestClientWithDummyNetwork()),
			expectedError: "",
		},
		{
			testBuilder:   newConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "networks.config.openshift.io \"cluster\" not found",
		},
	}

	for _, testCase := range testCases {
		network, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, clusterNetworkName, network.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestConfigExists(t *testing.T) {
	testCases := []struct {
		testBuilder *ConfigBuilder
		exists      bool
	}{
		{
			testBuilder: newConfigBuilder(buildTestClientWithDummyNetwork()),
			exists:      true,
		},
		{
			testBuilder: newConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestConfigValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		expectedError error
	}{
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: nil,
		},
		{
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: fmt.Errorf("error: received nil network.config builder"),
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: fmt.Errorf("can not redefine the undefined network.config"),
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: fmt.Errorf("network.config builder cannot have nil apiClient"),
		},
	}

	for _, testCase := range testCases {
		networkBuilder := newConfigBuilder(clients.GetTestClients(clients.TestClientParams{}))

		if testCase.builderNil {
			networkBuilder = nil
		}

		if testCase.definitionNil {
			networkBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			networkBuilder.apiClient = nil
		}

		valid, err := networkBuilder.validate()

		if testCase.expectedError != nil {
			assert.False(t, valid)
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.True(t, valid)
			assert.Nil(t, err)
		}
	}
}

// buildDummyNetwork builds a dummy network object. It uses the clusterNetworkName.
func buildDummyNetwork() *configv1.Network {
	return &configv1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterNetworkName,
		},
	}
}

func buildTestClientWithDummyNetwork() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyNetwork(),
		},
		SchemeAttachers: configTestSchemes,
	})
}

func newConfigBuilder(apiClient *clients.Settings) *ConfigBuilder {
	return &ConfigBuilder{
		apiClient:  apiClient.Client,
		Definition: buildDummyNetwork(),
	}
}
