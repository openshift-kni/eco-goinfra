package sriov

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"golang.org/x/exp/slices"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkBuilder provides struct for srIovNetwork object which contains connection to cluster and
// srIovNetwork definition.
type NetworkBuilder struct {
	// srIovNetwork definition. Used to create srIovNetwork object.
	Definition *srIovV1.SriovNetwork
	// Created srIovNetwork object.
	Object *srIovV1.SriovNetwork
	// Used in functions that define or mutate srIovNetwork definitions. errorMsg is processed before srIovNetwork
	// object is created.
	errorMsg string
	// apiClient opens api connection to the cluster.
	apiClient *clients.Settings
}

// NetworkAdditionalOptions additional options for SriovNetwork object.
type NetworkAdditionalOptions func(builder *NetworkBuilder) (*NetworkBuilder, error)

// NewNetworkBuilder creates new instance of Builder.
func NewNetworkBuilder(
	apiClient *clients.Settings, name, nsname, targetNsname, resName string) *NetworkBuilder {
	builder := NetworkBuilder{
		apiClient: apiClient,
		Definition: &srIovV1.SriovNetwork{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
			Spec: srIovV1.SriovNetworkSpec{
				ResourceName:     resName,
				NetworkNamespace: targetNsname,
			},
		},
	}

	if name == "" {
		builder.errorMsg = "SrIovNetwork 'name' cannot be empty"
	}

	if nsname == "" {
		builder.errorMsg = "SrIovNetwork 'nsname' cannot be empty"
	}

	if targetNsname == "" {
		builder.errorMsg = "SrIovNetwork 'targetNsname' cannot be empty"
	}

	if resName == "" {
		builder.errorMsg = "SrIovNetwork 'resName' cannot be empty"
	}

	return &builder
}

// WithVLAN sets vlan id in the SrIovNetwork definition. Allowed vlanId range is between 0-4094.
func (builder *NetworkBuilder) WithVLAN(vlanID uint16) *NetworkBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	if vlanID > 4094 {
		builder.errorMsg = "invalid vlanID, allowed vlanID values are between 0-4094"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.Vlan = int(vlanID)

	return builder
}

// WithSpoof sets spoof flag based on the given argument in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithSpoof(enabled bool) *NetworkBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	if enabled {
		builder.Definition.Spec.SpoofChk = "on"
	} else {
		builder.Definition.Spec.SpoofChk = "off"
	}

	return builder
}

// WithLinkState sets linkState parameters in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithLinkState(linkState string) *NetworkBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	allowedLinkStates := []string{"enable", "disable", "auto"}

	if !slices.Contains(allowedLinkStates, linkState) {
		builder.errorMsg = "invalid 'linkState' parameters"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.LinkState = linkState

	return builder
}

// WithMaxTxRate sets maxTxRate parameters in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithMaxTxRate(maxTxRate uint16) *NetworkBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	maxTxRateInt := int(maxTxRate)

	builder.Definition.Spec.MaxTxRate = &maxTxRateInt

	return builder
}

// WithMinTxRate sets minTxRate parameters in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithMinTxRate(minTxRate uint16) *NetworkBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	maxTxRateInt := int(minTxRate)
	builder.Definition.Spec.MaxTxRate = &maxTxRateInt

	return builder
}

// WithTrustFlag sets trust flag based on the given argument in the SrIoVNetwork definition spec.
func (builder *NetworkBuilder) WithTrustFlag(enabled bool) *NetworkBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	if enabled {
		builder.Definition.Spec.Trust = "on"
	} else {
		builder.Definition.Spec.Trust = "off"
	}

	return builder
}

// WithVlanQoS sets qoSClass parameters in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithVlanQoS(qoSClass uint16) *NetworkBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	if qoSClass > 7 {
		builder.errorMsg = "Invalid QoS class. Supported vlan QoS class values are between 0...7"
	}

	if builder.errorMsg != "" {
		return builder
	}

	qoSClassInt := int(qoSClass)

	builder.Definition.Spec.VlanQoS = qoSClassInt

	return builder
}

// WithIPAddressSupport sets ips capabilities in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithIPAddressSupport() *NetworkBuilder {
	return builder.withCapabilities("ips")
}

// WithMacAddressSupport sets mac capabilities in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithMacAddressSupport() *NetworkBuilder {
	return builder.withCapabilities("mac")
}

// WithStaticIpam sets static IPAM in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithStaticIpam() *NetworkBuilder {
	return builder.withIpam("static")
}

// WithOptions creates SriovNetwork with generic mutation options.
func (builder *NetworkBuilder) WithOptions(options ...NetworkAdditionalOptions) *NetworkBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting SriovNetwork additional options")

	for _, option := range options {
		if option != nil {
			builder, err := option(builder)

			if err != nil {
				glog.V(100).Infof("Error occurred in mutation function")

				builder.errorMsg = err.Error()

				return builder
			}
		}
	}

	return builder
}

