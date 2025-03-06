//go:build integration
// +build integration

package integration

import (
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/deployment"
	"github.com/openshift-kni/eco-goinfra/pkg/namespace"
	"github.com/openshift-kni/eco-goinfra/pkg/pod"
	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
)

const (
	namespacePrefix = "ecogoinfra-deployment"
	containerImage  = "nginx:latest"
	timeoutDuration = time.Duration(60) * time.Second
)

func TestDeploymentCreate(t *testing.T) {
	t.Parallel()
	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace  = "egi-deployment-create-ns"
		deploymentName = "create-test"
	)

	// Setup namespace for the test
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

	deploymentBuilder := deployment.NewBuilder(client, deploymentName, testNamespace, map[string]string{
		"app": "test",
	}, *containerDefinition)

	// Create a deployment in the namespace
	_, err = deploymentBuilder.CreateAndWaitUntilReady(timeoutDuration)
	assert.Nil(t, err)

	// Check if the deployment was created
	deploymentBuilder, err = deployment.Pull(client, deploymentName, testNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, deploymentBuilder.Object)
}

func TestDeploymentDelete(t *testing.T) {
	t.Parallel()
	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace  = "egi-deployment-delete-ns"
		deploymentName = "delete-test"
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

	deploymentBuilder := deployment.NewBuilder(client, deploymentName, testNamespace, map[string]string{
		"app": "test",
	}, *containerDefinition)

	// Create a deployment in the namespace
	_, err = deploymentBuilder.CreateAndWaitUntilReady(timeoutDuration)
	assert.Nil(t, err)

	defer func() {
		// Delete the deployment
		err = deploymentBuilder.DeleteAndWait(timeoutDuration)
		assert.Nil(t, err)

		// Check if the deployment was deleted
		deploymentBuilder, err = deployment.Pull(client, deploymentName, testNamespace)
		assert.Equal(t, "deployment object delete-test does not exist in namespace "+testNamespace, err.Error())
	}()

	// Check if the deployment was created
	deploymentBuilder, err = deployment.Pull(client, deploymentName, testNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, deploymentBuilder.Object)
}

func TestDeploymentWithReplicas(t *testing.T) {
	t.Parallel()
	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace  = "egi-deployment-replicas-ns"
		deploymentName = "replicas-test"
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

	// Create the namespace
	namespaceBuilder, err = namespaceBuilder.Create()

	// Defer the deletion of the namespace
	defer func() {
		// Delete the namespace
		err := namespaceBuilder.Delete()
		assert.Nil(t, err)
	}()

	testContainerBuilder := pod.NewContainerBuilder("test", containerImage, []string{"sleep", "3600"})
	containerDefinition, err := testContainerBuilder.GetContainerCfg()
	assert.Nil(t, err)

	// Change the container default security context to something that is allowed in the test environment
	testContainerBuilder.WithSecurityContext(&corev1.SecurityContext{
		RunAsUser:  nil,
		RunAsGroup: nil,
	})

	deploymentBuilder := deployment.NewBuilder(client, deploymentName, testNamespace, map[string]string{
		"app": "test",
	}, *containerDefinition).WithReplicas(2)

	// Create a deployment in the namespace
	_, err = deploymentBuilder.CreateAndWaitUntilReady(timeoutDuration)
	assert.Nil(t, err)

	// Check if the deployment was created
	deploymentBuilder, err = deployment.Pull(client, deploymentName, testNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, deploymentBuilder.Object)

	// Check if the deployment has the correct number of replicas
	assert.Equal(t, int32(2), *deploymentBuilder.Object.Spec.Replicas)
}
