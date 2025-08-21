package mco

import (
	mcv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const defaultMachineConfigName = "test-machine-config"

// buildDummyMachineConfig returns a MachineConfig with the provided name.
func buildDummyMachineConfig(name string) *mcv1.MachineConfig {
	return &mcv1.MachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestClientWithDummyMachineConfig returns a client with a dummy MachineConfig.
func buildTestClientWithDummyMachineConfig() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyMachineConfig(defaultMachineConfigName),
		},
	})
}

// buildValidMachineConfigTestBuilder returns a valid MCBuilder for testing.
func buildValidMachineConfigTestBuilder(apiClient *clients.Settings) *MCBuilder {
	return NewMCBuilder(apiClient, defaultMachineConfigName)
}
