package sriov

import (
	"fmt"
	"testing"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestNetworkList(t *testing.T) {
	testCases := []struct {
		testNetwork   []*NetworkBuilder
		nsName        string
		listOptions   []v1.ListOptions
		expectedError error
		client        bool
	}{
		{
			testNetwork:   []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			nsName:        defaultNetNsName,
			expectedError: nil,
			client:        true,
		},
		{
			testNetwork:   []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			nsName:        "",
			expectedError: fmt.Errorf("failed to list sriov networks, 'nsname' parameter is empty"),
			client:        true,
		},
		{
			testNetwork: []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			nsName:      defaultNetNsName,
			listOptions: []v1.ListOptions{{LabelSelector: "test"}},
			client:      true,
		},
		{
			testNetwork:   []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			nsName:        defaultNetNsName,
			listOptions:   []v1.ListOptions{{LabelSelector: "test"}, {Continue: "true"}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			testNetwork:   []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			nsName:        defaultNetNsName,
			expectedError: fmt.Errorf("failed to list sriov networks, 'apiClient' parameter is empty"),
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(buildDummySrIovNetworkObject())
		}

		netBuilders, err := List(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(netBuilders), len(testCase.testNetwork))
		}
	}
}

func TestNetworkCleanAllNetworksByTargetNamespace(t *testing.T) {
	testCases := []struct {
		testNetwork    []*NetworkBuilder
		operatorNsName string
		targetNsName   string
		listOptions    []v1.ListOptions
		client         bool
		expectedError  error
	}{
		{
			testNetwork:    []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			operatorNsName: defaultNetNsName,
			targetNsName:   defaultNetTargetNsName,
			client:         true,
		},
		{
			testNetwork:   []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			targetNsName:  defaultNetTargetNsName,
			expectedError: fmt.Errorf("failed to clean up sriov networks, 'operatornsname' parameter is empty"),
			client:        true,
		},
		{
			testNetwork:    []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			operatorNsName: defaultNetNsName,
			expectedError:  fmt.Errorf("failed to clean up sriov networks, 'targetnsname' parameter is empty"),
			client:         true,
		},
		{
			testNetwork:    []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			operatorNsName: defaultNetNsName,
			targetNsName:   defaultNetTargetNsName,
			listOptions:    []v1.ListOptions{{AllowWatchBookmarks: false}},
			client:         true,
		},
		{
			testNetwork:    []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			operatorNsName: defaultNetNsName,
			targetNsName:   defaultNetTargetNsName,
			listOptions:    []v1.ListOptions{{LabelSelector: "test"}, {Continue: "true"}},
			expectedError:  fmt.Errorf("error: more than one ListOptions was passed"),
			client:         true,
		},
		{
			testNetwork:    []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			operatorNsName: defaultNetNsName,
			targetNsName:   defaultNetTargetNsName,
			expectedError:  fmt.Errorf("failed to list sriov networks, 'apiClient' parameter is empty"),
			client:         false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(buildDummySrIovNetworkObject())
		}

		err := CleanAllNetworksByTargetNamespace(
			testSettings, testCase.operatorNsName, testCase.targetNsName, testCase.listOptions...)
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			netBuilders, err := List(testSettings, testCase.operatorNsName)
			assert.Nil(t, err)
			assert.Zero(t, len(netBuilders))
		}
	}
}

func TestListNetworkNodeState(t *testing.T) {
	testCases := []struct {
		testNetworkNodeStates []*srIovV1.SriovNetworkNodeState
		listOptions           []v1.ListOptions
		nsName                string
		client                bool
		expectedError         error
	}{
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodeState{buildNodeNetworkState("test", defaultNodeNsName),
				buildNodeNetworkState("test2", defaultNodeNsName)},
			nsName: defaultNodeNsName,
			client: true,
		},
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodeState{buildNodeNetworkState("test", defaultNodeNsName),
				buildNodeNetworkState("test2", defaultNodeNsName)},
			expectedError: fmt.Errorf("failed to list SriovNetworkNodeStates, 'nsname' parameter is empty"),
			client:        true,
		},
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodeState{buildNodeNetworkState("test", defaultNodeNsName),
				buildNodeNetworkState("test2", defaultNodeNsName)},
			nsName:      defaultNodeNsName,
			listOptions: []v1.ListOptions{{AllowWatchBookmarks: false}},
			client:      true,
		},
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodeState{buildNodeNetworkState("test", defaultNodeNsName),
				buildNodeNetworkState("test2", defaultNodeNsName)},
			nsName:        defaultNodeNsName,
			listOptions:   []v1.ListOptions{{LabelSelector: "test"}, {Continue: "true"}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodeState{buildNodeNetworkState("test", defaultNodeNsName),
				buildNodeNetworkState("test2", defaultNodeNsName)},
			nsName:        defaultNodeNsName,
			expectedError: fmt.Errorf("failed to list SriovNetworkNodeStates, 'apiClient' parameter is empty"),
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var (
			testSettings   *clients.Settings
			runtimeObjects []runtime.Object
		)

		for _, networkNodeState := range testCase.testNetworkNodeStates {
			runtimeObjects = append(runtimeObjects, networkNodeState)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(runtimeObjects)
		}

		builders, err := ListNetworkNodeState(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, len(testCase.testNetworkNodeStates), len(builders))
		}
	}
}

