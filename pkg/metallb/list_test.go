package metallb

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/schemes/metallb/frrtypes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListFrrNodeState(t *testing.T) {
	testCases := []struct {
		testFRRNodeStates []*frrtypes.FRRNodeState
		listOptions       []client.ListOptions
		nsName            string
		client            bool
		expectedError     error
	}{
		{
			testFRRNodeStates: []*frrtypes.FRRNodeState{buildFrrNodeState("test"),
				buildFrrNodeState("test2")},
			nsName: "testnamespace",
			client: true,
		},
		{
			testFRRNodeStates: []*frrtypes.FRRNodeState{buildFrrNodeState("test"),
				buildFrrNodeState("test2")},
			expectedError: fmt.Errorf("failed to list FrrNodeStates, 'nsname' parameter is empty"),
			client:        true,
			nsName:        "",
		},
		{
			testFRRNodeStates: []*frrtypes.FRRNodeState{buildFrrNodeState("test"),
				buildFrrNodeState("test2")},
			nsName:      "testnamespace",
			listOptions: []client.ListOptions{{Continue: "true"}},
			client:      true,
		},
		{
			testFRRNodeStates: []*frrtypes.FRRNodeState{buildFrrNodeState("test"),
				buildFrrNodeState("test2")},
			nsName:        "testnamespace",
			listOptions:   []client.ListOptions{{Namespace: "testnamespace"}, {Continue: "true"}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			testFRRNodeStates: []*frrtypes.FRRNodeState{buildFrrNodeState("test"),
				buildFrrNodeState("test2")},
			nsName:        "testnamespace",
			expectedError: fmt.Errorf("failed to list FrrNodeStates, 'apiClient' parameter is empty"),
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var (
			testSettings   *clients.Settings
			runtimeObjects []runtime.Object
		)

		for _, networkFRRNodeState := range testCase.testFRRNodeStates {
			runtimeObjects = append(runtimeObjects, networkFRRNodeState)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: frrTestSchemes,
			})
		}

		builders, err := ListFrrNodeState(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, len(testCase.testFRRNodeStates), len(builders))
		}
	}
}

func buildFrrNodeState(name string) *frrtypes.FRRNodeState {
	return &frrtypes.FRRNodeState{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "testnamespace",
		},
		Spec:   frrtypes.FRRNodeStateSpec{},
		Status: frrtypes.FRRNodeStateStatus{},
	}
}
