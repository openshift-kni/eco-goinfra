package idms

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestListImageDigestMirrorSets(t *testing.T) {
	testCases := []struct {
		idmsCount     int
		testClient    *clients.Settings
		options       []runtimeClient.ListOptions
		expectedError error
	}{
		{
			idmsCount:     0,
			testClient:    clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes}),
			options:       []runtimeClient.ListOptions{},
			expectedError: nil,
		},
		{
			idmsCount: 1,
			testClient: clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  []runtime.Object{generateImagDigestMirrorSet()},
				SchemeAttachers: testSchemes,
			}),
			options:       []runtimeClient.ListOptions{},
			expectedError: nil,
		},
		{
			idmsCount: 0,
			testClient: clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  []runtime.Object{generateImagDigestMirrorSet()},
				SchemeAttachers: testSchemes,
			}),
			options: []runtimeClient.ListOptions{
				{
					LabelSelector: labels.Everything(),
				},
				{
					Namespace: "test",
				},
			},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
		},
		{
			idmsCount:     0,
			testClient:    nil,
			options:       []runtimeClient.ListOptions{},
			expectedError: fmt.Errorf("apiClient cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		idmsBuilders, err := ListImageDigestMirrorSets(testCase.testClient, testCase.options...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.idmsCount, len(idmsBuilders))
		}
	}
}
