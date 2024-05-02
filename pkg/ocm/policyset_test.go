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

var (
	defaultPolicySetName   = "policyset-test"
	defaultPolicySetNsName = "test-ns"
)

func TestNewPolicySetBuilder(t *testing.T) {
	testCases := []struct {
		policySetName      string
		policySetNamespace string
		policyName         string
		expectedErrorText  string
	}{
		{
			policySetName:      defaultPolicySetName,
			policySetNamespace: defaultPolicySetNsName,
			policyName:         defaultPolicyName,
			expectedErrorText:  "",
		},
		{
			policySetName:      "",
			policySetNamespace: defaultPolicySetNsName,
			policyName:         defaultPolicyName,
			expectedErrorText:  "policyset's 'name' cannot be empty",
		},
		{
			policySetName:      defaultPolicySetName,
			policySetNamespace: "",
			policyName:         defaultPolicyName,
			expectedErrorText:  "policyset's 'nsname' cannot be empty",
		},
		{
			policySetName:      defaultPolicySetName,
			policySetNamespace: defaultPolicySetNsName,
			policyName:         "",
			expectedErrorText:  "policyset's 'policy' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		policySetBuilder := NewPolicySetBuilder(
			testSettings,
			testCase.policySetName,
			testCase.policySetNamespace,
			policiesv1beta1.NonEmptyString(testCase.policyName))
		assert.NotNil(t, policySetBuilder)
		assert.Equal(t, testCase.expectedErrorText, policySetBuilder.errorMsg)
		assert.Equal(t, testCase.policySetName, policySetBuilder.Definition.Name)
		assert.Equal(t, testCase.policySetNamespace, policySetBuilder.Definition.Namespace)
	}
}

func TestPullPolicySet(t *testing.T) {
	testCases := []struct {
		policySetName       string
		policySetNamespace  string
		addToRuntimeObjects bool
		client              bool
		expectedErrorText   string
	}{
		{
			policySetName:       defaultPolicySetName,
			policySetNamespace:  defaultPolicySetNsName,
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "",
		},
		{
			policySetName:       defaultPolicySetName,
			policySetNamespace:  defaultPolicySetNsName,
			addToRuntimeObjects: false,
			client:              true,
			expectedErrorText: fmt.Sprintf(
				"policyset object %s does not exist in namespace %s", defaultPolicySetName, defaultPolicySetNsName),
		},
		{
			policySetName:       "",
			policySetNamespace:  defaultPolicySetNsName,
			addToRuntimeObjects: false,
			client:              true,
			expectedErrorText:   "policyset's 'name' cannot be empty",
		},
		{
			policySetName:       defaultPolicySetName,
			policySetNamespace:  "",
			addToRuntimeObjects: false,
			client:              true,
			expectedErrorText:   "policyset's 'namespace' cannot be empty",
		},
		{
			policySetName:       defaultPolicySetName,
			policySetNamespace:  defaultPolicySetNsName,
			addToRuntimeObjects: false,
			client:              false,
			expectedErrorText:   "policyset's 'apiClient' cannot be empty",
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
				K8sMockObjects: runtimeObjects,
			})
		}

		policySetBuilder, err := PullPolicySet(testSettings, testPolicySet.Name, testPolicySet.Namespace)

		if testCase.expectedErrorText != "" {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedErrorText, err.Error())
		} else {
			assert.Nil(t, err)
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
			testBuilder: buildValidPolicySetTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
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
			testBuilder:       buildValidPolicySetTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
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
			testBuilder:   buildValidPolicySetTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPolicySetTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
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
		testBuilder := buildValidPolicySetTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

		// Create the builder rather than just adding it to the client so that the proper metadata is added and
		// the update will not fail.
		if testCase.alreadyExists {
			var err error

			testBuilder = buildValidPolicySetTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
			testBuilder, err = testBuilder.Create()
			assert.Nil(t, err)
		}

		assert.NotNil(t, testBuilder.Definition)
		assert.Empty(t, testBuilder.Definition.Spec.Description)

		testBuilder.Definition.Spec.Description = "test description"

		policySetBuilder, err := testBuilder.Update(testCase.force)
		assert.NotNil(t, testBuilder.Definition)

		if testCase.alreadyExists {
			assert.Nil(t, err)
			assert.Equal(t, testBuilder.Definition.Name, policySetBuilder.Definition.Name)
			assert.Equal(t, testBuilder.Definition.Spec.Description, policySetBuilder.Definition.Spec.Description)
		} else {
			assert.NotNil(t, err)
		}
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
		testSettings := clients.GetTestClients(clients.TestClientParams{})
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
