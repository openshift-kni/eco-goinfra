package ptp

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	ptpv1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/ptp/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPullPtpOperatorConfig(t *testing.T) {
	testCases := []struct {
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf(
				"ptpOperatorConfig object %s does not exist in namespace %s", PtpOperatorConfigName, PtpOperatorConfigNamespace),
		},
		{
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("ptpOperatorConfig 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testPtpOperatorConfig := buildDummyPtpOperatorConfig()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPtpOperatorConfig)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := PullPtpOperatorConfig(testSettings)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testPtpOperatorConfig.Name, testBuilder.Definition.Name)
			assert.Equal(t, testPtpOperatorConfig.Namespace, testBuilder.Definition.Namespace)
		}
	}
}

func TestPtpOperatorConfigGet(t *testing.T) {
	testCases := []struct {
		builder       *PtpOperatorConfigBuilder
		expectedError string
	}{
		{
			builder:       buildValidPtpOperatorConfigBuilder(buildTestClientWithDummyPtpOperatorConfig()),
			expectedError: "",
		},
		{
			builder:       buildValidPtpOperatorConfigBuilder(buildTestClientWithPtpScheme()),
			expectedError: fmt.Sprintf("ptpoperatorconfigs.ptp.openshift.io \"%s\" not found", PtpOperatorConfigName),
		},
		{
			builder:       buildInvalidPtpOperatorConfigBuilder(buildTestClientWithDummyPtpOperatorConfig()),
			expectedError: "test error",
		},
	}

	for _, testCase := range testCases {
		ptpOpConfig, err := testCase.builder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.NotNil(t, ptpOpConfig)
			assert.Equal(t, PtpOperatorConfigName, ptpOpConfig.Name)
			assert.Equal(t, PtpOperatorConfigNamespace, ptpOpConfig.Namespace)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestPtpOperatorConfigExists(t *testing.T) {
	testCases := []struct {
		builder *PtpOperatorConfigBuilder
		exists  bool
	}{
		{
			builder: buildValidPtpOperatorConfigBuilder(buildTestClientWithDummyPtpOperatorConfig()),
			exists:  true,
		},
		{
			builder: buildValidPtpOperatorConfigBuilder(buildTestClientWithPtpScheme()),
			exists:  false,
		},
		{
			builder: buildInvalidPtpOperatorConfigBuilder(buildTestClientWithDummyPtpOperatorConfig()),
			exists:  false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.builder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestPtpOperatorConfigUpdate(t *testing.T) {
	testCases := []struct {
		builder       *PtpOperatorConfigBuilder
		exists        bool
		expectedError error
	}{
		{
			builder:       buildValidPtpOperatorConfigBuilder(buildTestClientWithDummyPtpOperatorConfig()),
			exists:        true,
			expectedError: nil,
		},
		{
			builder:       buildValidPtpOperatorConfigBuilder(buildTestClientWithPtpScheme()),
			exists:        false,
			expectedError: fmt.Errorf("cannot update non-existent ptpOperatorConfig"),
		},
		{
			builder:       buildInvalidPtpOperatorConfigBuilder(buildTestClientWithDummyPtpOperatorConfig()),
			exists:        true,
			expectedError: fmt.Errorf("test error"),
		},
	}

	for _, testCase := range testCases {
		testCase.builder.Definition.Spec.DaemonNodeSelector = map[string]string{"updated": "true"}
		_, err := testCase.builder.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, "true", testCase.builder.Object.Spec.DaemonNodeSelector["updated"])
		}
	}
}

func TestPtpOperatorConfigWithEventConfig(t *testing.T) {
	testCases := []struct {
		eventConfig   ptpv1.PtpEventConfig
		builderValid  bool
		expectedError string
	}{
		{
			eventConfig: ptpv1.PtpEventConfig{
				TransportHost: "http://example.com:8080",
				ApiVersion:    "1.0",
			},
			builderValid:  true,
			expectedError: "",
		},
		{
			eventConfig: ptpv1.PtpEventConfig{
				TransportHost: "http://example.com:8080",
				ApiVersion:    "2.0",
			},
			builderValid:  true,
			expectedError: "",
		},
		{
			eventConfig: ptpv1.PtpEventConfig{
				TransportHost: "::::",
			},
			builderValid:  true,
			expectedError: "invalid TransportHost for PtpEventConfig: parse \"::::\": missing protocol scheme",
		},
		{
			eventConfig: ptpv1.PtpEventConfig{
				TransportHost: "http://example.com:8080",
				ApiVersion:    "3.0",
			},
			builderValid:  true,
			expectedError: "invalid ApiVersion for PtpEventConfig: must be 1.0 or start with 2.",
		},
		{
			eventConfig:   ptpv1.PtpEventConfig{},
			builderValid:  false,
			expectedError: "test error",
		},
	}

	for _, testCase := range testCases {
		var testBuilder *PtpOperatorConfigBuilder

		if testCase.builderValid {
			testBuilder = buildValidPtpOperatorConfigBuilder(buildTestClientWithPtpScheme())
		} else {
			testBuilder = buildInvalidPtpOperatorConfigBuilder(buildTestClientWithPtpScheme())
		}

		resultBuilder := testBuilder.WithEventConfig(testCase.eventConfig)
		assert.Equal(t, testCase.expectedError, resultBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, resultBuilder.Definition.Spec.EventConfig)
			assert.Equal(t, testCase.eventConfig, *resultBuilder.Definition.Spec.EventConfig)
		}
	}
}

func TestPtpOperatorConfigValidate(t *testing.T) {
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
			expectedError:   fmt.Errorf("error: received nil ptpOperatorConfig builder"),
		},
		{
			builderNil:      false,
			definitionNil:   true,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("%s", msg.UndefinedCrdObjectErrString("ptpOperatorConfig")),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    true,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("ptpOperatorConfig builder cannot have nil apiClient"),
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
		testBuilder := buildValidPtpOperatorConfigBuilder(buildTestClientWithPtpScheme())

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		if testCase.builderErrorMsg != "" {
			testBuilder.errorMsg = testCase.builderErrorMsg
		}

		valid, err := testBuilder.validate()
		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, testCase.expectedError == nil, valid)
	}
}

// buildDummyPtpOperatorConfig returns a PtpOperatorConfig with the default name and namespace.
func buildDummyPtpOperatorConfig() *ptpv1.PtpOperatorConfig {
	return &ptpv1.PtpOperatorConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PtpOperatorConfigName,
			Namespace: PtpOperatorConfigNamespace,
		},
		Spec: ptpv1.PtpOperatorConfigSpec{
			DaemonNodeSelector: make(map[string]string),
		},
	}
}

// buildTestClientWithDummyPtpOperatorConfig returns a client with a mock PtpOperatorConfig.
func buildTestClientWithDummyPtpOperatorConfig() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPtpOperatorConfig(),
		},
		SchemeAttachers: testSchemes,
	})
}

// buildInvalidPtpOperatorConfigBuilder returns a PtpOperatorConfigBuilder with an error message.
func buildInvalidPtpOperatorConfigBuilder(apiClient *clients.Settings) *PtpOperatorConfigBuilder {
	builder := buildValidPtpOperatorConfigBuilder(apiClient)
	builder.errorMsg = "test error"

	return builder
}

// buildValidPtpOperatorConfigBuilder returns a valid PtpOperatorConfigBuilder for testing.
func buildValidPtpOperatorConfigBuilder(apiClient *clients.Settings) *PtpOperatorConfigBuilder {
	return &PtpOperatorConfigBuilder{
		apiClient:  apiClient.Client,
		Definition: buildDummyPtpOperatorConfig(),
	}
}
