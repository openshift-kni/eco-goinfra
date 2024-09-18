package networkpolicy

import (
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
		networkPolicyExists bool
	}{
		{
			networkPolicyExists: true,
		},
		{
			networkPolicyExists: false,
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

		networkPolicyList, err := List(testSettings, "test-namespace")
		assert.NoError(t, err)

		if testCase.networkPolicyExists {
			assert.Len(t, networkPolicyList, 1)
		} else {
			assert.Nil(t, networkPolicyList)
		}
	}
}
