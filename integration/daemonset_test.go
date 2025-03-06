//go:build integration
// +build integration

package integration

import (
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/daemonset"
	"github.com/openshift-kni/eco-goinfra/pkg/namespace"
	"github.com/stretchr/testify/assert"
)

func TestDaemonsetCreate(t *testing.T) {
	t.Parallel()
	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace = "egi-daemonset-create-ns"
		daemonsetName = "create-test"
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

	containerDefinition, err := CreateTestContainerDefinition("test", containerImage, []string{"sleep", "3600"})
	assert.Nil(t, err)

	daemonsetBuilder := daemonset.NewBuilder(client, daemonsetName, testNamespace, map[string]string{
		"app": "test",
	}, *containerDefinition)

	// Create a daemonset in the namespace
	_, err = daemonsetBuilder.CreateAndWaitUntilReady(timeoutDuration)
	assert.Nil(t, err)

	// Check if the daemonset was created
	daemonsetBuilder, err = daemonset.Pull(client, daemonsetName, testNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, daemonsetBuilder.Object)
}

func TestDaemonsetDelete(t *testing.T) {
	t.Parallel()
	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace = "egi-daemonset-delete-ns"
		daemonsetName = "delete-test"
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

	containerDefinition, err := CreateTestContainerDefinition("test", containerImage, []string{"sleep", "3600"})
	assert.Nil(t, err)

	daemonsetBuilder := daemonset.NewBuilder(client, daemonsetName, testNamespace, map[string]string{
		"app": "test",
	}, *containerDefinition)

	// Create a daemonset in the namespace
	_, err = daemonsetBuilder.CreateAndWaitUntilReady(timeoutDuration)
	assert.Nil(t, err)

	// Check if the daemonset was created
	daemonsetBuilder, err = daemonset.Pull(client, daemonsetName, testNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, daemonsetBuilder.Object)

	// Delete the daemonset
	err = daemonsetBuilder.Delete()
	assert.Nil(t, err)

	// Check if the daemonset was deleted
	daemonsetBuilder, err = daemonset.Pull(client, daemonsetName, testNamespace)
	assert.NotNil(t, err)
	assert.Nil(t, daemonsetBuilder)
}
