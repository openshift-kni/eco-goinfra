package pfstatus

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/pfstatus/pfstatustypes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	pfStatusTestSchemes                = []clients.SchemeAttacher{pfstatustypes.AddToScheme}
	defaultPfStatusConfigurationName   = "pfstatusconfiguration"
	defaultPfStatusConfigurationNsName = "test-namespace"
)

func TestNewPfStatusConfigurationBuilder(t *testing.T) {
	generatePfStatusConfiguration := NewPfStatusConfigurationBuilder

	testCases := []struct {
		name          string
		namespace     string
		client        bool
		expectedError string
	}{
		{
			name:          defaultPfStatusConfigurationName,
			namespace:     defaultPfStatusConfigurationNsName,
			client:        true,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     defaultPfStatusConfigurationNsName,
			client:        true,
			expectedError: "pfStatusConfiguration 'name' cannot be empty",
		},
		{
			name:          defaultPfStatusConfigurationName,
			namespace:     "",
			client:        true,
			expectedError: "pfStatusConfiguration 'nsname' cannot be empty",
		},
		{
			name:          defaultPfStatusConfigurationName,
			namespace:     defaultPfStatusConfigurationNsName,
			client:        false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testPfStatusConfigurationBuilder := generatePfStatusConfiguration(testSettings, testCase.name, testCase.namespace)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, testPfStatusConfigurationBuilder.errorMsg)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.name, testPfStatusConfigurationBuilder.Definition.Name)
			}
		} else {
			assert.Nil(t, testPfStatusConfigurationBuilder)
		}
	}
}

