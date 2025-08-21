package main

import (
	"fmt"
	"log"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/icsp"
)

func main() {
	apiClient := clients.New("")

	var icspBuilderName = "testicsp"
	// Define new ICSPBuilder
	icspbuilder := icsp.NewICSPBuilder(apiClient, icspBuilderName, "mysource", []string{"mymirror.io"})
	// Delete ImageContentSourcePolicy if exists
	if icspbuilder.Exists() {
		_ = icspbuilder.Delete()
	}

	// Create new ImageContentSourcePolicy from icspbuilder.
	_, err := icspbuilder.Create()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Initial ImageContentSourcePolicy: ", icspbuilder.Object.Spec.RepositoryDigestMirrors)

	// Add new source to the ImageContentSourcePolicy definition and update the object
	_, err = icspbuilder.WithRepositoryDigestMirror(
		"newsource",
		[]string{"Mirror1.io", "Mirror2.io", "Mirror3.io"}).Update()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Updated ImageContentSourcePolicy: ", icspbuilder.Object.Spec.RepositoryDigestMirrors)
	// Delete ImageContentSourcePolicy after demonstration
	err = icspbuilder.Delete()
	if err != nil {
		log.Fatal(err)
	}
}
