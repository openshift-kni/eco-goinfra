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

const defaultValidatingName = "test-validating-webhook-configuration"

func TestPullValidatingConfiguration(t *testing.T) {
	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultValidatingName,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("validatingWebhookConfiguration 'name' cannot be empty"),
		},
		{
			name:                defaultValidatingName,
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("validatingWebhookConfiguration object %s does not exist", defaultValidatingName),
		},
		{
			name:                defaultValidatingName,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("validatingWebhookConfiguration 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, buildDummyValidatingConfiguration(defaultValidatingName))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := PullValidatingConfiguration(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
		}
	}
}

func TestValidatingConfigurationExists(t *testing.T) {
	testCases := []struct {
		testBuilder *ValidatingConfigurationBuilder
		exists      bool
	}{
		{
			testBuilder: newValidatingConfigurationBuilder(buildTestClientWithDummyValidatingConfiguration()),
			exists:      true,
		},
		{
			testBuilder: newValidatingConfigurationBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestValidatingConfigurationGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *ValidatingConfigurationBuilder
		expectedError string
	}{
		{
			testBuilder:   newValidatingConfigurationBuilder(buildTestClientWithDummyValidatingConfiguration()),
			expectedError: "",
		},
		{
			testBuilder: newValidatingConfigurationBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Sprintf("validatingwebhookconfigurations.admissionregistration.k8s.io \"%s\" not found",
				defaultValidatingName),
		},
	}

	for _, testCase := range testCases {
		validatingWebhookConfiguration, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, defaultValidatingName, validatingWebhookConfiguration.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestValidatingConfigurationDelete(t *testing.T) {
	testCases := []struct {
		testBuilder *ValidatingConfigurationBuilder
		expectedErr error
	}{
		{
			testBuilder: newValidatingConfigurationBuilder(buildTestClientWithDummyValidatingConfiguration()),
			expectedErr: nil,
		},
		{
			testBuilder: newValidatingConfigurationBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Delete()
		assert.Equal(t, testCase.expectedErr, err)

		if testCase.expectedErr == nil {
			assert.Equal(t, defaultValidatingName, testBuilder.Definition.Name)
		}
	}
}

func TestValidatingConfigurationUpdate(t *testing.T) {
	testCases := []struct {
		testBuilder *ValidatingConfigurationBuilder
		expectedErr error
	}{
		{
			testBuilder: newValidatingConfigurationBuilder(buildTestClientWithDummyValidatingConfiguration()),
			expectedErr: nil,
		},
		{
			testBuilder: newValidatingConfigurationBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErr: fmt.Errorf("cannot update non-existent validatingWebhookConfiguration"),
		},
	}

	for _, testCase := range testCases {
		assert.Empty(t, testCase.testBuilder.Definition.Webhooks)

		testCase.testBuilder.Definition.Webhooks = []admregv1.ValidatingWebhook{{}}

		testBuilder, err := testCase.testBuilder.Update()
		assert.Equal(t, testCase.expectedErr, err)

		if testCase.expectedErr == nil {
			assert.Len(t, testBuilder.Object.Webhooks, 1)
		}
	}
}

func TestValidatingConfigurationValidate(t *testing.T) {
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
			expectedError: fmt.Errorf("error: received nil validatingWebhookConfiguration builder"),
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			errorMsg:      "",
			expectedError: fmt.Errorf("can not redefine the undefined validatingWebhookConfiguration"),
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			errorMsg:      "",
			expectedError: fmt.Errorf("validatingWebhookConfiguration builder cannot have nil apiClient"),
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
		testBuilder := newValidatingConfigurationBuilder(clients.GetTestClients(clients.TestClientParams{}))

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

// buildDummyValidatingConfiguration returns a new ValidatingWebhookConfiguration with the given name.
func buildDummyValidatingConfiguration(name string) *admregv1.ValidatingWebhookConfiguration {
	return &admregv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestClientWithDummyValidatingConfiguration returns a new client with a mock ValidatingWebhookConfiguration
// object, using the default name.
func buildTestClientWithDummyValidatingConfiguration() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyValidatingConfiguration(defaultValidatingName),
		},
		SchemeAttachers: testSchemes,
	})
}

// newValidatingConfigurationBuilder returns a ValidatingConfigurationBuilder with the default name and the provided
// client.
func newValidatingConfigurationBuilder(apiClient *clients.Settings) *ValidatingConfigurationBuilder {
	return &ValidatingConfigurationBuilder{
		apiClient: apiClient.Client,
		Definition: &admregv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: defaultValidatingName,
			},
		},
	}
}
