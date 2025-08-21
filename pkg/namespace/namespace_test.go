package namespace

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestNamespaceRemoveLabels(t *testing.T) {
	testCases := []struct {
		labels        map[string]string
		expectedError string
	}{
		{
			labels:        map[string]string{"key1": "value1"},
			expectedError: "",
		},
		{
			labels:        map[string]string{},
			expectedError: "labels to be removed cannot be empty",
		},
	}
	for _, testCase := range testCases {
		testBuilder := buildValidTestNamespaceBuilderWithClient([]runtime.Object{}).WithMultipleLabels(testCase.labels)
		testBuilder = testBuilder.RemoveLabels(testCase.labels)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, 0, len(testBuilder.Definition.Labels), "Expected labels to be removed")
		}
	}
}

func buildValidTestNamespaceBuilderWithClient(objects []runtime.Object) *Builder {
	fakeClient := k8sfake.NewSimpleClientset(objects...)

	return NewBuilder(&clients.Settings{
		K8sClient:       fakeClient,
		CoreV1Interface: fakeClient.CoreV1(),
		AppsV1Interface: fakeClient.AppsV1(),
	}, "test-namespace")
}
