package nodes

import (
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultNodeName         = "test-node"
	defaultNodeLabel        = "node-role.kubernetes.io/control-plane"
	defaultExternalNetworks = `{"ipv4":"10.0.0.0/8","ipv6":"fd00::/8"}`
	defaultExternalIPv4     = "10.0.0.0/8"
)

// buildDummyNode returns a Node with the provided name.
func buildDummyNode(name string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestClientWithDummyNode returns a client with a dummy node.
func buildTestClientWithDummyNode() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{buildDummyNode(defaultNodeName)},
	})
}

func buildValidNodeTestBuilder(apiClient *clients.Settings) *Builder {
	return newNodeBuilder(apiClient, defaultNodeName)
}

// newNodeBuilder creates a new Builder instances for testing purposes.
func newNodeBuilder(apiClient *clients.Settings, name string) *Builder {
	if apiClient == nil {
		return nil
	}

	builder := Builder{
		apiClient:  apiClient.K8sClient,
		Definition: buildDummyNode(name),
	}

	return &builder
}
