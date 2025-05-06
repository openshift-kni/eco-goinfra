package oran

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	provisioningv1alpha1 "github.com/openshift-kni/oran-o2ims/api/provisioning/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultPRName            = "9c5372f3-ea1d-4a96-8157-b3b874a55cf9"
	defaultPRTemplateName    = "test-cluster-template"
	defaultPRTemplateVersion = "v4-18-0-1"
)

var (
	defaultPRCondition = metav1.Condition{
		Type:   string(provisioningv1alpha1.PRconditionTypes.Validated),
		Reason: string(provisioningv1alpha1.CRconditionReasons.Completed),
		Status: metav1.ConditionTrue,
	}

	provisioningTestSchemes = []clients.SchemeAttacher{
		provisioningv1alpha1.AddToScheme,
	}
)

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
			expectedError:   "provisioningRequest 'name' must be a valid UUID",
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
				assert.Equal(t, testCase.name, testBuilder.Definition.Spec.Name)
				assert.Equal(t, testCase.templateName, testBuilder.Definition.Spec.TemplateName)
				assert.Equal(t, testCase.templateVersion, testBuilder.Definition.Spec.TemplateVersion)
			}
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestPRWithTemplateParameter(t *testing.T) {
	testCases := []struct {
		key           string
		expectedError string
	}{
		{
			key:           "nodeClusterName",
			expectedError: "",
		},
		{
			key:           "",
			expectedError: "provisioningRequest TemplateParameter 'key' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		testBuilder = testBuilder.WithTemplateParameter(testCase.key, nil)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)
	}
}

func TestPRGetTemplateParameters(t *testing.T) {
	testCases := []struct {
		testBuilder   *ProvisioningRequestBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("provisioningRequest 'templateName' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		templateParams, err := testCase.testBuilder.GetTemplateParameters()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Empty(t, templateParams)
		}
	}
}

func TestPRWithTemplateParameters(t *testing.T) {
	testCases := []struct {
		testBuilder   *ProvisioningRequestBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "provisioningRequest 'templateName' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := testCase.testBuilder.WithTemplateParameters(nil)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, []byte("{}"), testBuilder.Definition.Spec.TemplateParameters.Raw)
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
		assert.Empty(t, testCase.testBuilder.Definition.Spec.Description)

		testCase.testBuilder.Definition.Spec.Description = "test"
		testCase.testBuilder.Definition.ResourceVersion = "999"

		testBuilder, err := testCase.testBuilder.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, "test", testBuilder.Object.Spec.Description)
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

func TestPRDeleteAndWait(t *testing.T) {
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
		err := testCase.testBuilder.DeleteAndWait(time.Second)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

func TestPRWaitForCondition(t *testing.T) {
	testCases := []struct {
		conditionMet  bool
		exists        bool
		valid         bool
		expectedError error
	}{
		{
			conditionMet:  true,
			exists:        true,
			valid:         true,
			expectedError: nil,
		},
		{
			conditionMet:  false,
			exists:        true,
			valid:         true,
			expectedError: context.DeadlineExceeded,
		},
		{
			conditionMet:  true,
			exists:        false,
			valid:         true,
			expectedError: fmt.Errorf("cannot wait for non-existent ProvisioningRequest"),
		},
		{
			conditionMet:  true,
			exists:        true,
			valid:         false,
			expectedError: fmt.Errorf("provisioningRequest 'templateName' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testBuilder    *ProvisioningRequestBuilder
		)

		if testCase.exists {
			provisioningRequest := buildDummyPR(defaultPRName)

			if testCase.conditionMet {
				provisioningRequest.Status.Conditions = append(provisioningRequest.Status.Conditions, defaultPRCondition)
			}

			runtimeObjects = append(runtimeObjects, provisioningRequest)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: provisioningTestSchemes,
		})

		if testCase.valid {
			testBuilder = buildValidPRTestBuilder(testSettings)
		} else {
			testBuilder = buildInvalidPRTestBuilder(testSettings)
		}

		_, err := testBuilder.WaitForCondition(defaultPRCondition, time.Second)
		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestPRWaitUntilFulfilled(t *testing.T) {
	testCases := []struct {
		fulfilled     bool
		exists        bool
		valid         bool
		expectedError error
	}{
		{
			fulfilled:     true,
			exists:        true,
			valid:         true,
			expectedError: nil,
		},
		{
			fulfilled:     false,
			exists:        true,
			valid:         true,
			expectedError: context.DeadlineExceeded,
		},
		{
			fulfilled:     true,
			exists:        false,
			valid:         true,
			expectedError: fmt.Errorf("cannot wait for non-existent ProvisioningRequest"),
		},
		{
			fulfilled:     true,
			exists:        true,
			valid:         false,
			expectedError: fmt.Errorf("provisioningRequest 'templateName' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testBuilder    *ProvisioningRequestBuilder
		)

		if testCase.exists {
			provisioningRequest := buildDummyPR(defaultPRName)

			if testCase.fulfilled {
				provisioningRequest.Status.ProvisioningStatus.ProvisioningPhase = provisioningv1alpha1.StateFulfilled
			}

			runtimeObjects = append(runtimeObjects, provisioningRequest)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: provisioningTestSchemes,
		})

		if testCase.valid {
			testBuilder = buildValidPRTestBuilder(testSettings)
		} else {
			testBuilder = buildInvalidPRTestBuilder(testSettings)
		}

		_, err := testBuilder.WaitUntilFulfilled(time.Second)
		assert.Equal(t, testCase.expectedError, err)
	}
}

// Since the rest of the WaitForPhaseAfter method is tested by the WaitUntilFulfilled test, we only cover the case here
// where the update time is non-zero. It should not affect coverage since WaitUntilFulfilled calls WaitForPhaseAfter.
func TestPRWaitForPhaseAfter(t *testing.T) {
	testDummyPR := buildDummyPR(defaultPRName)
	testDummyPR.Status.ProvisioningStatus.ProvisioningPhase = provisioningv1alpha1.StateFulfilled
	testDummyPR.Status.ProvisioningStatus.UpdateTime = metav1.Now()
	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{testDummyPR},
		SchemeAttachers: provisioningTestSchemes,
	})
	testBuilder := buildValidPRTestBuilder(testSettings)

	go func() {
		t.Helper()

		// Simulate a delay before updating the ProvisioningRequest's update time. Also wait on the test context
		// to avoid leaking the goroutine.
		select {
		case <-time.After(time.Second):
			testDummyPR.Status.ProvisioningStatus.UpdateTime = metav1.Now()
			err := testSettings.Update(t.Context(), testDummyPR)
			assert.NoError(t, err)
		case <-t.Context().Done():
		}
	}()

	// Since the method only polls every 3 seconds, timeout after 4 seconds to ensure that the second pull happens.
	err := testBuilder.WaitForPhaseAfter(provisioningv1alpha1.StateFulfilled, time.Now(), 4*time.Second)
	assert.NoError(t, err)
}

//nolint:unparam
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
