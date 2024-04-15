package ocm

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	placementrulev1 "open-cluster-management.io/multicloud-operators-subscription/pkg/apis/apps/placementrule/v1"
)

var (
	defaultPlacementRuleName   = "placementrule-test"
	defaultPlacementRuleNsName = "test-ns"
)

func TestNewPlacementRuleBuilder(t *testing.T) {
	testCases := []struct {
		placementRuleName      string
		placementRuleNamespace string
		expectedErrorText      string
	}{
		{
			placementRuleName:      defaultPlacementRuleName,
			placementRuleNamespace: defaultPlacementRuleNsName,
			expectedErrorText:      "",
		},
		{
			placementRuleName:      "",
			placementRuleNamespace: defaultPlacementRuleNsName,
			expectedErrorText:      "placementrule's 'name' cannot be empty",
		},
		{
			placementRuleName:      defaultPlacementRuleName,
			placementRuleNamespace: "",
			expectedErrorText:      "placementrule's 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		placementRuleBuilder := NewPlacementRuleBuilder(
			testSettings,
			testCase.placementRuleName,
			testCase.placementRuleNamespace)
		assert.NotNil(t, placementRuleBuilder)
		assert.Equal(t, testCase.expectedErrorText, placementRuleBuilder.errorMsg)
		assert.Equal(t, testCase.placementRuleName, placementRuleBuilder.Definition.Name)
		assert.Equal(t, testCase.placementRuleNamespace, placementRuleBuilder.Definition.Namespace)
	}
}

func TestPullPlacementRule(t *testing.T) {
	testCases := []struct {
		placementRuleName      string
		placementRuleNamespace string
		addToRuntimeObjects    bool
		client                 bool
		expectedErrorText      string
	}{
		{
			placementRuleName:      defaultPlacementRuleName,
			placementRuleNamespace: defaultPlacementRuleNsName,
			addToRuntimeObjects:    true,
			client:                 true,
			expectedErrorText:      "",
		},
		{
			placementRuleName:      defaultPlacementRuleName,
			placementRuleNamespace: defaultPlacementRuleNsName,
			addToRuntimeObjects:    false,
			client:                 true,
			expectedErrorText: fmt.Sprintf(
				"placementrule object %s doesn't exist in namespace %s", defaultPlacementRuleName, defaultPlacementRuleNsName),
		},
		{
			placementRuleName:      "",
			placementRuleNamespace: defaultPlacementRuleNsName,
			addToRuntimeObjects:    false,
			client:                 true,
			expectedErrorText:      "placementrule's 'name' cannot be empty",
		},
		{
			placementRuleName:      defaultPlacementRuleName,
			placementRuleNamespace: "",
			addToRuntimeObjects:    false,
			client:                 true,
			expectedErrorText:      "placementrule's 'namespace' cannot be empty",
		},
		{
			placementRuleName:      defaultPlacementRuleName,
			placementRuleNamespace: defaultPlacementRuleNsName,
			addToRuntimeObjects:    false,
			client:                 false,
			expectedErrorText:      "placementrule's 'apiClient' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testPlacementRule := buildDummyPlacementRule(testCase.placementRuleName, testCase.placementRuleNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPlacementRule)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		placementRuleBuilder, err := PullPlacementRule(testSettings, testPlacementRule.Name, testPlacementRule.Namespace)

		if testCase.expectedErrorText != "" {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedErrorText, err.Error())
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testPlacementRule.Name, placementRuleBuilder.Object.Name)
			assert.Equal(t, testPlacementRule.Namespace, placementRuleBuilder.Object.Namespace)
		}
	}
}

