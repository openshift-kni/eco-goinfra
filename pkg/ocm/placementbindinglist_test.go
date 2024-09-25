package ocm

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListPlacementBindingsInAllNamespaces(t *testing.T) {
	testCases := []struct {
		placementBindings []*PlacementBindingBuilder
		listOptions       []runtimeclient.ListOptions
		expectedError     error
		client            bool
	}{
		{
			placementBindings: []*PlacementBindingBuilder{
				buildValidPlacementBindingTestBuilder(buildTestClientWithDummyPlacementBinding()),
			},
			listOptions:   nil,
			expectedError: nil,
			client:        true,
		},
		{
			placementBindings: []*PlacementBindingBuilder{
				buildValidPlacementBindingTestBuilder(buildTestClientWithDummyPlacementBinding()),
			},
			listOptions:   []runtimeclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			expectedError: nil,
			client:        true,
		},
		{
			placementBindings: []*PlacementBindingBuilder{
				buildValidPlacementBindingTestBuilder(buildTestClientWithDummyPlacementBinding()),
			},
			listOptions: []runtimeclient.ListOptions{
				{LabelSelector: labels.NewSelector()},
				{LabelSelector: labels.NewSelector()},
			},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			placementBindings: []*PlacementBindingBuilder{
				buildValidPlacementBindingTestBuilder(buildTestClientWithDummyPlacementBinding()),
			},
			listOptions:   []runtimeclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			expectedError: fmt.Errorf("failed to list placementBindings, 'apiClient' parameter is nil"),
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithDummyPlacementBinding()
		}

		builders, err := ListPlacementBindingsInAllNamespaces(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.placementBindings), len(builders))
		}
	}
}
