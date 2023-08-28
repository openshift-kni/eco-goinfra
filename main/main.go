package main

import (
	"fmt"
	"log"

	"github.com/openshift-kni/eco-goinfra/pkg/argocd"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
)

func main() {
	apiClient := clients.New("/home/nkononov/Documents/work/auth/shai/kubeconfig")
	if apiClient == nil {
		fmt.Errorf("Error to load api")
	}
	application, err := argocd.Pull(apiClient, "clusters", "openshift-gitops")
	if err != nil {
		log.Print(err.Error())
	}

	log.Print("ArogCDApplication Resource !!!")
	log.Print(application.Definition)

	argoCdResource, err := argocd.ArgoCDPull(apiClient, "openshift-gitops", "openshift-gitops")

	if err != nil {
		log.Print(err.Error())
	}
	log.Print("ArogCD Resource !!!")
	log.Print(argoCdResource.Definition)
	//log.Print(err.Error())
	//log.Print(application.Definition)
	//log.Print("Hello")
}
