//go:build integration
// +build integration

package integration

import (
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/deployment"
	"github.com/openshift-kni/eco-goinfra/pkg/namespace"
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

	// Create a random namespace for the test
	randomNamespace := generateRandomNamespace(namespacePrefix)

	// Create the namespace in the cluster using the namespaces package
	_, err := namespace.NewBuilder(client, randomNamespace).Create()
	assert.Nil(t, err)

	// Defer the deletion of the namespace
	defer func() {
		// Delete the namespace
		err := namespace.NewBuilder(client, randomNamespace).Delete()
		assert.Nil(t, err)
	}()

	var deploymentName = "create-test"

	builder := deployment.NewBuilder(client, deploymentName, randomNamespace, map[string]string{
		"app": "test",
	}, corev1.Container{
		Name:  "test",
		Image: containerImage,
		Command: []string{
			"sleep",
			"3600",
		},
	})

	// Create a deployment in the namespace
	_, err = builder.CreateAndWaitUntilReady(timeoutDuration)
	assert.Nil(t, err)

	// Check if the deployment was created
	builder, err = deployment.Pull(client, deploymentName, randomNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, builder.Object)
}

func TestDeploymentDelete(t *testing.T) {
	t.Parallel()
	client := clients.New("")

	// Create a random namespace for the test
	randomNamespace := generateRandomNamespace(namespacePrefix)

	// Create the namespace in the cluster using the namespaces package
	_, err := namespace.NewBuilder(client, randomNamespace).Create()
	assert.Nil(t, err)

	// Defer the deletion of the namespace
	defer func() {
		// Delete the namespace
		err := namespace.NewBuilder(client, randomNamespace).Delete()
		assert.Nil(t, err)
	}()

	var deploymentName = "delete-test"

	builder := deployment.NewBuilder(client, deploymentName, randomNamespace, map[string]string{
		"app": "test",
	}, corev1.Container{
		Name:  "test",
		Image: containerImage,
		Command: []string{
			"sleep",
			"3600",
		},
	})

	// Create a deployment in the namespace
	_, err = builder.CreateAndWaitUntilReady(timeoutDuration)
	assert.Nil(t, err)

	// Check if the deployment was created
	builder, err = deployment.Pull(client, deploymentName, randomNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, builder.Object)

	// Delete the deployment
	err = builder.DeleteAndWait(timeoutDuration)
	assert.Nil(t, err)

	// Check if the deployment was deleted
	builder, err = deployment.Pull(client, deploymentName, randomNamespace)
	assert.Equal(t, "deployment object delete-test does not exist in namespace "+randomNamespace, err.Error())
}

func TestDeploymentWithReplicas(t *testing.T) {
	t.Parallel()
	client := clients.New("")

	// Create a random namespace for the test
	randomNamespace := generateRandomNamespace(namespacePrefix)

	// Create the namespace in the cluster using the namespaces package
	_, err := namespace.NewBuilder(client, randomNamespace).Create()
	assert.Nil(t, err)

	// Defer the deletion of the namespace
	defer func() {
		// Delete the namespace
		err := namespace.NewBuilder(client, randomNamespace).Delete()
		assert.Nil(t, err)
	}()

	var deploymentName = "replicas-test"

	builder := deployment.NewBuilder(client, deploymentName, randomNamespace, map[string]string{
		"app": "test",
	}, corev1.Container{
		Name:  "test",
		Image: containerImage,
		Command: []string{
			"sleep",
			"3600",
		},
	}).WithReplicas(2)

	// Create a deployment in the namespace
	_, err = builder.CreateAndWaitUntilReady(timeoutDuration)
	assert.Nil(t, err)

	// Check if the deployment was created
	builder, err = deployment.Pull(client, deploymentName, randomNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, builder.Object)

	// Check if the deployment has the correct number of replicas
	assert.Equal(t, int32(2), *builder.Object.Spec.Replicas)
}
