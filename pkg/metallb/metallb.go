package metallb

import (
	"context"
	"fmt"
	"net"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	mlbtypes "github.com/openshift-kni/eco-goinfra/pkg/schemes/metallb/mlboperator"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Builder provides struct for the MetalLb object containing connection to
// the cluster and the MetalLb definitions.
type Builder struct {
	Definition *mlbtypes.MetalLB
	Object     *mlbtypes.MetalLB
	apiClient  runtimeClient.Client
	errorMsg   string
}

// AdditionalOptions additional options for metallb object.
type AdditionalOptions func(builder *Builder) (*Builder, error)

// NewBuilder creates a new instance of Builder.
func NewBuilder(apiClient *clients.Settings, name, nsname string, nodeSelector map[string]string) *Builder {
	glog.V(100).Infof(
		"Initializing new metallb structure with the following params: %s, %s, %v",
		name, nsname, nodeSelector)

	err := apiClient.AttachScheme(mlbtypes.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add metallb scheme to client schemes")

		return nil
	}

	builder := Builder{
		apiClient: apiClient.Client,
		Definition: &mlbtypes.MetalLB{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
			Spec: mlbtypes.MetalLBSpec{
				SpeakerNodeSelector: nodeSelector,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the metallb is empty")

		builder.errorMsg = "metallb 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the metallb is empty")

		builder.errorMsg = "metallb 'nsname' cannot be empty"
	}

	if len(nodeSelector) < 1 {
		glog.V(100).Infof("The SpeakerNodeSelector of the metallb is empty")

		builder.errorMsg = "metallb 'nodeSelector' cannot be empty"
	}

	return &builder
}

// Pull retrieves an existing metallb.io object from the cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	glog.V(100).Infof(
		"Pulling metallb.io object name:%s in namespace: %s", name, nsname)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")

		return nil, fmt.Errorf("metallb 'apiClient' cannot be empty")
	}

	err := apiClient.AttachScheme(mlbtypes.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add metallb scheme to client schemes")

		return nil, err
	}

	builder := Builder{
		apiClient: apiClient.Client,
		Definition: &mlbtypes.MetalLB{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the metallb is empty")

		return nil, fmt.Errorf("metallb 'name' cannot be empty")
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the metallb is empty")

		return nil, fmt.Errorf("metallb 'nsname' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("metallb object %s does not exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Exists checks whether the given MetalLb exists.
func (builder *Builder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof(
		"Checking if MetalLb %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	if err != nil {
		glog.V(100).Infof("Failed to collect MetalLb object due to %s", err.Error())
	}

	return err == nil || !k8serrors.IsNotFound(err)
}

// Get returns MetalLb object if found.
func (builder *Builder) Get() (*mlbtypes.MetalLB, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof(
		"Collecting metallb object %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	metallb := &mlbtypes.MetalLB{}
	err := builder.apiClient.Get(context.TODO(),
		runtimeClient.ObjectKey{Name: builder.Definition.Name, Namespace: builder.Definition.Namespace}, metallb)

	if err != nil {
		glog.V(100).Infof(
			"metallb object %s does not exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)

		return nil, err
	}

	return metallb, nil
}

// Create makes a MetalLb in the cluster and stores the created object in struct.
func (builder *Builder) Create() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the metallb %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace,
	)

	var err error
	if !builder.Exists() {
		err = builder.apiClient.Create(context.TODO(), builder.Definition)

		if err != nil {
			glog.V(100).Infof("Failed to create MetalLb")

			return nil, err
		}

		builder.Object = builder.Definition

		if err != nil {
			return nil, err
		}
	}

	return builder, err
}

// Delete removes MetalLb object from a cluster.
func (builder *Builder) Delete() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting the metallb object %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace,
	)

	if !builder.Exists() {
		glog.V(100).Infof("MetalLb object %s does not exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)

		builder.Object = nil

		return builder, nil
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return builder, fmt.Errorf("can not delete metallb: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// Update renovates the existing MetalLb object with the MetalLb definition in builder.
func (builder *Builder) Update(force bool) (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("failed to update metallb, object does not exist on cluster")
	}

	glog.V(100).Infof("Updating the metallb object %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace,
	)

	if builder.errorMsg != "" {
		return nil, fmt.Errorf(builder.errorMsg)
	}

	builder.Definition.ResourceVersion = builder.Object.ResourceVersion
	builder.Definition.ObjectMeta.ResourceVersion = builder.Object.ObjectMeta.ResourceVersion

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		if force {
			glog.V(100).Infof(
				msg.FailToUpdateNotification("metallb", builder.Definition.Name, builder.Definition.Namespace))

			builder, err := builder.Delete()

			if err != nil {
				glog.V(100).Infof(
					msg.FailToUpdateError("metallb", builder.Definition.Name, builder.Definition.Namespace))

				return nil, err
			}

			return builder.Create()
		}
	}

	return builder, err
}

// RemoveLabel removes given label from metallb metadata.
func (builder *Builder) RemoveLabel(key string) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Removing label %s from metalLbIo %s", key, builder.Definition.Name)

	if key == "" {
		glog.V(100).Infof("Failed to remove empty label's key from metalLbIo %s", builder.Definition.Name)
		builder.errorMsg = "error to remove empty key from metalLbIo"
	}

	if builder.errorMsg != "" {
		return builder
	}

	delete(builder.Definition.Spec.SpeakerNodeSelector, key)

	return builder
}

// WithSpeakerNodeSelector adds the specified label to the MetalLbIo SpeakerNodeSelector.
func (builder *Builder) WithSpeakerNodeSelector(label map[string]string) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Adding label selector %v to metallb.io object %s",
		label, builder.Definition.Name,
	)

	if len(label) < 1 {
		builder.errorMsg = "can not accept empty label and redefine metallb NodeSelector"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.SpeakerNodeSelector = label

	return builder
}

// WithOptions creates metallb with generic mutation options.
func (builder *Builder) WithOptions(options ...AdditionalOptions) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting metallb additional options")

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

// WithFRRConfigAlwaysBlock adds specific routes to block from being advertised to the FRR nodes.
func (builder *Builder) WithFRRConfigAlwaysBlock(prefixes []string) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Adding prefixes to block in the metallb.io object %s", builder.Definition.Name)

	if len(prefixes) < 1 {
		glog.V(100).Infof("the frrConfiguration prefixes list is empty")

		builder.errorMsg = "can not accept empty prefix list for the metallb alwaysBlock mode"

		return builder
	}

	// Validate CIDR prefixes
	for _, prefix := range prefixes {
		if _, _, err := net.ParseCIDR(prefix); err != nil {
			glog.V(100).Infof("the frrConfiguration prefix %s is not a valid CIDR", prefix)
			builder.errorMsg = fmt.Sprintf("the prefix %s is not a valid CIDR", prefix)

			return builder
		}
	}

	builder.Definition.Spec.FRRK8SConfig = &mlbtypes.FRRK8SConfig{AlwaysBlock: prefixes}

	return builder
}

// GetMetalLbIoGVR returns metalLb's GroupVersionResource which could be used for Clean function.
func GetMetalLbIoGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: "metallb.io", Version: "v1beta1", Resource: "metallbs",
	}
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *Builder) validate() (bool, error) {
	resourceCRD := "MetalLB"

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