func TestPfStatusConfigurationCreate(t *testing.T) {
	testCases := []struct {
		testPfStatusConfiguration *PfStatusConfigurationBuilder
		expectedError             error
	}{
		{
			testPfStatusConfiguration: buildValidPfStatusConfigurationBuilder(
				buildPfStatusConfigurationTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testPfStatusConfiguration: buildInValidPfStatusConfigurationBuilder(
				buildPfStatusConfigurationTestClientWithDummyObject()),
			expectedError: fmt.Errorf("pfStatusConfiguration 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		testPfStatusConfigBuilder, err := testCase.testPfStatusConfiguration.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testPfStatusConfigBuilder.Definition.Name, testPfStatusConfigBuilder.Object.Name)
		}
	}
}

func TestPfStatusConfigurationGet(t *testing.T) {
	testCases := []struct {
		testPfStatusConfiguration *PfStatusConfigurationBuilder
		expectedError             string
	}{
		{
			testPfStatusConfiguration: buildValidPfStatusConfigurationBuilder(
				buildPfStatusConfigurationTestClientWithDummyObject()),
			expectedError: "",
		},
		{
			testPfStatusConfiguration: buildInValidPfStatusConfigurationBuilder(
				buildPfStatusConfigurationTestClientWithDummyObject()),
			expectedError: "pfStatusConfiguration 'name' cannot be empty",
		},
		{
			testPfStatusConfiguration: buildValidPfStatusConfigurationBuilder(
				clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "pflacpmonitors.pfstatusrelay.openshift.io \"pfstatusconfiguration\" not found",
		},
	}

	for _, testCase := range testCases {
		pfStatusConfig, err := testCase.testPfStatusConfiguration.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, pfStatusConfig.Name, testCase.testPfStatusConfiguration.Definition.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestPullPfStatusConfiguration(t *testing.T) {
	generatePfStatusConfiguration := func(name, namespace string) *pfstatustypes.PFLACPMonitor {
		return &pfstatustypes.PFLACPMonitor{
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
			name:                defaultPfStatusConfigurationName,
			namespace:           defaultPfStatusConfigurationNsName,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           defaultPfStatusConfigurationNsName,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("pfStatusConfiguration 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                defaultPfStatusConfigurationName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("pfStatusConfiguration 'namespace' cannot be empty"),
			client:              true,
		},
		{
			name:                defaultPfStatusConfigurationName,
			namespace:           defaultPfStatusConfigurationNsName,
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf("pfStatusConfiguration object pfstatusconfiguration does not" +
				" exist in namespace test-namespace"),
			client: true,
		},
		{
			name:                defaultPfStatusConfigurationName,
			namespace:           defaultPfStatusConfigurationNsName,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("pfStatusConfiguration 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testPfStatus := generatePfStatusConfiguration(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPfStatus)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: pfStatusTestSchemes,
			})
		}

		builderResult, err := PullPfStatusConfiguration(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

func TestPfStatusConfigurationExist(t *testing.T) {
	testCases := []struct {
		testPfStatusConfiguration *PfStatusConfigurationBuilder
		expectedStatus            bool
	}{
		{
			testPfStatusConfiguration: buildValidPfStatusConfigurationBuilder(
				buildPfStatusConfigurationTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testPfStatusConfiguration: buildInValidPfStatusConfigurationBuilder(
				buildPfStatusConfigurationTestClientWithDummyObject()),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testPfStatusConfiguration.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestPfStatusConfigurationDelete(t *testing.T) {
	testCases := []struct {
		testPfStatusConfiguration *PfStatusConfigurationBuilder
		expectedError             error
	}{
		{
			testPfStatusConfiguration: buildValidPfStatusConfigurationBuilder(
				buildPfStatusConfigurationTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testPfStatusConfiguration: buildInValidPfStatusConfigurationBuilder(
				buildPfStatusConfigurationTestClientWithDummyObject()),
			expectedError: fmt.Errorf("pfStatusConfiguration 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testPfStatusConfiguration.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testPfStatusConfiguration.Object)
		}
	}
}

func TestPfStatusConfigurationWithNodeSelector(t *testing.T) {
	testCases := []struct {
		testPfStatus  *PfStatusConfigurationBuilder
		nodeSelector  map[string]string
		expectedError string
	}{
		{
			testPfStatus: buildValidPfStatusConfigurationBuilder(buildPfStatusConfigurationTestClientWithDummyObject()),
			nodeSelector: map[string]string{"test": "test1"},
		},
		{
			testPfStatus:  buildValidPfStatusConfigurationBuilder(buildPfStatusConfigurationTestClientWithDummyObject()),
			nodeSelector:  map[string]string{},
			expectedError: "pfStatusConfiguration 'nodeSelector' cannot be empty map",
		},
	}

	for _, testCase := range testCases {
		pfStatusBuilder := testCase.testPfStatus.WithNodeSelector(testCase.nodeSelector)
		assert.Equal(t, testCase.expectedError, pfStatusBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.nodeSelector, pfStatusBuilder.Definition.Spec.NodeSelector)
		}
	}
}

func TestPfStatusConfigurationWithInterface(t *testing.T) {
	testCases := []struct {
		testPfStatus  *PfStatusConfigurationBuilder
		interfaceName string
		expectedError string
	}{
		{
			testPfStatus:  buildValidPfStatusConfigurationBuilder(buildPfStatusConfigurationTestClientWithDummyObject()),
			interfaceName: "ens3f0np0",
		},
		{
			testPfStatus:  buildValidPfStatusConfigurationBuilder(buildPfStatusConfigurationTestClientWithDummyObject()),
			interfaceName: "",
			expectedError: "interface can not be empty string",
		},
	}

	for _, testCase := range testCases {
		pfStatusBuilder := testCase.testPfStatus.WithInterface(testCase.interfaceName)
		assert.Equal(t, testCase.expectedError, pfStatusBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.interfaceName, pfStatusBuilder.Definition.Spec.Interfaces[0])
		}
	}
}

func TestPfStatusConfigurationPollingInterval(t *testing.T) {
	testCases := []struct {
		testPfStatus    *PfStatusConfigurationBuilder
		pollingInterval int
		expectedError   string
	}{
		{
			testPfStatus:    buildValidPfStatusConfigurationBuilder(buildPfStatusConfigurationTestClientWithDummyObject()),
			pollingInterval: 1000,
			expectedError:   "",
		},
		{
			testPfStatus:    buildValidPfStatusConfigurationBuilder(buildPfStatusConfigurationTestClientWithDummyObject()),
			pollingInterval: 10,
			expectedError:   "pfStatusConfiguration 'pollingInterval' value is not valid",
		},
		{
			testPfStatus:    buildValidPfStatusConfigurationBuilder(buildPfStatusConfigurationTestClientWithDummyObject()),
			pollingInterval: 65555,
			expectedError:   "pfStatusConfiguration 'pollingInterval' value is not valid",
		},
	}

	for _, testCase := range testCases {
		pfStatusBuilder := testCase.testPfStatus.WithPollingInterval(testCase.pollingInterval)
		assert.Equal(t, testCase.expectedError, pfStatusBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.pollingInterval, pfStatusBuilder.Definition.Spec.PollingInterval)
		}
	}
}

func buildValidPfStatusConfigurationBuilder(apiClient *clients.Settings) *PfStatusConfigurationBuilder {
	return NewPfStatusConfigurationBuilder(apiClient, defaultPfStatusConfigurationName, defaultPfStatusConfigurationNsName)
}

func buildInValidPfStatusConfigurationBuilder(apiClient *clients.Settings) *PfStatusConfigurationBuilder {
	return NewPfStatusConfigurationBuilder(apiClient, "", defaultPfStatusConfigurationNsName)
}

func buildPfStatusConfigurationTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPfStatusConfig(defaultPfStatusConfigurationName),
		},
		SchemeAttachers: pfStatusTestSchemes,
	})
}

// buildDummyPfStatusConfig returns a PfStatusConfiguration with the provided name.
func buildDummyPfStatusConfig(name string) *pfstatustypes.PFLACPMonitor {
	return &pfstatustypes.PFLACPMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: defaultPfStatusConfigurationNsName,
		},
	}
}
