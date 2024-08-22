package bmh

import (
	"context"
	"fmt"
	"testing"
	"time"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestBareMetalHostList(t *testing.T) {
	testCases := []struct {
		BareMetalHosts []*BmhBuilder
		nsName         string
		listOptions    []goclient.ListOptions
		expectedError  error
		client         bool
	}{
		{

			BareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:         "test-namespace",
			expectedError:  nil,
			client:         true,
		},
		{
			BareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:         "",
			expectedError:  fmt.Errorf("failed to list bareMetalHosts, 'nsname' parameter is empty"),
			client:         true,
		},
		{
			BareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:         "test-namespace",
			listOptions:    []goclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			client:         true,
			expectedError:  nil,
		},
		{
			BareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:         "test-namespace",
			listOptions:    []goclient.ListOptions{{LabelSelector: labels.NewSelector()}, {LabelSelector: labels.NewSelector()}},
			expectedError:  fmt.Errorf("error: more than one ListOptions was passed"),
			client:         true,
		},
		{
			BareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:         "test-namespace",
			expectedError:  fmt.Errorf("failed to list bareMetalHosts, 'apiClient' parameter is empty"),
			client:         false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyBmHostObject(bmhv1alpha1.StateProvisioned),
				SchemeAttachers: testSchemes,
			})
		}

		bmhBuilders, err := List(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.BareMetalHosts), len(bmhBuilders))
		}
	}
}

func TestBareMetalHostListInAllNamespaces(t *testing.T) {
	testCases := []struct {
		bareMetalHosts []*BmhBuilder
		listOptions    []goclient.ListOptions
		expectedError  error
		client         bool
	}{
		{
			bareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			listOptions:    nil,
			expectedError:  nil,
			client:         true,
		},
		{
			bareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			listOptions:    []goclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			expectedError:  nil,
			client:         true,
		},
		{
			bareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			listOptions:    []goclient.ListOptions{{LabelSelector: labels.NewSelector()}, {LabelSelector: labels.NewSelector()}},
			expectedError:  fmt.Errorf("error: more than one ListOptions was passed"),
			client:         true,
		},
		{
			bareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			listOptions:    []goclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			expectedError:  fmt.Errorf("failed to list bareMetalHosts, 'apiClient' parameter is empty"),
			client:         false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildBareMetalHostTestClientWithDummyObject()
		}

		bmhBuilders, err := ListInAllNamespaces(testSettings, testCase.listOptions...)
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.bareMetalHosts), len(bmhBuilders))
		}
	}
}

func TestBareMetalWaitForAllBareMetalHostsInGoodOperationalState(t *testing.T) {
	testCases := []struct {
		BareMetalHosts   []*BmhBuilder
		nsName           string
		listOptions      []goclient.ListOptions
		operationalState bmhv1alpha1.OperationalStatus
		expectedError    error
		client           bool
		expectedStatus   bool
	}{
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			operationalState: bmhv1alpha1.OperationalStatusOK,
			listOptions:      nil,
			expectedError:    nil,
			expectedStatus:   true,
			client:           true,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			operationalState: bmhv1alpha1.OperationalStatusDelayed,
			expectedError:    context.DeadlineExceeded,
			listOptions:      nil,
			expectedStatus:   false,
			client:           true,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "",
			operationalState: bmhv1alpha1.OperationalStatusOK,
			expectedError:    fmt.Errorf("failed to list bareMetalHosts, 'nsname' parameter is empty"),
			expectedStatus:   false,
			listOptions:      nil,
			client:           true,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			operationalState: bmhv1alpha1.OperationalStatusOK,
			expectedError:    fmt.Errorf("failed to list bareMetalHosts, 'apiClient' parameter is empty"),
			expectedStatus:   false,
			listOptions:      nil,
			client:           false,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			operationalState: bmhv1alpha1.OperationalStatusOK,
			expectedError:    nil,
			expectedStatus:   true,
			listOptions:      []goclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			client:           true,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			operationalState: bmhv1alpha1.OperationalStatusOK,
			expectedError:    fmt.Errorf("error: more than one ListOptions was passed"),
			expectedStatus:   false,
			listOptions: []goclient.ListOptions{
				{LabelSelector: labels.NewSelector()}, {LabelSelector: labels.NewSelector()}},
			client: true,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyBmHostObject(bmhv1alpha1.StateProvisioned, testCase.operationalState),
				SchemeAttachers: testSchemes,
			})
		}

		status, err := WaitForAllBareMetalHostsInGoodOperationalState(
			testSettings, testCase.nsName, 1*time.Second, testCase.listOptions...)
		assert.Equal(t, err, testCase.expectedError)
		assert.Equal(t, status, testCase.expectedStatus)
	}
}
