package keda

import (
	"fmt"
	"testing"

	kedav1alpha1 "github.com/kedacore/keda-olm-operator/apis/keda/v1alpha1"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultKedaControllerName      = "keda"
	defaultKedaControllerNamespace = "openshift-keda"
)

func TestPullKedaController(t *testing.T) {
	generateKedaController := func(name, namespace string) *kedav1alpha1.KedaController {
		return &kedav1alpha1.KedaController{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                "test",
			namespace:           "openshift-keda",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "openshift-keda",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("kedaController 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "test",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("kedaController 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "kedacontrollertest",
			namespace:           "openshift-keda",
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf("kedaController object kedacontrollertest does not exist " +
				"in namespace openshift-keda"),
			client: true,
		},
		{
			name:                "kedacontrollertest",
			namespace:           "openshift-keda",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("kedaController 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testKedaController := generateKedaController(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testKedaController)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := Pull(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testKedaController.Name, builderResult.Object.Name)
			assert.Nil(t, err)
		}
	}
}

func TestNewKedaControllerBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
	}{
		{
			name:          defaultKedaControllerName,
			namespace:     defaultKedaControllerNamespace,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     defaultKedaControllerNamespace,
			expectedError: "kedaController 'name' cannot be empty",
		},
		{
			name:          defaultKedaControllerName,
			namespace:     "",
			expectedError: "kedaController 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testKedaControllerBuilder := NewKedaControllerBuilder(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, testKedaControllerBuilder.errorMsg)
		assert.NotNil(t, testKedaControllerBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testKedaControllerBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testKedaControllerBuilder.Definition.Namespace)
		}
	}
}

