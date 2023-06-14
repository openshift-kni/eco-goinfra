package main

import (
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
)

func main() {
	// Init client struct. If path to kubeconfig is empty string then KUBECONFIG env var is used.
	apiClients := clients.New("")
	// If error occurs during apiClient initialization function returns nil pointer.
	if apiClients == nil {
		panic("Failed to load api client")
	}
}