func TestListPolicy(t *testing.T) {
	testCases := []struct {
		testNetworkNodeStates []*srIovV1.SriovNetworkNodePolicy
		listOptions           []v1.ListOptions
		nsName                string
		client                bool
		expectedError         error
	}{
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", defaultNodeNsName),
				buildDummySrIovPolicy("test1", defaultNodeNsName)},
			nsName: defaultNodeNsName,
			client: true,
		},
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", defaultNodeNsName),
				buildDummySrIovPolicy("test1", defaultNodeNsName)},
			expectedError: fmt.Errorf("failed to list SriovNetworkNodePolicies, 'nsname' parameter is empty"),
			client:        true,
		},
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", defaultNodeNsName),
				buildDummySrIovPolicy("test1", defaultNodeNsName)},
			nsName:      defaultNodeNsName,
			listOptions: []v1.ListOptions{{AllowWatchBookmarks: false}},
			client:      true,
		},
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", defaultNodeNsName),
				buildDummySrIovPolicy("test1", defaultNodeNsName)},
			nsName:        defaultNodeNsName,
			listOptions:   []v1.ListOptions{{LabelSelector: "test"}, {Continue: "true"}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", defaultNodeNsName),
				buildDummySrIovPolicy("test1", defaultNodeNsName)},
			nsName:        defaultNodeNsName,
			expectedError: fmt.Errorf("failed to list SriovNetworkNodePolicies, 'apiClient' parameter is empty"),
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var (
			testSettings   *clients.Settings
			runtimeObjects []runtime.Object
		)

		for _, networkNodeState := range testCase.testNetworkNodeStates {
			runtimeObjects = append(runtimeObjects, networkNodeState)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(runtimeObjects)
		}

		builders, err := ListPolicy(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, len(testCase.testNetworkNodeStates), len(builders))
		}
	}
}

func TestCleanAllNetworkNodePolicies(t *testing.T) {
	testCases := []struct {
		testNetworkPolicy []*srIovV1.SriovNetworkNodePolicy
		listOptions       []v1.ListOptions
		nsName            string
		client            bool
		expectedError     error
	}{
		{
			testNetworkPolicy: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", defaultNodeNsName),
				buildDummySrIovPolicy("test1", defaultNodeNsName)},
			nsName: defaultNodeNsName,
			client: true,
		},
		{
			testNetworkPolicy: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", defaultNodeNsName),
				buildDummySrIovPolicy("test1", defaultNodeNsName)},
			expectedError: fmt.Errorf("failed to clean up SriovNetworkNodePolicies, 'operatornsname' parameter is empty"),
			client:        true,
		},
		{
			testNetworkPolicy: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", defaultNodeNsName),
				buildDummySrIovPolicy("test1", defaultNodeNsName)},
			nsName:      defaultNodeNsName,
			listOptions: []v1.ListOptions{{AllowWatchBookmarks: false}},
			client:      true,
		},
		{
			testNetworkPolicy: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", defaultNodeNsName),
				buildDummySrIovPolicy("test1", defaultNodeNsName)},
			nsName:        defaultNodeNsName,
			listOptions:   []v1.ListOptions{{LabelSelector: "test"}, {Continue: "true"}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			testNetworkPolicy: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", defaultNodeNsName),
				buildDummySrIovPolicy("test1", defaultNodeNsName)},
			nsName:        defaultNodeNsName,
			expectedError: fmt.Errorf("failed to list SriovNetworkNodePolicies, 'apiClient' parameter is empty"),
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var (
			testSettings   *clients.Settings
			runtimeObjects []runtime.Object
		)

		for _, networkNodeState := range testCase.testNetworkPolicy {
			runtimeObjects = append(runtimeObjects, networkNodeState)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(runtimeObjects)
		}

		err := CleanAllNetworkNodePolicies(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			netBuilders, err := ListPolicy(testSettings, testCase.nsName)
			assert.Nil(t, err)
			assert.Zero(t, len(netBuilders))
		}
	}
}
