package oran

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	provisioningv1alpha1 "github.com/openshift-kni/oran-o2ims/api/provisioning/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultPRName            = "test-cluster"
	defaultPRTemplateName    = "test-cluster-template"
	defaultPRTemplateVersion = "v4-18-0-1"
)

var provisioningTestSchemes = []clients.SchemeAttacher{
	provisioningv1alpha1.AddToScheme,
}

func TestNewPRBuilder(t *testing.T) {
	testCases := []struct {
		name            string
		templateName    string
		templateVersion string
		client          bool
		expectedError   string
	}{
		{
			name:            defaultPRName,
			templateName:    defaultPRTemplateName,
			templateVersion: defaultPRTemplateVersion,
			client:          true,
			expectedError:   "",
		},
		{
			name:            "",
			templateName:    defaultPRTemplateName,
			templateVersion: defaultPRTemplateVersion,
			client:          true,
			expectedError:   "provisioningRequest 'name' cannot be empty",
		},
		{
			name:            defaultPRName,
			templateName:    "",
			templateVersion: defaultPRTemplateVersion,
			client:          true,
			expectedError:   "provisioningRequest 'templateName' cannot be empty",
		},
		{
			name:            defaultPRName,
			templateName:    defaultPRTemplateName,
			templateVersion: "",
			client:          true,
			expectedError:   "provisioningRequest 'templateVersion' cannot be empty",
		},
		{
			name:            defaultPRName,
			templateName:    defaultPRTemplateName,
			templateVersion: defaultPRTemplateVersion,
			client:          false,
			expectedError:   "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewPRBuilder(
			testSettings, testCase.name, testCase.templateName, testCase.templateVersion)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
				assert.Equal(t, testCase.templateName, testBuilder.Definition.Spec.TemplateName)
				assert.Equal(t, testCase.templateVersion, testBuilder.Definition.Spec.TemplateVersion)
			}
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestPullPR(t *testing.T) {
	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultPRName,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("provisioningRequest 'name' cannot be empty"),
		},
		{
			name:                defaultPRName,
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("provisioningRequest object %s does not exist", defaultPRName),
		},
		{
			name:                defaultPRName,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("provisioningRequest 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, buildDummyPR(defaultPRName))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: provisioningTestSchemes,
			})
		}

		testBuilder, err := PullPR(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
		}
	}
}

func TestPRGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *ProvisioningRequestBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidPRTestBuilder(buildTestClientWithDummyPR()),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidPRTestBuilder(buildTestClientWithDummyPR()),
			expectedError: "provisioningRequest 'templateName' cannot be empty",
		},
		{
			testBuilder:   buildValidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Sprintf("provisioningrequests.o2ims.provisioning.oran.org \"%s\" not found", defaultPRName),
		},
	}

	for _, testCase := range testCases {
		provisioningRequest, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, provisioningRequest.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestPRExists(t *testing.T) {
	testCases := []struct {
		testBuilder *ProvisioningRequestBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidPRTestBuilder(buildTestClientWithDummyPR()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidPRTestBuilder(buildTestClientWithDummyPR()),
			exists:      false,
		},
		{
			testBuilder: buildValidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestPRCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *ProvisioningRequestBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPRTestBuilder(buildTestClientWithDummyPR()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPRTestBuilder(buildTestClientWithDummyPR()),
			expectedError: fmt.Errorf("provisioningRequest 'templateName' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testBuilder.Definition.Name, testBuilder.Object.Name)
		}
	}
}

func TestPRUpdate(t *testing.T) {
	testCases := []struct {
		testBuilder   *ProvisioningRequestBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPRTestBuilder(buildTestClientWithDummyPR()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("cannot update non-existent provisioningRequest"),
		},
		{
			testBuilder:   buildInvalidPRTestBuilder(buildTestClientWithDummyPR()),
			expectedError: fmt.Errorf("provisioningRequest 'templateName' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		assert.Empty(t, testCase.testBuilder.Definition.Spec.Name)

		testCase.testBuilder.Definition.Spec.Name = "test"
		testCase.testBuilder.Definition.ResourceVersion = "999"

		testBuilder, err := testCase.testBuilder.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, "test", testBuilder.Object.Spec.Name)
		}
	}
}

func TestPRDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *ProvisioningRequestBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPRTestBuilder(buildTestClientWithDummyPR()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPRTestBuilder(buildTestClientWithDummyPR()),
			expectedError: fmt.Errorf("provisioningRequest 'templateName' cannot be empty"),
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

func buildDummyPR(name string) *provisioningv1alpha1.ProvisioningRequest {
	return &provisioningv1alpha1.ProvisioningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: provisioningv1alpha1.ProvisioningRequestSpec{
			TemplateName:    defaultPRTemplateName,
			TemplateVersion: defaultPRTemplateVersion,
		},
	}
}

func buildTestClientWithDummyPR() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPR(defaultPRName),
		},
		SchemeAttachers: provisioningTestSchemes,
	})
}

func buildValidPRTestBuilder(apiClient *clients.Settings) *ProvisioningRequestBuilder {
	return NewPRBuilder(apiClient, defaultPRName, defaultPRTemplateName, defaultPRTemplateVersion)
}

func buildInvalidPRTestBuilder(apiClient *clients.Settings) *ProvisioningRequestBuilder {
	return NewPRBuilder(apiClient, defaultPRName, "", defaultPRTemplateVersion)
}