func TestKedaControllerExists(t *testing.T) {
	testCases := []struct {
		testKedaController *KedaControllerBuilder
		expectedStatus     bool
	}{
		{
			testKedaController: buildValidKedaControllerBuilder(buildKedaControllerClientWithDummyObject()),
			expectedStatus:     true,
		},
		{
			testKedaController: buildInValidKedaControllerBuilder(buildKedaControllerClientWithDummyObject()),
			expectedStatus:     false,
		},
		{
			testKedaController: buildValidKedaControllerBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:     false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testKedaController.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestKedaControllerGet(t *testing.T) {
	testCases := []struct {
		testKedaController *KedaControllerBuilder
		expectedError      error
	}{
		{
			testKedaController: buildValidKedaControllerBuilder(buildKedaControllerClientWithDummyObject()),
			expectedError:      nil,
		},
		{
			testKedaController: buildInValidKedaControllerBuilder(buildKedaControllerClientWithDummyObject()),
			expectedError:      fmt.Errorf("kedacontrollers.keda.sh \"\" not found"),
		},
		{
			testKedaController: buildValidKedaControllerBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:      fmt.Errorf("kedacontrollers.keda.sh \"keda\" not found"),
		},
	}

	for _, testCase := range testCases {
		kedaControllerObj, err := testCase.testKedaController.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, kedaControllerObj.Name, testCase.testKedaController.Definition.Name)
			assert.Equal(t, kedaControllerObj.Namespace, testCase.testKedaController.Definition.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestKedaControllerCreate(t *testing.T) {
	testCases := []struct {
		testKedaController *KedaControllerBuilder
		expectedError      string
	}{
		{
			testKedaController: buildValidKedaControllerBuilder(buildKedaControllerClientWithDummyObject()),
			expectedError:      "",
		},
		{
			testKedaController: buildInValidKedaControllerBuilder(buildKedaControllerClientWithDummyObject()),
			expectedError:      " \"\" is invalid: metadata.name: Required value: name is required",
		},
		{
			testKedaController: buildValidKedaControllerBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:      "",
		},
	}

	for _, testCase := range testCases {
		testKedaControllerBuilder, err := testCase.testKedaController.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testKedaControllerBuilder.Definition.Name, testKedaControllerBuilder.Object.Name)
			assert.Equal(t, testKedaControllerBuilder.Definition.Namespace, testKedaControllerBuilder.Object.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestKedaControllerDelete(t *testing.T) {
	testCases := []struct {
		testKedaController *KedaControllerBuilder
		expectedError      error
	}{
		{
			testKedaController: buildValidKedaControllerBuilder(buildKedaControllerClientWithDummyObject()),
			expectedError:      nil,
		},
		{
			testKedaController: buildInValidKedaControllerBuilder(buildKedaControllerClientWithDummyObject()),
			expectedError:      nil,
		},
		{
			testKedaController: buildValidKedaControllerBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:      nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testKedaController.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testKedaController.Object)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestKedaControllerUpdate(t *testing.T) {
	testCases := []struct {
		testKedaController *KedaControllerBuilder
		expectedError      string
		watchNamespace     string
	}{
		{
			testKedaController: buildValidKedaControllerBuilder(buildKedaControllerClientWithDummyObject()),
			expectedError:      "",
			watchNamespace:     "keda",
		},
		{
			testKedaController: buildInValidKedaControllerBuilder(buildKedaControllerClientWithDummyObject()),
			expectedError:      " \"\" is invalid: metadata.name: Required value: name is required",
			watchNamespace:     "",
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, "", testCase.testKedaController.Definition.Spec.WatchNamespace)
		assert.Nil(t, nil, testCase.testKedaController.Object)
		testCase.testKedaController.WithWatchNamespace(testCase.watchNamespace)
		_, err := testCase.testKedaController.Update()

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.watchNamespace, testCase.testKedaController.Definition.Spec.WatchNamespace)
		}
	}
}

func TestKedaControllerWithAdmissionWebhooks(t *testing.T) {
	testCases := []struct {
		testAdmissionWebhooks kedav1alpha1.KedaAdmissionWebhooksSpec
		expectedError         bool
		expectedErrorText     string
	}{
		{
			testAdmissionWebhooks: kedav1alpha1.KedaAdmissionWebhooksSpec{
				LogLevel:   "info",
				LogEncoder: "console",
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testAdmissionWebhooks: kedav1alpha1.KedaAdmissionWebhooksSpec{},
			expectedError:         false,
			expectedErrorText:     "'admissionWebhooks' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidKedaControllerBuilder(buildKedaControllerClientWithDummyObject())

		result := testBuilder.WithAdmissionWebhooks(testCase.testAdmissionWebhooks)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testAdmissionWebhooks, result.Definition.Spec.AdmissionWebhooks)
		}
	}
}

func TestKedaControllerWithOperator(t *testing.T) {
	testCases := []struct {
		testOperator      kedav1alpha1.KedaOperatorSpec
		expectedError     bool
		expectedErrorText string
	}{
		{
			testOperator: kedav1alpha1.KedaOperatorSpec{
				LogLevel:   "info",
				LogEncoder: "console",
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testOperator:      kedav1alpha1.KedaOperatorSpec{},
			expectedError:     false,
			expectedErrorText: "'operator' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidKedaControllerBuilder(buildKedaControllerClientWithDummyObject())

		result := testBuilder.WithOperator(testCase.testOperator)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testOperator, result.Definition.Spec.Operator)
		}
	}
}

func TestKedaControllerWithMetricsServer(t *testing.T) {
	testCases := []struct {
		testMetricsServer kedav1alpha1.KedaMetricsServerSpec
		expectedError     bool
		expectedErrorText string
	}{
		{
			testMetricsServer: kedav1alpha1.KedaMetricsServerSpec{
				LogLevel: "0",
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testMetricsServer: kedav1alpha1.KedaMetricsServerSpec{},
			expectedError:     false,
			expectedErrorText: "'metricsServer' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidKedaControllerBuilder(buildKedaControllerClientWithDummyObject())

		result := testBuilder.WithMetricsServer(testCase.testMetricsServer)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testMetricsServer, result.Definition.Spec.MetricsServer)
		}
	}
}

func TestKedaControllerWithWatchNamespace(t *testing.T) {
	testCases := []struct {
		testWatchNamespace string
		expectedErrorText  string
	}{
		{
			testWatchNamespace: "test-app",
			expectedErrorText:  "",
		},
		{
			testWatchNamespace: "",
			expectedErrorText:  "'watchNamespace' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildInValidKedaControllerBuilder(buildKedaControllerClientWithDummyObject())

		result := testBuilder.WithWatchNamespace(testCase.testWatchNamespace)

		if testCase.expectedErrorText != "" {
			assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testWatchNamespace, result.Definition.Spec.WatchNamespace)
		}
	}
}

func buildValidKedaControllerBuilder(apiClient *clients.Settings) *KedaControllerBuilder {
	kedaControllerBuilder := NewKedaControllerBuilder(
		apiClient, defaultKedaControllerName, defaultKedaControllerNamespace)

	return kedaControllerBuilder
}

func buildInValidKedaControllerBuilder(apiClient *clients.Settings) *KedaControllerBuilder {
	kedaControllerBuilder := NewKedaControllerBuilder(
		apiClient, "", defaultKedaControllerNamespace)

	return kedaControllerBuilder
}

func buildKedaControllerClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyKedaController(),
	})
}

func buildDummyKedaController() []runtime.Object {
	return append([]runtime.Object{}, &kedav1alpha1.KedaController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultKedaControllerName,
			Namespace: defaultKedaControllerNamespace,
		},
	})
}
