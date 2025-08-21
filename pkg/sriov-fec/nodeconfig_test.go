package sriovfec

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	sriovfectypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/fec/fectypes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultNodeConfigName      = "test-node-config"
	defaultNodeConfigNamespace = "test-ns"
)

var testSchemes = []clients.SchemeAttacher{
	sriovfectypes.AddToScheme,
}

func TestFecNewNodeConfigBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		nsname        string
		client        bool
		expectedError string
	}{
		{
			name:          defaultNodeConfigName,
			nsname:        defaultNodeConfigNamespace,
			client:        true,
			expectedError: "",
		},
		{
			name:          "",
			nsname:        defaultNodeConfigNamespace,
			client:        true,
			expectedError: "sriovFecNodeConfig 'name' cannot be empty",
		},
		{
			name:          defaultNodeConfigName,
			nsname:        "",
			client:        true,
			expectedError: "sriovFecNodeConfig 'nsname' cannot be empty",
		},
		{
			name:          defaultNodeConfigName,
			nsname:        defaultNodeConfigNamespace,
			client:        false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewNodeConfigBuilder(testSettings, testCase.name, testCase.nsname, nil)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
				assert.Equal(t, testCase.nsname, testBuilder.Definition.Namespace)
			}
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestFecPullNodeConfig(t *testing.T) {
	testCases := []struct {
		name                string
		nsname              string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultNodeConfigName,
			nsname:              defaultNodeConfigNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			nsname:              defaultNodeConfigNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("sriovFecNodeConfig 'name' cannot be empty"),
		},
		{
			name:                defaultNodeConfigName,
			nsname:              "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("sriovFecNodeConfig 'nsname' cannot be empty"),
		},
		{
			name:                defaultNodeConfigName,
			nsname:              defaultNodeConfigNamespace,
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf(
				"sriovFecNodeConfig object %s does not exist in namespace %s", defaultNodeConfigName, defaultNodeConfigNamespace),
		},
		{
			name:                defaultNodeConfigName,
			nsname:              defaultNodeConfigNamespace,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("sriovFecNodeConfig 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testNodeConfig := buildDummyNodeConfig(testCase.name, testCase.nsname)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testNodeConfig)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := Pull(testSettings, testCase.name, testCase.nsname)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.nsname, testBuilder.Definition.Namespace)
		}
	}
}

func TestFecNodeConfigGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *NodeConfigBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidNodeConfigBuilder(buildTestClientWithDummyNodeConfig()),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidNodeConfigBuilder(buildTestClientWithDummyNodeConfig()),
			expectedError: "sriovFecNodeConfig 'nsname' cannot be empty",
		},
		{
			testBuilder:   buildValidNodeConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "sriovfecnodeconfigs.sriovfec.intel.com \"test-node-config\" not found",
		},
	}

	for _, testCase := range testCases {
		nodeConfig, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, nodeConfig.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestFecNodeConfigExists(t *testing.T) {
	testCases := []struct {
		testBuilder *NodeConfigBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidNodeConfigBuilder(buildTestClientWithDummyNodeConfig()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidNodeConfigBuilder(buildTestClientWithDummyNodeConfig()),
			exists:      false,
		},
		{
			testBuilder: buildValidNodeConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestFecNodeConfigCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *NodeConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidNodeConfigBuilder(buildTestClientWithDummyNodeConfig()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidNodeConfigBuilder(buildTestClientWithDummyNodeConfig()),
			expectedError: fmt.Errorf("sriovFecNodeConfig 'nsname' cannot be empty"),
		},
		{
			testBuilder:   buildValidNodeConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testBuilder.Definition.Name, testBuilder.Object.Name)
		}
	}
}

func TestFecNodeConfigDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *NodeConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidNodeConfigBuilder(buildTestClientWithDummyNodeConfig()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidNodeConfigBuilder(buildTestClientWithDummyNodeConfig()),
			expectedError: fmt.Errorf("sriovFecNodeConfig 'nsname' cannot be empty"),
		},
		{
			testBuilder:   buildValidNodeConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testBuilder.Object)
		}
	}
}

