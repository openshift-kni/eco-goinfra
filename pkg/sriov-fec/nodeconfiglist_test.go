package sriovfec

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestNodeConfigList(t *testing.T) {
	testCases := []struct {
		apiclient     bool
		nsname        string
		expectedError error
	}{
		{
			apiclient:     false,
			nsname:        defaultNodeConfigNamespace,
			expectedError: fmt.Errorf("failed to list SriovFecNodeConfig, 'apiClient' parameter is empty"),
		},
		{
			apiclient:     true,
			nsname:        "",
			expectedError: fmt.Errorf("failed to list SriovFecNodeConfig, 'nsname' parameter is empty"),
		},
		{
			apiclient:     true,
			nsname:        defaultNodeConfigNamespace,
			expectedError: nil,
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

		testBuilderList, err := List(testSettings, testCase.nsname)
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
