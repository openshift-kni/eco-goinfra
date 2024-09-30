package metallb

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/schemes/metallb/frrtypes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFrrNodeStateList(t *testing.T) {
	testCases := []struct {
		Definition    *frrtypes.FRRNodeState
		nodes         []*FrrNodeStateBuilder
		listOptions   []metav1.ListOptions
		client        bool
		expectedError error
	}{
		{
			nodes: []*FrrNodeStateBuilder{buildValidFrrNodeStateTestBuilder(
				buildTestFrrClientWithDummyNode(defaultNodeName))},
			listOptions:   nil,
			client:        true,
			expectedError: nil,
		},
		{
			nodes: []*FrrNodeStateBuilder{buildValidFrrNodeStateTestBuilder(
				buildTestFrrClientWithDummyNode(defaultNodeName))},
			listOptions:   []metav1.ListOptions{{LabelSelector: "test"}},
			client:        true,
			expectedError: nil,
		},
		{
			nodes: []*FrrNodeStateBuilder{buildValidFrrNodeStateTestBuilder(
				buildTestFrrClientWithDummyNode(defaultNodeName))},
			listOptions:   nil,
			client:        false,
			expectedError: fmt.Errorf("failed to list FrrNodeStates, 'apiClient' parameter is empty"),
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestFrrClientWithDummyNode(defaultNodeName)
		}

		frrNodes, err := ListFrrNodeState(testSettings)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.nodes), len(frrNodes))
		}
	}
}
