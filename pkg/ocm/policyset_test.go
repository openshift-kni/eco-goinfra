package ocm

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	policiesv1beta1 "open-cluster-management.io/governance-policy-propagator/api/v1beta1"
)

const (
	defaultPolicySetName   = "policyset-test"
	defaultPolicySetNsName = "test-ns"
)

var policySetTestSchemes = []clients.SchemeAttacher{
	policiesv1beta1.AddToScheme,
}

func TestNewPolicySetBuilder(t *testing.T) {
	testCases := []struct {
		policySetName      string
		policySetNamespace string
		policyName         string
		client             bool
		expectedErrorText  string
	}{
		{
			policySetName:      defaultPolicySetName,
			policySetNamespace: defaultPolicySetNsName,
			policyName:         defaultPolicyName,
			client:             true,
			expectedErrorText:  "",
		},
		{
			policySetName:      "",
			policySetNamespace: defaultPolicySetNsName,
			policyName:         defaultPolicyName,
			client:             true,
			expectedErrorText:  "policyset's 'name' cannot be empty",
		},
		{
			policySetName:      defaultPolicySetName,
			policySetNamespace: "",
			policyName:         defaultPolicyName,
			client:             true,
			expectedErrorText:  "policyset's 'nsname' cannot be empty",
		},
		{
			policySetName:      defaultPolicySetName,
			policySetNamespace: defaultPolicySetNsName,
			policyName:         "",
			client:             true,
			expectedErrorText:  "policyset's 'policy' cannot be empty",
		},
		{
			policySetName:      defaultPolicySetName,
			policySetNamespace: defaultPolicySetNsName,
			policyName:         defaultPolicyName,
			client:             false,
			expectedErrorText:  "",
		},
	}

	for _, testCase := range testCases {
		var client *clients.Settings

		if testCase.client {
			client = buildTestClientWithPolicySetScheme()
		}

		policySetBuilder := NewPolicySetBuilder(
			client,
			testCase.policySetName,
			testCase.policySetNamespace,
			policiesv1beta1.NonEmptyString(testCase.policyName))

		if testCase.client {
			assert.Equal(t, testCase.expectedErrorText, policySetBuilder.errorMsg)

			if testCase.expectedErrorText == "" {
				assert.Equal(t, testCase.expectedErrorText, policySetBuilder.errorMsg)
				assert.Equal(t, testCase.policySetName, policySetBuilder.Definition.Name)
				assert.Equal(t, testCase.policySetNamespace, policySetBuilder.Definition.Namespace)
			}
		} else {
			assert.Nil(t, policySetBuilder)
		}
	}
}

func TestPullPolicySet(t *testing.T) {
	testCases := []struct {
		policySetName       string
		policySetNamespace  string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			policySetName:       defaultPolicySetName,
			policySetNamespace:  defaultPolicySetNsName,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			policySetName:       defaultPolicySetName,
			policySetNamespace:  defaultPolicySetNsName,
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf(
				"policyset object %s does not exist in namespace %s", defaultPolicySetName, defaultPolicySetNsName),
		},
		{
			policySetName:       "",
			policySetNamespace:  defaultPolicySetNsName,
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("policyset's 'name' cannot be empty"),
		},
		{
			policySetName:       defaultPolicySetName,
			policySetNamespace:  "",
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("policyset's 'namespace' cannot be empty"),
		},
		{
			policySetName:       defaultPolicySetName,
			policySetNamespace:  defaultPolicySetNsName,
			addToRuntimeObjects: false,
			client:              false,
			expectedError:       fmt.Errorf("policyset's 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testPolicySet := buildDummyPolicySet(testCase.policySetName, testCase.policySetNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPolicySet)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: policySetTestSchemes,
			})
		}

		policySetBuilder, err := PullPolicySet(testSettings, testPolicySet.Name, testPolicySet.Namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testPolicySet.Name, policySetBuilder.Object.Name)
			assert.Equal(t, testPolicySet.Namespace, policySetBuilder.Object.Namespace)
		}
	}
}