func TestPlacementRuleExists(t *testing.T) {
	testCases := []struct {
		testBuilder *PlacementRuleBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidPlacementRuleTestBuilder(buildTestClientWithDummyPlacementRule()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidPlacementRuleTestBuilder(buildTestClientWithDummyPlacementRule()),
			exists:      false,
		},
		{
			testBuilder: buildValidPlacementRuleTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestPlacementRuleGet(t *testing.T) {
	testCases := []struct {
		testBuilder           *PlacementRuleBuilder
		expectedPlacementRule *placementrulev1.PlacementRule
	}{
		{
			testBuilder:           buildValidPlacementRuleTestBuilder(buildTestClientWithDummyPlacementRule()),
			expectedPlacementRule: buildDummyPlacementRule(defaultPlacementRuleName, defaultPlacementRuleNsName),
		},
		{
			testBuilder:           buildValidPlacementRuleTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedPlacementRule: nil,
		},
	}

	for _, testCase := range testCases {
		placementRule, err := testCase.testBuilder.Get()

		if testCase.expectedPlacementRule == nil {
			assert.Nil(t, placementRule)
			assert.True(t, k8serrors.IsNotFound(err))
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedPlacementRule.Name, placementRule.Name)
			assert.Equal(t, testCase.expectedPlacementRule.Namespace, placementRule.Namespace)
		}
	}
}

func TestPlacementRuleCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *PlacementRuleBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPlacementRuleTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPlacementRuleTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("placementrule's 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		placementRuleBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, placementRuleBuilder.Definition, placementRuleBuilder.Object)
		}
	}
}

func TestPlacementRuleDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *PlacementRuleBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPlacementRuleTestBuilder(buildTestClientWithDummyPlacementRule()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPlacementRuleTestBuilder(buildTestClientWithDummyPlacementRule()),
			expectedError: fmt.Errorf("placementrule's 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

func TestPlacementRuleUpdate(t *testing.T) {
	testCases := []struct {
		alreadyExists bool
		force         bool
	}{
		{
			alreadyExists: false,
			force:         false,
		},
		{
			alreadyExists: true,
			force:         false,
		},
		{
			alreadyExists: false,
			force:         true,
		},
		{
			alreadyExists: true,
			force:         true,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPlacementRuleTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

		// Create the builder rather than just adding it to the client so that the proper metadata is added and
		// the update won't fail.
		if testCase.alreadyExists {
			var err error

			testBuilder = buildValidPlacementRuleTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
			testBuilder, err = testBuilder.Create()
			assert.Nil(t, err)
		}

		assert.NotNil(t, testBuilder.Definition)
		assert.Empty(t, testBuilder.Definition.Spec.SchedulerName)

		testBuilder.Definition.Spec.SchedulerName = "test-scheduler"

		placementRuleBuilder, err := testBuilder.Update(testCase.force)
		assert.NotNil(t, testBuilder.Definition)

		if testCase.alreadyExists {
			assert.Nil(t, err)
			assert.Equal(t, testBuilder.Definition.Name, placementRuleBuilder.Definition.Name)
			assert.Equal(t, testBuilder.Definition.Spec.SchedulerName, placementRuleBuilder.Definition.Spec.SchedulerName)
		} else {
			assert.NotNil(t, err)
		}
	}
}

// buildDummyPlacementRule returns a PlacementRule with the provided name and namespace.
func buildDummyPlacementRule(name, nsname string) *placementrulev1.PlacementRule {
	return &placementrulev1.PlacementRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}

// buildTestClientWithDummyPlacementRule returns a client with a mock dummy PlacementRule.
func buildTestClientWithDummyPlacementRule() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPlacementRule(defaultPlacementRuleName, defaultPlacementRuleNsName),
		},
	})
}

// buildValidPlacementRuleTestBuilder returns a valid PlacementRuleBuilder for testing.
func buildValidPlacementRuleTestBuilder(apiClient *clients.Settings) *PlacementRuleBuilder {
	return NewPlacementRuleBuilder(apiClient, defaultPlacementRuleName, defaultPlacementRuleNsName)
}

// buildInvalidPlacementRuleTestBuilder returns an invalid PlacementRuleBuilder for testing.
func buildInvalidPlacementRuleTestBuilder(apiClient *clients.Settings) *PlacementRuleBuilder {
	return NewPlacementRuleBuilder(apiClient, defaultPlacementRuleName, "")
}
