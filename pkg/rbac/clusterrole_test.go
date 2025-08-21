package rbac

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestNewClusterRoleBuilder(t *testing.T) {
	testCases := []struct {
		name              string
		rule              rbacv1.PolicyRule
		expectedErrorText string
		client            bool
	}{
		{
			name: "test",
			rule: rbacv1.PolicyRule{
				Resources: []string{"pods"}, APIGroups: []string{"v1"}, Verbs: []string{"get"}},
			client:            true,
			expectedErrorText: "",
		},
		{
			name: "test",
			rule: rbacv1.PolicyRule{
				Resources: []string{"pods"}, APIGroups: []string{"v1"}, Verbs: []string{"get"}},
			client:            false,
			expectedErrorText: "clusterRole builder cannot have nil apiClient",
		},
		{
			name: "",
			rule: rbacv1.PolicyRule{
				Resources: []string{"pods"}, APIGroups: []string{"v1"}, Verbs: []string{"get"}},
			client:            true,
			expectedErrorText: "clusterrole 'name' cannot be empty",
		},
		{
			name: "test",
			rule: rbacv1.PolicyRule{
				Resources: []string{"pods"}, APIGroups: []string{"v1"}},
			client:            true,
			expectedErrorText: "clusterrole rule must contain at least one Verb entry",
		},
	}
	for _, testCase := range testCases {
		var (
			client *clients.Settings
		)

		if testCase.client {
			client = clients.GetTestClients(clients.TestClientParams{})
		}

		testClusterRole := NewClusterRoleBuilder(client, testCase.name, testCase.rule)
		if testCase.client {
			assert.NotNil(t, testClusterRole)
		}

		if len(testCase.expectedErrorText) > 0 {
			assert.Equal(t, testCase.expectedErrorText, testClusterRole.errorMsg)
		}
	}
}

func TestPullClusterRole(t *testing.T) {
	generateClusterRole := func(name string) *rbacv1.ClusterRole {
		return &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
	}
	testCases := []struct {
		name                string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
		client              bool
	}{
		{
			name:                "test",
			expectedError:       false,
			expectedErrorText:   "",
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "",
			expectedError:       true,
			addToRuntimeObjects: true,
			expectedErrorText:   "clusterrole 'name' cannot be empty",
			client:              true,
		},
		{
			name:                "test",
			expectedError:       true,
			expectedErrorText:   "clusterrole object test does not exist",
			addToRuntimeObjects: false,
			client:              true,
		},
		{
			name:                "test",
			expectedError:       true,
			expectedErrorText:   "the apiClient cannot be nil",
			addToRuntimeObjects: true,
			client:              false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testClusterRole := generateClusterRole(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testClusterRole)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := PullClusterRole(testSettings, testCase.name)

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testClusterRole.Name, builderResult.Object.Name)
			assert.Equal(t, testClusterRole.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestClusterRoleCreate(t *testing.T) {
	testCases := []struct {
		testClusterRole *ClusterRoleBuilder
		expectedError   error
	}{
		{
			testClusterRole: buildValidClusterRoleBuilder(buildTestClientWithClusterRoleDummyObject()),
			expectedError:   nil,
		},
		{
			testClusterRole: buildInvalidClusterRoleTestBuilder(buildTestClientWithClusterRoleDummyObject()),
			expectedError:   fmt.Errorf("clusterrole rule must contain at least one Verb entry"),
		},
		{
			testClusterRole: buildValidClusterRoleBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:   nil,
		},
	}

	for _, testCase := range testCases {
		clusterRoleBuilder, err := testCase.testClusterRole.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, clusterRoleBuilder.Definition.Name, clusterRoleBuilder.Object.Name)
		}
	}
}

func TestClusterRoleExist(t *testing.T) {
	testCases := []struct {
		testClusterRole *ClusterRoleBuilder
		expectedStatus  bool
	}{
		{
			testClusterRole: buildValidClusterRoleBuilder(buildTestClientWithClusterRoleDummyObject()),
			expectedStatus:  true,
		},
		{
			testClusterRole: buildInvalidClusterRoleTestBuilder(buildTestClientWithClusterRoleDummyObject()),
			expectedStatus:  false,
		},
		{
			testClusterRole: buildValidClusterRoleBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:  false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testClusterRole.Exists()
		assert.Equal(t, testCase.expectedStatus, exists)
	}
}

func TestClusterRoleDelete(t *testing.T) {
	testCases := []struct {
		testClusterRole *ClusterRoleBuilder
		expectedError   error
	}{
		{
			testClusterRole: buildValidClusterRoleBuilder(buildTestClientWithClusterRoleDummyObject()),
			expectedError:   nil,
		},
		{
			testClusterRole: buildInvalidClusterRoleTestBuilder(buildTestClientWithClusterRoleDummyObject()),
			expectedError:   fmt.Errorf("clusterrole rule must contain at least one Verb entry"),
		},
		{
			testClusterRole: buildValidClusterRoleBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:   nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testClusterRole.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testClusterRole.Object)
		}
	}
}

func TestClusterRoleUpdate(t *testing.T) {
	testCases := []struct {
		testClusterRole *ClusterRoleBuilder
		expectedError   error
	}{
		{
			testClusterRole: buildValidClusterRoleBuilder(buildTestClientWithClusterRoleDummyObject()),
			expectedError:   nil,
		},
		{
			testClusterRole: buildValidClusterRoleBuilder(buildTestClientWithClusterRoleDummyObject()),
			expectedError:   nil,
		},
		{
			testClusterRole: buildInvalidClusterRoleTestBuilder(buildTestClientWithClusterRoleDummyObject()),
			expectedError:   fmt.Errorf("clusterrole rule must contain at least one Verb entry"),
		},
		{
			testClusterRole: buildValidClusterRoleBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:   fmt.Errorf("clusterrole object test does not exist, fail to update"),
		},
	}
	for _, testCase := range testCases {
		assert.Empty(t, testCase.testClusterRole.Definition.Labels)
		assert.Nil(t, nil, testCase.testClusterRole.Object)
		testCase.testClusterRole.Definition.Labels = map[string]string{"test": "test"}
		testCase.testClusterRole.Definition.ObjectMeta.ResourceVersion = "999"
		roleBuilder, err := testCase.testClusterRole.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, map[string]string{"test": "test"}, testCase.testClusterRole.Object.Labels)
			assert.Equal(t, roleBuilder.Definition, roleBuilder.Object)
		}
	}
}

