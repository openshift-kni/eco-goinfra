package nmstate

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	assistedv1beta1 "github.com/openshift/assisted-service/api/v1beta1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NmStateConfigBuilder provides struct for the NMStateConfig object containing connection to
// the cluster and the NMStateConfig definitions.
type NmStateConfigBuilder struct {
	// NMStateConfig definition. Used to create NMStateConfig object with minimum set of required elements.
	Definition *assistedv1beta1.NMStateConfig
	// Created NMStateConfig object on the cluster.
	Object *assistedv1beta1.NMStateConfig
	// API client to interact with the cluster.
	apiClient *clients.Settings
	// errorMsg is processed before NMStateConfig object is created.
	errorMsg string
}

// NewBuilder creates a new instance of NMStateConfig Builder.
func NewNmStateConfigBuilder(apiClient *clients.Settings, name, namespace string) *NmStateConfigBuilder {
	glog.V(100).Infof("Initializing new NMStateConfig structure with the name: %s", name)

	builder := NmStateConfigBuilder{
		apiClient: apiClient,
		Definition: &assistedv1beta1.NMStateConfig{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the NMStateConfig is empty")

		builder.errorMsg = "NMStateConfig 'name' cannot be empty"
	}

	return &builder
}

// Exists checks whether the given NMStateConfig exists.
func (builder *NmStateConfigBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if NMStateConfig %s exists", builder.Definition.Name)

	var err error
	builder.Object, err = builder.Get()

	if err != nil {
		glog.V(100).Infof("Failed to collect NMStateConfig object due to %s", err.Error())
	}

	return err == nil || !k8serrors.IsNotFound(err)
}

// Get returns NMStateConfig object if found.
func (builder *NmStateConfigBuilder) Get() (*assistedv1beta1.NMStateConfig, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Collecting NMStateConfig object %s", builder.Definition.Name)

	nmStateConfig := &assistedv1beta1.NMStateConfig{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, nmStateConfig)

	if err != nil {
		glog.V(100).Infof("NMStateConfig object %s doesn't exist", builder.Definition.Name)

		return nil, err
	}

	return nmStateConfig, err
}

// Create makes a NMStateConfig in the cluster and stores the created object in struct.
func (builder *NmStateConfigBuilder) Create() (*NmStateConfigBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the NMStateConfig %s", builder.Definition.Name)

	var err error
	if !builder.Exists() {
		err = builder.apiClient.Create(context.TODO(), builder.Definition)
		if err == nil {
			builder.Object = builder.Definition
		}
	}

	return builder, err
}

// Delete removes NMStateConfig object from a cluster.
func (builder *NmStateConfigBuilder) Delete() (*NmStateConfigBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting the NMStateConfig object %s", builder.Definition.Name)

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return builder, fmt.Errorf("can not delete NMStateConfig: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *NmStateConfigBuilder) validate() (bool, error) {
	resourceCRD := "NMStateConfig"

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
