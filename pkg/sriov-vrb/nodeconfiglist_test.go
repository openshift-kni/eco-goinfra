package sriovvrb

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestVrbNodeConfigList(t *testing.T) {
	testCases := []struct {
		apiclient     bool
		nsname        string
		options       []client.ListOptions
		expectedError error
	}{
		{
			apiclient:     true,
			nsname:        defaultNodeConfigNamespace,
			options:       []client.ListOptions{},
			expectedError: nil,
		},
		{
			apiclient: true,
			nsname:    defaultNodeConfigNamespace,
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
			nsname:    defaultNodeConfigNamespace,
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
		{
			apiclient:     false,
			nsname:        defaultNodeConfigNamespace,
			options:       []client.ListOptions{},
			expectedError: fmt.Errorf("failed to list SriovVrbNodeConfig, 'apiClient' parameter is empty"),
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.apiclient {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyNodeConfigList(),
				SchemeAttachers: testSchemes,
			})
		}

		testBuilderList, err := ListNodeConfig(testSettings, testCase.nsname, testCase.options...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, len(buildDummyNodeConfigList()), len(testBuilderList))
		}
	}
}

func buildDummyNodeConfigList() []runtime.Object {
	return []runtime.Object{
		buildDummyNodeConfig("worker-0", "test-ns"),
		buildDummyNodeConfig("worker-1", "test-ns"),
	}
}
