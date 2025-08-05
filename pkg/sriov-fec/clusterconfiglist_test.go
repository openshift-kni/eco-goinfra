package sriovfec

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestClusterConfigList(t *testing.T) {
	testCases := []struct {
		apiclient     bool
		nsname        string
		options       []client.ListOptions
		expectedError error
	}{
		{
			apiclient:     false,
			nsname:        defaultClusterConfigNamespace,
			options:       []client.ListOptions{},
			expectedError: fmt.Errorf("failed to list SriovFecClusterConfig, 'apiClient' parameter is empty"),
		},
		{
			apiclient:     true,
			nsname:        "",
			options:       []client.ListOptions{},
			expectedError: fmt.Errorf("failed to list SriovFecClusterConfig, 'nsname' parameter is empty"),
		},
		{
			apiclient:     true,
			nsname:        defaultClusterConfigNamespace,
			options:       []client.ListOptions{},
			expectedError: nil,
		},
		{
			apiclient: true,
			nsname:    defaultClusterConfigNamespace,
			options: []client.ListOptions{
				{
					LabelSelector: nil,
					FieldSelector: nil,
				},
			},
			expectedError: nil,
		},
		{
			apiclient: true,
			nsname:    defaultClusterConfigNamespace,
			options: []client.ListOptions{
				{
					LabelSelector: nil,
					FieldSelector: nil,
				},
				{
					LabelSelector: nil,
					FieldSelector: nil,
				},
			},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.apiclient {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyClusterConfigList(),
				SchemeAttachers: testSchemes,
			})
		}

		testBuilderList, err := ListClusterConfig(testSettings, testCase.nsname, testCase.options...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, len(buildDummyClusterConfigList()), len(testBuilderList))
		}
	}
}

func buildDummyClusterConfigList() []runtime.Object {
	return []runtime.Object{
		buildDummyClusterConfig("cluster-config-0", "test-ns"),
		buildDummyClusterConfig("cluster-config-1", "test-ns"),
	}
}
