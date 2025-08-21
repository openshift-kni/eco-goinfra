package oran

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListAllocatedNodes(t *testing.T) {
	testCases := []struct {
		nodes         []*AllocatedNodeBuilder
		listOptions   []runtimeclient.ListOptions
		client        bool
		expectedError error
	}{
		{
			nodes:         []*AllocatedNodeBuilder{buildValidAllocatedNodeTestBuilder(buildTestClientWithDummyAllocatedNode())},
			listOptions:   nil,
			client:        true,
			expectedError: nil,
		},
		{
			nodes:         []*AllocatedNodeBuilder{buildValidAllocatedNodeTestBuilder(buildTestClientWithDummyAllocatedNode())},
			listOptions:   []runtimeclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			client:        true,
			expectedError: nil,
		},
		{
			nodes: []*AllocatedNodeBuilder{buildValidAllocatedNodeTestBuilder(buildTestClientWithDummyAllocatedNode())},
			listOptions: []runtimeclient.ListOptions{
				{LabelSelector: labels.NewSelector()},
				{LabelSelector: labels.NewSelector()},
			},
			client:        true,
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
		},
		{
			nodes:         []*AllocatedNodeBuilder{buildValidAllocatedNodeTestBuilder(buildTestClientWithDummyAllocatedNode())},
			listOptions:   nil,
			client:        false,
			expectedError: fmt.Errorf("failed to list allocatedNodes, 'apiClient' parameter is nil"),
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithDummyAllocatedNode()
		}

		builders, err := ListAllocatedNodes(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.nodes), len(builders))
		}
	}
}
