package nad

import (
	"fmt"
	"testing"

	nadV1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testSchemes = []clients.SchemeAttacher{
		nadV1.AddToScheme,
	}
	defaultNetName   = "nadtest"
	defaultNetNsName = "nadnamespace"
)

//nolint:funlen
func TestNADPull(t *testing.T) {
	generateNetwork := func(name, namespace string) *nadV1.NetworkAttachmentDefinition {
		return &nadV1.NetworkAttachmentDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: nadV1.NetworkAttachmentDefinitionSpec{},
		}
	}

	testCases := []struct {
		networkName         string
		networkNamespace    string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
		client              bool
	}{
		{
			networkName:         "test1",
			networkNamespace:    "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			networkName:         "test2",
			networkNamespace:    "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "networkattachmentdefinition object test2 does not exist in namespace test-namespace",
			client:              true,
		},
		{
			networkName:         "",
			networkNamespace:    "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "networkattachmentdefinition 'name' cannot be empty",
			client:              true,
		},
		{
			networkName:         "test3",
			networkNamespace:    "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "networkattachmentdefinition 'namespace' cannot be empty",
			client:              true,
		},
		{
			networkName:         "test3",
			networkNamespace:    "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "the apiClient cannot be nil",
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testNetwork := generateNetwork(testCase.networkName, testCase.networkNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testNetwork)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		// Test the Pull method
		builderResult, err := Pull(testSettings, testNetwork.Name, testNetwork.Namespace)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			// Check the error message
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testNetwork.Name, builderResult.Object.Name)
			assert.Equal(t, testNetwork.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestNADNewNetworkBuilder(t *testing.T) {
	testCases := []struct {
		networkName       string
		networkNamespace  string
		expectedErrorText string
		client            bool
	}{
		{
			networkName:      "test1",
			networkNamespace: "test-namespace",
			client:           true,
		},
		{
			networkName:       "",
			networkNamespace:  "test-namespace",
			expectedErrorText: "NAD name is empty",
			client:            true,
		},
		{
			networkName:       "test1",
			networkNamespace:  "",
			expectedErrorText: "NAD namespace is empty",
			client:            true,
		},
		{
			networkName:      "test1",
			networkNamespace: "test-namespace",
			client:           false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings
		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testNetworkStructure := NewBuilder(
			testSettings, testCase.networkName, testCase.networkNamespace)
		if testCase.client {
			assert.NotNil(t, testNetworkStructure)
		}

		if len(testCase.expectedErrorText) > 0 {
			assert.Equal(t, testCase.expectedErrorText, testNetworkStructure.errorMsg)
		}
	}
}

func TestNADGet(t *testing.T) {
	testCases := []struct {
		networkBuilder *Builder
		expectedError  error
	}{
		{
			networkBuilder: buildValidNADNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			networkBuilder: buildInvalidNADNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedError:  fmt.Errorf("NAD namespace is empty"),
		},
		{
			networkBuilder: buildValidNADNetworkTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("networkattachmentdefinitions.k8s.cni.cncf.io \"nadtest\" not found"),
		},
	}

	for _, testCase := range testCases {
		network, err := testCase.networkBuilder.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, network.Name, testCase.networkBuilder.Definition.Name)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestNADCreate(t *testing.T) {
	testCases := []struct {
		testNetwork   *Builder
		expectedError error
	}{
		{
			testNetwork:   buildValidNADNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testNetwork:   buildInvalidNADNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("NAD namespace is empty"),
		},
		{
			testNetwork:   buildValidNADNetworkTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		netBuilder, err := testCase.testNetwork.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, netBuilder.Definition, netBuilder.Object)
		}
	}
}

func TestNADDelete(t *testing.T) {
	testCases := []struct {
		testNetwork   *Builder
		expectedError error
	}{
		{
			testNetwork:   buildValidNADNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testNetwork:   buildInvalidNADNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("NAD namespace is empty"),
		},
		{
			testNetwork:   buildValidNADNetworkTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testNetwork.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testNetwork.Object)
		}
	}
}

func TestNADExist(t *testing.T) {
	testCases := []struct {
		testNetwork    *Builder
		expectedStatus bool
	}{
		{
			testNetwork:    buildValidNADNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testNetwork:    buildInvalidNADNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testNetwork:    buildValidNADNetworkTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testNetwork.Exists()
		assert.Equal(t, testCase.expectedStatus, exists)
	}
}

func TestNADUpdate(t *testing.T) {
	testCases := []struct {
		testNetwork   *Builder
		expectedError error
	}{
		{
			testNetwork:   buildValidNADNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testNetwork:   buildInvalidNADNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("NAD namespace is empty"),
		},
		{
			testNetwork:   buildValidNADNetworkTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("failed to update NetworkAttachmentDefinition, object does not exist on cluster"),
		},
	}
	for _, testCase := range testCases {
		assert.Equal(t, "", testCase.testNetwork.Definition.Spec.Config)

		masterPlugin, err := NewMasterMacVlanPlugin("test").GetMasterPluginConfig()
		assert.Nil(t, err)
		assert.Nil(t, nil, testCase.testNetwork.Object)
		testCase.testNetwork.WithMasterPlugin(masterPlugin)
		testCase.testNetwork.Definition.ResourceVersion = "999"
		netBuilder, err := testCase.testNetwork.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, `{"cniVersion":"0.3.1","name":"test","type":"macvlan"}`,
				testCase.testNetwork.Object.Spec.Config)
			assert.Equal(t, netBuilder.Definition, netBuilder.Object)
		}
	}
}

func TestNADWithMasterPlugin(t *testing.T) {
	testCases := []struct {
		testNetwork   *Builder
		name          string
		plugin        *MasterPlugin
		expectedError string
	}{
		{
			testNetwork: buildValidNADNetworkTestBuilder(buildTestClientWithDummyObject()),
			plugin:      &MasterPlugin{CniVersion: "0.3.1", Name: "test", Type: "ipvlan"},
		},
		{
			testNetwork:   buildValidNADNetworkTestBuilder(buildTestClientWithDummyObject()),
			plugin:        nil,
			expectedError: "error 'masterPlugin' is empty",
		},
		{
			testNetwork:   buildInvalidNADNetworkTestBuilder(buildTestClientWithDummyObject()),
			plugin:        &MasterPlugin{CniVersion: "0.3.1", Name: "test", Type: "ipvlan"},
			expectedError: "NAD namespace is empty",
		},
	}

	for _, testCase := range testCases {
		builder := testCase.testNetwork.WithMasterPlugin(testCase.plugin)
		assert.Equal(t, testCase.expectedError, builder.errorMsg)

		if testCase.expectedError == "" {
			assert.Contains(t, builder.Definition.Spec.Config, testCase.plugin.Name)
		}
	}
}

func TestNADWithPlugins(t *testing.T) {
	testCases := []struct {
		testNetwork   *Builder
		name          string
		plugins       []Plugin
		expectedError string
	}{
		{
			testNetwork: buildValidNADNetworkTestBuilder(buildTestClientWithDummyObject()),
			name:        "test",
			plugins:     []Plugin{{Name: "test"}, {Name: "test2"}},
		},
		{
			testNetwork:   buildInvalidNADNetworkTestBuilder(buildTestClientWithDummyObject()),
			name:          "test",
			plugins:       []Plugin{{Name: "test"}, {Name: "test2"}},
			expectedError: "NAD namespace is empty",
		},
	}

	for _, testCase := range testCases {
		builder := testCase.testNetwork.WithPlugins(testCase.name, &testCase.plugins)
		assert.Equal(t, testCase.expectedError, builder.errorMsg)

		if testCase.expectedError == "" {
			for _, plugin := range testCase.plugins {
				assert.Contains(t, builder.Definition.Spec.Config, plugin.Name)
			}
		}
	}
}

func TestNADGetString(t *testing.T) {
	testCases := []struct {
		testNetwork   *Builder
		name          string
		plugins       []Plugin
		expectedError error
	}{
		{
			testNetwork:   buildValidNADNetworkTestBuilder(buildTestClientWithDummyObject()),
			name:          "test",
			expectedError: nil,
		},
		{
			testNetwork:   buildInvalidNADNetworkTestBuilder(buildTestClientWithDummyObject()),
			name:          "test",
			expectedError: fmt.Errorf("NAD namespace is empty"),
		},
	}

	for _, testCase := range testCases {
		testNetworkString, err := testCase.testNetwork.GetString()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotEmpty(t, testNetworkString)
		}
	}
}

// buildValidTestBuilder returns a valid Builder for testing purposes.
func buildValidNADNetworkTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultNetName, defaultNetNsName)
}

// buildValidTestBuilder returns a valid Builder for testing purposes.
func buildInvalidNADNetworkTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultNetName, "")
}

func buildTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyNADNetworkObject(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyNADNetworkObject() []runtime.Object {
	return append([]runtime.Object{}, &nadV1.NetworkAttachmentDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultNetName,
			Namespace: defaultNetNsName,
		},
		Spec: nadV1.NetworkAttachmentDefinitionSpec{},
	})
}
