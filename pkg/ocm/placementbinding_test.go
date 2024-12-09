package ocm

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	policiesv1 "open-cluster-management.io/governance-policy-propagator/api/v1"
)

const (
	defaultPlacementBindingName   = "placementbinding-test"
	defaultPlacementBindingNsName = "test-ns"
)

var (
	defaultPlacementBindingRef = policiesv1.PlacementSubject{
		Name:     "placementrule-test",
		APIGroup: "apps.open-cluster-management.io",
		Kind:     "PlacementRule",
	}
	defaultPlacementBindingSubject = policiesv1.Subject{
		Name:     "policyset-test",
		APIGroup: "policy.open-cluster-management.io",
		Kind:     "PolicySet",
	}
	placementBindingTestSchemes = []clients.SchemeAttacher{
		policiesv1.AddToScheme,
	}
)

func TestNewPlacementBindingBuilder(t *testing.T) {
	// No test cases for invalid refs or subjects, those have their own unit tests.
	testCases := []struct {
		placementBindingName      string
		placementBindingNamespace string
		placementBindingRef       policiesv1.PlacementSubject
		placementBindingSubject   policiesv1.Subject
		client                    bool
		expectedErrorText         string
	}{
		{
			placementBindingName:      defaultPlacementBindingName,
			placementBindingNamespace: defaultPlacementBindingNsName,
			placementBindingRef:       defaultPlacementBindingRef,
			placementBindingSubject:   defaultPlacementBindingSubject,
			client:                    true,
			expectedErrorText:         "",
		},
		{
			placementBindingName:      "",
			placementBindingNamespace: defaultPlacementBindingNsName,
			placementBindingRef:       defaultPlacementBindingRef,
			placementBindingSubject:   defaultPlacementBindingSubject,
			client:                    true,
			expectedErrorText:         "placementBinding's 'name' cannot be empty",
		},
		{
			placementBindingName:      defaultPlacementBindingName,
			placementBindingNamespace: "",
			placementBindingRef:       defaultPlacementBindingRef,
			placementBindingSubject:   defaultPlacementBindingSubject,
			client:                    true,
			expectedErrorText:         "placementBinding's 'nsname' cannot be empty",
		},
		{
			placementBindingName:      defaultPlacementBindingName,
			placementBindingNamespace: defaultPlacementBindingNsName,
			placementBindingRef:       defaultPlacementBindingRef,
			placementBindingSubject:   defaultPlacementBindingSubject,
			client:                    false,
			expectedErrorText:         "",
		},
	}

	for _, testCase := range testCases {
		var client *clients.Settings

		if testCase.client {
			client = buildTestClientWithPlacementBindingScheme()
		}

		placementBindingBuilder := NewPlacementBindingBuilder(
			client,
			testCase.placementBindingName,
			testCase.placementBindingNamespace,
			testCase.placementBindingRef,
			testCase.placementBindingSubject)

		if testCase.client {
			assert.Equal(t, testCase.expectedErrorText, placementBindingBuilder.errorMsg)

			if testCase.expectedErrorText == "" {
				assert.Equal(t, testCase.expectedErrorText, placementBindingBuilder.errorMsg)
				assert.Equal(t, testCase.placementBindingRef, placementBindingBuilder.Definition.PlacementRef)
				assert.Equal(t, []policiesv1.Subject{testCase.placementBindingSubject}, placementBindingBuilder.Definition.Subjects)
			}
		} else {
			assert.Nil(t, placementBindingBuilder)
		}
	}
}

func TestPullPlacementBinding(t *testing.T) {
	testCases := []struct {
		placementBindingName      string
		placementBindingNamespace string
		addToRuntimeObjects       bool
		client                    bool
		expectedError             error
	}{
		{
			placementBindingName:      defaultPlacementBindingName,
			placementBindingNamespace: defaultPlacementBindingNsName,
			addToRuntimeObjects:       true,
			client:                    true,
			expectedError:             nil,
		},
		{
			placementBindingName:      defaultPlacementBindingName,
			placementBindingNamespace: defaultPlacementBindingNsName,
			addToRuntimeObjects:       false,
			client:                    true,
			expectedError: fmt.Errorf(
				"placementBinding object %s does not exist in namespace %s",
				defaultPlacementBindingName,
				defaultPlacementBindingNsName),
		},
		{
			placementBindingName:      "",
			placementBindingNamespace: defaultPlacementBindingNsName,
			addToRuntimeObjects:       false,
			client:                    true,
			expectedError:             fmt.Errorf("placementBinding's 'name' cannot be empty"),
		},
		{
			placementBindingName:      defaultPlacementBindingName,
			placementBindingNamespace: "",
			addToRuntimeObjects:       false,
			client:                    true,
			expectedError:             fmt.Errorf("placementBinding's 'namespace' cannot be empty"),
		},
		{
			placementBindingName:      defaultPlacementBindingName,
			placementBindingNamespace: defaultPlacementBindingNsName,
			addToRuntimeObjects:       false,
			client:                    false,
			expectedError:             fmt.Errorf("placementBinding's 'apiClient' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testPlacementBinding := buildDummyPlacementBinding(testCase.placementBindingName, testCase.placementBindingNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPlacementBinding)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: placementBindingTestSchemes,
			})
		}

		placementBindingBuilder, err := PullPlacementBinding(
			testSettings, testPlacementBinding.Name, testPlacementBinding.Namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testPlacementBinding.Name, placementBindingBuilder.Object.Name)
			assert.Equal(t, testPlacementBinding.Namespace, placementBindingBuilder.Object.Namespace)
		}
	}
}

