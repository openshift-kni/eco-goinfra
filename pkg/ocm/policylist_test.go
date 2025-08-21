package ocm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	policiesv1 "open-cluster-management.io/governance-policy-propagator/api/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListPoliciesInAllNamespaces(t *testing.T) {
	testCases := []struct {
		policies      []*PolicyBuilder
		listOptions   []runtimeclient.ListOptions
		expectedError error
		client        bool
	}{
		{
			policies: []*PolicyBuilder{
				buildValidPolicyTestBuilder(buildTestClientWithDummyPolicy()),
			},
			listOptions:   nil,
			expectedError: nil,
			client:        true,
		},
		{
			policies: []*PolicyBuilder{
				buildValidPolicyTestBuilder(buildTestClientWithDummyPolicy()),
			},
			listOptions:   []runtimeclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			expectedError: nil,
			client:        true,
		},
		{
			policies: []*PolicyBuilder{
				buildValidPolicyTestBuilder(buildTestClientWithDummyPolicy()),
			},
			listOptions: []runtimeclient.ListOptions{
				{LabelSelector: labels.NewSelector()},
				{LabelSelector: labels.NewSelector()},
			},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			policies: []*PolicyBuilder{
				buildValidPolicyTestBuilder(buildTestClientWithDummyPolicy()),
			},
			listOptions:   []runtimeclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			expectedError: fmt.Errorf("failed to list policies, 'apiClient' parameter is nil"),
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithDummyPolicy()
		}

		builders, err := ListPoliciesInAllNamespaces(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.policies), len(builders))
		}
	}
}

func TestWaitForAllPoliciesComplianceState(t *testing.T) {
	testCases := []struct {
		compliant     bool
		client        bool
		listOptions   []runtimeclient.ListOptions
		expectedError error
	}{
		{
			compliant:     true,
			client:        true,
			listOptions:   nil,
			expectedError: nil,
		},
		{
			compliant:     false,
			client:        true,
			listOptions:   nil,
			expectedError: context.DeadlineExceeded,
		},
		{
			compliant:     true,
			client:        false,
			listOptions:   nil,
			expectedError: fmt.Errorf("failed to wait for policies compliance state, 'apiClient' parameter is nil"),
		},
		{
			compliant: true,
			client:    true,
			listOptions: []runtimeclient.ListOptions{
				{LabelSelector: labels.NewSelector()},
				{LabelSelector: labels.NewSelector()},
			},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			policy := buildDummyPolicy(defaultPolicyName, defaultPolicyNsName)

			if testCase.compliant {
				policy.Status.ComplianceState = policiesv1.Compliant
			}

			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  []runtime.Object{policy},
				SchemeAttachers: policyTestSchemes,
			})
		}

		err := WaitForAllPoliciesComplianceState(testSettings, policiesv1.Compliant, time.Second, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)
	}
}
