package mco

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListMC(t *testing.T) {
	testCases := []struct {
		machineConfigs []*MCBuilder
		listOptions    []runtimeclient.ListOptions
		expectedError  error
		client         bool
	}{
		{
			machineConfigs: []*MCBuilder{buildValidMachineConfigTestBuilder(buildTestClientWithDummyMachineConfig())},
			listOptions:    nil,
			expectedError:  nil,
			client:         true,
		},
		{
			machineConfigs: []*MCBuilder{buildValidMachineConfigTestBuilder(buildTestClientWithDummyMachineConfig())},
			listOptions:    []runtimeclient.ListOptions{{Continue: "test"}},
			expectedError:  nil,
			client:         true,
		},
		{
			machineConfigs: []*MCBuilder{buildValidMachineConfigTestBuilder(buildTestClientWithDummyMachineConfig())},
			listOptions:    []runtimeclient.ListOptions{{Continue: "test"}, {Continue: "test"}},
			expectedError:  fmt.Errorf("error: more than one ListOptions was passed"),
			client:         true,
		},
		{
			machineConfigs: []*MCBuilder{buildValidMachineConfigTestBuilder(buildTestClientWithDummyMachineConfig())},
			listOptions:    []runtimeclient.ListOptions{{Continue: "test"}},
			expectedError:  fmt.Errorf("failed to list MachineConfigs, 'apiClient' parameter is empty"),
			client:         false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithDummyMachineConfig()
		}

		mcBuilders, err := ListMC(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.machineConfigs), len(mcBuilders))
		}
	}
}
