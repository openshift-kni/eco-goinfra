package metallb

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/schemes/metallb/frrtypes"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// FrrNodeStateBuilder provides struct for FrrNodeState object which contains connection to cluster and
// frrconfiguration definitions.
type FrrNodeStateBuilder struct {
	// Dynamically discovered FrrNodeState object.
	Objects *frrtypes.FRRNodeState
	// apiClient opens api connection to the cluster.
	apiClient runtimeClient.Client
	// nodeName defines on what node FrrNodeState resource should be queried.
	nodeName string
	// nsName defines metallb operator namespace.
	nsName string
	// errorMsg used in discovery function before sending api request to cluster.
	errorMsg string
}

// NewFrrNodeStateBuilder creates new instance of FrrNodeStateBuilder.
func NewFrrNodeStateBuilder(apiClient *clients.Settings, nodeName, nsname string) *FrrNodeStateBuilder {
	glog.V(100).Infof(
		"Initializing new FrrNodeStateBuilder structure with the following params: %s, %s",
		nodeName, nsname)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient cannot be nil")

		return nil
	}

	err := apiClient.AttachScheme(frrtypes.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add metallb scheme to client schemes")

		return nil
	}

	builder := &FrrNodeStateBuilder{
		apiClient: apiClient.Client,
		nodeName:  nodeName,
		nsName:    nsname,
	}

	if nodeName == "" {
		glog.V(100).Infof("The name of the nodeName is empty")

		builder.errorMsg = "FrrNodeState 'nodeName' is empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the FrrNodeState is empty")

		builder.errorMsg = "FrrNodeState 'nsname' is empty"
	}

	return builder
}

// Discover method gets the FrrNodeState items and stores them in the FrrNodeStateBuilder struct.
func (builder *FrrNodeStateBuilder) Discover() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Getting the FrrNodeState object in namespace %s for node %s",
		builder.nsName, builder.nodeName)

	frrNodeState := &frrtypes.FRRNodeState{}
	err := builder.apiClient.Get(context.TODO(),
		runtimeClient.ObjectKey{Name: builder.nodeName, Namespace: builder.nsName}, frrNodeState)

	if err == nil {
		builder.Objects = frrNodeState
	}

	return err
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *FrrNodeStateBuilder) validate() (bool, error) {
	resourceCRD := "frrnodestate"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.apiClient == nil {
		glog.V(100).Infof("The %s builder apiclient is nil", resourceCRD)

		builder.errorMsg = fmt.Sprintf("%s builder cannot have nil apiClient", resourceCRD)
	}

	if builder.errorMsg != "" {
		glog.V(100).Infof("The %s builder has error message: %s", resourceCRD, builder.errorMsg)

		return false, fmt.Errorf(builder.errorMsg)
	}

	return true, nil
}
