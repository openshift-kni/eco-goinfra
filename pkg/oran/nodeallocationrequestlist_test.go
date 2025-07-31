package oran

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListNodeAllocationRequests(t *testing.T) {
	testCases := []struct {
		nodeAllocationRequests []*NARBuilder
		listOptions            []runtimeclient.ListOptions
		client                 bool
		expectedError          error
	}{
		{
			nodeAllocationRequests: []*NARBuilder{buildValidNARTestBuilder(buildTestClientWithDummyNAR())},
			listOptions:            nil,
			client:                 true,
			expectedError:          nil,
		},
		{
			nodeAllocationRequests: []*NARBuilder{buildValidNARTestBuilder(buildTestClientWithDummyNAR())},
			listOptions:            []runtimeclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			client:                 true,
			expectedError:          nil,
		},
		{
			nodeAllocationRequests: []*NARBuilder{buildValidNARTestBuilder(buildTestClientWithDummyNAR())},
			listOptions: []runtimeclient.ListOptions{
				{LabelSelector: labels.NewSelector()},
				{LabelSelector: labels.NewSelector()},
			},
			client:        true,
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
		},
		{
			nodeAllocationRequests: []*NARBuilder{buildValidNARTestBuilder(buildTestClientWithDummyNAR())},
			listOptions:            nil,
			client:                 false,
			expectedError:          fmt.Errorf("failed to list nodeAllocationRequests, 'apiClient' parameter is nil"),
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithDummyNAR()
		}

		builders, err := ListNodeAllocationRequests(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.nodeAllocationRequests), len(builders))
		}
	}
}
