package cgu

import (
	"fmt"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
)

func TestCguListInAllNamespaces(t *testing.T) {
	testCases := []struct {
		testCGU       []*CguBuilder
		listOptions   []client.ListOptions
		expectedError error
		client        bool
	}{
		{
			testCGU:       []*CguBuilder{buildValidCguTestBuilder(buildTestClientWithDummyCguObject())},
			listOptions:   nil,
			expectedError: nil,
			client:        true,
		},
		{
			testCGU:       []*CguBuilder{buildValidCguTestBuilder(buildTestClientWithDummyCguObject())},
			listOptions:   []client.ListOptions{{Namespace: "test"}},
			expectedError: nil,
			client:        true,
		},
		{
			testCGU:       []*CguBuilder{buildValidCguTestBuilder(buildTestClientWithDummyCguObject())},
			listOptions:   []client.ListOptions{{Namespace: "test"}, {Continue: "true"}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			testCGU:       []*CguBuilder{buildValidCguTestBuilder(buildTestClientWithDummyCguObject())},
			listOptions:   nil,
			expectedError: fmt.Errorf("failed to list cgu objects, 'apiClient' parameter is empty"),
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyCguObject(),
				SchemeAttachers: testSchemes,
			})
		}

		cguBuilders, err := ListInAllNamespaces(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(cguBuilders), len(testCase.testCGU))
		}
	}
}
