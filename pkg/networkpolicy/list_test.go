package networkpolicy

import (
	"errors"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestList(t *testing.T) {
	generateNetworkPolicy := func(name, namespace string) *netv1.NetworkPolicy {
		return &netv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    map[string]string{"demo": name},
			},
		}
	}

	testCases := []struct {
		networkPolicyExists        bool
		testNamespace              string
		listOptions                []metav1.ListOptions
		expetedError               error
		expectedNumNetworkPolicies int
	}{
		{ // NetworkPolicy does exist
			networkPolicyExists:        true,
			testNamespace:              "test-namespace",
			expetedError:               nil,
			listOptions:                []metav1.ListOptions{},
			expectedNumNetworkPolicies: 1,
		},
		{ // NetworkPolicy does not exist
			networkPolicyExists:        false,
			testNamespace:              "test-namespace",
			expetedError:               nil,
			listOptions:                []metav1.ListOptions{},
			expectedNumNetworkPolicies: 0,
		},
		{ // Missing namespace parameter
			networkPolicyExists:        true,
			testNamespace:              "",
			expetedError:               errors.New("failed to list networkpolicies, 'nsname' parameter is empty"),
			listOptions:                []metav1.ListOptions{},
			expectedNumNetworkPolicies: 0,
		},
		{ // More than one ListOptions was passed
			networkPolicyExists:        true,
			testNamespace:              "test-namespace",
			expetedError:               errors.New("error: more than one ListOptions was passed"),
			listOptions:                []metav1.ListOptions{{}, {}},
			expectedNumNetworkPolicies: 0,
		},
		{ // Valid number of list options
			networkPolicyExists:        true,
			testNamespace:              "test-namespace",
			expetedError:               nil,
			listOptions:                []metav1.ListOptions{{}},
			expectedNumNetworkPolicies: 1,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.networkPolicyExists {
			runtimeObjects = append(runtimeObjects, generateNetworkPolicy("test-networkpolicy", "test-namespace"))
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		networkPolicyList, err := List(testSettings, testCase.testNamespace, testCase.listOptions...)
		assert.Equal(t, testCase.expetedError, err)
		assert.Equal(t, testCase.expectedNumNetworkPolicies, len(networkPolicyList))
	}
}
