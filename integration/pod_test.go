//go:build integration
// +build integration

package integration

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/namespace"
	"github.com/openshift-kni/eco-goinfra/pkg/pod"
	"github.com/stretchr/testify/assert"
)

func TestPodCreate(t *testing.T) {
	t.Parallel()
	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace = CreateRandomNamespace()
		podName       = "create-test"
	)

	// Create a namespace in the cluster using the namespaces package
	namespaceBuilder, err := namespace.NewBuilder(client, testNamespace).Create()
	assert.Nil(t, err)

	// Defer the deletion of the namespace
	defer func() {
		// Delete the namespace
		err := namespaceBuilder.Delete()
		assert.Nil(t, err)
	}()

	testContainerBuilder := pod.NewContainerBuilder("test", containerImage, []string{"sleep", "3600"})
	containerDefinition, err := testContainerBuilder.GetContainerCfg()
	assert.Nil(t, err)

	podBuilder := pod.NewBuilder(client, podName, testNamespace, containerImage)
	podBuilder = podBuilder.RedefineDefaultContainer(*containerDefinition)

	// Create a pod in the namespace
	_, err = podBuilder.CreateAndWaitUntilRunning(timeoutDuration)
	assert.Nil(t, err)

	// Check if the pod was created
	podBuilder, err = pod.Pull(client, podName, testNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, podBuilder.Object)
}

func TestPodDelete(t *testing.T) {
	t.Parallel()
	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace = CreateRandomNamespace()
		podName       = "delete-test"
	)

	// Create a namespace in the cluster using the namespaces package
	namespaceBuilder, err := namespace.NewBuilder(client, testNamespace).Create()
	assert.Nil(t, err)

	// Defer the deletion of the namespace
	defer func() {
		// Delete the namespace
		err := namespaceBuilder.Delete()
		assert.Nil(t, err)
	}()

	testContainerBuilder := pod.NewContainerBuilder("test", containerImage, []string{"sleep", "3600"})
	containerDefinition, err := testContainerBuilder.GetContainerCfg()
	assert.Nil(t, err)

	podBuilder := pod.NewBuilder(client, podName, testNamespace, containerImage)
	podBuilder = podBuilder.RedefineDefaultContainer(*containerDefinition)

	// Create a pod in the namespace
	_, err = podBuilder.CreateAndWaitUntilRunning(timeoutDuration)
	assert.Nil(t, err)

	defer func() {
		_, err = podBuilder.DeleteAndWait(timeoutDuration)
		assert.Nil(t, err)

		// Check if the pod was deleted
		podBuilder, err = pod.Pull(client, podName, testNamespace)
		assert.EqualError(t, err, fmt.Sprintf("pod object %s does not exist in namespace %s", podName, testNamespace))
	}()

	// Check if the pod was created
	podBuilder, err = pod.Pull(client, podName, testNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, podBuilder.Object)
}

func TestPodExecCommand(t *testing.T) {
	t.Parallel()
	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace = CreateRandomNamespace()
		podName       = "exec-test"
	)

	// Create a namespace in the cluster using the namespaces package
	namespaceBuilder, err := namespace.NewBuilder(client, testNamespace).Create()
	assert.Nil(t, err)

	// Defer the deletion of the namespace
	defer func() {
		// Delete the namespace
		err := namespaceBuilder.Delete()
		assert.Nil(t, err)
	}()

	testContainerBuilder := pod.NewContainerBuilder("test", containerImage, []string{"sleep", "3600"})
	containerDefinition, err := testContainerBuilder.GetContainerCfg()
	assert.Nil(t, err)

	podBuilder := pod.NewBuilder(client, podName, testNamespace, containerImage)
	podBuilder = podBuilder.RedefineDefaultContainer(*containerDefinition)

	// Create a pod in the namespace
	podBuilder, err = podBuilder.CreateAndWaitUntilRunning(timeoutDuration)
	assert.Nil(t, err)

	// Check if the pod was created
	podBuilder, err = pod.Pull(client, podName, testNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, podBuilder.Object)

	// Execute a command in the pod
	buffer, err := podBuilder.ExecCommand([]string{"sh", "-c", "echo f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2"})
	assert.Nil(t, err)
	assert.Equal(t, "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2\r\n", buffer.String())
}
