package ocm

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListPlacementrulesInAllNamespaces(t *testing.T) {
	testCases := []struct {
		placementRules []*PlacementRuleBuilder
		listOptions    []runtimeclient.ListOptions
		expectedError  error
		client         bool
	}{
		{
			placementRules: []*PlacementRuleBuilder{
				buildValidPlacementRuleTestBuilder(buildTestClientWithDummyPlacementRule()),
			},
			listOptions:   nil,
			expectedError: nil,
			client:        true,
		},
		{
			placementRules: []*PlacementRuleBuilder{
				buildValidPlacementRuleTestBuilder(buildTestClientWithDummyPlacementRule()),
			},
			listOptions:   []runtimeclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			expectedError: nil,
			client:        true,
		},
		{
			placementRules: []*PlacementRuleBuilder{
				buildValidPlacementRuleTestBuilder(buildTestClientWithDummyPlacementRule()),
			},
			listOptions: []runtimeclient.ListOptions{
				{LabelSelector: labels.NewSelector()},
				{LabelSelector: labels.NewSelector()},
			},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			placementRules: []*PlacementRuleBuilder{
				buildValidPlacementRuleTestBuilder(buildTestClientWithDummyPlacementRule()),
			},
			listOptions:   []runtimeclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			expectedError: fmt.Errorf("failed to list placementrules, 'apiClient' parameter is nil"),
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithDummyPlacementRule()
		}

		builders, err := ListPlacementrulesInAllNamespaces(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.placementRules), len(builders))
		}
	}
}
