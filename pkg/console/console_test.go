package console

import (
	"testing"

	"fmt"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultConsoleName = "console"
)

var (
	consoleTestSchemes = []clients.SchemeAttacher{
		configv1.Install,
	}
)

func TestConsoleNewBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		client        bool
		expectedError string
	}{
		{
			name:          defaultConsoleName,
			client:        true,
			expectedError: "",
		},
		{
			name:          "",
			client:        true,
			expectedError: "console 'name' cannot be empty",
		},
		{
			name:          defaultConsoleName,
			client:        false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewBuilder(testSettings, testCase.name)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			}
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestConsolePull(t *testing.T) {
	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultConsoleName,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("console 'name' cannot be empty"),
		},
		{
			name:                defaultConsoleName,
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("console object %s does not exist", defaultConsoleName),
		},
		{
			name:                defaultConsoleName,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("console 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, buildDummyConsole(defaultConsoleName))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: consoleTestSchemes,
			})
		}

		testBuilder, err := Pull(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
		}
	}
}

func TestConsoleGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *Builder
		expectedError string
	}{
		{
			testBuilder:   buildValidConsoleTestBuilder(buildTestClientWithDummyConsole()),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidConsoleTestBuilder(buildTestClientWithDummyConsole()),
			expectedError: "console 'name' cannot be empty",
		},
		{
			testBuilder:   buildValidConsoleTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Sprintf("consoles.config.openshift.io \"%s\" not found", defaultConsoleName),
		},
	}

	for _, testCase := range testCases {
		console, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, console.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestConsoleExists(t *testing.T) {
	testCases := []struct {
		testBuilder *Builder
		exists      bool
	}{
		{
			testBuilder: buildValidConsoleTestBuilder(buildTestClientWithDummyConsole()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidConsoleTestBuilder(buildTestClientWithDummyConsole()),
			exists:      false,
		},
		{
			testBuilder: buildValidConsoleTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestConsoleCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *Builder
		expectedError error
	}{
		{
			testBuilder:   buildValidConsoleTestBuilder(buildTestClientWithDummyConsole()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidConsoleTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidConsoleTestBuilder(buildTestClientWithDummyConsole()),
			expectedError: fmt.Errorf("console 'name' cannot be empty"),
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

func TestConsoleUpdate(t *testing.T) {
	testCases := []struct {
		testBuilder   *Builder
		expectedError error
	}{
		{
			testBuilder:   buildValidConsoleTestBuilder(buildTestClientWithDummyConsole()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidConsoleTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("cannot update non-existent console"),
		},
		{
			testBuilder:   buildInvalidConsoleTestBuilder(buildTestClientWithDummyConsole()),
			expectedError: fmt.Errorf("console 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		assert.Empty(t, testCase.testBuilder.Definition.Spec.Authentication.LogoutRedirect)

		testCase.testBuilder.Definition.Spec.Authentication.LogoutRedirect = "test"
		testCase.testBuilder.Definition.ResourceVersion = "999"

		testBuilder, err := testCase.testBuilder.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, "test", testBuilder.Object.Spec.Authentication.LogoutRedirect)
		}
	}
}

func TestConsoleDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *Builder
		expectedError error
	}{
		{
			testBuilder:   buildValidConsoleTestBuilder(buildTestClientWithDummyConsole()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidConsoleTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidConsoleTestBuilder(buildTestClientWithDummyConsole()),
			expectedError: fmt.Errorf("console 'name' cannot be empty"),
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

func TestConsoleValidate(t *testing.T) {
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
			expectedError:   fmt.Errorf("error: received nil console builder"),
		},
		{
			builderNil:      false,
			definitionNil:   true,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("can not redefine the undefined console"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    true,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("console builder cannot have nil apiClient"),
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
		consoleBuilder := buildValidConsoleTestBuilder(buildTestClientWithDummyConsole())

		if testCase.builderNil {
			consoleBuilder = nil
		}

		if testCase.definitionNil {
			consoleBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			consoleBuilder.apiClient = nil
		}

		if testCase.builderErrorMsg != "" {
			consoleBuilder.errorMsg = testCase.builderErrorMsg
		}

		valid, err := consoleBuilder.validate()
		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, testCase.expectedError == nil, valid)
	}
}

// buildDummyConsole returns a dummy console object using the given name.
func buildDummyConsole(name string) *configv1.Console {
	return &configv1.Console{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestClientWithDummyConsole returns a test client with a dummy console object.
func buildTestClientWithDummyConsole() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyConsole(defaultConsoleName),
		},
		SchemeAttachers: consoleTestSchemes,
	})
}

// buildValidConsoleTestBuilder returns a valid console test builder.
func buildValidConsoleTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultConsoleName)
}

// buildInvalidConsoleTestBuilder returns an invalid console test builder with an empty name.
func buildInvalidConsoleTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "")
}