func TestFecNodeConfigUpdate(t *testing.T) {
	testCases := []struct {
		testBuilder   *NodeConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidNodeConfigBuilder(buildTestClientWithDummyNodeConfig()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidNodeConfigBuilder(buildTestClientWithDummyNodeConfig()),
			expectedError: fmt.Errorf("sriovFecNodeConfig 'nsname' cannot be empty"),
		},
		{
			testBuilder:   buildValidNodeConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("cannot update non-existent SriovFecNodeConfig"),
		},
	}

	for _, testCase := range testCases {
		assert.False(t, testCase.testBuilder.Definition.Spec.DrainSkip)

		testCase.testBuilder.Definition.Spec.DrainSkip = true
		testCase.testBuilder.Definition.ResourceVersion = "999"

		testBuilder, err := testCase.testBuilder.Update(false)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.True(t, testBuilder.Object.Spec.DrainSkip)
		}
	}
}

func TestFecNodeConfigWithOptions(t *testing.T) {
	testCases := []struct {
		testBuilder   *NodeConfigBuilder
		options       AdditionalOptions
		expectedError string
	}{
		{
			testBuilder: buildValidNodeConfigBuilder(buildTestClientWithDummyNodeConfig()),
			options: func(builder *NodeConfigBuilder) (*NodeConfigBuilder, error) {
				builder.Definition.Spec.DrainSkip = true

				return builder, nil
			},
			expectedError: "",
		},
		{
			testBuilder: buildInvalidNodeConfigBuilder(buildTestClientWithDummyNodeConfig()),
			options: func(builder *NodeConfigBuilder) (*NodeConfigBuilder, error) {
				return builder, nil
			},
			expectedError: "sriovFecNodeConfig 'nsname' cannot be empty",
		},
		{
			testBuilder: buildValidNodeConfigBuilder(buildTestClientWithDummyNodeConfig()),
			options: func(builder *NodeConfigBuilder) (*NodeConfigBuilder, error) {
				return builder, fmt.Errorf("error adding additional option")
			},
			expectedError: "error adding additional option",
		},
	}

	for _, testCase := range testCases {
		testBuilder := testCase.testBuilder.WithOptions(testCase.options)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.True(t, testBuilder.Definition.Spec.DrainSkip)
		}
	}
}

func TestFecNodeConfigGetGVR(t *testing.T) {
	gvr := GetSriovFecNodeConfigIoGVR()
	assert.Equal(t, APIGroup, gvr.Group)
	assert.Equal(t, APIVersion, gvr.Version)
	assert.Equal(t, NodeConfigsResource, gvr.Resource)
}

// buildDummyNodeConfig returns a NodeConfig with the provided name and namespace.
func buildDummyNodeConfig(name, nsname string) *sriovfectypes.SriovFecNodeConfig {
	return &sriovfectypes.SriovFecNodeConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}

// buildTestClientWithDummyNodeConfig returns a client with a dummy NodeConfig. It uses the default name and namespace.
func buildTestClientWithDummyNodeConfig() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyNodeConfig(defaultNodeConfigName, defaultNodeConfigNamespace),
		},
		SchemeAttachers: testSchemes,
	})
}

// buildValidNodeConfigBuilder returns a valid NodeConfigBuilder for testing.
func buildValidNodeConfigBuilder(apiClient *clients.Settings) *NodeConfigBuilder {
	return NewNodeConfigBuilder(apiClient, defaultNodeConfigName, defaultNodeConfigNamespace, nil)
}

// buildInvalidNodeConfigBuilder returns an invalid NodeConfigBuilder for testing, missing the nsname.
func buildInvalidNodeConfigBuilder(apiClient *clients.Settings) *NodeConfigBuilder {
	return NewNodeConfigBuilder(apiClient, defaultNodeConfigName, "", nil)
}
