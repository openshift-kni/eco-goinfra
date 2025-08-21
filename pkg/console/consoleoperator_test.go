package console

import (
	"fmt"
	"testing"

	"github.com/golang/glog"
	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	consoleOperatorGVK = schema.GroupVersionKind{
		Group:   "operator.openshift.io",
		Version: "v1",
		Kind:    "Console",
	}
	defaultConsoleOperatorName = "cluster"
	defaultPluginsList         = []string{"monitoring-plugin", "nmstate-console-plugin"}
)

func TestConsoleOperatorPull(t *testing.T) {
	generateConsoleOperator := func(name string) *operatorv1.Console {
		return &operatorv1.Console{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
	}

	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                "test",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("the consoleOperator 'consoleOperatorName' cannot be empty"),
			client:              true,
		},
		{
			name:                "consoletest",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("the consoleOperator object consoletest does not exist"),
			client:              true,
		},
		{
			name:                "consoletest",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("consoleOperator 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testConsoleOperator := generateConsoleOperator(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testConsoleOperator)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := PullConsoleOperator(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.Equal(t, testConsoleOperator.Name, builderResult.Object.Name)
		}
	}
}

func TestConsoleOperatorExist(t *testing.T) {
	testCases := []struct {
		testConsoleOperator *ConsoleOperatorBuilder
		expectedStatus      bool
	}{
		{
			testConsoleOperator: buildValidConsoleOperatorBuilder(buildConsoleOperatorClientWithDummyObject()),
			expectedStatus:      true,
		},
		{
			testConsoleOperator: buildInValidConsoleOperatorBuilder(buildConsoleOperatorClientWithDummyObject()),
			expectedStatus:      false,
		},
		{
			testConsoleOperator: buildValidConsoleOperatorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:      false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testConsoleOperator.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestConsoleOperatorGet(t *testing.T) {
	testCases := []struct {
		testConsoleOperator *ConsoleOperatorBuilder
		expectedError       error
	}{
		{
			testConsoleOperator: buildValidConsoleOperatorBuilder(buildConsoleOperatorClientWithDummyObject()),
			expectedError:       nil,
		},
		{
			testConsoleOperator: buildInValidConsoleOperatorBuilder(buildConsoleOperatorClientWithDummyObject()),
			expectedError:       fmt.Errorf("the consoleOperator 'name' cannot be empty"),
		},
		{
			testConsoleOperator: buildValidConsoleOperatorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:       fmt.Errorf("consoles.operator.openshift.io \"cluster\" not found"),
		},
	}

	for _, testCase := range testCases {
		consoleOperatorObj, err := testCase.testConsoleOperator.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, consoleOperatorObj, testCase.testConsoleOperator.Definition)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestConsoleOperatorUpdate(t *testing.T) {
	testCases := []struct {
		testConsoleOperator *ConsoleOperatorBuilder
		expectedError       string
		plugins             []string
	}{
		{
			testConsoleOperator: buildValidConsoleOperatorBuilder(buildConsoleOperatorClientWithDummyObject()),
			expectedError:       "",
			plugins:             []string{"odf-console"},
		},
		{
			testConsoleOperator: buildValidConsoleOperatorBuilder(buildConsoleOperatorClientWithDummyObject()),
			expectedError:       "",
			plugins:             defaultPluginsList,
		},
		{
			testConsoleOperator: buildInValidConsoleOperatorBuilder(buildConsoleOperatorClientWithDummyObject()),
			expectedError:       "the consoleOperator 'name' cannot be empty",
			plugins:             defaultPluginsList,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, defaultPluginsList, testCase.testConsoleOperator.Definition.Spec.Plugins)
		assert.Nil(t, nil, testCase.testConsoleOperator.Object)
		testCase.testConsoleOperator.WithPlugins(testCase.plugins, true)
		_, err := testCase.testConsoleOperator.Update()

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.plugins, testCase.testConsoleOperator.Definition.Spec.Plugins)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestConsoleOperatorWithPlugins(t *testing.T) {
	testCases := []struct {
		testPlugin          []string
		testRedefine        bool
		expectedPluginsList []string
		expectedError       bool
		expectedErrorText   string
	}{
		{
			testPlugin:          []string{"test-new-plugin"},
			testRedefine:        true,
			expectedPluginsList: []string{"test-new-plugin"},
			expectedError:       false,
			expectedErrorText:   "",
		},
		{
			testPlugin:          []string{"test-new-plugin"},
			testRedefine:        false,
			expectedPluginsList: []string{"monitoring-plugin", "nmstate-console-plugin", "test-new-plugin"},
			expectedError:       false,
			expectedErrorText:   "",
		},
		{
			testPlugin:          defaultPluginsList,
			testRedefine:        false,
			expectedPluginsList: []string{"monitoring-plugin", "nmstate-console-plugin"},
			expectedError:       false,
			expectedErrorText:   "",
		},
		{
			testPlugin:          []string{},
			testRedefine:        false,
			expectedPluginsList: []string{"monitoring-plugin", "nmstate-console-plugin"},
			expectedError:       false,
			expectedErrorText:   "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidConsoleOperatorBuilder(buildConsoleOperatorClientWithDummyObject())

		result := testBuilder.WithPlugins(testCase.testPlugin, testCase.testRedefine)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.expectedPluginsList, result.Definition.Spec.Plugins)
		}
	}
}

func buildValidConsoleOperatorBuilder(apiClient *clients.Settings) *ConsoleOperatorBuilder {
	return newConsoleOperatorBuilder(apiClient, defaultConsoleOperatorName, defaultPluginsList)
}

func buildInValidConsoleOperatorBuilder(apiClient *clients.Settings) *ConsoleOperatorBuilder {
	return newConsoleOperatorBuilder(apiClient, "", defaultPluginsList)
}

func buildConsoleOperatorClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyConsoleOperator(),
		GVK:            []schema.GroupVersionKind{consoleOperatorGVK},
	})
}

func buildDummyConsoleOperator() []runtime.Object {
	return append([]runtime.Object{}, &operatorv1.Console{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultConsoleOperatorName,
		},
		Spec: operatorv1.ConsoleSpec{
			Plugins: defaultPluginsList,
		},
	})
}

// newBuilder method creates new instance of builder (for the unit test propose only).
func newConsoleOperatorBuilder(apiClient *clients.Settings, name string, pluginsList []string) *ConsoleOperatorBuilder {
	glog.V(100).Infof("Initializing new ConsoleOperatorBuilder structure with the name: %s", name)

	builder := &ConsoleOperatorBuilder{
		apiClient: apiClient.Client,
		Definition: &operatorv1.Console{
			ObjectMeta: metav1.ObjectMeta{
				Name:            name,
				ResourceVersion: "999",
			},
			Spec: operatorv1.ConsoleSpec{
				Plugins: pluginsList,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the consoleOperator is empty")

		builder.errorMsg = "the consoleOperator 'name' cannot be empty"

		return builder
	}

	if len(pluginsList) == 0 {
		glog.V(100).Infof("The pluginsList of the consoleOperator is empty")

		builder.errorMsg = "the consoleOperator 'pluginsList' cannot be empty"

		return builder
	}

	return builder
}
