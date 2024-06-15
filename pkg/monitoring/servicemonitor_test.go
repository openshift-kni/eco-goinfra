package monitoring

import (
	"fmt"
	"testing"

	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultServiceMonitorName      = "test-monitor-name"
	defaultServiceMonitorNamespace = "test-monitor-namespace"
)

func TestPullServiceMonitor(t *testing.T) {
	generateServiceMonitor := func(name, namespace string) *monv1.ServiceMonitor {
		return &monv1.ServiceMonitor{
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
			name:                defaultServiceMonitorName,
			namespace:           defaultServiceMonitorNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           defaultServiceMonitorNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("serviceMonitor 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                defaultServiceMonitorName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("serviceMonitor 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "mon-test",
			namespace:           defaultServiceMonitorNamespace,
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf("serviceMonitor object mon-test does not exist " +
				"in namespace test-monitor-namespace"),
			client: true,
		},
		{
			name:                "mon-test",
			namespace:           defaultServiceMonitorNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("serviceMonitor 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testServiceMonitor := generateServiceMonitor(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testServiceMonitor)
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
			assert.Equal(t, testServiceMonitor.Name, builderResult.Object.Name)
			assert.Equal(t, testServiceMonitor.Namespace, builderResult.Object.Namespace)
			assert.Nil(t, err)
		}
	}
}

func TestNewServiceMonitorBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
		client        bool
	}{
		{
			name:          defaultServiceMonitorName,
			namespace:     defaultServiceMonitorNamespace,
			expectedError: "",
			client:        true,
		},
		{
			name:          "",
			namespace:     defaultServiceMonitorNamespace,
			expectedError: "serviceMonitor 'name' cannot be empty",
			client:        true,
		},
		{
			name:          defaultServiceMonitorName,
			namespace:     "",
			expectedError: "serviceMonitor 'nsname' cannot be empty",
			client:        true,
		},
		{
			name:          defaultServiceMonitorName,
			namespace:     defaultServiceMonitorNamespace,
			expectedError: "",
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings
		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testServiceMonitorBuilder := NewBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedError == "" {
			if testCase.client {
				assert.Equal(t, testCase.name, testServiceMonitorBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testServiceMonitorBuilder.Definition.Namespace)
			} else {
				assert.Nil(t, testServiceMonitorBuilder)
			}
		} else {
			assert.Equal(t, testCase.expectedError, testServiceMonitorBuilder.errorMsg)
			assert.NotNil(t, testServiceMonitorBuilder.Definition)
		}
	}
}

func TestServiceMonitorExists(t *testing.T) {
	testCases := []struct {
		testServiceMonitor *Builder
		expectedStatus     bool
	}{
		{
			testServiceMonitor: buildValidServiceMonitorBuilder(buildServiceMonitorClientWithDummyObject()),
			expectedStatus:     true,
		},
		{
			testServiceMonitor: buildInValidServiceMonitorBuilder(buildServiceMonitorClientWithDummyObject()),
			expectedStatus:     false,
		},
		{
			testServiceMonitor: buildValidServiceMonitorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:     false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testServiceMonitor.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestServiceMonitorGet(t *testing.T) {
	testCases := []struct {
		testServiceMonitor *Builder
		expectedError      error
	}{
		{
			testServiceMonitor: buildValidServiceMonitorBuilder(buildServiceMonitorClientWithDummyObject()),
			expectedError:      nil,
		},
		{
			testServiceMonitor: buildInValidServiceMonitorBuilder(buildServiceMonitorClientWithDummyObject()),
			expectedError:      fmt.Errorf("serviceMonitor 'name' cannot be empty"),
		},
		{
			testServiceMonitor: buildValidServiceMonitorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:      fmt.Errorf("servicemonitors.monitoring.coreos.com \"test-monitor-name\" not found"),
		},
	}

	for _, testCase := range testCases {
		serviceMonitorObj, err := testCase.testServiceMonitor.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, serviceMonitorObj.Name, testCase.testServiceMonitor.Definition.Name)
			assert.Equal(t, serviceMonitorObj.Namespace, testCase.testServiceMonitor.Definition.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestServiceMonitorCreate(t *testing.T) {
	testCases := []struct {
		testServiceMonitor *Builder
		expectedError      string
	}{
		{
			testServiceMonitor: buildValidServiceMonitorBuilder(buildServiceMonitorClientWithDummyObject()),
			expectedError:      "",
		},
		{
			testServiceMonitor: buildInValidServiceMonitorBuilder(buildServiceMonitorClientWithDummyObject()),
			expectedError:      "serviceMonitor 'name' cannot be empty",
		},
		{
			testServiceMonitor: buildValidServiceMonitorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:      "",
		},
	}

	for _, testCase := range testCases {
		serviceMonitorObj, err := testCase.testServiceMonitor.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, serviceMonitorObj.Definition.Name, serviceMonitorObj.Object.Name)
			assert.Equal(t, serviceMonitorObj.Definition.Namespace, serviceMonitorObj.Object.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestServiceMonitorDelete(t *testing.T) {
	testCases := []struct {
		testServiceMonitor *Builder
		expectedError      error
	}{
		{
			testServiceMonitor: buildValidServiceMonitorBuilder(buildServiceMonitorClientWithDummyObject()),
			expectedError:      nil,
		},
		{
			testServiceMonitor: buildValidServiceMonitorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:      nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testServiceMonitor.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testServiceMonitor.Object)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestServiceMonitorUpdate(t *testing.T) {
	testCases := []struct {
		testServiceMonitor *Builder
		expectedError      string
		testLabels         map[string]string
	}{
		{
			testServiceMonitor: buildValidServiceMonitorBuilder(buildServiceMonitorClientWithDummyObject()),
			testLabels: map[string]string{"first-test-label-key": "first-test-label-value",
				"second-test-label-key": ""},
			expectedError: "",
		},
		{
			testServiceMonitor: buildValidServiceMonitorBuilder(buildServiceMonitorClientWithDummyObject()),
			testLabels:         map[string]string{},
			expectedError:      "labels can not be empty",
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, map[string]string(nil), testCase.testServiceMonitor.Definition.Labels)
		testCase.testServiceMonitor.WithLabels(testCase.testLabels)
		_, err := testCase.testServiceMonitor.Update()

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.testLabels,
				testCase.testServiceMonitor.Definition.Labels)
		}
	}
}

func TestServiceMonitorWithEndpoints(t *testing.T) {
	testCases := []struct {
		testEndpoints     []monv1.Endpoint
		expectedErrorText string
	}{
		{
			testEndpoints: []monv1.Endpoint{{
				Port:   "http",
				Scheme: "http",
			}},
			expectedErrorText: "",
		},
		{
			testEndpoints: []monv1.Endpoint{{
				Port:   "http",
				Scheme: "http",
			},
				{
					Port:   "sctp",
					Scheme: "sctp",
				}},
			expectedErrorText: "",
		},
		{
			testEndpoints:     []monv1.Endpoint{},
			expectedErrorText: "'endpoints' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidServiceMonitorBuilder(buildServiceMonitorClientWithDummyObject())

		result := testBuilder.WithEndpoints(testCase.testEndpoints)

		if testCase.expectedErrorText != "" {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testEndpoints, result.Definition.Spec.Endpoints)
		}
	}
}

func TestServiceMonitorWithLabel(t *testing.T) {
	testCases := []struct {
		testLabel      map[string]string
		expectedErrMsg string
	}{
		{
			testLabel:      map[string]string{"test-label-key": "test-label-value"},
			expectedErrMsg: "",
		},
		{
			testLabel: map[string]string{"test-label1-key": "test-label1-value",
				"test-label2-key": "test-label2-value",
				"test-label3-key": "test-label3-value",
			},
			expectedErrMsg: "",
		},
		{
			testLabel:      map[string]string{"test-label-key": ""},
			expectedErrMsg: "",
		},
		{
			testLabel:      map[string]string{"": "test-label-value"},
			expectedErrMsg: "can not apply a labels with an empty key",
		},
		{
			testLabel:      map[string]string{},
			expectedErrMsg: "labels can not be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidServiceMonitorBuilder(buildServiceMonitorClientWithDummyObject())

		testBuilder.WithLabels(testCase.testLabel)

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.testLabel, testBuilder.Definition.Labels)
		}
	}
}

func TestServiceMonitorWithSelector(t *testing.T) {
	testCases := []struct {
		testSelector   map[string]string
		expectedErrMsg string
	}{
		{
			testSelector:   map[string]string{"test-selector-key": "test-selector-value"},
			expectedErrMsg: "",
		},
		{
			testSelector:   map[string]string{"test-selector-key": ""},
			expectedErrMsg: "",
		},
		{
			testSelector:   map[string]string{"": "test-selector-value"},
			expectedErrMsg: "can not apply a selector with an empty key",
		},
		{
			testSelector:   map[string]string{},
			expectedErrMsg: "selector can not be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidServiceMonitorBuilder(buildServiceMonitorClientWithDummyObject())

		testBuilder.WithSelector(testCase.testSelector)

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.testSelector, testBuilder.Definition.Spec.Selector.MatchLabels)
		}
	}
}

func TestServiceMonitorWithNamespaceSelector(t *testing.T) {
	testCases := []struct {
		testNamespaceSelector []string
		expectedErrMsg        string
	}{
		{
			testNamespaceSelector: []string{"test-ns-selector"},
			expectedErrMsg:        "",
		},
		{
			testNamespaceSelector: []string{"test-ns-selector1", "test-ns-selector2"},
			expectedErrMsg:        "",
		},
		{
			testNamespaceSelector: []string{},
			expectedErrMsg:        "namespaceSelector can not be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidServiceMonitorBuilder(buildServiceMonitorClientWithDummyObject())

		testBuilder.WithNamespaceSelector(testCase.testNamespaceSelector)

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.testNamespaceSelector, testBuilder.Definition.Spec.NamespaceSelector.MatchNames)
		}
	}
}

func buildValidServiceMonitorBuilder(apiClient *clients.Settings) *Builder {
	serviceMonitorBuilder := NewBuilder(
		apiClient, defaultServiceMonitorName, defaultServiceMonitorNamespace)

	return serviceMonitorBuilder
}

func buildInValidServiceMonitorBuilder(apiClient *clients.Settings) *Builder {
	serviceMonitorBuilder := NewBuilder(
		apiClient, "", defaultServiceMonitorNamespace)

	return serviceMonitorBuilder
}

func buildServiceMonitorClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyServiceMonitor(),
	})
}

func buildDummyServiceMonitor() []runtime.Object {
	return append([]runtime.Object{}, &monv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultServiceMonitorName,
			Namespace: defaultServiceMonitorNamespace,
		},
	})
}
