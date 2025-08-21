package cgu

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCguListInAllNamespaces(t *testing.T) {
	testCases := []struct {
		testCGU       []*CguBuilder
		listOptions   []metav1.ListOptions
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
			listOptions:   []metav1.ListOptions{{LabelSelector: "test"}},
			expectedError: nil,
			client:        true,
		},
		{
			testCGU:       []*CguBuilder{buildValidCguTestBuilder(buildTestClientWithDummyCguObject())},
			listOptions:   []metav1.ListOptions{{LabelSelector: "test"}, {Continue: "true"}},
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
				K8sMockObjects: buildDummyCguObject(),
			})
		}

		cguBuilders, err := ListInAllNamespaces(testSettings, testCase.listOptions...)
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(cguBuilders), len(testCase.testCGU))
		}
	}
}
