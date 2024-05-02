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

var (
	defaultPlacementBindingName   = "placementbinding-test"
	defaultPlacementBindingNsName = "test-ns"
	defaultPlacementBindingRef    = policiesv1.PlacementSubject{
		Name:     "placementrule-test",
		APIGroup: "apps.open-cluster-management.io",
		Kind:     "PlacementRule",
	}
	defaultPlacementBindingSubject = policiesv1.Subject{
		Name:     "policyset-test",
		APIGroup: "policy.open-cluster-management.io",
		Kind:     "PolicySet",
	}
)

func TestNewPlacementBindingBuilder(t *testing.T) {
	// No test cases for invalid refs or subjects, those have their own unit tests.
	testCases := []struct {
		placementBindingName      string
		placementBindingNamespace string
		placementBindingRef       policiesv1.PlacementSubject
		placementBindingSubject   policiesv1.Subject
		expectedErrorText         string
	}{
		{
			placementBindingName:      defaultPlacementBindingName,
			placementBindingNamespace: defaultPlacementBindingNsName,
			placementBindingRef:       defaultPlacementBindingRef,
			placementBindingSubject:   defaultPlacementBindingSubject,
			expectedErrorText:         "",
		},
		{
			placementBindingName:      "",
			placementBindingNamespace: defaultPlacementBindingNsName,
			placementBindingRef:       defaultPlacementBindingRef,
			placementBindingSubject:   defaultPlacementBindingSubject,
			expectedErrorText:         "placementBinding's 'name' cannot be empty",
		},
		{
			placementBindingName:      defaultPlacementBindingName,
			placementBindingNamespace: "",
			placementBindingRef:       defaultPlacementBindingRef,
			placementBindingSubject:   defaultPlacementBindingSubject,
			expectedErrorText:         "placementBinding's 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		placementBindingBuilder := NewPlacementBindingBuilder(
			testSettings,
			testCase.placementBindingName,
			testCase.placementBindingNamespace,
			testCase.placementBindingRef,
			testCase.placementBindingSubject)

		assert.NotNil(t, placementBindingBuilder)
		assert.Equal(t, testCase.expectedErrorText, placementBindingBuilder.errorMsg)
		assert.Equal(t, testCase.placementBindingRef, placementBindingBuilder.Definition.PlacementRef)
		assert.Equal(t, []policiesv1.Subject{testCase.placementBindingSubject}, placementBindingBuilder.Definition.Subjects)
	}
}

func TestPullPlacementBinding(t *testing.T) {
	testCases := []struct {
		placementBindingName      string
		placementBindingNamespace string
		addToRuntimeObjects       bool
		client                    bool
		expectedErrorText         string
	}{
		{
			placementBindingName:      defaultPlacementBindingName,
			placementBindingNamespace: defaultPlacementBindingNsName,
			addToRuntimeObjects:       true,
			client:                    true,
			expectedErrorText:         "",
		},
		{
			placementBindingName:      defaultPlacementBindingName,
			placementBindingNamespace: defaultPlacementBindingNsName,
			addToRuntimeObjects:       false,
			client:                    true,
			expectedErrorText: fmt.Sprintf(
				"placementBinding object %s doesn't exist in namespace %s",
				defaultPlacementBindingName,
				defaultPlacementBindingNsName),
		},
		{
			placementBindingName:      "",
			placementBindingNamespace: defaultPlacementBindingNsName,
			addToRuntimeObjects:       false,
			client:                    true,
			expectedErrorText:         "placementBinding's 'name' cannot be empty",
		},
		{
			placementBindingName:      defaultPlacementBindingName,
			placementBindingNamespace: "",
			addToRuntimeObjects:       false,
			client:                    true,
			expectedErrorText:         "placementBinding's 'namespace' cannot be empty",
		},
		{
			placementBindingName:      defaultPlacementBindingName,
			placementBindingNamespace: defaultPlacementBindingNsName,
			addToRuntimeObjects:       false,
			client:                    false,
			expectedErrorText:         "placementBinding's 'apiClient' cannot be empty",
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
				K8sMockObjects: runtimeObjects,
			})
		}

		placementBindingBuilder, err := PullPlacementBinding(
			testSettings, testPlacementBinding.Name, testPlacementBinding.Namespace)

		if testCase.expectedErrorText != "" {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedErrorText, err.Error())
		} else {
			assert.Nil(t, err)
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
			testBuilder: buildValidPlacementBindingTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
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
			testBuilder:              buildValidPlacementBindingTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
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
			testBuilder:   buildValidPlacementBindingTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPlacementBindingTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
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
		testBuilder := buildValidPlacementBindingTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

		// Create the builder rather than just adding it to the client so that the proper metadata is added and
		// the update will not fail.
		if testCase.alreadyExists {
			var err error

			testBuilder = buildValidPlacementBindingTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
			testBuilder, err = testBuilder.Create()
			assert.Nil(t, err)
		}

		assert.NotNil(t, testBuilder.Definition)
		assert.Empty(t, testBuilder.Definition.SubFilter)

		testBuilder.Definition.SubFilter = "restricted"

		placementBindingBuilder, err := testBuilder.Update(testCase.force)
		assert.NotNil(t, testBuilder.Definition)

		if testCase.alreadyExists {
			assert.Nil(t, err)
			assert.Equal(t, testBuilder.Definition.Name, placementBindingBuilder.Definition.Name)
			assert.Equal(t, testBuilder.Definition.SubFilter, placementBindingBuilder.Definition.SubFilter)
		} else {
			assert.NotNil(t, err)
		}
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
		testSettings := clients.GetTestClients(clients.TestClientParams{})
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
