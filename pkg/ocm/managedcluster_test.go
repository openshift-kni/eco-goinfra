package ocm

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ocm/clusterv1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const defaultManagedClusterName = "managedcluster-test"

var clusterTestSchemes = []clients.SchemeAttacher{
	clusterv1.Install,
}

func TestNewManagedClusterBuilder(t *testing.T) {
	testCases := []struct {
		managedClusterName string
		client             bool
		expectedErrorText  string
	}{
		{
			managedClusterName: defaultManagedClusterName,
			client:             true,
			expectedErrorText:  "",
		},
		{
			managedClusterName: "",
			client:             true,
			expectedErrorText:  "managedCluster 'name' cannot be empty",
		},
		{
			managedClusterName: defaultManagedClusterName,
			client:             false,
			expectedErrorText:  "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithManagedClusterScheme()
		}

		managedClusterBuilder := NewManagedClusterBuilder(testSettings, testCase.managedClusterName)

		if testCase.client {
			assert.Equal(t, testCase.expectedErrorText, managedClusterBuilder.errorMsg)

			if testCase.expectedErrorText == "" {
				assert.Equal(t, testCase.managedClusterName, managedClusterBuilder.Definition.Name)
			}
		} else {
			assert.Nil(t, managedClusterBuilder)
		}
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
		managedClusterBuilder := buildInvalidManagedClusterTestBuilder(buildTestClientWithManagedClusterScheme())
		if testCase.valid {
			managedClusterBuilder = buildValidManagedClusterTestBuilder(buildTestClientWithManagedClusterScheme())
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
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: clusterTestSchemes,
			})
		}

		managedClusterBuilder, err := PullManagedCluster(testSettings, testManagedCluster.Name)

		if testCase.expectedErrorText == "" {
			assert.Nil(t, err)
			assert.Equal(t, testManagedCluster.Name, managedClusterBuilder.Object.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedErrorText)
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
			testBuilder:   buildValidManagedClusterTestBuilder(buildTestClientWithManagedClusterScheme()),
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
		testBuilder := buildValidManagedClusterTestBuilder(buildTestClientWithManagedClusterScheme())

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

func TestManagedClusterValidate(t *testing.T) {
	testCases := []struct {
		builderNil      bool
		definitionNil   bool
		apiClientNil    bool
		builderErrorMsg string
		expectedError   error
	}{
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   nil,
		},
		{
			builderNil:      true,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("error: received nil managedCluster builder"),
		},
		{
			builderNil:      false,
			definitionNil:   true,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("can not redefine the undefined managedCluster"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    true,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("managedCluster builder cannot have nil apiClient"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "test error",
			expectedError:   fmt.Errorf("test error"),
		},
	}

	for _, testCase := range testCases {
		managedClusterBuilder := buildValidManagedClusterTestBuilder(buildTestClientWithManagedClusterScheme())

		if testCase.builderNil {
			managedClusterBuilder = nil
		}

		if testCase.definitionNil {
			managedClusterBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			managedClusterBuilder.apiClient = nil
		}

		if testCase.builderErrorMsg != "" {
			managedClusterBuilder.errorMsg = testCase.builderErrorMsg
		}

		valid, err := managedClusterBuilder.validate()

		if testCase.expectedError != nil {
			assert.False(t, valid)
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.True(t, valid)
			assert.Nil(t, err)
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
		SchemeAttachers: clusterTestSchemes,
	})
}

// buildTestClientWithManagedClusterScheme returns a client with no objects but the ManagedCluster scheme.
func buildTestClientWithManagedClusterScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: clusterTestSchemes,
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
