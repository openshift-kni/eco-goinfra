package ocm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	policiesv1 "open-cluster-management.io/governance-policy-propagator/api/v1"
)

const (
	defaultPolicyName            = "policy-test"
	defaultPolicyNsName          = "test-ns"
	defaultPolicyMessage         = "wrong type for value; expected string; got int"
	defaultPolicyExpectedMessage = "wrong type for value"
)

var policyTestSchemes = []clients.SchemeAttacher{
	policiesv1.AddToScheme,
}

func TestNewPolicyBuilder(t *testing.T) {
	testCases := []struct {
		policyName        string
		policyNamespace   string
		policyTemplate    *policiesv1.PolicyTemplate
		client            bool
		expectedErrorText string
	}{
		{
			policyName:        defaultPolicyName,
			policyNamespace:   defaultPolicyNsName,
			policyTemplate:    &policiesv1.PolicyTemplate{},
			client:            true,
			expectedErrorText: "",
		},
		{
			policyName:        "",
			policyNamespace:   defaultPolicyNsName,
			policyTemplate:    &policiesv1.PolicyTemplate{},
			client:            true,
			expectedErrorText: "policy 'name' cannot be empty",
		},
		{
			policyName:        defaultPolicyName,
			policyNamespace:   "",
			policyTemplate:    &policiesv1.PolicyTemplate{},
			client:            true,
			expectedErrorText: "policy 'nsname' cannot be empty",
		},
		{
			policyName:        defaultPolicyName,
			policyNamespace:   defaultPolicyNsName,
			policyTemplate:    nil,
			client:            true,
			expectedErrorText: "policy 'template' cannot be nil",
		},
		{
			policyName:        defaultPolicyName,
			policyNamespace:   defaultPolicyNsName,
			policyTemplate:    &policiesv1.PolicyTemplate{},
			client:            false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		var client *clients.Settings

		if testCase.client {
			client = buildTestClientWithPolicyScheme()
		}

		policyBuilder := NewPolicyBuilder(client, testCase.policyName, testCase.policyNamespace, testCase.policyTemplate)

		if testCase.client {
			assert.Equal(t, testCase.expectedErrorText, policyBuilder.errorMsg)

			if testCase.expectedErrorText == "" {
				assert.Equal(t, testCase.policyName, policyBuilder.Definition.Name)
				assert.Equal(t, testCase.policyNamespace, policyBuilder.Definition.Namespace)
				assert.Equal(t, testCase.policyTemplate, policyBuilder.Definition.Spec.PolicyTemplates[0])
			}
		} else {
			assert.Nil(t, policyBuilder)
		}
	}
}

func TestPullPolicy(t *testing.T) {
	testCases := []struct {
		policyName          string
		policyNamespace     string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			policyName:          defaultPolicyName,
			policyNamespace:     defaultPolicyNsName,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			policyName:          defaultPolicyName,
			policyNamespace:     defaultPolicyNsName,
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf(
				"policy object %s does not exist in namespace %s", defaultPolicyName, defaultPolicyNsName),
		},
		{
			policyName:          "",
			policyNamespace:     defaultPolicyNsName,
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("policy's 'name' cannot be empty"),
		},
		{
			policyName:          defaultPolicyName,
			policyNamespace:     "",
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("policy's 'namespace' cannot be empty"),
		},
		{
			policyName:          defaultPolicyName,
			policyNamespace:     defaultPolicyNsName,
			addToRuntimeObjects: false,
			client:              false,
			expectedError:       fmt.Errorf("policy 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testPolicy := buildDummyPolicy(testCase.policyName, testCase.policyNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPolicy)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: policyTestSchemes,
			})
		}

		policyBuilder, err := PullPolicy(testSettings, testPolicy.Name, testPolicy.Namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testPolicy.Name, policyBuilder.Definition.Name)
			assert.Equal(t, testPolicy.Namespace, policyBuilder.Definition.Namespace)
		}
	}
}

func TestPolicyExists(t *testing.T) {
	testCases := []struct {
		testBuilder *PolicyBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicy()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidPolicyTestBuilder(buildTestClientWithDummyPolicy()),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestPolicyGet(t *testing.T) {
	testCases := []struct {
		testBuilder    *PolicyBuilder
		expectedPolicy *policiesv1.Policy
	}{
		{
			testBuilder:    buildValidPolicyTestBuilder(buildTestClientWithDummyPolicy()),
			expectedPolicy: buildDummyPolicy(defaultPolicyName, defaultPolicyNsName),
		},
		{
			testBuilder:    buildValidPolicyTestBuilder(buildTestClientWithPolicyScheme()),
			expectedPolicy: nil,
		},
	}

	for _, testCase := range testCases {
		policy, err := testCase.testBuilder.Get()

		if testCase.expectedPolicy == nil {
			assert.Nil(t, policy)
			assert.True(t, k8serrors.IsNotFound(err))
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedPolicy.Name, policy.Name)
			assert.Equal(t, testCase.expectedPolicy.Namespace, policy.Namespace)
		}
	}
}

func TestPolicyCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *PolicyBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPolicyTestBuilder(buildTestClientWithPolicyScheme()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPolicyTestBuilder(buildTestClientWithPolicyScheme()),
			expectedError: fmt.Errorf("policy 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		policyBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, policyBuilder.Definition, policyBuilder.Object)
		}
	}
}

func TestPolicyDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *PolicyBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPolicyTestBuilder(buildTestClientWithDummyPolicy()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidPolicyTestBuilder(buildTestClientWithPolicyScheme()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPolicyTestBuilder(buildTestClientWithDummyPolicy()),
			expectedError: fmt.Errorf("policy 'nsname' cannot be empty"),
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

func TestPolicyUpdate(t *testing.T) {
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
		testBuilder := buildValidPolicyTestBuilder(buildTestClientWithPolicyScheme())

		// Create the builder rather than just adding it to the client so that the proper metadata is added and
		// the update will not fail.
		if testCase.alreadyExists {
			var err error

			testBuilder = buildValidPolicyTestBuilder(buildTestClientWithPolicyScheme())
			testBuilder, err = testBuilder.Create()
			assert.Nil(t, err)
		}

		assert.NotNil(t, testBuilder.Definition)
		assert.False(t, testBuilder.Definition.Spec.Disabled)

		testBuilder.Definition.Spec.Disabled = true

		policyBuilder, err := testBuilder.Update(testCase.force)
		assert.NotNil(t, testBuilder.Definition)

		if testCase.alreadyExists {
			assert.Nil(t, err)
			assert.Equal(t, testBuilder.Definition.Name, policyBuilder.Definition.Name)
			assert.Equal(t, testBuilder.Definition.Spec.Disabled, policyBuilder.Definition.Spec.Disabled)
		} else {
			assert.NotNil(t, err)
		}
	}
}

func TestWithRemediationAction(t *testing.T) {
	testCases := []struct {
		action            policiesv1.RemediationAction
		expectedErrorText string
	}{
		{
			action:            "Inform",
			expectedErrorText: "",
		},
		{
			action:            "inform",
			expectedErrorText: "",
		},
		{
			action:            "Enforce",
			expectedErrorText: "",
		},
		{
			action:            "enforce",
			expectedErrorText: "",
		},
		{
			action:            "",
			expectedErrorText: "remediation action in policy spec must be either 'Inform' or 'Enforce'",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithPolicyScheme()
		policyBuilder := buildValidPolicyTestBuilder(testSettings).WithRemediationAction(testCase.action)
		assert.Equal(t, testCase.expectedErrorText, policyBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.action, policyBuilder.Definition.Spec.RemediationAction)
		}
	}
}

func TestWithAdditionalPolicyTemplate(t *testing.T) {
	testCases := []struct {
		policyTemplate    *policiesv1.PolicyTemplate
		expectedErrorText string
	}{
		{
			policyTemplate:    &policiesv1.PolicyTemplate{},
			expectedErrorText: "",
		},
		{
			policyTemplate:    nil,
			expectedErrorText: "policy template in policy policytemplates cannot be nil",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithPolicyScheme()
		policyBuilder := buildValidPolicyTestBuilder(testSettings).WithAdditionalPolicyTemplate(testCase.policyTemplate)
		assert.Equal(t, testCase.expectedErrorText, policyBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(
				t, []*policiesv1.PolicyTemplate{{}, testCase.policyTemplate}, policyBuilder.Definition.Spec.PolicyTemplates)
		}
	}
}

func TestPolicyWaitUntilDeleted(t *testing.T) {
	testCases := []struct {
		testBuilder   *PolicyBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPolicyTestBuilder(buildTestClientWithPolicyScheme()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidPolicyTestBuilder(buildTestClientWithDummyPolicy()),
			expectedError: context.DeadlineExceeded,
		},
		{
			testBuilder:   buildInvalidPolicyTestBuilder(buildTestClientWithDummyPolicy()),
			expectedError: fmt.Errorf("policy 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.WaitUntilDeleted(time.Second)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

func TestPolicyWaitUntilComplianceState(t *testing.T) {
	testCases := []struct {
		state policiesv1.ComplianceState
	}{
		{
			state: policiesv1.Compliant,
		},
		{
			state: policiesv1.NonCompliant,
		},
		{
			state: policiesv1.Pending,
		},
	}

	for _, testCase := range testCases {
		dummyPolicy := buildDummyPolicy(defaultPolicyName, defaultPolicyNsName)
		dummyPolicy.Status.ComplianceState = testCase.state

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  []runtime.Object{dummyPolicy},
			SchemeAttachers: policyTestSchemes,
		})

		policyBuilder := buildValidPolicyTestBuilder(testSettings)
		err := policyBuilder.WaitUntilComplianceState(testCase.state, 5*time.Second)

		assert.Nil(t, err)
	}
}

func TestPolicyWaitForStatusMessageToContain(t *testing.T) {
	testCases := []struct {
		expectedMessage string
		valid           bool
		exists          bool
		hasMessage      bool
		expectedError   error
	}{
		{
			expectedMessage: defaultPolicyExpectedMessage,
			valid:           true,
			exists:          true,
			hasMessage:      true,
			expectedError:   nil,
		},
		{
			expectedMessage: "",
			valid:           true,
			exists:          true,
			hasMessage:      true,
			expectedError:   fmt.Errorf("policy expectedMessage is empty"),
		},
		{
			expectedMessage: defaultPolicyExpectedMessage,
			valid:           false,
			exists:          true,
			hasMessage:      true,
			expectedError:   fmt.Errorf("policy 'nsname' cannot be empty"),
		},
		{
			expectedMessage: defaultPolicyExpectedMessage,
			valid:           true,
			exists:          false,
			hasMessage:      true,
			expectedError: fmt.Errorf(
				"policy object %s does not exist in namespace %s", defaultPolicyName, defaultPolicyNsName),
		},
		{
			expectedMessage: defaultPolicyExpectedMessage,
			valid:           true,
			exists:          true,
			hasMessage:      false,
			expectedError:   context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			policyBuilder  *PolicyBuilder
		)

		if testCase.exists {
			policy := buildDummyPolicy(defaultPolicyName, defaultPolicyNsName)

			if testCase.hasMessage {
				policy.Status.Details = []*policiesv1.DetailsPerTemplate{
					{History: []policiesv1.ComplianceHistory{{Message: defaultPolicyMessage}}},
				}
			}

			runtimeObjects = append(runtimeObjects, policy)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: policyTestSchemes,
		})

		if testCase.valid {
			policyBuilder = buildValidPolicyTestBuilder(testSettings)
		} else {
			policyBuilder = buildInvalidPolicyTestBuilder(testSettings)
		}

		_, err := policyBuilder.WaitForStatusMessageToContain(testCase.expectedMessage, time.Second)
		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestPolicyValidate(t *testing.T) {
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
			expectedError:   fmt.Errorf("error: received nil policy builder"),
		},
		{
			builderNil:      false,
			definitionNil:   true,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("can not redefine the undefined policy"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    true,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("policy builder cannot have nil apiClient"),
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
		policyBuilder := buildValidPolicyTestBuilder(buildTestClientWithPolicyScheme())

		if testCase.builderNil {
			policyBuilder = nil
		}

		if testCase.definitionNil {
			policyBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			policyBuilder.apiClient = nil
		}

		if testCase.builderErrorMsg != "" {
			policyBuilder.errorMsg = testCase.builderErrorMsg
		}

		valid, err := policyBuilder.validate()

		if testCase.expectedError != nil {
			assert.False(t, valid)
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.True(t, valid)
			assert.Nil(t, err)
		}
	}
}

// buildDummyPolicy returns a Policy with the provided name and namespace.
func buildDummyPolicy(name, nsname string) *policiesv1.Policy {
	return &policiesv1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
		Spec: policiesv1.PolicySpec{
			PolicyTemplates: []*policiesv1.PolicyTemplate{{}},
		},
	}
}

// buildTestClientWithDummyPolicy returns a client with a mock dummy policy.
func buildTestClientWithDummyPolicy() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPolicy(defaultPolicyName, defaultPolicyNsName),
		},
		SchemeAttachers: policyTestSchemes,
	})
}

// buildTestClientWithPolicyScheme returns a client with no objects but the Policy scheme attached.
func buildTestClientWithPolicyScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: policyTestSchemes,
	})
}

// buildValidPolicyTestBuilder returns a valid PolicyBuilder for testing.
func buildValidPolicyTestBuilder(apiClient *clients.Settings) *PolicyBuilder {
	return NewPolicyBuilder(apiClient, defaultPolicyName, defaultPolicyNsName, &policiesv1.PolicyTemplate{})
}

// buildInvalidPolicyTestBuilder returns an invalid PolicyBuilder for testing.
func buildInvalidPolicyTestBuilder(apiClient *clients.Settings) *PolicyBuilder {
	return NewPolicyBuilder(apiClient, defaultPolicyName, "", &policiesv1.PolicyTemplate{})
}