func TestClusterRoleWithRules(t *testing.T) {
	testCases := []struct {
		rule              []rbacv1.PolicyRule
		expectedError     bool
		expectedErrorText string
	}{
		{
			rule: []rbacv1.PolicyRule{{Resources: []string{"pods"},
				APIGroups: []string{"v1"}, Verbs: []string{"get"}}},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			rule: []rbacv1.PolicyRule{
				{Resources: []string{"pods"},
					APIGroups: []string{"v1"}, Verbs: []string{"get"}},
				{Resources: []string{"pods"},
					APIGroups: []string{"v1"}, Verbs: []string{"get"}},
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			rule: []rbacv1.PolicyRule{{Resources: []string{"pods"},
				APIGroups: []string{"v1"}}},
			expectedError:     true,
			expectedErrorText: "clusterrole rule must contain at least one Verb entry",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidClusterRoleBuilder(buildTestClientWithClusterRoleDummyObject())

		result := testBuilder.WithRules(testCase.rule)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.rule, result.Definition.Rules[1:])
		}
	}
}

func TestClusterRoleWithOptions(t *testing.T) {
	testSettings := buildTestClientWithDummyObject()
	testBuilder := buildValidRoleBuilder(testSettings).WithOptions(
		func(builder *RoleBuilder) (*RoleBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidRoleBuilder(testSettings).WithOptions(
		func(builder *RoleBuilder) (*RoleBuilder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

// buildInvalidClusterRoleTestBuilder returns a valid Builder for testing purposes.
func buildValidClusterRoleBuilder(apiClient *clients.Settings) *ClusterRoleBuilder {
	return NewClusterRoleBuilder(
		apiClient,
		defaultRoleName,
		rbacv1.PolicyRule{Resources: []string{"pods"}, APIGroups: []string{"v1"}, Verbs: []string{"get"}})
}

// buildValidClusterRoleBuilder returns a valid Builder for testing purposes.
func buildInvalidClusterRoleTestBuilder(apiClient *clients.Settings) *ClusterRoleBuilder {
	return NewClusterRoleBuilder(apiClient, defaultRoleName, rbacv1.PolicyRule{})
}

func buildTestClientWithClusterRoleDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyClusterRoleObject(),
	})
}

func buildDummyClusterRoleObject() []runtime.Object {
	return append([]runtime.Object{}, &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultRoleName,
		},
	})
}