func TestPolicySetExists(t *testing.T) {
	testCases := []struct {
		testBuilder *PolicySetBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidPolicySetTestBuilder(buildTestClientWithDummyPolicySet()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidPolicySetTestBuilder(buildTestClientWithDummyPolicySet()),
			exists:      false,
		},
		{
			testBuilder: buildValidPolicySetTestBuilder(buildTestClientWithPolicySetScheme()),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestPolicySetGet(t *testing.T) {
	testCases := []struct {
		testBuilder       *PolicySetBuilder
		expectedPolicySet *policiesv1beta1.PolicySet
	}{
		{
			testBuilder:       buildValidPolicySetTestBuilder(buildTestClientWithDummyPolicySet()),
			expectedPolicySet: buildDummyPolicySet(defaultPolicySetName, defaultPolicySetNsName),
		},
		{
			testBuilder:       buildValidPolicySetTestBuilder(buildTestClientWithPolicySetScheme()),
			expectedPolicySet: nil,
		},
	}

	for _, testCase := range testCases {
		policySet, err := testCase.testBuilder.Get()

		if testCase.expectedPolicySet == nil {
			assert.Nil(t, policySet)
			assert.True(t, k8serrors.IsNotFound(err))
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedPolicySet.Name, policySet.Name)
			assert.Equal(t, testCase.expectedPolicySet.Namespace, policySet.Namespace)
		}
	}
}

func TestPolicySetCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *PolicySetBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPolicySetTestBuilder(buildTestClientWithPolicySetScheme()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPolicySetTestBuilder(buildTestClientWithPolicySetScheme()),
			expectedError: fmt.Errorf("policyset's 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		policySetBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, policySetBuilder.Definition, policySetBuilder.Object)
		}
	}
}

func TestPolicySetDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *PolicySetBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPolicySetTestBuilder(buildTestClientWithDummyPolicySet()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPolicySetTestBuilder(buildTestClientWithDummyPolicySet()),
			expectedError: fmt.Errorf("policyset's 'nsname' cannot be empty"),
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

func TestPolicySetUpdate(t *testing.T) {
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
		var err error

		testBuilder := buildValidPolicySetTestBuilder(buildTestClientWithPolicySetScheme())
		testBuilder, err = testBuilder.Create()
		assert.Nil(t, err)

		assert.NotNil(t, testBuilder.Definition)
		assert.Empty(t, testBuilder.Definition.Spec.Description)

		testBuilder.Definition.Spec.Description = "test description"

		policySetBuilder, err := testBuilder.Update(testCase.force)
		assert.NotNil(t, testBuilder.Definition)

		assert.Nil(t, err)
		assert.Equal(t, testBuilder.Definition.Name, policySetBuilder.Definition.Name)
		assert.Equal(t, testBuilder.Definition.Spec.Description, policySetBuilder.Definition.Spec.Description)
	}
}

func TestPolicySetWithPolicy(t *testing.T) {
	testCases := []struct {
		policyName        policiesv1beta1.NonEmptyString
		expectedErrorText string
	}{
		{
			policyName:        "",
			expectedErrorText: "policy in PolicySet Policies spec cannot be empty",
		},
		{
			policyName:        policiesv1beta1.NonEmptyString(defaultPolicyName),
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithPolicySetScheme()
		policySetBuilder := buildValidPolicySetTestBuilder(testSettings).WithAdditionalPolicy(testCase.policyName)
		assert.Equal(t, testCase.expectedErrorText, policySetBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(
				t,
				[]policiesv1beta1.NonEmptyString{policiesv1beta1.NonEmptyString(defaultPolicyName), testCase.policyName},
				policySetBuilder.Definition.Spec.Policies)
		}
	}
}

func TestPolicySetValidate(t *testing.T) {
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
			expectedError:   fmt.Errorf("error: received nil policySet builder"),
		},
		{
			builderNil:      false,
			definitionNil:   true,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("can not redefine the undefined policySet"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    true,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("policySet builder cannot have nil apiClient"),
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
		policySetBuilder := buildValidPolicySetTestBuilder(buildTestClientWithPolicySetScheme())

		if testCase.builderNil {
			policySetBuilder = nil
		}

		if testCase.definitionNil {
			policySetBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			policySetBuilder.apiClient = nil
		}

		if testCase.builderErrorMsg != "" {
			policySetBuilder.errorMsg = testCase.builderErrorMsg
		}

		valid, err := policySetBuilder.validate()

		if testCase.expectedError != nil {
			assert.False(t, valid)
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.True(t, valid)
			assert.Nil(t, err)
		}
	}
}

// buildDummyPolicySet returns a PolicySet with the provided name and namespace.
func buildDummyPolicySet(name, nsname string) *policiesv1beta1.PolicySet {
	return &policiesv1beta1.PolicySet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}

// buildTestClientWithDummyPolicySet returns a client with a mock dummy PolicySet.
func buildTestClientWithDummyPolicySet() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPolicySet(defaultPolicySetName, defaultPolicySetNsName),
		},
		SchemeAttachers: policySetTestSchemes,
	})
}

// buildTestClientWithPolicySetScheme returns a client with no objects but the PolicySet scheme attached.
func buildTestClientWithPolicySetScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: policySetTestSchemes,
	})
}

// buildValidPolicySetTestBuilder returns a valid PolicySetBuilder for testing.
func buildValidPolicySetTestBuilder(apiClient *clients.Settings) *PolicySetBuilder {
	return NewPolicySetBuilder(
		apiClient, defaultPolicySetName, defaultPolicySetNsName, policiesv1beta1.NonEmptyString(defaultPolicyName))
}

// buildInvalidPolicySetTestBuilder returns an invalid PolicySetBuilder for testing.
func buildInvalidPolicySetTestBuilder(apiClient *clients.Settings) *PolicySetBuilder {
	return NewPolicySetBuilder(apiClient, defaultPolicySetName, "", policiesv1beta1.NonEmptyString(defaultPolicyName))
}
