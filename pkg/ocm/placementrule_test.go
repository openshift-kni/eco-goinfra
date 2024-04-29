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

const (
	defaultPlacementRuleName   = "placementrule-test"
	defaultPlacementRuleNsName = "test-ns"
)

var placementRuleTestSchemes = []clients.SchemeAttacher{
	placementrulev1.AddToScheme,
}

func TestNewPlacementRuleBuilder(t *testing.T) {
	testCases := []struct {
		placementRuleName      string
		placementRuleNamespace string
		client                 bool
		expectedErrorText      string
	}{
		{
			placementRuleName:      defaultPlacementRuleName,
			placementRuleNamespace: defaultPlacementRuleNsName,
			client:                 true,
			expectedErrorText:      "",
		},
		{
			placementRuleName:      "",
			placementRuleNamespace: defaultPlacementRuleNsName,
			client:                 true,
			expectedErrorText:      "placementrule's 'name' cannot be empty",
		},
		{
			placementRuleName:      defaultPlacementRuleName,
			placementRuleNamespace: "",
			client:                 true,
			expectedErrorText:      "placementrule's 'nsname' cannot be empty",
		},
		{
			placementRuleName:      defaultPlacementRuleName,
			placementRuleNamespace: defaultPlacementRuleNsName,
			client:                 false,
			expectedErrorText:      "",
		},
	}

	for _, testCase := range testCases {
		var client *clients.Settings

		if testCase.client {
			client = buildTestClientWithPlacementRuleScheme()
		}

		placementRuleBuilder := NewPlacementRuleBuilder(client, testCase.placementRuleName, testCase.placementRuleNamespace)

		if testCase.client {
			assert.Equal(t, testCase.expectedErrorText, placementRuleBuilder.errorMsg)

			if testCase.expectedErrorText == "" {
				assert.Equal(t, testCase.placementRuleName, placementRuleBuilder.Definition.Name)
				assert.Equal(t, testCase.placementRuleNamespace, placementRuleBuilder.Definition.Namespace)
			}
		} else {
			assert.Nil(t, placementRuleBuilder)
		}
	}
}

func TestPullPlacementRule(t *testing.T) {
	testCases := []struct {
		placementRuleName      string
		placementRuleNamespace string
		addToRuntimeObjects    bool
		client                 bool
		expectedError          error
	}{
		{
			placementRuleName:      defaultPlacementRuleName,
			placementRuleNamespace: defaultPlacementRuleNsName,
			addToRuntimeObjects:    true,
			client:                 true,
			expectedError:          nil,
		},
		{
			placementRuleName:      defaultPlacementRuleName,
			placementRuleNamespace: defaultPlacementRuleNsName,
			addToRuntimeObjects:    false,
			client:                 true,
			expectedError: fmt.Errorf(
				"placementrule object %s does not exist in namespace %s", defaultPlacementRuleName, defaultPlacementRuleNsName),
		},
		{
			placementRuleName:      "",
			placementRuleNamespace: defaultPlacementRuleNsName,
			addToRuntimeObjects:    false,
			client:                 true,
			expectedError:          fmt.Errorf("placementrule's 'name' cannot be empty"),
		},
		{
			placementRuleName:      defaultPlacementRuleName,
			placementRuleNamespace: "",
			addToRuntimeObjects:    false,
			client:                 true,
			expectedError:          fmt.Errorf("placementrule's 'namespace' cannot be empty"),
		},
		{
			placementRuleName:      defaultPlacementRuleName,
			placementRuleNamespace: defaultPlacementRuleNsName,
			addToRuntimeObjects:    false,
			client:                 false,
			expectedError:          fmt.Errorf("placementrule's 'apiClient' cannot be empty"),
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
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: placementRuleTestSchemes,
			})
		}

		placementRuleBuilder, err := PullPlacementRule(testSettings, testPlacementRule.Name, testPlacementRule.Namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
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
			testBuilder: buildValidPlacementRuleTestBuilder(buildTestClientWithPlacementRuleScheme()),
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
			testBuilder:           buildValidPlacementRuleTestBuilder(buildTestClientWithPlacementRuleScheme()),
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
			testBuilder:   buildValidPlacementRuleTestBuilder(buildTestClientWithPlacementRuleScheme()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPlacementRuleTestBuilder(buildTestClientWithPlacementRuleScheme()),
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
		force bool
	}{
		{
			force: false,
		},
		{
			force: true,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPlacementRuleTestBuilder(buildTestClientWithPlacementRuleScheme())

		// Create the builder rather than just adding it to the client so that the proper metadata is added and
		// the update will not fail.
		var err error

		testBuilder = buildValidPlacementRuleTestBuilder(buildTestClientWithPlacementRuleScheme())
		testBuilder, err = testBuilder.Create()
		assert.Nil(t, err)

		assert.NotNil(t, testBuilder.Definition)
		assert.Empty(t, testBuilder.Definition.Spec.SchedulerName)

		testBuilder.Definition.Spec.SchedulerName = "test-scheduler"

		placementRuleBuilder, err := testBuilder.Update(testCase.force)
		assert.NotNil(t, testBuilder.Definition)

		assert.Nil(t, err)
		assert.Equal(t, testBuilder.Definition.Name, placementRuleBuilder.Definition.Name)
		assert.Equal(t, testBuilder.Definition.Spec.SchedulerName, placementRuleBuilder.Definition.Spec.SchedulerName)
	}
}

func TestPlacementRuleValidate(t *testing.T) {
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
			expectedError:   fmt.Errorf("error: received nil placementRule builder"),
		},
		{
			builderNil:      false,
			definitionNil:   true,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("can not redefine the undefined placementRule"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    true,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("placementRule builder cannot have nil apiClient"),
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
		placementRuleBuilder := buildValidPlacementRuleTestBuilder(buildTestClientWithPlacementRuleScheme())

		if testCase.builderNil {
			placementRuleBuilder = nil
		}

		if testCase.definitionNil {
			placementRuleBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			placementRuleBuilder.apiClient = nil
		}

		if testCase.builderErrorMsg != "" {
			placementRuleBuilder.errorMsg = testCase.builderErrorMsg
		}

		valid, err := placementRuleBuilder.validate()

		if testCase.expectedError != nil {
			assert.False(t, valid)
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.True(t, valid)
			assert.Nil(t, err)
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
		SchemeAttachers: placementRuleTestSchemes,
	})
}

// buildTestClientWithPlacementRuleScheme returns a client with no objects but the PlacementRule scheme attached.
func buildTestClientWithPlacementRuleScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: placementRuleTestSchemes,
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
