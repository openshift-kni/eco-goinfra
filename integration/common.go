// Package integration provides common functions for integration tests.
package integration

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/namespace"
	"github.com/openshift-kni/eco-goinfra/pkg/pod"
	corev1 "k8s.io/api/core/v1"
)

// CreateRandomNamespace generates a random namespace name for testing.
func CreateRandomNamespace() string {
	// Generate a random namespace name with rand library
	return fmt.Sprintf("test-namespace-%d", rand.Intn(100000))
}

// PreEmptiveNamespaceDeleteAndSetup deletes the namespace preemptively and sets it up for the test.
func PreEmptiveNamespaceDeleteAndSetup(namespace string, namespaceBuilder *namespace.Builder) error {
	if namespaceBuilder == nil {
		return errors.New("namespaceBuilder is nil")
	}

	// Preemptively delete the namespace before the test.
	err := namespaceBuilder.DeleteAndWait(time.Duration(30) * time.Second)
	if err != nil {
		return err
	}

	// Create the namespace
	_, err = namespaceBuilder.Create()
	if err != nil {
		return err
	}

	return nil
}

// CreateTestContainerDefinition creates a container definition for the test.
func CreateTestContainerDefinition(containerName string, containerImage string,
	command []string) (*corev1.Container, error) {
	testContainerBuilder := pod.NewContainerBuilder(containerName, containerImage, command)

	// Change the container default security context to something that is allowed in the test environment
	testContainerBuilder.WithSecurityContext(&corev1.SecurityContext{
		RunAsUser:  nil,
		RunAsGroup: nil,
	})

	containerDefinition, err := testContainerBuilder.GetContainerCfg()
	if err != nil {
		return nil, err
	}

	return containerDefinition, nil
}
