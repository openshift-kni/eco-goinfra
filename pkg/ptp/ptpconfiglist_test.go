package ptp

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListPtpConfigs(t *testing.T) {
	testCases := []struct {
		ptpConfigs    []*PtpConfigBuilder
		listOptions   []runtimeclient.ListOptions
		client        bool
		expectedError error
	}{
		{
			ptpConfigs:    []*PtpConfigBuilder{buildValidPtpConfigBuilder(buildTestClientWithDummyPtpConfig())},
			listOptions:   nil,
			client:        true,
			expectedError: nil,
		},
		{
			ptpConfigs:    []*PtpConfigBuilder{buildValidPtpConfigBuilder(buildTestClientWithDummyPtpConfig())},
			listOptions:   []runtimeclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			client:        true,
			expectedError: nil,
		},
		{
			ptpConfigs: []*PtpConfigBuilder{buildValidPtpConfigBuilder(buildTestClientWithDummyPtpConfig())},
			listOptions: []runtimeclient.ListOptions{
				{LabelSelector: labels.NewSelector()},
				{LabelSelector: labels.NewSelector()},
			},
			client:        true,
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
		},
		{
			ptpConfigs:    []*PtpConfigBuilder{buildValidPtpConfigBuilder(buildTestClientWithDummyPtpConfig())},
			listOptions:   nil,
			client:        false,
			expectedError: fmt.Errorf("failed to list PtpConfigs, 'apiClient' parameter is nil"),
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithDummyPtpConfig()
		}

		builders, err := ListPtpConfigs(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, len(testCase.ptpConfigs), len(builders))
		}
	}
}
