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

var (
	defaultRoleName   = "test"
	defaultRoleNsName = "testns"
)

func TestNewRoleBuilder(t *testing.T) {
	testCases := []struct {
		name              string
		nsName            string
		rule              rbacv1.PolicyRule
		expectedErrorText string
		client            bool
	}{
		{
			name:   "test",
			nsName: "testns",
			rule: rbacv1.PolicyRule{
				Resources: []string{"pods"}, APIGroups: []string{"v1"}, Verbs: []string{"get"}},
			client:            true,
			expectedErrorText: "",
		},
		{
			name:   "test",
			nsName: "testns",
			rule: rbacv1.PolicyRule{
				Resources: []string{"pods"}, APIGroups: []string{"v1"}, Verbs: []string{"get"}},
			client:            false,
			expectedErrorText: "role builder cannot have nil apiClient",
		},
		{
			name:   "",
			nsName: "testns",
			rule: rbacv1.PolicyRule{
				Resources: []string{"pods"}, APIGroups: []string{"v1"}, Verbs: []string{"get"}},
			client:            true,
			expectedErrorText: "role 'name' cannot be empty",
		},
		{
			name:   "test",
			nsName: "",
			rule: rbacv1.PolicyRule{
				Resources: []string{"pods"}, APIGroups: []string{"v1"}, Verbs: []string{"get"}},
			client:            true,
			expectedErrorText: "role 'nsname' cannot be empty",
		},
		{
			name:   "test",
			nsName: "testns",
			rule: rbacv1.PolicyRule{
				Resources: []string{"pods"}, APIGroups: []string{"v1"}},
			client:            true,
			expectedErrorText: "role must contain at least one Verb",
		},
	}
	for _, testCase := range testCases {
		var (
			client *clients.Settings
		)

		if testCase.client {
			client = clients.GetTestClients(clients.TestClientParams{})
		}

		testPolicy := NewRoleBuilder(client, testCase.name, testCase.nsName, testCase.rule)
		if testCase.client {
			assert.NotNil(t, testPolicy)

			if len(testCase.expectedErrorText) > 0 {
				assert.Equal(t, testCase.expectedErrorText, testPolicy.errorMsg)
			}
		} else {
			assert.Nil(t, testPolicy)
		}
	}
}

func TestPullRole(t *testing.T) {
	generateRole := func(name, nsName string) *rbacv1.Role {
		return &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsName,
			},
		}
	}
	testCases := []struct {
		name                string
		nsName              string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
		client              bool
	}{
		{
			name:                "test",
			nsName:              "test",
			expectedError:       false,
			expectedErrorText:   "",
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "",
			nsName:              "test",
			expectedError:       true,
			addToRuntimeObjects: true,
			expectedErrorText:   "role 'name' cannot be empty",
			client:              true,
		},
		{
			name:                "test",
			nsName:              "",
			expectedError:       true,
			expectedErrorText:   "role 'namespace' cannot be empty",
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "test",
			nsName:              "test",
			expectedError:       true,
			expectedErrorText:   "role object test does not exist in namespace test",
			addToRuntimeObjects: false,
			client:              true,
		},
		{
			name:                "test",
			nsName:              "test",
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

		testRole := generateRole(testCase.name, testCase.nsName)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testRole)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := PullRole(testSettings, testCase.name, testCase.nsName)

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testRole.Name, builderResult.Object.Name)
			assert.Equal(t, testRole.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestRoleCreate(t *testing.T) {
	testCases := []struct {
		testRole      *RoleBuilder
		expectedError error
	}{
		{
			testRole:      buildValidRoleBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testRole:      buildInvalidRoleTestBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("role must contain at least one Verb"),
		},
		{
			testRole:      buildValidRoleBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		roleBuilder, err := testCase.testRole.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, roleBuilder.Definition.Name, roleBuilder.Object.Name)
		}
	}
}

func TestRoleExist(t *testing.T) {
	testCases := []struct {
		testRole       *RoleBuilder
		expectedStatus bool
	}{
		{
			testRole:       buildValidRoleBuilder(buildTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testRole:       buildInvalidRoleTestBuilder(buildTestClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testRole:       buildValidRoleBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testRole.Exists()
		assert.Equal(t, testCase.expectedStatus, exists)
	}
}

func TestRoleDelete(t *testing.T) {
	testCases := []struct {
		testRole      *RoleBuilder
		expectedError error
	}{
		{
			testRole:      buildValidRoleBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testRole:      buildInvalidRoleTestBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("role must contain at least one Verb"),
		},
		{
			testRole:      buildValidRoleBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testRole.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testRole.Object)
		}
	}
}

func TestRoleUpdate(t *testing.T) {
	testCases := []struct {
		testRole      *RoleBuilder
		expectedError error
	}{
		{
			testRole:      buildValidRoleBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testRole:      buildValidRoleBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testRole:      buildInvalidRoleTestBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("role must contain at least one Verb"),
		},
		{
			testRole:      buildValidRoleBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("role object test does not exist, fail to update"),
		},
	}
	for _, testCase := range testCases {
		assert.Empty(t, testCase.testRole.Definition.Labels)
		assert.Nil(t, nil, testCase.testRole.Object)
		testCase.testRole.Definition.Labels = map[string]string{"test": "test"}
		testCase.testRole.Definition.ResourceVersion = "999"
		roleBuilder, err := testCase.testRole.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, map[string]string{"test": "test"}, testCase.testRole.Object.Labels)
			assert.Equal(t, roleBuilder.Definition, roleBuilder.Object)
		}
	}
}

func TestRoleWithRules(t *testing.T) {
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
				{Resources: []string{"pods"}, APIGroups: []string{"v1"}, Verbs: []string{"get"}},
				{Resources: []string{"pods"}, APIGroups: []string{"v1"}, Verbs: []string{"get"}},
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			rule: []rbacv1.PolicyRule{{Resources: []string{"pods"},
				APIGroups: []string{"v1"}}},
			expectedError:     true,
			expectedErrorText: "role must contain at least one Verb",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidRoleBuilder(buildTestClientWithDummyObject())

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

func TestWithOptions(t *testing.T) {
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

// buildValidTestBuilder returns a valid Builder for testing purposes.
func buildValidRoleBuilder(apiClient *clients.Settings) *RoleBuilder {
	return NewRoleBuilder(
		apiClient,
		defaultRoleName,
		defaultRoleNsName,
		rbacv1.PolicyRule{Resources: []string{"pods"}, APIGroups: []string{"v1"}, Verbs: []string{"get"}})
}

// buildValidTestBuilder returns a valid Builder for testing purposes.
func buildInvalidRoleTestBuilder(apiClient *clients.Settings) *RoleBuilder {
	return NewRoleBuilder(apiClient, defaultRoleName, defaultRoleNsName, rbacv1.PolicyRule{})
}

func buildTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyRoleObject(),
	})
}

func buildDummyRoleObject() []runtime.Object {
	return append([]runtime.Object{}, &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultRoleName,
			Namespace: defaultRoleNsName,
		},
	})
}
