package webhook

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	admregv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const defaultMutatingName = "test-mutating-webhook-configuration"

var testSchemes = []clients.SchemeAttacher{
	admregv1.AddToScheme,
}

func TestPullMutatingConfiguration(t *testing.T) {
	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultMutatingName,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("mutatingWebhookConfiguration 'name' cannot be empty"),
		},
		{
			name:                defaultMutatingName,
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("mutatingWebhookConfiguration object %s does not exist", defaultMutatingName),
		},
		{
			name:                defaultMutatingName,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("mutatingWebhookConfiguration 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, buildDummyMutatingConfiguration(defaultMutatingName))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := PullMutatingConfiguration(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
		}
	}
}

func TestMutatingConfigurationExists(t *testing.T) {
	testCases := []struct {
		testBuilder *MutatingConfigurationBuilder
		exists      bool
	}{
		{
			testBuilder: newMutatingConfigurationBuilder(buildTestClientWithDummyMutatingConfiguration()),
			exists:      true,
		},
		{
			testBuilder: newMutatingConfigurationBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestMutatingConfigurationGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *MutatingConfigurationBuilder
		expectedError string
	}{
		{
			testBuilder:   newMutatingConfigurationBuilder(buildTestClientWithDummyMutatingConfiguration()),
			expectedError: "",
		},
		{
			testBuilder: newMutatingConfigurationBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Sprintf("mutatingwebhookconfigurations.admissionregistration.k8s.io \"%s\" not found",
				defaultMutatingName),
		},
	}

	for _, testCase := range testCases {
		mutatingWebhookConfiguration, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, defaultMutatingName, mutatingWebhookConfiguration.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestMutatingConfigurationDelete(t *testing.T) {
	testCases := []struct {
		testBuilder *MutatingConfigurationBuilder
		expectedErr error
	}{
		{
			testBuilder: newMutatingConfigurationBuilder(buildTestClientWithDummyMutatingConfiguration()),
			expectedErr: nil,
		},
		{
			testBuilder: newMutatingConfigurationBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Delete()
		assert.Equal(t, testCase.expectedErr, err)

		if testCase.expectedErr == nil {
			assert.Equal(t, defaultMutatingName, testBuilder.Definition.Name)
		}
	}
}

func TestMutatingConfigurationUpdate(t *testing.T) {
	testCases := []struct {
		testBuilder *MutatingConfigurationBuilder
		expectedErr error
	}{
		{
			testBuilder: newMutatingConfigurationBuilder(buildTestClientWithDummyMutatingConfiguration()),
			expectedErr: nil,
		},
		{
			testBuilder: newMutatingConfigurationBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErr: fmt.Errorf("cannot update non-existent mutatingWebhookConfiguration"),
		},
	}

	for _, testCase := range testCases {
		assert.Empty(t, testCase.testBuilder.Definition.Webhooks)

		testCase.testBuilder.Definition.Webhooks = []admregv1.MutatingWebhook{{}}

		testBuilder, err := testCase.testBuilder.Update()
		assert.Equal(t, testCase.expectedErr, err)

		if testCase.expectedErr == nil {
			assert.Len(t, testBuilder.Object.Webhooks, 1)
		}
	}
}

func TestMutatingConfigurationValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		errorMsg      string
		expectedError error
	}{
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			errorMsg:      "",
			expectedError: nil,
		},
		{
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			errorMsg:      "",
			expectedError: fmt.Errorf("error: received nil mutatingWebhookConfiguration builder"),
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			errorMsg:      "",
			expectedError: fmt.Errorf("can not redefine the undefined mutatingWebhookConfiguration"),
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			errorMsg:      "",
			expectedError: fmt.Errorf("mutatingWebhookConfiguration builder cannot have nil apiClient"),
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			errorMsg:      "test error",
			expectedError: fmt.Errorf("test error"),
		},
	}

	for _, testCase := range testCases {
		testBuilder := newMutatingConfigurationBuilder(clients.GetTestClients(clients.TestClientParams{}))

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		if testCase.errorMsg != "" {
			testBuilder.errorMsg = testCase.errorMsg
		}

		valid, err := testBuilder.validate()

		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, testCase.expectedError == nil, valid)
	}
}

// buildDummyMutatingConfiguration returns a new MutatingWebhookConfiguration with the given name.
func buildDummyMutatingConfiguration(name string) *admregv1.MutatingWebhookConfiguration {
	return &admregv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestClientWithDummyMutatingConfiguration returns a new client with a mock MutatingWebhookConfiguration object,
// using the default name.
func buildTestClientWithDummyMutatingConfiguration() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyMutatingConfiguration(defaultMutatingName),
		},
		SchemeAttachers: testSchemes,
	})
}

// newMutatingConfigurationBuilder returns a MutatingConfigurationBuilder with the default name and the provided client.
func newMutatingConfigurationBuilder(apiClient *clients.Settings) *MutatingConfigurationBuilder {
	return &MutatingConfigurationBuilder{
		apiClient: apiClient.Client,
		Definition: &admregv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: defaultMutatingName,
			},
		},
	}
}
