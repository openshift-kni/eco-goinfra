package metallb

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	frrtypes "github.com/openshift-kni/eco-goinfra/pkg/schemes/metallb/frrtypes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	frrConfigurationKind = "FRRConfiguration"
)

// FrrConfigurationBuilder provides struct for the FrrConfiguration object containing connection to
// the cluster and the FrrConfiguration definitions.
type FrrConfigurationBuilder struct {
	Definition *frrtypes.FRRConfiguration
	Object     *frrtypes.FRRConfiguration
	apiClient  *clients.Settings
	errorMsg   string
}

// NewFrrConfigurationBuilder creates a new instance of FRRConfiguration.
func NewFrrConfigurationBuilder(
	apiClient *clients.Settings, name, nsname, peerIP string, localASN, remoteASN uint32) *FrrConfigurationBuilder {
	glog.V(100).Infof(
		"Initializing new Frrconfiguration structure with the following params: %s, %s %s %d %d",
		name, nsname, peerIP, localASN, remoteASN)

	builder := FrrConfigurationBuilder{
		apiClient: apiClient,
		Definition: &frrtypes.FRRConfiguration{
			TypeMeta: metav1.TypeMeta{
				Kind:       frrConfigurationKind,
				APIVersion: fmt.Sprintf("%s/%s", FRRAPIGroup, APIVersion),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			}, Spec: frrtypes.FRRConfigurationSpec{
				BGP: frrtypes.BGPConfig{
					Routers: []frrtypes.Router{
						{
							ASN: localASN,
							Neighbors: []frrtypes.Neighbor{
								{
									Address: peerIP,
									ASN:     remoteASN,
								},
							},
						},
					},
				},
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the frrConfiguration is empty")

		builder.errorMsg = "frrConfiguration 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the frrConfiguration is empty")

		builder.errorMsg = "frrConfiguration 'nsname' cannot be empty"
	}

	if net.ParseIP(peerIP) == nil {
		glog.V(100).Infof("The peerIP for the frrConfiguration bgp peer contains invalid ip address %s", peerIP)

		builder.errorMsg = "frrConfiguration 'peerIP' of the BGPPeer contains invalid ip address"
	}

	if strconv.Itoa(int(localASN)) == "" {
		glog.V(100).Infof("The localASN of the frrConfiguration can not be empty")

		builder.errorMsg = "frrConfiguration 'localASN' cannot be empty"
	}

	if strconv.Itoa(int(remoteASN)) == "" {
		glog.V(100).Infof("The remoteASN of the frrConfiguration can not be empty")

		builder.errorMsg = "frrConfiguration 'remoteASN' cannot be empty"
	}

	return &builder
}

// Create makes a FrrConfiguration in the cluster and stores the created object in struct.
func (builder *FrrConfigurationBuilder) Create() (*FrrConfigurationBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the FrrConfiguration %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace,
	)

	var err error
	if !builder.Exists() {
		unstructuredFRRConfiguration, err := runtime.DefaultUnstructuredConverter.ToUnstructured(builder.Definition)

		if err != nil {
			glog.V(100).Infof("Failed to convert structured FRRConfiguration to unstructured object")

			return nil, err
		}

		unsObject, err := builder.apiClient.Resource(
			GetFrrConfigurationGVR()).Namespace(builder.Definition.Namespace).Create(
			context.TODO(), &unstructured.Unstructured{Object: unstructuredFRRConfiguration}, metav1.CreateOptions{})

		if err != nil {
			glog.V(100).Infof("Failed to create FRRConfiguration")

			return nil, err
		}

		builder.Object, err = builder.convertToStructured(unsObject)

		if err != nil {
			return nil, err
		}
	}

	return builder, err
}

// Get returns FRRConfiguration object if found.
func (builder *FrrConfigurationBuilder) Get() (*frrtypes.FRRConfiguration, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof(
		"Collecting FRRConfiguration object %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	unsObject, err := builder.apiClient.Resource(
		GetFrrConfigurationGVR()).Namespace(builder.Definition.Namespace).Get(
		context.TODO(), builder.Definition.Name, metav1.GetOptions{})

	if err != nil {
		glog.V(100).Infof(
			"Failed to Unmarshal FRRConfiguration: unstructured object to structure in namespace %s",
			builder.Definition.Namespace)

		return nil, err
	}

	return builder.convertToStructured(unsObject)
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *FrrConfigurationBuilder) validate() (bool, error) {
	resourceCRD := "FRRConfiguration"

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

func (builder *FrrConfigurationBuilder) convertToStructured(unsObject *unstructured.Unstructured) (
	*frrtypes.FRRConfiguration, error) {
	frrConfiguration := &frrtypes.FRRConfiguration{}

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unsObject.Object, frrConfiguration)
	if err != nil {
		glog.V(100).Infof(
			"Failed to convert from unstructured to FrrConfiguration object in namespace %s",
			builder.Definition.Namespace)

		return nil, err
	}

	return frrConfiguration, err
}

