//go:build integration
// +build integration

package integration

import (
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/namespace"
	"github.com/openshift-kni/eco-goinfra/pkg/networkpolicy"
	"github.com/stretchr/testify/assert"
)

func TestNetworkPolicyCreate(t *testing.T) {
	t.Parallel()

	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace     = "egi-networkpolicy-create-test"
		testNetworkPolicy = "egi-networkpolicy-create-test"
	)

	// Create a namespace in the cluster using the namespaces package
	namespaceBuilder := namespace.NewBuilder(client, testNamespace)
	assert.Nil(t, PreEmptiveNamespaceDeleteAndSetup(testNamespace, namespaceBuilder))

	// Defer the deletion of the namespace
	defer func() {
		// Delete the namespace
		err := namespaceBuilder.Delete()
		assert.Nil(t, err)
	}()

	// Create a network policy in the cluster using the network policies package
	networkPolicyBuilder := networkpolicy.NewNetworkPolicyBuilder(client, testNetworkPolicy, testNamespace)
	assert.NotNil(t, networkPolicyBuilder)

	// Add a pod selector to the network policy
	networkPolicyBuilder.WithPodSelector(map[string]string{"app": "test"})
	assert.NotNil(t, networkPolicyBuilder)

	// Create the network policy
	result, err := networkPolicyBuilder.Create()
	assert.Nil(t, err)
	assert.NotNil(t, result)

	// Check if the network policy was created
	networkPolicyBuilder, err = networkpolicy.Pull(client, testNetworkPolicy, testNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, networkPolicyBuilder.Object)
	assert.Equal(t, testNetworkPolicy, networkPolicyBuilder.Object.GetName())
	assert.Equal(t, testNamespace, networkPolicyBuilder.Object.GetNamespace())
	assert.Equal(t, networkPolicyBuilder.Object.Spec.PodSelector.MatchLabels["app"], "test")
}

func TestNetworkPolicyDelete(t *testing.T) {
	t.Parallel()

	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace     = "egi-networkpolicy-delete-test"
		testNetworkPolicy = "egi-networkpolicy-delete-test"
	)

	// Create a namespace in the cluster using the namespaces package
	namespaceBuilder := namespace.NewBuilder(client, testNamespace)
	assert.Nil(t, PreEmptiveNamespaceDeleteAndSetup(testNamespace, namespaceBuilder))

	// Defer the deletion of the namespace
	defer func() {
		// Delete the namespace
		err := namespaceBuilder.Delete()
		assert.Nil(t, err)
	}()

	// Create a network policy in the cluster using the network policies package
	networkPolicyBuilder := networkpolicy.NewNetworkPolicyBuilder(client, testNetworkPolicy, testNamespace)
	assert.NotNil(t, networkPolicyBuilder)

	// Add a pod selector to the network policy
	networkPolicyBuilder.WithPodSelector(map[string]string{"app": "test"})
	assert.NotNil(t, networkPolicyBuilder)

	// Create the network policy
	result, err := networkPolicyBuilder.Create()
	assert.Nil(t, err)
	assert.NotNil(t, result)

	// Check if the network policy was created
	networkPolicyBuilder, err = networkpolicy.Pull(client, testNetworkPolicy, testNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, networkPolicyBuilder.Object)
	assert.Equal(t, testNetworkPolicy, networkPolicyBuilder.Object.GetName())
	assert.Equal(t, testNamespace, networkPolicyBuilder.Object.GetNamespace())
	assert.Equal(t, networkPolicyBuilder.Object.Spec.PodSelector.MatchLabels["app"], "test")

	// Delete the network policy
	err = networkPolicyBuilder.Delete()
	assert.Nil(t, err)

	// Check if the network policy was deleted
	networkPolicyBuilder, err = networkpolicy.Pull(client, testNetworkPolicy, testNamespace)
	assert.NotNil(t, err)
}
