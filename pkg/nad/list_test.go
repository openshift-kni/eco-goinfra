package nad

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
)

func TestListNad(t *testing.T) {
	testCases := []struct {
		nad           []*Builder
		nsName        string
		expectedError error
		client        bool
	}{
		{
			nad: []*Builder{
				buildValidNADNetworkTestBuilder(buildTestClientWithDummyObject())},
			nsName:        "nadnamespace",
			expectedError: nil,
			client:        true,
		},
		{
			nad: []*Builder{
				buildValidNADNetworkTestBuilder(buildTestClientWithDummyObject())},
			nsName:        "",
			expectedError: fmt.Errorf("failed to list NADs, 'nsname' parameter is empty"),
			client:        true,
		},
		{
			nad: []*Builder{
				buildValidNADNetworkTestBuilder(buildTestClientWithDummyObject())},
			nsName:        "nadnamespace",
			expectedError: fmt.Errorf("nadList 'apiClient' cannot be empty"),
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyNADNetworkObject(),
				SchemeAttachers: testSchemes,
			})
		}

		nadBuilders, err := List(testSettings, testCase.nsName)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, len(nadBuilders), len(testCase.nad))
		}
	}
}
