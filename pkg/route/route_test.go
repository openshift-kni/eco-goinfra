package route

import (
	"fmt"
	"testing"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
)

func TestNewBuilder(t *testing.T) {
	testcases := []struct {
		name           string
		namespace      string
		targetService  string
		expectedErrMsg string
	}{
		{
			name:           "route-test-name",
			namespace:      "route-test-namespace",
			targetService:  "route-test-service",
			expectedErrMsg: "",
		},
		{
			name:           "",
			namespace:      "route-test-namespace",
			targetService:  "route-test-service",
			expectedErrMsg: "route 'name' cannot be empty",
		},
		{
			name:           "route-test-name",
			namespace:      "",
			targetService:  "route-test-service",
			expectedErrMsg: "route 'nsname' cannot be empty",
		},
		{
			name:           "route-test-name",
			namespace:      "route-test-namespace",
			targetService:  "",
			expectedErrMsg: "route 'serviceName' cannot be empty",
		},
	}

	for _, test := range testcases {
		testBuilder := NewBuilder(clients.GetTestClients(clients.TestClientParams{}),
			test.name, test.namespace, test.targetService)
		assert.Equal(t, test.expectedErrMsg, testBuilder.errorMsg)
	}
}

func TestPull(t *testing.T) {
	generateRoute := func(name, namespace string) *routev1.Route {
		return &routev1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: routev1.RouteSpec{
				To: routev1.RouteTargetReference{
					Kind: "Service",
					Name: "route-test-service",
				},
			},
		}
	}

	testCases := []struct {
		name                string
		namespace           string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			name:                "route-test-name",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       false,
			expectedErrorText:   "",
		},
		{
			name:                "route-test-name2",
			namespace:           "test-namespace2",
			addToRuntimeObjects: false,
			expectedError:       true,
			expectedErrorText:   "route object route-test-name2 does not exist in namespace test-namespace2",
		},
		{
			name:                "",
			namespace:           "test-namespace3",
			addToRuntimeObjects: false,
			expectedError:       true,
			expectedErrorText:   "route 'name' cannot be empty",
		},
		{
			name:                "route-test-name4",
			namespace:           "",
			addToRuntimeObjects: false,
			expectedError:       true,
			expectedErrorText:   "route 'namespace' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testRoute := generateRoute(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testRoute)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		// Test the Pull method
		builderResult, err := Pull(testSettings, testCase.name, testCase.namespace)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			// Check the error message
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

// buildValidTestBuilder returns a valid Builder for testing purposes.
func buildValidTestBuilder() *Builder {
	return NewBuilder(clients.GetTestClients(clients.TestClientParams{}),
		"route-test-name", "route-test-namespace", "route-test-service")
}

func TestWithTargetPortNumber(t *testing.T) {
	testBuilder := buildValidTestBuilder()

	testBuilder.WithTargetPortNumber(8080)

	assert.Equal(t, int32(8080), testBuilder.Definition.Spec.Port.TargetPort.IntVal)
}

func TestWithTargetPortName(t *testing.T) {
	testCases := []struct {
		name           string
		expectedErrMsg string
	}{
		{
			"",
			"route target port name cannot be empty string",
		},
		{
			"8080-target",
			"",
		},
	}

	for _, test := range testCases {
		testBuilder := buildValidTestBuilder()
		testBuilder.WithTargetPortName(test.name)

		assert.Equal(t, test.expectedErrMsg, testBuilder.errorMsg)
	}
}

func TestWithHostDomain(t *testing.T) {
	testCases := []struct {
		hostDomain     string
		expectedErrMsg string
	}{
		{
			"",
			"route host domain cannot be empty string",
		},
		{
			"app.demo-server.dummy.domain.com",
			"",
		},
	}

	for _, test := range testCases {
		testBuilder := buildValidTestBuilder()
		testBuilder.WithHostDomain(test.hostDomain)

		assert.Equal(t, test.expectedErrMsg, testBuilder.errorMsg)
	}
}

func TestWithWildCardPolicy(t *testing.T) {
	testCases := []struct {
		policy         string
		expectedErrMsg string
	}{
		{
			"",
			fmt.Sprintf("received unsupported route wildcardPolicy: supported policies %v", supportedWildCardPolicies()),
		},
		{
			"Any",
			fmt.Sprintf("received unsupported route wildcardPolicy: supported policies %v", supportedWildCardPolicies()),
		},
		{
			"Subdomain",
			"",
		},
		{
			"None",
			"",
		},
	}

	for _, test := range testCases {
		testBuilder := buildValidTestBuilder()
		testBuilder.WithWildCardPolicy(test.policy)

		assert.Equal(t, test.expectedErrMsg, testBuilder.errorMsg)
	}
}
