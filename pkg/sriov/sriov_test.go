package sriov

import (
	"testing"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPull(t *testing.T) {
	generateNetwork := func(name, namespace string) *srIovV1.SriovNetwork {
		return &srIovV1.SriovNetwork{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: srIovV1.SriovNetworkSpec{},
		}
	}

	testCases := []struct {
		networkName         string
		networkNamespace    string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			networkName:         "test1",
			networkNamespace:    "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: true,
		},
		{
			networkName:         "test2",
			networkNamespace:    "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "sriovnetwork object test2 doesn't exist in namespace test-namespace",
		},
		{
			networkName:         "",
			networkNamespace:    "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "sriovnetwork 'name' cannot be empty",
		},
		{
			networkName:         "test3",
			networkNamespace:    "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "sriovnetwork 'namespace' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testNetwork := generateNetwork(testCase.networkName, testCase.networkNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testNetwork)
		}

		testSettings = clients.GetTestClients(runtimeObjects)
		// Test the Pull method
		builderResult, err := PullNetwork(testSettings, testNetwork.Name, testNetwork.Namespace)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			// Check the error message
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testNetwork.Name, builderResult.Object.Name)
			assert.Equal(t, testNetwork.Namespace, builderResult.Object.Namespace)
		}
	}
}
