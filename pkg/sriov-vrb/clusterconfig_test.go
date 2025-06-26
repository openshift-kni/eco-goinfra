package sriovvrb

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	sriovvrbtypes "github.com/openshift-kni/eco-goinfra/pkg/schemes/fec/vrbtypes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultClusterConfigName      = "test-cluster-config"
	defaultClusterConfigNamespace = "test-ns"
)

func TestVrbNewClusterConfigBuilder(t *testing.T) {
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
			expectedError: "SriovVrbClusterConfig 'name' cannot be empty",
		},
		{
			name:          defaultClusterConfigName,
			nsname:        "",
			client:        true,
			expectedError: "SriovVrbClusterConfig 'nsname' cannot be empty",
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

func TestVrbPullClusterConfig(t *testing.T) {
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
			expectedError:       fmt.Errorf("SriovVrbClusterConfig 'name' cannot be empty"),
		},
		{
			name:                defaultClusterConfigName,
			nsname:              "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("SriovVrbClusterConfig 'nsname' cannot be empty"),
		},
		{
			name:                defaultClusterConfigName,
			nsname:              defaultClusterConfigNamespace,
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf(
				"SriovVrbClusterConfig object %s does not exist in namespace %s",
				defaultClusterConfigName, defaultClusterConfigNamespace),
		},
		{
			name:                defaultClusterConfigName,
			nsname:              defaultClusterConfigNamespace,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("SriovVrbClusterConfig 'apiClient' cannot be nil"),
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

func TestVrbClusterConfigGet(t *testing.T) {
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
			expectedError: "SriovVrbClusterConfig 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		clusterConfig, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.NotNil(t, clusterConfig)
			assert.Nil(t, err)
			assert.Equal(t, defaultClusterConfigName, clusterConfig.Name)
			assert.Equal(t, defaultClusterConfigNamespace, clusterConfig.Namespace)
		} else {
			assert.Nil(t, clusterConfig)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), testCase.expectedError)
		}
	}
}

func TestVrbClusterConfigExists(t *testing.T) {
	testCases := []struct {
		testBuilder   *ClusterConfigBuilder
		expectedExist bool
	}{
		{
			testBuilder:   buildValidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			expectedExist: true,
		},
		{
			testBuilder:   buildValidClusterConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedExist: false,
		},
		{
			testBuilder:   buildInvalidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			expectedExist: false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.expectedExist, exists)
	}
}

func TestVrbClusterConfigCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *ClusterConfigBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidClusterConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidClusterConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "SriovVrbClusterConfig 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Create()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.NotNil(t, testBuilder)
			assert.NotNil(t, testBuilder.Object)
		} else {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), testCase.expectedError)
		}
	}
}

func TestVrbClusterConfigDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *ClusterConfigBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			expectedError: "",
		},
		{
			testBuilder:   buildValidClusterConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			expectedError: "SriovVrbClusterConfig 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Delete()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.NotNil(t, testBuilder)
			assert.Nil(t, testBuilder.Object)
		} else {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), testCase.expectedError)
		}
	}
}

func TestVrbClusterConfigUpdate(t *testing.T) {
	testCases := []struct {
		testBuilder   *ClusterConfigBuilder
		force         bool
		expectedError string
	}{
		{
			testBuilder:   buildValidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			force:         false,
			expectedError: "",
		},
		{
			testBuilder:   buildValidClusterConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			force:         false,
			expectedError: "cannot update non-existent SriovVrbClusterConfig",
		},
		{
			testBuilder:   buildInvalidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			force:         false,
			expectedError: "SriovVrbClusterConfig 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Update(testCase.force)

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.NotNil(t, testBuilder)
		} else {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), testCase.expectedError)
		}
	}
}

func TestVrbClusterConfigWithOptions(t *testing.T) {
	testCases := []struct {
		testBuilder   *ClusterConfigBuilder
		options       ClusterAdditionalOptions
		expectedError string
	}{
		{
			testBuilder: buildValidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			options: func(builder *ClusterConfigBuilder) (*ClusterConfigBuilder, error) {
				// Add some test modification
				builder.Definition.Spec = sriovvrbtypes.SriovVrbClusterConfigSpec{}

				return builder, nil
			},
			expectedError: "",
		},
		{
			testBuilder: buildInvalidClusterConfigBuilder(buildTestClientWithDummyClusterConfig()),
			options: func(builder *ClusterConfigBuilder) (*ClusterConfigBuilder, error) {
				return builder, nil
			},
			expectedError: "SriovVrbClusterConfig 'nsname' cannot be empty",
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
			// Verify the option was applied
			assert.NotNil(t, testBuilder.Definition.Spec)
		}
	}
}

func TestVrbClusterConfigGetGVR(t *testing.T) {
	gvr := GetSriovVrbClusterConfigIoGVR()
	assert.Equal(t, APIGroup, gvr.Group)
	assert.Equal(t, APIVersion, gvr.Version)
	assert.Equal(t, ClusterConfigsResource, gvr.Resource)
}

// buildDummyClusterConfig returns a ClusterConfig with the provided name and namespace.
func buildDummyClusterConfig(name, nsname string) *sriovvrbtypes.SriovVrbClusterConfig {
	return &sriovvrbtypes.SriovVrbClusterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
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
