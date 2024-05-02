package nmstate

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v2"

	"github.com/golang/glog"
	"golang.org/x/exp/slices"

	nmstateV1alpha1 "github.com/nmstate/kubernetes-nmstate/api/v1alpha1"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// allowedInterfaceTypes represents all allowed types for interface.
	allowedInterfaceTypes = []string{"ethernet", "bond", "ovs-bridge", "unknown",
		"vlan", "vxlan", "linux-bridge", "team", "veth"}
)

// StateBuilder provides struct for the NodeNetworkState object containing connection to the cluster.
type StateBuilder struct {
	// Created NodeNetworkState object on the cluster.
	Object *nmstateV1alpha1.NodeNetworkState
	// API client to interact with the cluster.
	apiClient *clients.Settings
	// errorMsg is processed before NodeNetworkState object is created.
	errorMsg string
}

// Exists checks whether the given NodeNetworkState exists.
func (builder *StateBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if NodeNetworkState %s exists", builder.Object.Name)

	var err error
	builder.Object, err = builder.Get()

	if err != nil {
		glog.V(100).Infof("Failed to collect NodeNetworkState object due to %s", err.Error())
	}

	return err == nil || !k8serrors.IsNotFound(err)
}

// Get returns NodeNetworkState object if found.
func (builder *StateBuilder) Get() (*nmstateV1alpha1.NodeNetworkState, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Collecting NodeNetworkState object %s", builder.Object.Name)

	nodeNetworkState := &nmstateV1alpha1.NodeNetworkState{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name: builder.Object.Name,
	}, nodeNetworkState)

	if err != nil {
		glog.V(100).Infof("NodeNetworkState object %s does not exist", builder.Object.Name)

		return nil, err
	}

	return nodeNetworkState, err
}

// GetTotalVFs returns total-vfs under the given interface.
func (builder *StateBuilder) GetTotalVFs(sriovInterfaceName string) (int, error) {
	if valid, err := builder.validate(); !valid {
		return 0, err
	}

	glog.V(100).Infof(
		"Getting total-vfs under interface %s from NodeNetworkState %s",
		sriovInterfaceName, builder.Object.Name)

	if sriovInterfaceName == "" {
		glog.V(100).Infof("The sriovInterfaceName can not be empty string")

		return 0, fmt.Errorf("the sriovInterfaceName is empty sting")
	}

	var CurrentState DesiredState

	err := yaml.Unmarshal(builder.Object.Status.CurrentState.Raw, &CurrentState)
	if err != nil {
		return 0, fmt.Errorf("failed to Unmarshal NMState state")
	}

	for _, interfaceFromCurrentState := range CurrentState.Interfaces {
		if interfaceFromCurrentState.Name == sriovInterfaceName {
			return *interfaceFromCurrentState.Ethernet.Sriov.TotalVfs, nil
		}
	}

	return 0, fmt.Errorf("failed to find interface %s", sriovInterfaceName)
}

// GetInterfaceType returns Interface with the given interface name and given type.
func (builder *StateBuilder) GetInterfaceType(interfaceName, interfaceType string) (NetworkInterface, error) {
	if valid, err := builder.validate(); !valid {
		return NetworkInterface{}, err
	}

	glog.V(100).Infof(
		"Getting interface %s with type %s from NodeNetworkState %s",
		interfaceName, interfaceType, builder.Object.Name)

	if interfaceName == "" {
		glog.V(100).Infof("The interfaceName can not be empty string")

		return NetworkInterface{}, fmt.Errorf("the interfaceName is empty sting")
	}

	if !slices.Contains(allowedInterfaceTypes, interfaceType) {
		glog.V(100).Infof("error to add type %s, allowed types are %v", interfaceType, allowedInterfaceTypes)

		return NetworkInterface{}, fmt.Errorf("invalid interfaceType parameter")
	}

	var CurrentState DesiredState

	err := yaml.Unmarshal(builder.Object.Status.CurrentState.Raw, &CurrentState)
	if err != nil {
		return NetworkInterface{}, fmt.Errorf("failed to Unmarshal NMState state")
	}

	for _, interf := range CurrentState.Interfaces {
		if interf.Name == interfaceName && interf.Type == interfaceType {
			return interf, nil
		}
	}

	return NetworkInterface{}, fmt.Errorf("failed to find interface %s or it is not a %s type",
		interfaceName, interfaceType)
}

// GetSriovVfs returns all configured VFs  under the given SR-IOV interface.
func (builder *StateBuilder) GetSriovVfs(sriovInterfaceName string) ([]Vf, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof(
		"Getting all configured VFs under interface %s from NodeNetworkState %s",
		sriovInterfaceName, builder.Object.Name)

	if sriovInterfaceName == "" {
		glog.V(100).Infof("The sriovInterfaceName can not be empty string")

		return nil, fmt.Errorf("the sriovInterfaceName is empty sting")
	}

	var CurrentState DesiredState

	err := yaml.Unmarshal(builder.Object.Status.CurrentState.Raw, &CurrentState)
	if err != nil {
		return nil, fmt.Errorf("failed to Unmarshal NMState state")
	}

	for _, networkInterface := range CurrentState.Interfaces {
		if networkInterface.Name == sriovInterfaceName && networkInterface.Ethernet.Sriov.Vfs != nil {
			return networkInterface.Ethernet.Sriov.Vfs, nil
		}
	}

	return nil, fmt.Errorf("failed to find interface %s "+
		"or SR-IOV VFs are not configured on it", sriovInterfaceName)
}

// PullNodeNetworkState retrieves an existing NodeNetworkState object from the cluster.
func PullNodeNetworkState(apiClient *clients.Settings, name string) (*StateBuilder, error) {
	glog.V(100).Infof("Pulling NodeNetworkState object name:%s", name)

	stateBuilder := StateBuilder{
		apiClient: apiClient,
		Object: &nmstateV1alpha1.NodeNetworkState{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the NodeNetworkState is empty")

		stateBuilder.errorMsg = "NodeNetworkState 'name' cannot be empty"
	}

	if !stateBuilder.Exists() {
		return nil, fmt.Errorf("NodeNetworkState object %s does not exist", name)
	}

	return &stateBuilder, nil
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *StateBuilder) validate() (bool, error) {
	resourceCRD := "NodeNetworkState"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Object == nil {
		glog.V(100).Infof("The %s is undefined", resourceCRD)

		builder.errorMsg = msg.UndefinedCrdObjectErrString(resourceCRD)
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
