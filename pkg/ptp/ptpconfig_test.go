package ptp

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	ptpv1 "github.com/openshift/ptp-operator/api/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultPtpConfigName      = "test-ptp-config"
	defaultPtpConfigNamespace = "test-ns"
)

var testSchemes = []clients.SchemeAttacher{
	ptpv1.AddToScheme,
}

func TestNewPtpConfigBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		nsname        string
		client        bool
		expectedError string
	}{
		{
			name:          defaultPtpConfigName,
			nsname:        defaultPtpConfigNamespace,
			client:        true,
			expectedError: "",
		},
		{
			name:          "",
			nsname:        defaultPtpConfigNamespace,
			client:        true,
			expectedError: "ptpConfig 'name' cannot be empty",
		},
		{
			name:          defaultPtpConfigName,
			nsname:        "",
			client:        true,
			expectedError: "ptpConfig 'nsname' cannot be empty",
		},
		{
			name:          defaultPtpConfigName,
			nsname:        defaultPtpConfigNamespace,
			client:        false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithPtpScheme()
		}

		testBuilder := NewPtpConfigBuilder(testSettings, testCase.name, testCase.nsname)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
				assert.Equal(t, testCase.nsname, testBuilder.Definition.Namespace)
			}
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestPullPtpConfig(t *testing.T) {
	testCases := []struct {
		name                string
		nsname              string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultPtpConfigName,
			nsname:              defaultPtpConfigNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			nsname:              defaultPtpConfigNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("ptpConfig 'name' cannot be empty"),
		},
		{
			name:                defaultPtpConfigName,
			nsname:              "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("ptpConfig 'nsname' cannot be empty"),
		},
		{
			name:                defaultPtpConfigName,
			nsname:              defaultPtpConfigNamespace,
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf(
				"ptpConfig object %s does not exist in namespace %s", defaultPtpConfigName, defaultPtpConfigNamespace),
		},
		{
			name:                defaultPtpConfigName,
			nsname:              defaultPtpConfigNamespace,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("ptpConfig 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testPtpConfig := buildDummyPtpConfig(testCase.name, testCase.nsname)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPtpConfig)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := PullPtpConfig(testSettings, testCase.name, testCase.nsname)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testPtpConfig.Name, testBuilder.Definition.Name)
			assert.Equal(t, testPtpConfig.Namespace, testBuilder.Definition.Namespace)
		}
	}
}

func TestPtpConfigGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *PtpConfigBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidPtpConfigBuilder(buildTestClientWithDummyPtpConfig()),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidPtpConfigBuilder(buildTestClientWithDummyPtpConfig()),
			expectedError: "ptpConfig 'nsname' cannot be empty",
		},
		{
			testBuilder:   buildValidPtpConfigBuilder(buildTestClientWithPtpScheme()),
			expectedError: "ptpconfigs.ptp.openshift.io \"test-ptp-config\" not found",
		},
	}

	for _, testCase := range testCases {
		ptpConfig, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, ptpConfig.Name)
			assert.Equal(t, testCase.testBuilder.Definition.Namespace, ptpConfig.Namespace)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestPtpConfigExists(t *testing.T) {
	testCases := []struct {
		testBuilder *PtpConfigBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidPtpConfigBuilder(buildTestClientWithDummyPtpConfig()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidPtpConfigBuilder(buildTestClientWithDummyPtpConfig()),
			exists:      false,
		},
		{
			testBuilder: buildValidPtpConfigBuilder(buildTestClientWithPtpScheme()),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestPtpConfigCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *PtpConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPtpConfigBuilder(buildTestClientWithPtpScheme()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidPtpConfigBuilder(buildTestClientWithDummyPtpConfig()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPtpConfigBuilder(buildTestClientWithPtpScheme()),
			expectedError: fmt.Errorf("ptpConfig 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testBuilder.Definition.Name, testBuilder.Object.Name)
			assert.Equal(t, testBuilder.Definition.Namespace, testBuilder.Object.Namespace)
		}
	}
}

func TestPtpConfigUpdate(t *testing.T) {
	testCases := []struct {
		alreadyExists bool
		expectedError error
	}{
		{
			alreadyExists: false,
			expectedError: fmt.Errorf("cannot update non-existent ptpConfig"),
		},
		{
			alreadyExists: true,
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithPtpScheme()

		if testCase.alreadyExists {
			testSettings = buildTestClientWithDummyPtpConfig()
		}

		testBuilder := buildValidPtpConfigBuilder(testSettings)

		assert.NotNil(t, testBuilder.Definition)
		assert.Empty(t, testBuilder.Definition.Spec.Profile)

		testBuilder.Definition.Spec.Profile = []ptpv1.PtpProfile{{}}

		testBuilder, err := testBuilder.Update(false)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotEmpty(t, testBuilder.Object.Spec.Profile)
		}
	}
}

func TestPtpConfigDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *PtpConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPtpConfigBuilder(buildTestClientWithDummyPtpConfig()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidPtpConfigBuilder(buildTestClientWithPtpScheme()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPtpConfigBuilder(buildTestClientWithDummyPtpConfig()),
			expectedError: fmt.Errorf("ptpConfig 'nsname' cannot be empty"),
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

func TestPtpConfigValidate(t *testing.T) {
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
			expectedError:   fmt.Errorf("error: received nil ptpConfig builder"),
		},
		{
			builderNil:      false,
			definitionNil:   true,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("can not redefine the undefined ptpConfig"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    true,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("ptpConfig builder cannot have nil apiClient"),
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
		testBuilder := buildValidPtpConfigBuilder(buildTestClientWithPtpScheme())

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

// buildDummyPtpConfig returns a PtpConfig with the provided name and namespace.
func buildDummyPtpConfig(name, namespace string) *ptpv1.PtpConfig {
	return &ptpv1.PtpConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

// buildTestClientWithDummyPtpConfig returns a client with a mock PtpConfig.
func buildTestClientWithDummyPtpConfig() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPtpConfig(defaultPtpConfigName, defaultPtpConfigNamespace),
		},
		SchemeAttachers: testSchemes,
	})
}

// buildTestClientWithPtpScheme returns a client with no objects but the ptp v1 scheme attached.
func buildTestClientWithPtpScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: testSchemes,
	})
}

// buildValidPtpConfigBuilder returns a valid PtpConfigBuilder for testing.
func buildValidPtpConfigBuilder(apiClient *clients.Settings) *PtpConfigBuilder {
	return NewPtpConfigBuilder(apiClient, defaultPtpConfigName, defaultPtpConfigNamespace)
}

// buildInvalidPtpConfigBuilder returns an invalid PtpConfigBuilder for testing.
func buildInvalidPtpConfigBuilder(apiClient *clients.Settings) *PtpConfigBuilder {
	return NewPtpConfigBuilder(apiClient, defaultPtpConfigName, "")
}