// Delete removes FrrConfigurationBuilder object from a cluster.
func (builder *FrrConfigurationBuilder) Delete() (*FrrConfigurationBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting the FrrConfiguration object %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace,
	)

	if !builder.Exists() {
		return builder, fmt.Errorf("FrrConfiguration cannot be deleted because it does not exist")
	}

	err := builder.apiClient.Resource(
		GetFrrConfigurationGVR()).Namespace(builder.Definition.Namespace).Delete(
		context.TODO(), builder.Definition.Name, metav1.DeleteOptions{})

	if err != nil {
		return builder, fmt.Errorf("can not delete FRRConfiguration: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// GetFrrConfigurationGVR returns FrrConfiguration GroupVersionResource which could be used for Clean function.
func GetFrrConfigurationGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: FRRAPIGroup, Version: APIVersion, Resource: "frrconfigurations",
	}
}

// Exists checks whether the given FrrConfiguration exists.
func (builder *FrrConfigurationBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof(
		"Checking if FrrConfiguration %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// WithBGPPassword defines the password used between BGP peers to form adjacency.
func (builder *FrrConfigurationBuilder) WithBGPPassword(bgpPassword string) *FrrConfigurationBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Define FRRConfiguration %s in namespace %s with the password: %s",
		builder.Definition.Name, builder.Definition.Namespace, bgpPassword)

	if bgpPassword == "" {
		glog.V(100).Infof("The bgpPassword should not be blank")

		builder.errorMsg = fmt.Sprintf("the bgpPassword %s is an empty string", bgpPassword)
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.BGP.Routers[0].Neighbors[0].Password = bgpPassword

	return builder
}

// WithToReceiveModeAll allows all incoming routes.
func (builder *FrrConfigurationBuilder) WithToReceiveModeAll(mode string) *FrrConfigurationBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Allows the FRR nodes to receive all routes in namespace %s with mode type: all",
		builder.Definition.Namespace)

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.BGP.Routers[0].Neighbors[0].ToReceive.Allowed.Mode = frrtypes.AllowMode(mode)

	return builder
}

// WithToReceiveModeFiltered allows all defined prefixes to be received.
func (builder *FrrConfigurationBuilder) WithToReceiveModeFiltered(
	neigh frrtypes.Neighbor, prefixes []string) *FrrConfigurationBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"AllowMode 'filtered' receives only routes with in the prefix %s", prefixes)

	for _, prefix := range prefixes {
		glog.V(100).Infof("AllowMode 'filtered' receives only routes within the prefix %s", prefix)

		if _, _, err := net.ParseCIDR(prefix); err != nil {
			glog.V(100).Infof("The prefix of the frrConfiguration BGP peer contains an invalid IP "+
				"address: %s", prefix)

			builder.errorMsg = "frrConfiguration 'peerIP' of the BGPPeer contains an invalid IP address"

			break
		}
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.BGP.Routers[0].Neighbors[0] = neigh

	// Assign the list of valid prefixes
	neigh.ToReceive.Allowed.Prefixes = make([]frrtypes.PrefixSelector, len(prefixes))
	for i, prefix := range prefixes {
		neigh.ToReceive.Allowed.Prefixes[i] = frrtypes.PrefixSelector{Prefix: prefix}
	}

	return builder
}

// WithEBGPMultiHop sets the EBGP multihop setting on the remote BGP peer.
func (builder *FrrConfigurationBuilder) WithEBGPMultiHop(ebgpMultiHop bool) *FrrConfigurationBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Sets the EBGP multihop setting to true in the FrrConfiguration %s",
		builder.Definition.Name)

	builder.Definition.Spec.BGP.Routers[0].Neighbors[0].EBGPMultiHop = true

	return builder
}

// WithHoldTime defines the holdTime placed in the FrrConfiguration spec.
func (builder *FrrConfigurationBuilder) WithHoldTime(holdTime metav1.Duration) *FrrConfigurationBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Creating FrrConfiguration BGPPeer %s in namespace %s with this holdTime: %s",
		builder.Definition.Name, builder.Definition.Namespace, holdTime)

	builder.Definition.Spec.BGP.Routers[0].Neighbors[0].HoldTime = &holdTime

	return builder
}

// WithKeepalive defines the keepAliveTime placed in the FrrConfiguration spec.
func (builder *FrrConfigurationBuilder) WithKeepalive(keepAlive metav1.Duration) *FrrConfigurationBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Creating aFrrConfiguration BGPPeer %s in namespace %s with this keepalive: %s",
		builder.Definition.Name, builder.Definition.Namespace, keepAlive)

	builder.Definition.Spec.BGP.Routers[0].Neighbors[0].KeepaliveTime = &keepAlive

	return builder
}

// WithConnectTime defines the time to wait until trying to connect to a BGP peer.
func (builder *FrrConfigurationBuilder) WithConnectTime(connectTime metav1.Duration) *FrrConfigurationBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Define hold time. The bgp connection will attempt every %s to establish a connection with the "+
			"BGP Peer", connectTime)

	duration := connectTime.Duration

	if duration < time.Second || duration > 65535*time.Second {
		glog.V(100).Infof("A valid connect time is between 1-65535")

		builder.errorMsg = "frrConfiguration 'connectTime' value is not valid"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.BGP.Routers[0].Neighbors[0].ConnectTime = &connectTime

	return builder
}
