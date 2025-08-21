//go:build integration
// +build integration

package integration

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/configmap"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/namespace"
	"github.com/stretchr/testify/assert"
)

func TestConfigmapCreate(t *testing.T) {
	t.Parallel()
	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace = "egi-configmap-create-ns"
		configmapName = "create-test"
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

	configmapBuilder := configmap.NewBuilder(client, configmapName, testNamespace)

	// Create a configmap in the namespace
	_, err := configmapBuilder.Create()
	assert.Nil(t, err)

	// Check if the configmap was created
	configmapBuilder, err = configmap.Pull(client, configmapName, testNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, configmapBuilder.Object)
}

func TestConfigmapDelete(t *testing.T) {
	t.Parallel()
	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace = "egi-configmap-delete-ns"
		configmapName = "delete-test"
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

	configmapBuilder := configmap.NewBuilder(client, configmapName, testNamespace)

	// Create a configmap in the namespace
	_, err := configmapBuilder.Create()
	assert.Nil(t, err)

	// Check if the configmap was created
	configmapBuilder, err = configmap.Pull(client, configmapName, testNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, configmapBuilder.Object)

	// Delete the configmap
	err = configmapBuilder.Delete()
	assert.Nil(t, err)

	// Check if the configmap was deleted
	configmapBuilder, err = configmap.Pull(client, configmapName, testNamespace)
	assert.NotNil(t, err)
	assert.Nil(t, configmapBuilder)
}
