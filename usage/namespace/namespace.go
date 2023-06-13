package main

import (
	"fmt"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/namespace"
)

func createExample() {
	// Init client struct. If path to kubeconfig is empty string then KUBECONFIG env var is used.
	apiClients := clients.New("")
	// If error occurs during apiClient initialization function returns nil pointer.
	if apiClients == nil {
		panic("Failed to load api client")
	}
	// Init namespace struct. NewBuilder function require minimum set of parameters.
	exampleNamespace := namespace.NewBuilder(apiClients, "example")
	// Mutate exampleNamespace struct and add labels to the namespace object
	exampleNamespace = exampleNamespace.WithLabel("examplekey", "examplevalue")
	// Create namespace on cluster
	_, err := exampleNamespace.Create()
	if err != nil {
		fmt.Print(err.Error())
		panic("Failed to create namespace on cluster")
	}
}

func deleteExistingExample() {
	// Init client struct. If path to kubeconfig is empty string then KUBECONFIG env var is used.
	apiClients := clients.New("")
	// If error occurs during apiClient initialization function returns nil pointer.
	if apiClients == nil {
		panic("Failed to load api client")
	}
	// Pull existing namespace from cluster
	exampleNamespace, err := namespace.Pull(apiClients, "example")
	if err != nil {
		fmt.Print(err.Error())
		panic("Failed to Pull namespace from cluster")
	}
	// Delete namespace
	err = exampleNamespace.Delete()
	if err != nil {
		fmt.Print(err.Error())
		panic("Failed to delete namespace from cluster")
	}
}

func updateExistingExample() {
	// Init client struct. If path to kubeconfig is empty string then KUBECONFIG env var is used.
	apiClients := clients.New("")
	// If error occurs during apiClient initialization function returns nil pointer.
	if apiClients == nil {
		panic("Failed to load api client")
	}
	// Pull existing namespace from cluster
	exampleNamespace, err := namespace.Pull(apiClients, "example")
	if err != nil {
		fmt.Print(err.Error())
		panic("Failed to Pull namespace from cluster")
	}
	// Update object definition using With* function
	exampleNamespace = exampleNamespace.WithLabel("key", "value")
	// Update object on cluster
	_, err = exampleNamespace.Update()
	if err != nil {
		fmt.Print(err.Error())
		panic("Failed to Update namespace from cluster")
	}
}

func createOneLinerExample() {
	// Init client struct. If path to kubeconfig is empty string then KUBECONFIG env var is used.
	apiClients := clients.New("")
	// If error occurs during apiClient initialization function returns nil pointer.
	if apiClients == nil {
		panic("Failed to load api client")
	}

	// Create namespace with labels using single line
	_, err := namespace.NewBuilder(apiClients, "example").WithLabel("key", "value").Create()
	if err != nil {
		fmt.Print(err.Error())
		panic("Failed to create namespace on cluster")
	}
}

func main() {
	// Example of how to create namespace using ./pkg/namespace.
	createExample()
	// Example of how to delete existing namespace on cluster using ./pkg/namespace.
	deleteExistingExample()
	// Update existing object using With mutation function.
	updateExistingExample()
	// Create namespace using one line.
	createOneLinerExample()
}
