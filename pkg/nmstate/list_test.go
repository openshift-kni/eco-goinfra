package nmstate

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListPolicy(t *testing.T) {
	testCases := []struct {
		testPolicy    []*PolicyBuilder
		listOptions   []client.ListOptions
		runtimeObject bool
		expectedError error
		client        bool
	}{
		{
			testPolicy:    []*PolicyBuilder{buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())},
			expectedError: nil,
			client:        true,
			runtimeObject: true,
		},
		{
			testPolicy:    []*PolicyBuilder{buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())},
			expectedError: fmt.Errorf("failed to list sriov networks, 'apiClient' parameter is empty"),
			client:        false,
			runtimeObject: true,
		},
		{
			testPolicy:    []*PolicyBuilder{buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())},
			listOptions:   []client.ListOptions{{Continue: "test"}},
			client:        true,
			runtimeObject: true,
		},
		{
			testPolicy:    []*PolicyBuilder{buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())},
			listOptions:   []client.ListOptions{{Continue: "test"}, {Namespace: "test"}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
			runtimeObject: true,
		},
		{
			testPolicy:    []*PolicyBuilder{},
			expectedError: nil,
			client:        true,
			runtimeObject: false,
		},
	}
	for _, testCase := range testCases {
		var (
			testSettings   *clients.Settings
			runtimeObjects []runtime.Object
		)

		if testCase.runtimeObject {
			runtimeObjects = buildDummyPolicyObject()
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: v1TestSchemes,
			})
		}

		netBuilders, err := ListPolicy(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(netBuilders), len(testCase.testPolicy))
		}
	}
}

func TestCleanAllNMStatePolicies(t *testing.T) {
	testCases := []struct {
		testPolicy    []*PolicyBuilder
		listOptions   []client.ListOptions
		runtimeObject bool
		expectedError error
		client        bool
	}{
		{
			testPolicy:    []*PolicyBuilder{buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())},
			expectedError: nil,
			client:        true,
			runtimeObject: true,
		},
		{
			testPolicy:    []*PolicyBuilder{buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())},
			expectedError: fmt.Errorf("failed to list sriov networks, 'apiClient' parameter is empty"),
			client:        false,
			runtimeObject: true,
		},
		{
			testPolicy:    []*PolicyBuilder{buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())},
			listOptions:   []client.ListOptions{{Continue: "test"}},
			client:        true,
			runtimeObject: true,
		},
		{
			testPolicy:    []*PolicyBuilder{buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())},
			listOptions:   []client.ListOptions{{Continue: "test"}, {Namespace: "test"}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
			runtimeObject: true,
		},
		{
			testPolicy:    []*PolicyBuilder{},
			expectedError: nil,
			client:        true,
			runtimeObject: false,
		},
	}
	for _, testCase := range testCases {
		var (
			testSettings   *clients.Settings
			runtimeObjects []runtime.Object
		)

		if testCase.runtimeObject {
			runtimeObjects = buildDummyPolicyObject()
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: v1TestSchemes,
			})
		}

		err := CleanAllNMStatePolicies(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)
		netBuilders, err := ListPolicy(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, len(netBuilders), 0)
	}
}
