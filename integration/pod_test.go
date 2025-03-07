//go:build integration
// +build integration

package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/namespace"
	"github.com/openshift-kni/eco-goinfra/pkg/pod"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestPodCreate(t *testing.T) {
	t.Parallel()
	client := clients.New("")
	assert.NotNil(t, client)

	const (
		testNamespace = "egi-pod-create-ns"
		podName       = "create-test"
	)

	// Create a namespace in the cluster using the namespaces package
	namespaceBuilder := namespace.NewBuilder(client, testNamespace)
	assert.NotNil(t, namespaceBuilder)

	// Preemptively delete the namespace before the test
	err := namespaceBuilder.DeleteAndWait(time.Duration(30) * time.Second)
	assert.Nil(t, err)

	// Create the namespace
	namespaceBuilder, err = namespaceBuilder.Create()
	assert.Nil(t, err)

	// Defer the deletion of the namespace
	defer func() {
		// Delete the namespace
		err := namespaceBuilder.Delete()
		assert.Nil(t, err)
	}()

	testContainerBuilder := pod.NewContainerBuilder("test", containerImage, []string{"sleep", "3600"}).
		WithSecurityContext(&corev1.SecurityContext{
			RunAsUser:  nil,
			RunAsGroup: nil,
		})
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

	const (
		testNamespace = "egi-pod-delete-ns"
		podName       = "delete-test"
	)

	// Create a namespace in the cluster using the namespaces package
	namespaceBuilder := namespace.NewBuilder(client, testNamespace)
	assert.NotNil(t, namespaceBuilder)

	// Preemptively delete the namespace before the test
	err := namespaceBuilder.DeleteAndWait(time.Duration(30) * time.Second)
	assert.Nil(t, err)

	// Create the namespace
	namespaceBuilder, err = namespaceBuilder.Create()
	assert.Nil(t, err)

	// Defer the deletion of the namespace
	defer func() {
		// Delete the namespace
		err := namespaceBuilder.Delete()
		assert.Nil(t, err)
	}()

	testContainerBuilder := pod.NewContainerBuilder("test", containerImage, []string{"sleep", "3600"}).
		WithSecurityContext(&corev1.SecurityContext{
			RunAsUser:  nil,
			RunAsGroup: nil,
		})
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

	const (
		testNamespace = "egi-pod-exec-ns"
		podName       = "exec-test"
	)

	// Create a namespace in the cluster using the namespaces package
	namespaceBuilder := namespace.NewBuilder(client, testNamespace)
	assert.NotNil(t, namespaceBuilder)

	// Preemptively delete the namespace before the test
	err := namespaceBuilder.DeleteAndWait(time.Duration(30) * time.Second)
	assert.Nil(t, err)

	// Create the namespace
	namespaceBuilder, err = namespaceBuilder.Create()
	assert.Nil(t, err)

	// Defer the deletion of the namespace
	defer func() {
		// Delete the namespace
		err := namespaceBuilder.Delete()
		assert.Nil(t, err)
	}()

	testContainerBuilder := pod.NewContainerBuilder("test", containerImage, []string{"sleep", "3600"}).
		WithSecurityContext(&corev1.SecurityContext{
			RunAsUser:  nil,
			RunAsGroup: nil,
		})
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
	buffer, err := podBuilder.ExecCommand([]string{"sh", "-c", "echo test | sha256sum"})
	assert.Nil(t, err)
	assert.Equal(t, "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2  -\r\n", buffer.String())
}
