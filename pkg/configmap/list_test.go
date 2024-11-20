package configmap

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestList(t *testing.T) {
	testCases := []struct {
		nsname         string
		client         bool
		configMaps     []runtime.Object
		configmapCount int
		expectedError  error
	}{
		{
			nsname: "test-namespace",
			client: true,
			configMaps: []runtime.Object{
				generateConfigMap("test-name1", "test-namespace"),
				generateConfigMap("test-name2", "test-namespace"),
			},
			configmapCount: 2,
			expectedError:  nil,
		},
		{
			nsname: "test-namespace",
			client: true,
			configMaps: []runtime.Object{
				generateConfigMap("test-name1", "test-namespace"),
				generateConfigMap("test-name2", "test-namespace2"),
			},
			configmapCount: 1,
			expectedError:  nil,
		},
		{
			nsname:        "",
			client:        true,
			configMaps:    []runtime.Object{},
			expectedError: fmt.Errorf("failed to list configmaps, 'nsname' parameter is empty"),
		},
		{
			nsname:        "test-namespace",
			client:        false,
			configMaps:    []runtime.Object{},
			expectedError: fmt.Errorf("the apiClient cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			testSettings *clients.Settings
		)

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: testCase.configMaps,
			})
		}

		configmapBuilders, err := List(testSettings, testCase.nsname)

		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, configmapBuilders)
			assert.Equal(t, testCase.configmapCount, len(configmapBuilders))
		}
	}
}

func TestListInAllNamespaces(t *testing.T) {
	testCases := []struct {
		client         bool
		configMaps     []runtime.Object
		configmapCount int
		expectedError  error
	}{
		{
			client: true,
			configMaps: []runtime.Object{
				generateConfigMap("test-name1", "test-namespace"),
				generateConfigMap("test-name2", "test-namespace"),
			},
			configmapCount: 2,
			expectedError:  nil,
		},
		{
			client: true,
			configMaps: []runtime.Object{
				generateConfigMap("test-name1", "test-namespace"),
				generateConfigMap("test-name2", "test-namespace2"),
			},
			configmapCount: 2,
			expectedError:  nil,
		},
		{
			client:        false,
			configMaps:    []runtime.Object{},
			expectedError: fmt.Errorf("the apiClient cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			testSettings *clients.Settings
		)

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: testCase.configMaps,
			})
		}

		configmapBuilders, err := ListInAllNamespaces(testSettings)

		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, configmapBuilders)
			assert.Equal(t, testCase.configmapCount, len(configmapBuilders))
		}
	}
}
