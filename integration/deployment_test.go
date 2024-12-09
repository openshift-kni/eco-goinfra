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

func TestDeploymentWithNodeSelector(t *testing.T) {
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

	var deploymentName = "node-selector-test"

	builder := deployment.NewBuilder(client, deploymentName, randomNamespace, map[string]string{
		"app": "test",
	}, corev1.Container{
		Name:  "test",
		Image: containerImage,
		Command: []string{
			"sleep",
			"3600",
		},
	}).WithNodeSelector(map[string]string{
		"node-role.kubernetes.io/worker": "worker",
	})

	// Create a deployment in the namespace
	_, err = builder.CreateAndWaitUntilReady(timeoutDuration)
	assert.Nil(t, err)

	// Check if the deployment was created
	builder, err = deployment.Pull(client, deploymentName, randomNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, builder.Object)

	// Check if the deployment has the node selector
	assert.Equal(t, map[string]string{
		"node-role.kubernetes.io/worker": "worker",
	}, builder.Object.Spec.Template.Spec.NodeSelector)
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

func TestDeploymentWithAdditionalContainerSpecs(t *testing.T) {
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

	var deploymentName = "additional-container-specs-test"

	builder := deployment.NewBuilder(client, deploymentName, randomNamespace, map[string]string{
		"app": "test",
	}, corev1.Container{
		Name:  "test",
		Image: containerImage,
		Command: []string{
			"sleep",
			"3600",
		},
	}).WithAdditionalContainerSpecs([]corev1.Container{
		{
			Name:  "additional-container",
			Image: containerImage,
			Command: []string{
				"sleep",
				"3600",
			},
		},
	})

	// Create a deployment in the namespace
	_, err = builder.CreateAndWaitUntilReady(timeoutDuration)
	assert.Nil(t, err)

	// Check if the deployment was created
	builder, err = deployment.Pull(client, deploymentName, randomNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, builder.Object)

	// Check if the deployment has the additional container
	assert.Len(t, builder.Object.Spec.Template.Spec.Containers, 2)
}

func TestDeploymentWithSecurityContext(t *testing.T) {
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

	var deploymentName = "security-context-test"

	var falseVar = false

	builder := deployment.NewBuilder(client, deploymentName, randomNamespace, map[string]string{
		"app": "test",
	}, corev1.Container{
		Name:  "test",
		Image: containerImage,
		Command: []string{
			"sleep",
			"3600",
		},
	}).WithSecurityContext(&corev1.PodSecurityContext{
		RunAsNonRoot: &falseVar,
	})

	// Create a deployment in the namespace
	_, err = builder.Create()
	assert.Nil(t, err)

	// Check if the deployment was created
	builder, err = deployment.Pull(client, deploymentName, randomNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, builder.Object)

	// Check if the deployment has the security context
	assert.NotNil(t, builder.Object.Spec.Template.Spec.SecurityContext)
	assert.Equal(t, false, *builder.Object.Spec.Template.Spec.SecurityContext.RunAsNonRoot)

	// NOTE: The container will not spawn on OCP because of the security context warning.
}

func TestDeploymentWithLabel(t *testing.T) {
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

	var deploymentName = "label-test"

	builder := deployment.NewBuilder(client, deploymentName, randomNamespace, map[string]string{
		"app": "test",
	}, corev1.Container{
		Name:  "test",
		Image: containerImage,
		Command: []string{
			"sleep",
			"3600",
		},
	}).WithLabel("key", "value")

	// Create a deployment in the namespace
	_, err = builder.Create()
	assert.Nil(t, err)

	// Check if the deployment was created
	builder, err = deployment.Pull(client, deploymentName, randomNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, builder.Object)

	// Check if the deployment has the label
	assert.NotNil(t, builder.Object.Spec.Template.Labels)
	for key, value := range builder.Object.Spec.Template.Labels {
		if key == "key" {
			assert.Equal(t, "value", value)
		}
	}
}

func TestDeploymentWithServiceAccountName(t *testing.T) {
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

	var deploymentName = "service-account-name-test"

	builder := deployment.NewBuilder(client, deploymentName, randomNamespace, map[string]string{
		"app": "test",
	}, corev1.Container{
		Name:  "test",
		Image: containerImage,
		Command: []string{
			"sleep",
			"3600",
		},
	}).WithServiceAccountName("default")

	// Create a deployment in the namespace
	_, err = builder.Create()
	assert.Nil(t, err)

	// Check if the deployment was created
	builder, err = deployment.Pull(client, deploymentName, randomNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, builder.Object)

	// Check if the deployment has the service account name
	assert.Equal(t, "default", builder.Object.Spec.Template.Spec.ServiceAccountName)
}

func TestDeploymentWithVolume(t *testing.T) {
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

	var deploymentName = "volume-test"

	builder := deployment.NewBuilder(client, deploymentName, randomNamespace, map[string]string{
		"app": "test",
	}, corev1.Container{
		Name:  "test",
		Image: containerImage,
		Command: []string{
			"sleep",
			"3600",
		},
	}).WithVolume(corev1.Volume{
		Name: "test-volume",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	// Create a deployment in the namespace
	_, err = builder.Create()
	assert.Nil(t, err)

	// Check if the deployment was created
	builder, err = deployment.Pull(client, deploymentName, randomNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, builder.Object)

	// Check if the deployment has the volume
	assert.Len(t, builder.Object.Spec.Template.Spec.Volumes, 1)
	assert.Equal(t, "test-volume", builder.Object.Spec.Template.Spec.Volumes[0].Name)
}

func TestDeploymentWithSchedulerName(t *testing.T) {
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

	var deploymentName = "scheduler-name-test"

	builder := deployment.NewBuilder(client, deploymentName, randomNamespace, map[string]string{
		"app": "test",
	}, corev1.Container{
		Name:  "test",
		Image: containerImage,
		Command: []string{
			"sleep",
			"3600",
		},
	}).WithSchedulerName("default-scheduler")

	// Create a deployment in the namespace
	_, err = builder.Create()
	assert.Nil(t, err)

	// Check if the deployment was created
	builder, err = deployment.Pull(client, deploymentName, randomNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, builder.Object)

	// Check if the deployment has the scheduler name
	assert.Equal(t, "default-scheduler", builder.Object.Spec.Template.Spec.SchedulerName)
}

func TestDeploymentWithToleration(t *testing.T) {
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

	var deploymentName = "toleration-test"

	builder := deployment.NewBuilder(client, deploymentName, randomNamespace, map[string]string{
		"app": "test",
	}, corev1.Container{
		Name:  "test",
		Image: containerImage,
		Command: []string{
			"sleep",
			"3600",
		},
	}).WithToleration(corev1.Toleration{
		Key:      "key",
		Operator: corev1.TolerationOpEqual,
		Value:    "value",
		Effect:   corev1.TaintEffectNoSchedule,
	})

	// Create a deployment in the namespace
	_, err = builder.Create()
	assert.Nil(t, err)

	// Check if the deployment was created
	builder, err = deployment.Pull(client, deploymentName, randomNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, builder.Object)

	// Check if the deployment has the toleration
	assert.Len(t, builder.Object.Spec.Template.Spec.Tolerations, 1)
	assert.Equal(t, "key", builder.Object.Spec.Template.Spec.Tolerations[0].Key)
	assert.Equal(t, corev1.TolerationOpEqual, builder.Object.Spec.Template.Spec.Tolerations[0].Operator)
	assert.Equal(t, "value", builder.Object.Spec.Template.Spec.Tolerations[0].Value)
	assert.Equal(t, corev1.TaintEffectNoSchedule, builder.Object.Spec.Template.Spec.Tolerations[0].Effect)
}
