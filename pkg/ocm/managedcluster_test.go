package ocm

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

var defaultManagedClusterName = "managedcluster-test"

func TestNewManagedClusterBuilder(t *testing.T) {
	testCases := []struct {
		managedClusterName string
		expectedErrorText  string
	}{
		{
			managedClusterName: defaultManagedClusterName,
			expectedErrorText:  "",
		},
		{
			managedClusterName: "",
			expectedErrorText:  "managedCluster 'name' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		managedClusterBuilder := NewManagedClusterBuilder(testSettings, testCase.managedClusterName)

		assert.NotNil(t, managedClusterBuilder)
		assert.Equal(t, testCase.expectedErrorText, managedClusterBuilder.errorMsg)
	}
}

func TestManagedClusterWithOptions(t *testing.T) {
	testCases := []struct {
		valid             bool
		options           ManagedClusterAdditionalOptions
		expectedErrorText string
	}{
		{
			valid: true,
			options: func(builder *ManagedClusterBuilder) (*ManagedClusterBuilder, error) {
				builder.Definition.Spec.HubAcceptsClient = true

				return builder, nil
			},
			expectedErrorText: "",
		},
		{
			valid: false,
			options: func(builder *ManagedClusterBuilder) (*ManagedClusterBuilder, error) {
				return builder, nil
			},
			expectedErrorText: "managedCluster 'name' cannot be empty",
		},
		{
			valid: true,
			options: func(builder *ManagedClusterBuilder) (*ManagedClusterBuilder, error) {
				return builder, fmt.Errorf("error adding additional option")
			},
			expectedErrorText: "error adding additional option",
		},
	}

	for _, testCase := range testCases {
		managedClusterBuilder := buildInvalidManagedClusterTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		if testCase.valid {
			managedClusterBuilder = buildValidManagedClusterTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		}

		managedClusterBuilder = managedClusterBuilder.WithOptions(testCase.options)

		assert.NotNil(t, managedClusterBuilder)
		assert.Equal(t, testCase.expectedErrorText, managedClusterBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.True(t, managedClusterBuilder.Definition.Spec.HubAcceptsClient)
		}
	}
}

func TestPullManagedCluster(t *testing.T) {
	testCases := []struct {
		managedClusterName  string
		addToRuntimeObjects bool
		client              bool
		expectedErrorText   string
	}{
		{
			managedClusterName:  defaultManagedClusterName,
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "",
		},
		{
			managedClusterName:  defaultManagedClusterName,
			addToRuntimeObjects: false,
			client:              true,
			expectedErrorText:   fmt.Sprintf("managedCluster object %s does not exist", defaultManagedClusterName),
		},
		{
			managedClusterName:  "",
			addToRuntimeObjects: false,
			client:              true,
			expectedErrorText:   "managedCluster 'name' cannot be empty",
		},
		{
			managedClusterName:  defaultManagedClusterName,
			addToRuntimeObjects: false,
			client:              false,
			expectedErrorText:   "managedCluster 'apiClient' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testManagedCluster := buildDummyManagedCluster(testCase.managedClusterName)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testManagedCluster)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		managedClusterBuilder, err := PullManagedCluster(testSettings, testManagedCluster.Name)

		if testCase.expectedErrorText != "" {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedErrorText, err.Error())
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testManagedCluster.Name, managedClusterBuilder.Object.Name)
		}
	}
}

func TestManagedClusterExists(t *testing.T) {
	testCases := []struct {
		testBuilder *ManagedClusterBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidManagedClusterTestBuilder(buildTestClientWithDummyManagedCluster()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidManagedClusterTestBuilder(buildTestClientWithDummyManagedCluster()),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestManagedClusterDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *ManagedClusterBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidManagedClusterTestBuilder(buildTestClientWithDummyManagedCluster()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidManagedClusterTestBuilder(buildTestClientWithDummyManagedCluster()),
			expectedError: fmt.Errorf("managedCluster 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

func TestManagedClusterUpdate(t *testing.T) {
	testCases := []struct {
		alreadyExists bool
	}{
		{
			alreadyExists: false,
		},
		{
			alreadyExists: true,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidManagedClusterTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

		if testCase.alreadyExists {
			testBuilder = buildValidManagedClusterTestBuilder(buildTestClientWithDummyManagedCluster())
		}

		assert.NotNil(t, testBuilder.Definition)
		assert.False(t, testBuilder.Definition.Spec.HubAcceptsClient)

		testBuilder.Definition.Spec.HubAcceptsClient = true

		managedClusterBuilder, err := testBuilder.Update()
		assert.NotNil(t, testBuilder.Definition)

		if testCase.alreadyExists {
			assert.Nil(t, err)
			assert.Equal(t, testBuilder.Definition.Name, managedClusterBuilder.Definition.Name)
			assert.Equal(t, testBuilder.Definition.Spec.HubAcceptsClient, managedClusterBuilder.Definition.Spec.HubAcceptsClient)
		} else {
			assert.NotNil(t, err)
		}
	}
}

// buildDummyManagedCluster returns a ManagedCluster with the provided name.
func buildDummyManagedCluster(name string) *clusterv1.ManagedCluster {
	return &clusterv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestClientWithDummyManagedCluster returns a client with a mock dummy ManagedCluster.
func buildTestClientWithDummyManagedCluster() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyManagedCluster(defaultManagedClusterName),
		},
	})
}

// buildValidManagedClusterTestBuilder returns a valid ManagedCluster for testing.
func buildValidManagedClusterTestBuilder(apiClient *clients.Settings) *ManagedClusterBuilder {
	return NewManagedClusterBuilder(apiClient, defaultManagedClusterName)
}

// buildInvalidManagedClusterTestBuilder returns an invalid ManagedCluster for testing.
func buildInvalidManagedClusterTestBuilder(apiClient *clients.Settings) *ManagedClusterBuilder {
	return NewManagedClusterBuilder(apiClient, "")
}
