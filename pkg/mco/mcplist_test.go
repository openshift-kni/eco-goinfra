package mco

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const defaultMCPListLabel = "test-label-value"

func TestListMCP(t *testing.T) {
	testCases := []struct {
		mcPools       []*MCPBuilder
		listOptions   []runtimeclient.ListOptions
		expectedError error
		client        bool
	}{
		{
			mcPools:       []*MCPBuilder{buildValidMCPTestBuilder(buildTestClientWithDummyMCP())},
			listOptions:   nil,
			expectedError: nil,
			client:        true,
		},
		{
			mcPools:       []*MCPBuilder{buildValidMCPTestBuilder(buildTestClientWithDummyMCP())},
			listOptions:   []runtimeclient.ListOptions{{Continue: "test"}},
			expectedError: nil,
			client:        true,
		},
		{
			mcPools:       []*MCPBuilder{buildValidMCPTestBuilder(buildTestClientWithDummyMCP())},
			listOptions:   []runtimeclient.ListOptions{{Continue: "test"}, {Continue: "test"}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			mcPools:       []*MCPBuilder{buildValidMCPTestBuilder(buildTestClientWithDummyMCP())},
			listOptions:   []runtimeclient.ListOptions{{Continue: "test"}},
			expectedError: fmt.Errorf("failed to list MachineConfigPools, 'apiClient' parameter is empty"),
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithDummyMCP()
		}

		mcBuilders, err := ListMCP(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.mcPools), len(mcBuilders))
		}
	}
}

func TestListMCPByMachineConfigSelector(t *testing.T) {
	testCases := []struct {
		client        bool
		hasLabel      bool
		expectedError error
	}{
		{
			client:        true,
			hasLabel:      true,
			expectedError: nil,
		},
		{
			client:        false,
			hasLabel:      true,
			expectedError: fmt.Errorf("failed to list MachineConfigPools, 'apiClient' parameter is empty"),
		},
		{
			client:   true,
			hasLabel: false,
			expectedError: fmt.Errorf(
				"cannot find MachineConfigPool that targets machineConfig with label: %s", defaultMCPListLabel),
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			mcp := buildDummyMCP(defaultMCPName)

			if testCase.hasLabel {
				mcp.Spec.MachineConfigSelector = &metav1.LabelSelector{
					MatchLabels: map[string]string{"test": defaultMCPListLabel},
				}
			}

			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  []runtime.Object{mcp},
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := ListMCPByMachineConfigSelector(testSettings, defaultMCPListLabel)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, testBuilder)
			assert.Equal(t, defaultMCPName, testBuilder.Definition.Name)
		}
	}
}

func TestListMCPWaitToBeStableFor(t *testing.T) {
	testCases := []struct {
		client        bool
		stable        bool
		expectedError error
	}{
		{
			client:        true,
			stable:        true,
			expectedError: nil,
		},
		{
			client:        false,
			stable:        true,
			expectedError: fmt.Errorf("failed to list MachineConfigPools, 'apiClient' parameter is empty"),
		},
		{
			client:        true,
			stable:        false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			mcp := buildDummyMCP(defaultMCPName)

			if !testCase.stable {
				mcp.Status.DegradedMachineCount = 1
			}

			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  []runtime.Object{mcp},
				SchemeAttachers: testSchemes,
			})
		}

		err := ListMCPWaitToBeStableFor(testSettings, 500*time.Millisecond, time.Second)
		assert.Equal(t, testCase.expectedError, err)
	}
}
