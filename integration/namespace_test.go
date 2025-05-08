//go:build integration
// +build integration

package integration

import (
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/namespace"
	"github.com/openshift-kni/eco-goinfra/pkg/pod"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestNamespaceCreate(t *testing.T) {
	t.Parallel()

	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace = CreateRandomNamespace()
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

	// Check if the namespace was created
	namespaceBuilder, err := namespace.Pull(client, testNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, namespaceBuilder.Object)
}

func TestNamespaceDelete(t *testing.T) {
	t.Parallel()

	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace = "egi-namespace-delete-test"
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

	// Delete the namespace
	err := namespaceBuilder.Delete()
	assert.Nil(t, err)

	// Check if the namespace was deleted
	namespaceBuilder, err = namespace.Pull(client, testNamespace)
	assert.Nil(t, err)
}

func TestNamespaceCleanObjects(t *testing.T) {
	t.Parallel()

	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace = "egi-namespace-clean-objects-test"
	)

	// Create a namespace in the cluster using the namespaces package
	namespaceBuilder := namespace.NewBuilder(client, testNamespace)
	assert.Nil(t, PreEmptiveNamespaceDeleteAndSetup(testNamespace, namespaceBuilder))

	// Create some test pods in the namespace
	podBuilder := pod.NewBuilder(client, "test-pod", testNamespace, "nginx:latest")

	// Create a pod in the namespace
	_, err := podBuilder.Create()
	assert.Nil(t, err)

	// Defer the deletion of the namespace
	defer func() {
		// Delete the namespace
		err := namespaceBuilder.Delete()
		assert.Nil(t, err)
	}()

	// GVR to clean
	gvrObjects := []schema.GroupVersionResource{pod.GetGVR()}

	// Clean objects in the namespace
	err = namespaceBuilder.CleanObjects(5*time.Second, gvrObjects...)
	assert.Nil(t, err)

	// Check if the objects were cleaned
	_, err = pod.Pull(client, "test-pod", testNamespace)
	assert.NotNil(t, err)
	assert.Equal(t, "pod object test-pod does not exist in namespace egi-namespace-clean-objects-test", err.Error())
}
