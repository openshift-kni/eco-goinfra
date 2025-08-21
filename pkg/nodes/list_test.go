package nodes

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestNodesList(t *testing.T) {
	testCases := []struct {
		nodes         []*Builder
		listOptions   []metav1.ListOptions
		client        bool
		expectedError error
	}{
		{
			nodes:         []*Builder{buildValidNodeTestBuilder(buildTestClientWithDummyNode())},
			listOptions:   nil,
			client:        true,
			expectedError: nil,
		},
		{
			nodes:         []*Builder{buildValidNodeTestBuilder(buildTestClientWithDummyNode())},
			listOptions:   []metav1.ListOptions{{LabelSelector: "test"}},
			client:        true,
			expectedError: nil,
		},
		{
			nodes:         []*Builder{buildValidNodeTestBuilder(buildTestClientWithDummyNode())},
			listOptions:   []metav1.ListOptions{{LabelSelector: "test"}, {LabelSelector: "test"}},
			client:        true,
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
		},
		{
			nodes:         []*Builder{buildValidNodeTestBuilder(buildTestClientWithDummyNode())},
			listOptions:   nil,
			client:        false,
			expectedError: fmt.Errorf("failed to list node objects, 'apiClient' parameter is empty"),
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithDummyNode()
		}

		nodeBuilders, err := List(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(nodeBuilders), len(testCase.nodes))
		}
	}
}

func TestNodesListExternalIPv4Networks(t *testing.T) {
	testCases := []struct {
		client        bool
		expectedError error
	}{
		{
			client:        true,
			expectedError: nil,
		},
		{
			client:        false,
			expectedError: fmt.Errorf("failed to list node objects, 'apiClient' parameter is empty"),
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			node := buildDummyNode(defaultNodeName)
			node.Annotations = map[string]string{
				ovnExternalAddresses: defaultExternalNetworks,
			}
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: []runtime.Object{node},
			})
		}

		ipv4Networks, err := ListExternalIPv4Networks(testSettings)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, []string{defaultExternalIPv4}, ipv4Networks)
		}
	}
}

func TestNodesWaitForAllNodesAreReady(t *testing.T) {
	testCases := []struct {
		client        bool
		ready         bool
		expectedError error
	}{
		{
			client:        true,
			ready:         true,
			expectedError: nil,
		},
		{
			client:        false,
			ready:         true,
			expectedError: fmt.Errorf("failed to list node objects, 'apiClient' parameter is empty"),
		},
		{
			client:        true,
			ready:         false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: []runtime.Object{buildDummyNodeWithReadiness(defaultNodeName, testCase.ready)},
			})
		}

		ready, err := WaitForAllNodesAreReady(testSettings, time.Second)
		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, testCase.expectedError == nil, ready)
	}
}

func TestNodesWaitForAllNodesToReboot(t *testing.T) {
	// There's no way to test for success without editing the node while the function is running so only failures
	// are covered here.
	testCases := []struct {
		client        bool
		rebooted      bool
		expectedError error
	}{
		{
			client:        false,
			rebooted:      true,
			expectedError: fmt.Errorf("failed to list node objects, 'apiClient' parameter is empty"),
		},
		{
			client:        true,
			rebooted:      false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: []runtime.Object{buildDummyNodeWithReadiness(defaultNodeName, testCase.rebooted)},
			})
		}

		ready, err := WaitForAllNodesToReboot(testSettings, time.Second)
		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, testCase.expectedError == nil, ready)
	}
}

// buildDummyNodeWithReadiness returns a dummy node with the specified value for the ready condition.
func buildDummyNodeWithReadiness(name string, ready bool) *corev1.Node {
	node := buildDummyNode(name)

	readyStatus := corev1.ConditionFalse
	if ready {
		readyStatus = corev1.ConditionTrue
	}

	node.Status.Conditions = []corev1.NodeCondition{{
		Type:   corev1.NodeReady,
		Status: readyStatus,
	}}

	return node
}
