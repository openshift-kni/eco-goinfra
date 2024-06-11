package mco

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultLabelSelector = "test-label-selector"

func TestListMC(t *testing.T) {
	testCases := []struct {
		machineConfigs []*MCBuilder
		listOptions    []metav1.ListOptions
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
			listOptions:    []metav1.ListOptions{{LabelSelector: defaultLabelSelector}},
			expectedError:  nil,
			client:         true,
		},
		{
			machineConfigs: []*MCBuilder{buildValidMachineConfigTestBuilder(buildTestClientWithDummyMachineConfig())},
			listOptions:    []metav1.ListOptions{{LabelSelector: defaultLabelSelector}, {LabelSelector: defaultLabelSelector}},
			expectedError:  fmt.Errorf("error: more than one ListOptions was passed"),
			client:         true,
		},
		{
			machineConfigs: []*MCBuilder{buildValidMachineConfigTestBuilder(buildTestClientWithDummyMachineConfig())},
			listOptions:    []metav1.ListOptions{{LabelSelector: defaultLabelSelector}},
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
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.machineConfigs), len(mcBuilders))
		}
	}
}
