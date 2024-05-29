package ocm

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	mceV1 "github.com/stolostron/backplane-operator/api/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var defaultMultiClusterEngineName = "mce-test"

func TestNewMultiClusterEngineBuilder(t *testing.T) {
	testCases := []struct {
		multiClusterEngineName string
		client                 bool
		expectedErrorText      string
	}{
		{
			multiClusterEngineName: defaultMultiClusterEngineName,
			client:                 true,
			expectedErrorText:      "",
		},
		{
			multiClusterEngineName: "",
			client:                 true,
			expectedErrorText:      "multiclusterengine 'name' cannot be empty",
		},
		{
			multiClusterEngineName: defaultMultiClusterEngineName,
			client:                 false,
			expectedErrorText:      "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		multiClusterEngineBuilder := NewMultiClusterEngineBuilder(testSettings, testCase.multiClusterEngineName)

		if !testCase.client {
			assert.Nil(t, multiClusterEngineBuilder)
		}

		if testCase.client {
			assert.NotNil(t, multiClusterEngineBuilder)
			assert.Equal(t, testCase.expectedErrorText, multiClusterEngineBuilder.errorMsg)
		}
	}
}

func TestMultiClusterEngineWithOptions(t *testing.T) {
	testCases := []struct {
		valid             bool
		options           MultiClusterEngineAdditionalOptions
		expectedErrorText string
	}{
		{
			valid: true,
			options: func(builder *MultiClusterEngineBuilder) (*MultiClusterEngineBuilder, error) {
				builder.Definition.Spec.TargetNamespace = "rhacm"

				return builder, nil
			},
			expectedErrorText: "",
		},
		{
			valid: false,
			options: func(builder *MultiClusterEngineBuilder) (*MultiClusterEngineBuilder, error) {
				return builder, nil
			},
			expectedErrorText: "multiclusterengine 'name' cannot be empty",
		},
		{
			valid: true,
			options: func(builder *MultiClusterEngineBuilder) (*MultiClusterEngineBuilder, error) {
				return builder, fmt.Errorf("error adding additional option")
			},
			expectedErrorText: "error adding additional option",
		},
	}

	for _, testCase := range testCases {
		multiClusterEngineBuilder := buildInvalidMultiClusterEngineTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		if testCase.valid {
			multiClusterEngineBuilder = buildValidMultiClusterEngineTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		}

		multiClusterEngineBuilder = multiClusterEngineBuilder.WithOptions(testCase.options)

		assert.NotNil(t, multiClusterEngineBuilder)
		assert.Equal(t, testCase.expectedErrorText, multiClusterEngineBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, multiClusterEngineBuilder.Definition.Spec.TargetNamespace, "rhacm")
		}
	}
}

func TestPullMultiClusterEngine(t *testing.T) {
	testCases := []struct {
		multiClusterEngineName string
		addToRuntimeObjects    bool
		client                 bool
		expectedErrorText      string
	}{
		{
			multiClusterEngineName: defaultMultiClusterEngineName,
			addToRuntimeObjects:    true,
			client:                 true,
			expectedErrorText:      "",
		},
		{
			multiClusterEngineName: defaultMultiClusterEngineName,
			addToRuntimeObjects:    false,
			client:                 true,
			expectedErrorText:      fmt.Sprintf("multiclusterengine object %s does not exist", defaultMultiClusterEngineName),
		},
		{
			multiClusterEngineName: "",
			addToRuntimeObjects:    false,
			client:                 true,
			expectedErrorText:      "multiclusterengine 'name' cannot be empty",
		},
		{
			multiClusterEngineName: defaultMultiClusterEngineName,
			addToRuntimeObjects:    false,
			client:                 false,
			expectedErrorText:      "multiclusterengine 'apiclient' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testMultiClusterEngine := buildDummyMultiClusterEngine(testCase.multiClusterEngineName)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testMultiClusterEngine)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		multiClusterEngineBuilder, err := PullMultiClusterEngine(testSettings, testCase.multiClusterEngineName)

		if testCase.expectedErrorText != "" {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedErrorText, err.Error())
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testMultiClusterEngine.Name, multiClusterEngineBuilder.Object.Name)
		}
	}
}

func TestMultiClusterEngineExists(t *testing.T) {
	testCases := []struct {
		testBuilder *MultiClusterEngineBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidMultiClusterEngineTestBuilder(buildTestClientWithDummyMultiClusterEngine()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidMultiClusterEngineTestBuilder(buildTestClientWithDummyMultiClusterEngine()),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestMultiClusterEngineDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *MultiClusterEngineBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidMultiClusterEngineTestBuilder(buildTestClientWithDummyMultiClusterEngine()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidMultiClusterEngineTestBuilder(buildTestClientWithDummyMultiClusterEngine()),
			expectedError: fmt.Errorf("multiclusterengine 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

func TestMultiClusterEngineUpdate(t *testing.T) {
	testCases := []struct {
		alreadyExists bool
	}{
		// {
		// 	alreadyExists: false,
		// },
		{
			alreadyExists: true,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidMultiClusterEngineTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

		if testCase.alreadyExists {
			testBuilder = buildValidMultiClusterEngineTestBuilder(buildTestClientWithDummyMultiClusterEngine())
		}

		assert.NotNil(t, testBuilder.Definition)
		assert.Equal(t, testBuilder.Definition.Spec.TargetNamespace, "")

		testBuilder.Definition.Spec.TargetNamespace = "rhacm"

		multiClusterEngineBuilder, err := testBuilder.Update()
		assert.NotNil(t, testBuilder.Definition)

		if testCase.alreadyExists {
			assert.Nil(t, err)
			assert.Equal(t, testBuilder.Definition.Name, multiClusterEngineBuilder.Definition.Name)
			assert.Equal(t, testBuilder.Definition.Spec.TargetNamespace, multiClusterEngineBuilder.Definition.Spec.TargetNamespace)
		} else {
			assert.NotNil(t, err)
		}
	}
}

// buildDummyMultiClusterEngine returns a MultiClusterEngine with the provided name.
func buildDummyMultiClusterEngine(name string) *mceV1.MultiClusterEngine {
	return &mceV1.MultiClusterEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestClientWithDummyMultiClusterEngine returns a client with a mock dummy MultiClusterEngine.
func buildTestClientWithDummyMultiClusterEngine() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyMultiClusterEngine(defaultMultiClusterEngineName),
		},
	})
}

// buildValidMultiClusterEngineTestBuilder returns a valid MultiClusterEngine for testing.
func buildValidMultiClusterEngineTestBuilder(apiClient *clients.Settings) *MultiClusterEngineBuilder {
	return NewMultiClusterEngineBuilder(apiClient, defaultMultiClusterEngineName)
}

// buildInvalidMultiClusterEngineTestBuilder returns an invalid MultiClusterEngine for testing.
func buildInvalidMultiClusterEngineTestBuilder(apiClient *clients.Settings) *MultiClusterEngineBuilder {
	return NewMultiClusterEngineBuilder(apiClient, "")
}
