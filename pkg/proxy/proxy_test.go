package proxy

import (
	"fmt"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var testSchemes = []clients.SchemeAttacher{
	configv1.Install,
}

func TestPullProxy(t *testing.T) {
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
			expectedError:       fmt.Errorf("proxy object %s does not exist", clusterProxyName),
		},
		{
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("proxy 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, buildDummyProxy())
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := Pull(testSettings)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, clusterProxyName, testBuilder.Definition.Name)
		}
	}
}

func TestProxyGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *Builder
		expectedError string
	}{
		{
			testBuilder:   newProxyBuilder(buildTestClientWithDummyProxy()),
			expectedError: "",
		},
		{
			testBuilder:   newProxyBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Sprintf("proxies.config.openshift.io \"%s\" not found", clusterProxyName),
		},
	}

	for _, testCase := range testCases {
		proxy, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, clusterProxyName, proxy.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestProxyExists(t *testing.T) {
	testCases := []struct {
		testBuilder *Builder
		exists      bool
	}{
		{
			testBuilder: newProxyBuilder(buildTestClientWithDummyProxy()),
			exists:      true,
		},
		{
			testBuilder: newProxyBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestProxyValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		expectedError error
	}{
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: nil,
		},
		{
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: fmt.Errorf("error: received nil proxy builder"),
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: fmt.Errorf("can not redefine the undefined proxy"),
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: fmt.Errorf("proxy builder cannot have nil apiClient"),
		},
	}

	for _, testCase := range testCases {
		proxyBuilder := newProxyBuilder(clients.GetTestClients(clients.TestClientParams{}))

		if testCase.builderNil {
			proxyBuilder = nil
		}

		if testCase.definitionNil {
			proxyBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			proxyBuilder.apiClient = nil
		}

		valid, err := proxyBuilder.validate()

		if testCase.expectedError != nil {
			assert.False(t, valid)
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.True(t, valid)
			assert.Nil(t, err)
		}
	}
}

// buildDummyProxy returns a new Proxy with the clusterProxyName.
func buildDummyProxy() *configv1.Proxy {
	return &configv1.Proxy{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterProxyName,
		},
	}
}

// buildTestClientWithDummyProxy returns a new client with a mock Proxy object.
func buildTestClientWithDummyProxy() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{buildDummyProxy()},
		SchemeAttachers: testSchemes,
	})
}

// newProxyBuilder returns a Builder for a Proxy with the provided client.
func newProxyBuilder(apiClient *clients.Settings) *Builder {
	return &Builder{
		apiClient:  apiClient.Client,
		Definition: buildDummyProxy(),
	}
}
