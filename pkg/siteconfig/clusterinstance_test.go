package siteconfig

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	siteconfigv1alpha1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/siteconfig/v1alpha1"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
)

var (
	// defaultClusterInstanceCondition  is a variable which depicts default value of ClusterInstance condition types.
	defaultClusterInstanceCondition = metav1.Condition{Type: "", Status: metav1.ConditionTrue}
)

const (
	testClusterInstance = "test-cluster-instance"
)

var testSchemes = []clients.SchemeAttacher{
	siteconfigv1alpha1.AddToScheme,
}

func TestNewClusterInstanceBuilder(t *testing.T) {
	testCases := []struct {
		name              string
		namespace         string
		client            bool
		expectedErrorText string
	}{
		{
			name:              testClusterInstance,
			namespace:         testClusterInstance,
			client:            true,
			expectedErrorText: "",
		},
		{
			name:              "",
			namespace:         testClusterInstance,
			client:            true,
			expectedErrorText: "clusterinstance 'name' cannot be empty",
		},
		{
			name:              testClusterInstance,
			namespace:         "",
			client:            true,
			expectedErrorText: "clusterinstance 'nsname' cannot be empty",
		},
		{
			name:              testClusterInstance,
			namespace:         testClusterInstance,
			client:            false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testClusterInstanceStructure := NewCIBuilder(
			testSettings,
			testCase.name,
			testCase.namespace)

		if testCase.client {
			assert.NotNil(t, testClusterInstanceStructure)
			assert.Equal(t, testCase.expectedErrorText, testClusterInstanceStructure.errorMsg)
		} else {
			assert.Nil(t, testClusterInstanceStructure)
		}
	}
}

func TestClusterInstancePull(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		client        bool
		exists        bool
		expectedError error
	}{
		{
			name:          testClusterInstance,
			namespace:     testClusterInstance,
			client:        true,
			exists:        true,
			expectedError: nil,
		},
		{
			name:          "",
			namespace:     testClusterInstance,
			client:        true,
			exists:        true,
			expectedError: fmt.Errorf("clusterinstance 'name' cannot be empty"),
		},
		{
			name:          testClusterInstance,
			namespace:     "",
			client:        true,
			exists:        true,
			expectedError: fmt.Errorf("clusterinstance 'nsname' cannot be empty"),
		},
		{
			name:          testClusterInstance,
			namespace:     testClusterInstance,
			client:        false,
			exists:        true,
			expectedError: fmt.Errorf("apiClient cannot be nil"),
		},
		{
			name:      testClusterInstance,
			namespace: testClusterInstance,
			client:    true,
			exists:    false,
			expectedError: fmt.Errorf(
				"clusterinstance object %s does not exist in namespace %s",
				testClusterInstance, testClusterInstance),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testClient     *clients.Settings
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateClusterInstance())
		}

		if testCase.client {
			testClient = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := PullClusterInstance(testClient, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testBuilder.Definition.Namespace)
		}
	}
}

func TestClusterInstanceWithExtraManifests(t *testing.T) {
	testCases := []struct {
		extramanifest    string
		expectedErrorMsg string
	}{
		{
			extramanifest:    "ci-extra-manifest",
			expectedErrorMsg: "",
		},
		{
			extramanifest:    "",
			expectedErrorMsg: "clusterinstance extramanifest cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateClusterInstanceBuilderWithFakeObjects([]runtime.Object{})

		testBuilder.WithExtraManifests(testCase.extramanifest)
		assert.Equal(t, testCase.expectedErrorMsg, testBuilder.errorMsg)

		if testCase.expectedErrorMsg == "" {
			assert.Equal(t, testCase.extramanifest, testBuilder.Definition.Spec.ExtraManifestsRefs[0].Name)
		}
	}
}

func TestClusterInstanceWithExtraLabels(t *testing.T) {
	testCases := []struct {
		testKey        string
		testLabels     map[string]string
		expectedErrMsg string
		emptyLabels    bool
	}{
		{
			testKey:        "test-key",
			testLabels:     map[string]string{"test-label-key": "test-label-value"},
			expectedErrMsg: "",
			emptyLabels:    false,
		},
		{
			testKey:        "",
			testLabels:     map[string]string{"test-label-key": "test-label-value"},
			expectedErrMsg: "can not apply empty key",
			emptyLabels:    false,
		},
		{
			testKey:        "test-key",
			testLabels:     map[string]string{"": "test-label-value"},
			expectedErrMsg: "can not apply a labels with an empty key",
			emptyLabels:    false,
		},
		{
			testKey:        "test-key",
			testLabels:     map[string]string{},
			expectedErrMsg: "labels can not be empty",
			emptyLabels:    true,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidClusterInstanceTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

		if testCase.emptyLabels {
			testBuilder.Definition.Spec.ExtraLabels = nil
		}

		testBuilder.WithExtraLabels(testCase.testKey, testCase.testLabels)

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.testLabels, testBuilder.Definition.Spec.ExtraLabels[testCase.testKey])
		}
	}
}

