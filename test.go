package main

import (
	"fmt"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/ocm"
)

func main() {
	pc, err := ocm.PullPolicySet(clients.New(""), "demo-policyset", "default")
	if err != nil {
		fmt.Println("err = ", err)

	}
	fmt.Println("policySet = ", pc.Object.Name)

	list, err := ocm.ListPolicieSetsInAllNamespaces(clients.New(""))
	fmt.Println("list = ", list[0].Object.Name)

	pc.Delete()

	list, err = ocm.ListPolicieSetsInAllNamespaces(clients.New(""))
	fmt.Println("list after delete = ", list)
}