// PullNetwork pulls existing sriovnetwork from cluster.
func PullNetwork(apiClient *clients.Settings, name, nsname string) (*NetworkBuilder, error) {
	glog.V(100).Infof("Pulling existing sriovnetwork name %s under namespace %s from cluster", name, nsname)

	builder := NetworkBuilder{
		apiClient: apiClient,
		Definition: &srIovV1.SriovNetwork{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the sriovnetwork is empty")

		builder.errorMsg = "sriovnetwork 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the sriovnetwork is empty")

		builder.errorMsg = "sriovnetwork 'namespace' cannot be empty"
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("sriovnetwork object %s doesn't exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Create generates SrIovNetwork in a cluster and stores the created object in struct.
func (builder *NetworkBuilder) Create() (*NetworkBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	if !builder.Exists() {
		var err error
		builder.Object, err = builder.apiClient.SriovNetworks(builder.Definition.Namespace).Create(
			context.TODO(), builder.Definition, metaV1.CreateOptions{},
		)

		if err != nil {
			return nil, err
		}
	}

	return builder, nil
}

// Delete removes SrIovNetwork object.
func (builder *NetworkBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	if !builder.Exists() {
		return nil
	}

	err := builder.apiClient.SriovNetworks(builder.Definition.Namespace).Delete(
		context.TODO(), builder.Object.Name, metaV1.DeleteOptions{})

	if err != nil {
		return err
	}

	builder.Object = nil

	return err
}

// Exists checks whether the given SrIovNetwork object exists in a cluster.
func (builder *NetworkBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	var err error
	builder.Object, err = builder.apiClient.SriovNetworks(builder.Definition.Namespace).Get(
		context.Background(), builder.Definition.Name, metaV1.GetOptions{})

	return err == nil || !k8serrors.IsNotFound(err)
}

// List returns sriov networks in the given namespace.
func List(apiClient *clients.Settings, nsname string, options metaV1.ListOptions) ([]*NetworkBuilder, error) {
	glog.V(100).Infof("Listing sriov networks in the namespace %s with the options %v", nsname, options)

	if nsname == "" {
		glog.V(100).Infof("sriov network 'nsname' parameter can not be empty")

		return nil, fmt.Errorf("failed to list sriov networks, 'nsname' parameter is empty")
	}

	networkList, err := apiClient.SriovNetworks(nsname).List(context.Background(), options)

	if err != nil {
		glog.V(100).Infof("Failed to list sriov networks in the namespace %s due to %s", nsname, err.Error())

		return nil, err
	}

	var networkObjects []*NetworkBuilder

	for _, runningNetwork := range networkList.Items {
		copiedNetwork := runningNetwork
		networkBuilder := &NetworkBuilder{
			apiClient:  apiClient,
			Object:     &copiedNetwork,
			Definition: &copiedNetwork,
		}

		networkObjects = append(networkObjects, networkBuilder)
	}

	return networkObjects, nil
}

// GetSriovNetworksGVR returns SriovNetwork's GroupVersionResource which could be used for Clean function.
func GetSriovNetworksGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: "sriovnetwork.openshift.io", Version: "v1", Resource: "sriovnetworks",
	}
}

func (builder *NetworkBuilder) withCapabilities(capability string) *NetworkBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	builder.Definition.Spec.Capabilities = fmt.Sprintf(`{ "%s": true }`, capability)

	return builder
}

func (builder *NetworkBuilder) withIpam(ipamType string) *NetworkBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	if ipamType == "" {
		glog.V(100).Infof("sriov network 'ipamType' parameter can not be empty")

		builder.errorMsg = "failed to configure IPAM, 'ipamType' parameter is empty"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.IPAM = fmt.Sprintf(`{ "type": "%s" }`, ipamType)

	return builder
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *NetworkBuilder) validate() (bool, error) {
	resourceCRD := "SriovNetwork"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Definition == nil {
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

// CleanAllNetworksByTargetNamespace deletes all networks matched by their NetworkNamespace spec.
func CleanAllNetworksByTargetNamespace(
	apiClient *clients.Settings,
	operatornsname string,
	targetnsname string,
	options metaV1.ListOptions) error {
	glog.V(100).Infof("Cleaning up sriov networks in the %s namespace with %s NetworkNamespace spec",
		operatornsname, targetnsname)

	if operatornsname == "" {
		glog.V(100).Infof("'operatornsname' parameter can not be empty")

		return fmt.Errorf("failed to clean up sriov networks, 'operatornsname' parameter is empty")
	}

	if targetnsname == "" {
		glog.V(100).Infof("'targetnsname' parameter can not be empty")

		return fmt.Errorf("failed to clean up sriov networks, 'targetnsname' parameter is empty")
	}

	networks, err := List(apiClient, operatornsname, options)

	if err != nil {
		glog.V(100).Infof("Failed to list sriov networks in namespace: %s", operatornsname)

		return err
	}

	for _, network := range networks {
		if network.Object.Spec.NetworkNamespace == targetnsname {
			err = network.Delete()
			if err != nil {
				glog.V(100).Infof("Failed to delete sriov networks: %s", network.Object.Name)

				return err
			}
		}
	}

	return nil
}