func TestClusterInstanceWaitForCondition(t *testing.T) {
	testCases := []struct {
		condition     metav1.Condition
		exists        bool
		conditionMet  bool
		valid         bool
		expectedError error
	}{
		{
			condition:     defaultClusterInstanceCondition,
			exists:        true,
			conditionMet:  true,
			valid:         true,
			expectedError: nil,
		},
		{
			condition:    defaultClusterInstanceCondition,
			exists:       false,
			conditionMet: true,
			valid:        true,
			expectedError: fmt.Errorf("clusterinstance object %s does not exist in namespace %s",
				testClusterInstance, testClusterInstance),
		},
		{
			condition:     defaultClusterInstanceCondition,
			exists:        true,
			conditionMet:  false,
			valid:         true,
			expectedError: context.DeadlineExceeded,
		},
		{
			condition:     defaultClusterInstanceCondition,
			exists:        true,
			conditionMet:  true,
			valid:         false,
			expectedError: fmt.Errorf("clusterinstance 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects         []runtime.Object
			clusterInstanceBuilder *CIBuilder
		)

		if testCase.exists {
			clusterinstance := generateClusterInstance()

			if testCase.conditionMet {
				clusterinstance.Status.Conditions = append(clusterinstance.Status.Conditions, testCase.condition)
			}

			runtimeObjects = append(runtimeObjects, clusterinstance)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: testSchemes,
		})

		if testCase.valid {
			clusterInstanceBuilder = buildValidClusterInstanceTestBuilder(testSettings)
		} else {
			clusterInstanceBuilder = buildInvalidClusterInstanceTestBuilder(testSettings)
		}

		_, err := clusterInstanceBuilder.WaitForCondition(testCase.condition, time.Second)
		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestClusterInstanceGet(t *testing.T) {
	testCases := []struct {
		exists bool
	}{
		{
			exists: true,
		},
		{
			exists: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateClusterInstance())
		}

		testBuilder := generateClusterInstanceBuilderWithFakeObjects(runtimeObjects)

		clusterinstance, err := testBuilder.Get()
		if testCase.exists {
			assert.Nil(t, err)
			assert.NotNil(t, clusterinstance)
		} else {
			assert.NotNil(t, err)
			assert.Nil(t, clusterinstance)
		}
	}
}

func TestClusterInstanceCreate(t *testing.T) {
	testCases := []struct {
		exists bool
	}{
		{
			exists: true,
		},
		{
			exists: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateClusterInstance())
		}

		testBuilder := generateClusterInstanceBuilderWithFakeObjects(runtimeObjects)

		result, err := testBuilder.Create()
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testClusterInstance, result.Definition.Name)
		assert.Equal(t, testClusterInstance, result.Definition.Namespace)
	}
}

func TestClusterInstanceUpdate(t *testing.T) {
	testCases := []struct {
		exists        bool
		expectedError error
	}{
		{
			exists:        true,
			expectedError: nil,
		},
		{
			exists:        false,
			expectedError: fmt.Errorf("cannot update non-existent clusterinstance"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateClusterInstance())
		}

		testBuilder := generateClusterInstanceBuilderWithFakeObjects(runtimeObjects)

		testBuilder.Definition.Spec.ClusterName = "test-clustername"

		clusterinstance, err := testBuilder.Update(true)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, clusterinstance.Object.Spec.ClusterName, "test-clustername")
		}
	}
}

func TestClusterInstanceDelete(t *testing.T) {
	testCases := []struct {
		exists        bool
		expectedError error
	}{
		{
			exists:        true,
			expectedError: nil,
		},
		{
			exists:        false,
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateClusterInstance())
		}

		testBuilder := generateClusterInstanceBuilderWithFakeObjects(runtimeObjects)

		err := testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testBuilder.Object)
		}
	}
}
func TestClusterInstanceExists(t *testing.T) {
	testCases := []struct {
		exists bool
	}{
		{
			exists: true,
		},
		{
			exists: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateClusterInstance())
		}

		testBuilder := generateClusterInstanceBuilderWithFakeObjects(runtimeObjects)

		assert.Equal(t, testCase.exists, testBuilder.Exists())
	}
}

func TestClusterInstanceValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		expectedError string
	}{
		{
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "error: received nil ClusterInstance builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined ClusterInstance",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "ClusterInstance builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateClusterInstanceBuilderWithFakeObjects([]runtime.Object{})

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		result, err := testBuilder.validate()
		if testCase.expectedError != "" {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedError, err.Error())
			assert.False(t, result)
		} else {
			assert.Nil(t, err)
			assert.True(t, result)
		}
	}
}

func generateClusterInstanceBuilderWithFakeObjects(objects []runtime.Object) *CIBuilder {
	return &CIBuilder{
		apiClient: clients.GetTestClients(
			clients.TestClientParams{K8sMockObjects: objects, SchemeAttachers: testSchemes}).Client,
		Definition: generateClusterInstance(),
	}
}

func generateClusterInstance() *siteconfigv1alpha1.ClusterInstance {
	return &siteconfigv1alpha1.ClusterInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testClusterInstance,
			Namespace: testClusterInstance,
		},
	}
}

// buildValidClusterInstanceTestBuilder returns a valid ClusterInstanceBuilder for testing purposes.
func buildValidClusterInstanceTestBuilder(apiClient *clients.Settings) *CIBuilder {
	return NewCIBuilder(
		apiClient,
		testClusterInstance,
		testClusterInstance)
}

// buildInvalidClusterInstanceTestBuilder returns an invalid ClusterInstanceBuilder for testing purposes.
func buildInvalidClusterInstanceTestBuilder(apiClient *clients.Settings) *CIBuilder {
	return NewCIBuilder(
		apiClient,
		testClusterInstance,
		"",
	)
}