func TestPlacementBindingExists(t *testing.T) {
	testCases := []struct {
		testBuilder *PlacementBindingBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidPlacementBindingTestBuilder(buildTestClientWithDummyPlacementBinding()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidPlacementBindingTestBuilder(buildTestClientWithDummyPlacementBinding()),
			exists:      false,
		},
		{
			testBuilder: buildValidPlacementBindingTestBuilder(buildTestClientWithPlacementBindingScheme()),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestPlacementBindingGet(t *testing.T) {
	testCases := []struct {
		testBuilder              *PlacementBindingBuilder
		expectedPlacementBinding *policiesv1.PlacementBinding
	}{
		{
			testBuilder:              buildValidPlacementBindingTestBuilder(buildTestClientWithDummyPlacementBinding()),
			expectedPlacementBinding: buildDummyPlacementBinding(defaultPlacementBindingName, defaultPlacementBindingNsName),
		},
		{
			testBuilder:              buildValidPlacementBindingTestBuilder(buildTestClientWithPlacementBindingScheme()),
			expectedPlacementBinding: nil,
		},
	}

	for _, testCase := range testCases {
		placementBinding, err := testCase.testBuilder.Get()

		if testCase.expectedPlacementBinding == nil {
			assert.Nil(t, placementBinding)
			assert.True(t, k8serrors.IsNotFound(err))
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedPlacementBinding.Name, placementBinding.Name)
			assert.Equal(t, testCase.expectedPlacementBinding.Namespace, placementBinding.Namespace)
		}
	}
}

func TestPlacementBindingCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *PlacementBindingBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPlacementBindingTestBuilder(buildTestClientWithPlacementBindingScheme()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPlacementBindingTestBuilder(buildTestClientWithPlacementBindingScheme()),
			expectedError: fmt.Errorf("placementBinding's 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		placementBindingBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, placementBindingBuilder.Definition, placementBindingBuilder.Object)
		}
	}
}

func TestPlacementBindingDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *PlacementBindingBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPlacementBindingTestBuilder(buildTestClientWithDummyPlacementBinding()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPlacementBindingTestBuilder(buildTestClientWithDummyPlacementBinding()),
			expectedError: fmt.Errorf("placementBinding's 'nsname' cannot be empty"),
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

func TestPlacementBindingUpdate(t *testing.T) {
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
		// Create the builder rather than just adding it to the client so that the proper metadata is added and
		// the update will not fail.
		var err error

		testBuilder := buildValidPlacementBindingTestBuilder(buildTestClientWithPlacementBindingScheme())
		testBuilder, err = testBuilder.Create()
		assert.Nil(t, err)

		assert.NotNil(t, testBuilder.Definition)
		assert.Empty(t, testBuilder.Definition.SubFilter)

		testBuilder.Definition.SubFilter = "restricted"

		placementBindingBuilder, err := testBuilder.Update(testCase.force)
		assert.NotNil(t, testBuilder.Definition)

		assert.Nil(t, err)
		assert.Equal(t, testBuilder.Definition.Name, placementBindingBuilder.Definition.Name)
		assert.Equal(t, testBuilder.Definition.SubFilter, placementBindingBuilder.Definition.SubFilter)
	}
}

func TestWithAdditionalSubject(t *testing.T) {
	testCases := []struct {
		subject           policiesv1.Subject
		expectedErrorText string
	}{
		{
			subject:           defaultPlacementBindingSubject,
			expectedErrorText: "",
		},
		{
			subject: policiesv1.Subject{
				Name:     "",
				APIGroup: "policy.open-cluster-management.io",
				Kind:     "PolicySet",
			},
			expectedErrorText: "placementBinding's 'Subject.Name' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithPlacementBindingScheme()
		placementBindingBuilder := buildValidPlacementBindingTestBuilder(testSettings).WithAdditionalSubject(testCase.subject)
		assert.Equal(t, testCase.expectedErrorText, placementBindingBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(
				t,
				[]policiesv1.Subject{defaultPlacementBindingSubject, testCase.subject},
				placementBindingBuilder.Definition.Subjects)
		} else {
			assert.Equal(t, []policiesv1.Subject{defaultPlacementBindingSubject}, placementBindingBuilder.Definition.Subjects)
		}
	}
}

func TestValidatePlacementRef(t *testing.T) {
	testCases := []struct {
		ref               policiesv1.PlacementSubject
		expectedErrorText string
	}{
		{
			ref:               defaultPlacementBindingRef,
			expectedErrorText: "",
		},
		{
			ref: policiesv1.PlacementSubject{
				Name:     "",
				APIGroup: "apps.open-cluster-management.io",
				Kind:     "PlacementRule",
			},
			expectedErrorText: "placementBinding's 'PlacementRef.Name' cannot be empty",
		},
		{
			ref: policiesv1.PlacementSubject{
				Name:     "placementrule-test",
				APIGroup: "",
				Kind:     "PlacementRule",
			},
			expectedErrorText: "placementBinding's 'PlacementRef.APIGroup' must be a valid option",
		},
		{
			ref: policiesv1.PlacementSubject{
				Name:     "placementrule-test",
				APIGroup: "apps.open-cluster-management.io",
				Kind:     "",
			},
			expectedErrorText: "placementBinding's 'PlacementRef.Kind' must be a valid option",
		},
	}

	for _, testCase := range testCases {
		err := validatePlacementRef(testCase.ref)
		assert.Equal(t, testCase.expectedErrorText, err)
	}
}

func TestValidateSubject(t *testing.T) {
	testCases := []struct {
		subject           policiesv1.Subject
		expectedErrorText string
	}{
		{
			subject:           defaultPlacementBindingSubject,
			expectedErrorText: "",
		},
		{
			subject: policiesv1.Subject{
				Name:     "",
				APIGroup: "policy.open-cluster-management.io",
				Kind:     "PolicySet",
			},
			expectedErrorText: "placementBinding's 'Subject.Name' cannot be empty",
		},
		{
			subject: policiesv1.Subject{
				Name:     "policyset-test",
				APIGroup: "",
				Kind:     "PolicySet",
			},
			expectedErrorText: "placementBinding's 'Subject.APIGroup' must be 'policy.open-cluster-management.io'",
		},
		{
			subject: policiesv1.Subject{
				Name:     "policyset-test",
				APIGroup: "policy.open-cluster-management.io",
				Kind:     "",
			},
			expectedErrorText: "placementBinding's 'Subject.Kind' must be a valid option",
		},
	}

	for _, testCase := range testCases {
		err := validateSubject(testCase.subject)
		assert.Equal(t, testCase.expectedErrorText, err)
	}
}

func TestPlacementBindingValidate(t *testing.T) {
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
			expectedError:   fmt.Errorf("error: received nil PlacementBinding builder"),
		},
		{
			builderNil:      false,
			definitionNil:   true,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("can not redefine the undefined PlacementBinding"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    true,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("PlacementBinding builder cannot have nil apiClient"),
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
		placementBindingBuilder := buildValidPlacementBindingTestBuilder(buildTestClientWithPlacementBindingScheme())

		if testCase.builderNil {
			placementBindingBuilder = nil
		}

		if testCase.definitionNil {
			placementBindingBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			placementBindingBuilder.apiClient = nil
		}

		if testCase.builderErrorMsg != "" {
			placementBindingBuilder.errorMsg = testCase.builderErrorMsg
		}

		valid, err := placementBindingBuilder.validate()

		if testCase.expectedError != nil {
			assert.False(t, valid)
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.True(t, valid)
			assert.Nil(t, err)
		}
	}
}

// buildDummyPlacementBinding returns a PlacementBinding with the provided name and namespace.
func buildDummyPlacementBinding(name, nsname string) *policiesv1.PlacementBinding {
	return &policiesv1.PlacementBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
		PlacementRef: defaultPlacementBindingRef,
		Subjects:     []policiesv1.Subject{defaultPlacementBindingSubject},
	}
}

// buildTestClientWithDummyPlacementBinding returns a client with a mock dummy PlacementBinding.
func buildTestClientWithDummyPlacementBinding() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPlacementBinding(defaultPlacementBindingName, defaultPlacementBindingNsName),
		},
		SchemeAttachers: placementBindingTestSchemes,
	})
}

// buildTestClientWithPlacementBindingScheme returns a client with no objects but the PlacementBinding scheme attached.
func buildTestClientWithPlacementBindingScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: placementBindingTestSchemes,
	})
}

// buildValidPlacementBindingTestBuilder returns a valid PlacementBindingBuilder for testing.
func buildValidPlacementBindingTestBuilder(apiClient *clients.Settings) *PlacementBindingBuilder {
	return NewPlacementBindingBuilder(
		apiClient,
		defaultPlacementBindingName,
		defaultPlacementBindingNsName,
		defaultPlacementBindingRef,
		defaultPlacementBindingSubject)
}

// buildInvalidPlacementBindingTestBuilder returns an invalid PlacementBindingBuilder for testing.
func buildInvalidPlacementBindingTestBuilder(apiClient *clients.Settings) *PlacementBindingBuilder {
	return NewPlacementBindingBuilder(
		apiClient, defaultPlacementBindingName, "", defaultPlacementBindingRef, defaultPlacementBindingSubject)
}
