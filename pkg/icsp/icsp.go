package icsp

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	v1alpha1 "github.com/openshift/api/operator/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ICSPBuilder provides struct for the ImageContentSourcePolicy object with connection to the cluster.
type ICSPBuilder struct {
	// ImageContentSourcePolicy definition. Used to create ImageContentSourcePolicy object.
	Definition *v1alpha1.ImageContentSourcePolicy
	// Created ImageContentSourcePolicy object.
	Object *v1alpha1.ImageContentSourcePolicy
	// Used in functions that defines or mutates ImageContentSourcePolicy definition.
	// errorMsg is processed before the ImageContentSourcePolicy object is created.
	apiClient runtimeclient.Client
	errorMsg  string
}

// AdditionalOptions additional options for ImageContentSourcePolicy object.
type AdditionalOptions func(builder *ICSPBuilder) (*ICSPBuilder, error)

// NewICSPBuilder creates a new instance of ICSPBuilder.
func NewICSPBuilder(apiClient *clients.Settings, name, source string, mirrors []string) *ICSPBuilder {
	glog.V(100).Infof(
		"Initializing new ICSPBuilder structure with the following params: "+
			"name: %s, source: %s, mirrors: %v",
		name, source, mirrors)

	if apiClient == nil {
		glog.V(100).Info("ImageContentSourcePolicy apiClient cannot be nil")

		return nil
	}

	err := apiClient.AttachScheme(v1alpha1.Install)
	if err != nil {
		glog.V(100).Info("Failed to add operator v1alpha1 scheme to client schemes")

		return nil
	}

	icspBuilder := &ICSPBuilder{
		apiClient: apiClient.Client,
		Definition: &v1alpha1.ImageContentSourcePolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: v1alpha1.ImageContentSourcePolicySpec{
				RepositoryDigestMirrors: []v1alpha1.RepositoryDigestMirrors{
					{
						Source:  source,
						Mirrors: mirrors,
					},
				},
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the ImageContentSourcePolicy is empty")

		icspBuilder.errorMsg = "ImageContentSourcePolicy 'name' cannot be empty"

		return icspBuilder
	}

	if source == "" {
		glog.V(100).Infof("The Source of the ImageContentSourcePolicy is empty")

		icspBuilder.errorMsg = "ImageContentSourcePolicy 'source' cannot be empty"

		return icspBuilder
	}

	if len(mirrors) == 0 {
		glog.V(100).Infof("The mirrors of the ImageContentSourcePolicy are empty")

		icspBuilder.errorMsg = "ImageContentSourcePolicy 'mirrors' cannot be empty"

		return icspBuilder
	}

	return icspBuilder
}

// Pull pulls object definition from cluster to ICSPBuilder struct.
func Pull(apiClient *clients.Settings, name string) (*ICSPBuilder, error) {
	glog.V(100).Infof("Pulling existing ImageContentSourcePolicy: %s", name)

	if apiClient == nil {
		glog.V(100).Info("ImageContentSourcePolicy apiClient cannot be nil")

		return nil, fmt.Errorf("imageContentSourcePolicy 'apiClient' cannot be nil")
	}

	err := apiClient.AttachScheme(v1alpha1.Install)
	if err != nil {
		glog.V(100).Info("Failed to add operator v1alpha1 scheme to client schemes")

		return nil, err
	}

	builder := ICSPBuilder{
		apiClient: apiClient.Client,
		Definition: &v1alpha1.ImageContentSourcePolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	if name == "" {
		glog.V(100).Info("The name of the ImageContentSourcePolicy is empty")

		return nil, fmt.Errorf("imageContentSourcePolicy 'name' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("imageContentSourcePolicy object %s does not exist", name)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Get returns the ImageContentSourcePolicy if found.
func (builder *ICSPBuilder) Get() (*v1alpha1.ImageContentSourcePolicy, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting ImageContentSourcePolicy object %s", builder.Definition.Name)

	icsp := &v1alpha1.ImageContentSourcePolicy{}
	err := builder.apiClient.Get(context.TODO(), runtimeclient.ObjectKey{Name: builder.Definition.Name}, icsp)

	if err != nil {
		glog.V(100).Infof("ImageContentSourcePolicy object %s does not exist", builder.Definition.Name)

		return nil, err
	}

	return icsp, nil
}

// Exists check if object exists in the cluster.
func (builder *ICSPBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if ImageContentSourcePolicy %s exists", builder.Definition.Name)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Create makes a ImageContentSourcePolicy in the cluster and stores the created object in struct.
func (builder *ICSPBuilder) Create() (*ICSPBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating ImageContentSourcePolicy %s", builder.Definition.Name)

	var err error
	if !builder.Exists() {
		err = builder.apiClient.Create(context.TODO(), builder.Definition)
		if err == nil {
			builder.Object = builder.Definition
		}
	}

	return builder, err
}

// Delete removes an ImageContentSourcePolicy.
func (builder *ICSPBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting ImageContentSourcePolicy %s", builder.Definition.Name)

	if !builder.Exists() {
		glog.V(100).Infof("ImageContentSourcePolicy %s cannot be deleted because it does not exist", builder.Definition.Name)

		builder.Object = nil

		return nil
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Object)
	if err != nil {
		return err
	}

	builder.Object = nil

	return nil
}

// Update renovates the existing ImageContentSourcePolicy object with the definition in ICSPbuilder.
func (builder *ICSPBuilder) Update() (*ICSPBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof(
		"Updating the ImageContentSourcePolicy %s with the definition in the ICSPbuilder", builder.Definition.Name)

	if !builder.Exists() {
		glog.V(100).Infof("ImageContentSourcePolicy %s does not exist", builder.Definition.Name)

		return nil, fmt.Errorf("cannot update non-existent ImageContentSourcePolicy")
	}

	builder.Definition.ResourceVersion = builder.Object.ResourceVersion
	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		glog.V(100).Infof("Failed to update ImageContentSourcePolicy %s", builder.Definition.Name)

		return nil, err
	}

	builder.Object = builder.Definition

	return builder, nil
}

// WithRepositoryDigestMirror adds new RepositoryDigestMirror.
func (builder *ICSPBuilder) WithRepositoryDigestMirror(source string, mirrors []string) *ICSPBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Adding RepositoryDigestMirror with source %s and mirrors %v to ImageContentSourcePolicy %s",
		source, mirrors, builder.Definition.Name)

	if source == "" {
		glog.V(100).Infof("The source is empty")

		builder.errorMsg = "'source' cannot be empty"

		return builder
	}

	if len(mirrors) == 0 {
		glog.V(100).Infof("Mirrors is empty")

		builder.errorMsg = "'mirrors' cannot be empty"

		return builder
	}

	builder.Definition.Spec.RepositoryDigestMirrors = append(
		builder.Definition.Spec.RepositoryDigestMirrors, v1alpha1.RepositoryDigestMirrors{Source: source, Mirrors: mirrors})

	return builder
}

// WithOptions creates ImageContentPolicy with generic mutation options.
func (builder *ICSPBuilder) WithOptions(options ...AdditionalOptions) *ICSPBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting ImageContentPolicy additional options")

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

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *ICSPBuilder) validate() (bool, error) {
	resourceCRD := "ImageContentSourcePolicy"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Definition == nil {
		glog.V(100).Infof("The %s is undefined", resourceCRD)

		return false, fmt.Errorf(msg.UndefinedCrdObjectErrString(resourceCRD))
	}

	if builder.apiClient == nil {
		glog.V(100).Infof("The %s builder apiclient is nil", resourceCRD)

		return false, fmt.Errorf("%s builder cannot have nil apiClient", resourceCRD)
	}

	if builder.errorMsg != "" {
		glog.V(100).Infof("The %s builder has error message: %s", resourceCRD, builder.errorMsg)

		return false, fmt.Errorf(builder.errorMsg)
	}

	return true, nil
}
