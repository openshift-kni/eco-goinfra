package servicemesh

import (
	"fmt"
	"testing"

	istiov2 "maistra.io/api/core/v2"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultControlPlaneName = "basic"
	dummyIPAddress          = "10.22.22.1"
	emptyIPAddress          = ""
	emptyAddonsConfig       = (*istiov2.AddonsConfig)(nil)
	isTrue                  = true
	isFalse                 = false
	istiov2TestSchemes      = []clients.SchemeAttacher{
		istiov2.AddToScheme,
	}
)

func TestPullControlPlane(t *testing.T) {
	generateControlPlane := func(name, namespace string) *istiov2.ServiceMeshControlPlane {
		return &istiov2.ServiceMeshControlPlane{
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
			namespace:           "istio-system",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "istio-system",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("serviceMeshControlPlane 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "test",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("serviceMeshControlPlane 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "smocptest",
			namespace:           "istio-system",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("serviceMeshControlPlane object smocptest does not exist in namespace istio-system"),
			client:              true,
		},
		{
			name:                "smocptest",
			namespace:           "istio-system",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("serviceMeshControlPlane 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testControlPlane := generateControlPlane(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testControlPlane)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: istiov2TestSchemes,
			})
		}

		builderResult, err := PullControlPlane(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testControlPlane.Name, builderResult.Object.Name)
			assert.Equal(t, testControlPlane.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestNewControlPlaneBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
	}{
		{
			name:          defaultControlPlaneName,
			namespace:     defaultServiceMeshNamespace,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     defaultServiceMeshNamespace,
			expectedError: "serviceMeshControlPlane 'name' cannot be empty",
		},
		{
			name:          defaultControlPlaneName,
			namespace:     "",
			expectedError: "serviceMeshControlPlane 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testControlPlaneBuilder := NewControlPlaneBuilder(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, testControlPlaneBuilder.errorMsg)
		assert.NotNil(t, testControlPlaneBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testControlPlaneBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testControlPlaneBuilder.Definition.Namespace)
		}
	}
}

func TestControlPlaneExists(t *testing.T) {
	testCases := []struct {
		testControlPlane *ControlPlaneBuilder
		expectedStatus   bool
	}{
		{
			testControlPlane: buildValidControlPlaneBuilder(buildControlPlaneClientWithDummyObject()),
			expectedStatus:   true,
		},
		{
			testControlPlane: buildInValidControlPlaneBuilder(buildControlPlaneClientWithDummyObject()),
			expectedStatus:   false,
		},
		{
			testControlPlane: buildValidControlPlaneBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:   false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testControlPlane.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestControlPlaneGet(t *testing.T) {
	testCases := []struct {
		testControlPlane *ControlPlaneBuilder
		expectedError    error
	}{
		{
			testControlPlane: buildValidControlPlaneBuilder(buildControlPlaneClientWithDummyObject()),
			expectedError:    nil,
		},
		{
			testControlPlane: buildInValidControlPlaneBuilder(buildControlPlaneClientWithDummyObject()),
			expectedError:    fmt.Errorf("serviceMeshControlPlane 'name' cannot be empty"),
		},
		{
			testControlPlane: buildValidControlPlaneBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:    fmt.Errorf("servicemeshcontrolplanes.maistra.io \"basic\" not found"),
		},
	}

	for _, testCase := range testCases {
		controlPlaneObj, err := testCase.testControlPlane.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.testControlPlane.Definition, controlPlaneObj)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestControlPlaneCreate(t *testing.T) {
	testCases := []struct {
		testControlPlane *ControlPlaneBuilder
		expectedError    string
	}{
		{
			testControlPlane: buildValidControlPlaneBuilder(buildControlPlaneClientWithDummyObject()),
			expectedError:    "",
		},
		{
			testControlPlane: buildInValidControlPlaneBuilder(buildControlPlaneClientWithDummyObject()),
			expectedError:    "serviceMeshControlPlane 'name' cannot be empty",
		},
		{
			testControlPlane: buildValidControlPlaneBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:    "resourceVersion can not be set for Create requests",
		},
	}

	for _, testCase := range testCases {
		testControlPlaneBuilder, err := testCase.testControlPlane.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testControlPlaneBuilder.Definition, testControlPlaneBuilder.Object)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestControlPlaneDelete(t *testing.T) {
	testCases := []struct {
		testControlPlane *ControlPlaneBuilder
		expectedError    error
	}{
		{
			testControlPlane: buildValidControlPlaneBuilder(buildControlPlaneClientWithDummyObject()),
			expectedError:    nil,
		},
		{
			testControlPlane: buildInValidControlPlaneBuilder(buildControlPlaneClientWithDummyObject()),
			expectedError:    fmt.Errorf("serviceMeshControlPlane 'name' cannot be empty"),
		},
		{
			testControlPlane: buildValidControlPlaneBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:    nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testControlPlane.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testControlPlane.Object)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestControlPlaneUpdate(t *testing.T) {
	testCases := []struct {
		testControlPlane *ControlPlaneBuilder
		expectedError    string
		addonEnablement  bool
	}{
		{
			testControlPlane: buildValidControlPlaneBuilder(buildControlPlaneClientWithDummyObject()),
			expectedError:    "",
			addonEnablement:  false,
		},
		{
			testControlPlane: buildInValidControlPlaneBuilder(buildControlPlaneClientWithDummyObject()),
			expectedError:    "serviceMeshControlPlane 'name' cannot be empty",
			addonEnablement:  false,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, emptyAddonsConfig, testCase.testControlPlane.Definition.Spec.Addons)
		assert.Nil(t, nil, testCase.testControlPlane.Object)
		testCase.testControlPlane.WithGrafanaAddon(testCase.addonEnablement, &istiov2.GrafanaInstallConfig{}, "")
		_, err := testCase.testControlPlane.Update(true)

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.addonEnablement,
				*testCase.testControlPlane.Definition.Spec.Addons.Grafana.Enablement.Enabled)
		}
	}
}

func TestControlPlaneWithAllAddonsDisabled(t *testing.T) {
	testCases := []struct {
		testControlPlane  bool
		expectedError     bool
		expectedErrorText string
	}{
		{
			testControlPlane:  false,
			expectedError:     false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidControlPlaneBuilder(buildControlPlaneClientWithDummyObject())

		result := testBuilder.WithAllAddonsDisabled()

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testControlPlane, *result.Definition.Spec.Addons.Prometheus.Enablement.Enabled)
			assert.Equal(t, testCase.testControlPlane, *result.Definition.Spec.Addons.Grafana.Enablement.Enabled)
			assert.Equal(t, testCase.testControlPlane, *result.Definition.Spec.Addons.Kiali.Enablement.Enabled)
			assert.Equal(t, testCase.testControlPlane, *result.Definition.Spec.Addons.ThreeScale.Enablement.Enabled)
		}
	}
}

func TestControlPlaneWithGrafanaAddon(t *testing.T) {
	testCases := []struct {
		testEnablement    bool
		testInstall       *istiov2.GrafanaInstallConfig
		testAddress       string
		expectedError     bool
		expectedErrorText string
	}{
		{
			testEnablement: true,
			testInstall: &istiov2.GrafanaInstallConfig{
				SelfManaged: false,
			},
			testAddress:       dummyIPAddress,
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testEnablement: false,
			testInstall: &istiov2.GrafanaInstallConfig{
				SelfManaged: false,
			},
			testAddress:       dummyIPAddress,
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testEnablement:    true,
			testInstall:       nil,
			testAddress:       dummyIPAddress,
			expectedError:     true,
			expectedErrorText: "the Grafana addon 'grafanaInstallConfig' cannot be empty when Grafana addon is enabled",
		},
		{
			testEnablement:    false,
			testInstall:       nil,
			testAddress:       dummyIPAddress,
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testEnablement: false,
			testInstall: &istiov2.GrafanaInstallConfig{
				SelfManaged: false,
			},
			testAddress:       emptyIPAddress,
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testEnablement: true,
			testInstall: &istiov2.GrafanaInstallConfig{
				SelfManaged: false,
			},
			testAddress:       emptyIPAddress,
			expectedError:     true,
			expectedErrorText: "the Grafana addon 'address' cannot be empty when Grafana addon is enabled",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidControlPlaneBuilder(buildControlPlaneClientWithDummyObject())

		result := testBuilder.WithGrafanaAddon(testCase.testEnablement,
			testCase.testInstall, testCase.testAddress)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testEnablement, *result.Definition.Spec.Addons.Grafana.Enablement.Enabled)

			if testCase.testEnablement {
				assert.Equal(t, testCase.testInstall, result.Definition.Spec.Addons.Grafana.Install)
				assert.Equal(t, testCase.testAddress, *result.Definition.Spec.Addons.Grafana.Address)
			}
		}
	}
}

func TestControlPlaneWithJaegerAddon(t *testing.T) {
	testCases := []struct {
		testName          string
		testInstall       *istiov2.JaegerInstallConfig
		expectedError     bool
		expectedErrorText string
	}{
		{
			testName: "jaegertest",
			testInstall: &istiov2.JaegerInstallConfig{
				Ingress: &istiov2.JaegerIngressConfig{
					Enablement: istiov2.Enablement{
						Enabled: &isTrue,
					},
				},
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testName: "jaegertest",
			testInstall: &istiov2.JaegerInstallConfig{
				Ingress: &istiov2.JaegerIngressConfig{
					Enablement: istiov2.Enablement{
						Enabled: &isFalse,
					},
				},
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testName: "",
			testInstall: &istiov2.JaegerInstallConfig{
				Ingress: &istiov2.JaegerIngressConfig{
					Enablement: istiov2.Enablement{
						Enabled: &isTrue,
					},
				},
			},
			expectedError:     true,
			expectedErrorText: "the Jaeger addon 'name' cannot be empty",
		},
		{
			testName:          "jaegertest",
			testInstall:       nil,
			expectedError:     true,
			expectedErrorText: "the Jaeger addon 'jaegerInstallConfig' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidControlPlaneBuilder(buildControlPlaneClientWithDummyObject())

		result := testBuilder.WithJaegerAddon(testCase.testName, testCase.testInstall)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testName, result.Definition.Spec.Addons.Jaeger.Name)
			assert.Equal(t, testCase.testInstall, result.Definition.Spec.Addons.Jaeger.Install)
		}
	}
}

//nolint:funlen
func TestControlPlaneWithKialiAddon(t *testing.T) {
	testCases := []struct {
		testEnablement    bool
		testName          string
		testInstall       *istiov2.KialiInstallConfig
		expectedError     bool
		expectedErrorText string
	}{
		{
			testEnablement: true,
			testName:       "testkiali",
			testInstall: &istiov2.KialiInstallConfig{
				Dashboard: &istiov2.KialiDashboardConfig{
					ViewOnly:         &isTrue,
					EnableGrafana:    &isTrue,
					EnablePrometheus: &isTrue,
					EnableTracing:    &isTrue,
				},
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testEnablement: false,
			testName:       "testkiali",
			testInstall: &istiov2.KialiInstallConfig{
				Dashboard: &istiov2.KialiDashboardConfig{
					ViewOnly:         &isTrue,
					EnableGrafana:    &isTrue,
					EnablePrometheus: &isTrue,
					EnableTracing:    &isTrue,
				},
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testEnablement: true,
			testName:       "testkiali",
			testInstall: &istiov2.KialiInstallConfig{
				Dashboard: &istiov2.KialiDashboardConfig{
					ViewOnly:         &isFalse,
					EnableGrafana:    &isFalse,
					EnablePrometheus: &isFalse,
					EnableTracing:    &isFalse,
				},
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testEnablement:    false,
			testName:          "testkiali",
			testInstall:       nil,
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testEnablement: false,
			testName:       "",
			testInstall: &istiov2.KialiInstallConfig{
				Dashboard: &istiov2.KialiDashboardConfig{
					ViewOnly:         &isFalse,
					EnableGrafana:    &isFalse,
					EnablePrometheus: &isFalse,
					EnableTracing:    &isFalse,
				},
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testEnablement: true,
			testName:       "",
			testInstall: &istiov2.KialiInstallConfig{
				Dashboard: &istiov2.KialiDashboardConfig{
					ViewOnly:         &isFalse,
					EnableGrafana:    &isFalse,
					EnablePrometheus: &isFalse,
					EnableTracing:    &isFalse,
				},
			},
			expectedError:     true,
			expectedErrorText: "the Kiali addon 'name' cannot be empty when Kiali addon is enabled",
		},
		{
			testEnablement:    true,
			testName:          "testkiali",
			testInstall:       nil,
			expectedError:     true,
			expectedErrorText: "the Kiali addon 'kialiInstallConfig' cannot be empty when Kiali addon is enabled",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidControlPlaneBuilder(buildControlPlaneClientWithDummyObject())

		result := testBuilder.WithKialiAddon(testCase.testEnablement,
			testCase.testName, testCase.testInstall)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testEnablement, *result.Definition.Spec.Addons.Kiali.Enablement.Enabled)

			if testCase.testEnablement {
				assert.Equal(t, testCase.testInstall, result.Definition.Spec.Addons.Kiali.Install)
				assert.Equal(t, testCase.testName, result.Definition.Spec.Addons.Kiali.Name)
			}
		}
	}
}

//nolint:funlen
func TestControlPlaneWithPrometheusAddon(t *testing.T) {
	testCases := []struct {
		testEnablement            bool
		testScrape                bool
		testMetricsExpiryDuration string
		testAddress               string
		testInstall               *istiov2.PrometheusInstallConfig
		expectedError             bool
		expectedErrorText         string
	}{
		{
			testEnablement:            true,
			testScrape:                true,
			testMetricsExpiryDuration: "100",
			testAddress:               dummyIPAddress,
			testInstall: &istiov2.PrometheusInstallConfig{
				SelfManaged:    false,
				ScrapeInterval: "5",
				UseTLS:         &isFalse,
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testEnablement:            false,
			testScrape:                true,
			testMetricsExpiryDuration: "100",
			testAddress:               dummyIPAddress,
			testInstall: &istiov2.PrometheusInstallConfig{
				SelfManaged:    false,
				ScrapeInterval: "5",
				UseTLS:         &isFalse,
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testEnablement:            true,
			testScrape:                false,
			testMetricsExpiryDuration: "100",
			testAddress:               dummyIPAddress,
			testInstall: &istiov2.PrometheusInstallConfig{
				SelfManaged:    false,
				ScrapeInterval: "5",
				UseTLS:         &isFalse,
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testEnablement:            false,
			testScrape:                true,
			testMetricsExpiryDuration: "",
			testAddress:               dummyIPAddress,
			testInstall: &istiov2.PrometheusInstallConfig{
				SelfManaged:    false,
				ScrapeInterval: "5",
				UseTLS:         &isFalse,
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testEnablement:            false,
			testScrape:                true,
			testMetricsExpiryDuration: "100",
			testAddress:               emptyIPAddress,
			testInstall: &istiov2.PrometheusInstallConfig{
				SelfManaged:    false,
				ScrapeInterval: "5",
				UseTLS:         &isFalse,
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testEnablement:            false,
			testScrape:                true,
			testMetricsExpiryDuration: "100",
			testAddress:               dummyIPAddress,
			testInstall:               nil,
			expectedError:             false,
			expectedErrorText:         "",
		},
		{
			testEnablement:            true,
			testScrape:                true,
			testMetricsExpiryDuration: "",
			testAddress:               dummyIPAddress,
			testInstall: &istiov2.PrometheusInstallConfig{
				SelfManaged:    false,
				ScrapeInterval: "5",
				UseTLS:         &isFalse,
			},
			expectedError:     true,
			expectedErrorText: "the Prometheus addon 'metricsExpiryDuration' cannot be empty when Prometheus addon is enabled",
		},
		{
			testEnablement:            true,
			testScrape:                true,
			testMetricsExpiryDuration: "100",
			testAddress:               emptyIPAddress,
			testInstall: &istiov2.PrometheusInstallConfig{
				SelfManaged:    false,
				ScrapeInterval: "5",
				UseTLS:         &isFalse,
			},
			expectedError:     true,
			expectedErrorText: "the Prometheus addon 'address' cannot be empty when Prometheus addon is enabled",
		},
		{
			testEnablement:            true,
			testScrape:                true,
			testMetricsExpiryDuration: "100",
			testAddress:               dummyIPAddress,
			testInstall:               nil,
			expectedError:             true,
			expectedErrorText: "the Prometheus addon 'prometheusInstallConfig' cannot be empty " +
				"when Prometheus addon is enabled",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidControlPlaneBuilder(buildControlPlaneClientWithDummyObject())

		result := testBuilder.WithPrometheusAddon(testCase.testEnablement, testCase.testScrape,
			testCase.testMetricsExpiryDuration, testCase.testAddress, testCase.testInstall)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testEnablement, *result.Definition.Spec.Addons.Prometheus.Enablement.Enabled)

			if testCase.testEnablement {
				assert.Equal(t, testCase.testScrape, *result.Definition.Spec.Addons.Prometheus.Scrape)
				assert.Equal(t, testCase.testMetricsExpiryDuration,
					result.Definition.Spec.Addons.Prometheus.MetricsExpiryDuration)
				assert.Equal(t, testCase.testAddress, *result.Definition.Spec.Addons.Prometheus.Address)
				assert.Equal(t, testCase.testInstall, result.Definition.Spec.Addons.Prometheus.Install)
			}
		}
	}
}

func TestControlPlaneWithGatewaysEnablement(t *testing.T) {
	testCases := []struct {
		testEnablement    bool
		expectedError     bool
		expectedErrorText string
	}{
		{
			testEnablement:    true,
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testEnablement:    false,
			expectedError:     false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidControlPlaneBuilder(buildControlPlaneClientWithDummyObject())

		result := testBuilder.WithGatewaysEnablement(testCase.testEnablement)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testEnablement, *result.Definition.Spec.Gateways.Enablement.Enabled)
		}
	}
}

func buildValidControlPlaneBuilder(apiClient *clients.Settings) *ControlPlaneBuilder {
	controlPlaneBuilder := NewControlPlaneBuilder(
		apiClient, defaultControlPlaneName, defaultServiceMeshNamespace)
	controlPlaneBuilder.Definition.ResourceVersion = "999"

	return controlPlaneBuilder
}

func buildInValidControlPlaneBuilder(apiClient *clients.Settings) *ControlPlaneBuilder {
	controlPlaneBuilder := NewControlPlaneBuilder(
		apiClient, "", defaultServiceMeshNamespace)
	controlPlaneBuilder.Definition.ResourceVersion = "999"

	return controlPlaneBuilder
}

func buildControlPlaneClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyControlplane(),
		SchemeAttachers: istiov2TestSchemes,
	})
}

func buildDummyControlplane() []runtime.Object {
	return append([]runtime.Object{}, &istiov2.ServiceMeshControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultControlPlaneName,
			Namespace: defaultServiceMeshNamespace,
		},
	})
}
