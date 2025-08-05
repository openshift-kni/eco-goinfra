package sriovfec

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	sriovfectypes "github.com/openshift-kni/eco-goinfra/pkg/schemes/fec/fectypes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultClusterConfigName      = "test-cluster-config"
	defaultClusterConfigNamespace = "test-ns"
)

func TestFecNewClusterConfigBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		nsname        string
		client        bool
		expectedError string
	}{
		{
			name:          defaultClusterConfigName,
			nsname:        defaultClusterConfigNamespace,
			client:        true,
			expectedError: "",
		},
		{
			name:          "",
			nsname:        defaultClusterConfigNamespace,
			client:        true,
			expectedError: "SriovFecClusterConfig 'name' cannot be empty",
		},
		{
			name:          defaultClusterConfigName,
			nsname:        "",
			client:        true,
			expectedError: "SriovFecClusterConfig 'nsname' cannot be empty",
		},
		{
			name:          defaultClusterConfigName,
			nsname:        defaultClusterConfigNamespace,
			client:        false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewClusterConfigBuilder(testSettings, testCase.name, testCase.nsname)

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

func TestFecPullClusterConfig(t *testing.T) {
	testCases := []struct {
		name                string
		nsname              string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultClusterConfigName,
			nsname:              defaultClusterConfigNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			nsname:              defaultClusterConfigNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("SriovFecClusterConfig 'name' cannot be empty"),
		},
		{
			name:                defaultClusterConfigName,
			nsname:              "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("SriovFecClusterConfig 'nsname' cannot be empty"),
		},
		{
			name:                defaultClusterConfigName,
			nsname:              defaultClusterConfigNamespace,
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf(
				"SriovFecClusterConfig object %s does not exist in namespace %s",
				defaultClusterConfigName, defaultClusterConfigNamespace),
		},
		{
			name:                defaultClusterConfigName,
			nsname:              defaultClusterConfigNamespace,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("SriovFecClusterConfig 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testClusterConfig := buildDummyClusterConfig(testCase.name, testCase.nsname)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testClusterConfig)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := PullClusterConfig(testSettings, testCase.name, testCase.nsname)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.nsname, testBuilder.Definition.Namespace)
		}
	}
}

func TestFecClusterConfigGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *ClusterConfigBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			expectedError: "SriovFecClusterConfig 'nsname' cannot be empty",
		},
		{
			testBuilder:   buildValidClusterConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "sriovfecclusterconfigs.sriovfec.intel.com \"test-cluster-config\" not found",
		},
	}

	for _, testCase := range testCases {
		clusterConfig, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, clusterConfig.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestFecClusterConfigExists(t *testing.T) {
	testCases := []struct {
		testBuilder *ClusterConfigBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			exists:      false,
		},
		{
			testBuilder: buildValidClusterConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestFecClusterConfigCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *ClusterConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			expectedError: fmt.Errorf("SriovFecClusterConfig 'nsname' cannot be empty"),
		},
		{
			testBuilder:   buildValidClusterConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
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

func TestFecClusterConfigDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *ClusterConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			expectedError: fmt.Errorf("SriovFecClusterConfig 'nsname' cannot be empty"),
		},
		{
			testBuilder:   buildValidClusterConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
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

func TestFecClusterConfigUpdate(t *testing.T) {
	testCases := []struct {
		testBuilder   *ClusterConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			expectedError: fmt.Errorf("SriovFecClusterConfig 'nsname' cannot be empty"),
		},
		{
			testBuilder:   buildValidClusterConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("cannot update non-existent SriovFecClusterConfig"),
		},
	}

	for _, testCase := range testCases {
		testCase.testBuilder.Definition.ResourceVersion = "999"

		testBuilder, err := testCase.testBuilder.Update(false)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, testBuilder.Object)
		}
	}
}

func TestFecClusterConfigWithOptions(t *testing.T) {
	testCases := []struct {
		testBuilder   *ClusterConfigBuilder
		options       ClusterAdditionalOptions
		expectedError string
	}{
		{
			testBuilder: buildValidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			options: func(builder *ClusterConfigBuilder) (*ClusterConfigBuilder, error) {
				builder.Definition.Spec.Priority = 10

				return builder, nil
			},
			expectedError: "",
		},
		{
			testBuilder: buildInvalidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			options: func(builder *ClusterConfigBuilder) (*ClusterConfigBuilder, error) {
				return builder, nil
			},
			expectedError: "SriovFecClusterConfig 'nsname' cannot be empty",
		},
		{
			testBuilder: buildValidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			options: func(builder *ClusterConfigBuilder) (*ClusterConfigBuilder, error) {
				return builder, fmt.Errorf("error adding additional option")
			},
			expectedError: "error adding additional option",
		},
	}

	for _, testCase := range testCases {
		testBuilder := testCase.testBuilder.WithOptions(testCase.options)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, 10, testBuilder.Definition.Spec.Priority)
		}
	}
}

func TestFecClusterConfigGetGVR(t *testing.T) {
	gvr := GetSriovFecClusterConfigIoGVR()
	assert.Equal(t, APIGroup, gvr.Group)
	assert.Equal(t, APIVersion, gvr.Version)
	assert.Equal(t, ClusterConfigsResource, gvr.Resource)
}

// buildDummyClusterConfig returns a ClusterConfig with the provided name and namespace.
func buildDummyClusterConfig(name, nsname string) *sriovfectypes.SriovFecClusterConfig {
	return &sriovfectypes.SriovFecClusterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
		Spec: sriovfectypes.SriovFecClusterConfigSpec{
			Priority: 1,
		},
	}
}

// buildTestClientWithDummyClusterConfig returns a client with a dummy ClusterConfig.
// It uses the default name and namespace.
func buildTestClientWithDummyClusterConfig() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyClusterConfig(defaultClusterConfigName, defaultClusterConfigNamespace),
		},
		SchemeAttachers: testSchemes,
	})
}

// buildValidClusterConfigBuilder returns a valid ClusterConfigBuilder for testing.
func buildValidClusterConfigBuilder(apiClient *clients.Settings) *ClusterConfigBuilder {
	return NewClusterConfigBuilder(apiClient, defaultClusterConfigName, defaultClusterConfigNamespace)
}

// buildInvalidClusterConfigBuilder returns an invalid ClusterConfigBuilder for testing, missing the nsname.
func buildInvalidClusterConfigBuilder(apiClient *clients.Settings) *ClusterConfigBuilder {
	return NewClusterConfigBuilder(apiClient, defaultClusterConfigName, "")
}
