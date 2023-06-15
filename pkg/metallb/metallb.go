package metallb

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/golang/glog"
	"github.com/metallb/metallb-operator/api/v1beta1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Builder provides struct for the MetalLb object containing connection to
// the cluster and the MetalLb definitions.
type Builder struct {
	Definition *v1beta1.MetalLB
	Object     *v1beta1.MetalLB
	apiClient  *clients.Settings
	errorMsg   string
}

// AdditionalOptions additional options for metallb object.
type AdditionalOptions func(builder *Builder) (*Builder, error)

// NewBuilder creates a new instance of Builder.
func NewBuilder(apiClient *clients.Settings, name, nsname string, label map[string]string) *Builder {
	glog.V(100).Infof(
		"Initializing new metallb structure with the following params: %s, %s, %v",
		name, nsname, label)

	builder := Builder{
		apiClient: apiClient,
		Definition: &v1beta1.MetalLB{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			}, Spec: v1beta1.MetalLBSpec{
				SpeakerNodeSelector: label,
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

	return &builder
}

// Pull retrieves an existing metallb.io object from the cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	glog.V(100).Infof(
		"Pulling metallb.io object name:%s in namespace: %s", name, nsname)

	builder := Builder{
		apiClient: apiClient,
		Definition: &v1beta1.MetalLB{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
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

	if !builder.Exists() {
		return nil, fmt.Errorf("metallb oject %s doesn't exist in namespace %s", name, nsname)
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
func (builder *Builder) Get() (*v1beta1.MetalLB, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof(
		"Collecting metallb object %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	metalLb := &v1beta1.MetalLB{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, metalLb)

	if err != nil {
		glog.V(100).Infof(
			"metallb object %s doesn't exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)

		return nil, err
	}

	return metalLb, err
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
		if err == nil {
			builder.Object = builder.Definition
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
		return builder, fmt.Errorf("metallb cannot be deleted because it does not exist")
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

	glog.V(100).Infof("Updating the metallb object %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace,
	)

	if builder.errorMsg != "" {
		return nil, fmt.Errorf(builder.errorMsg)
	}

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		if force {
			glog.V(100).Infof(
				"Failed to update the metallb object %s in namespace %s. "+
					"Note: Force flag set, executed delete/create methods instead",
				builder.Definition.Name, builder.Definition.Namespace,
			)

			builder, err := builder.Delete()

			if err != nil {
				glog.V(100).Infof(
					"Failed to update the metallb object %s in namespace %s, "+
						"due to error in delete function",
					builder.Definition.Name, builder.Definition.Namespace,
				)

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
