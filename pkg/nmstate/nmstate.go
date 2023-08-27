package nmstate

import (
	"context"
	"fmt"

	"github.com/golang/glog"

	nmstateV1 "github.com/nmstate/kubernetes-nmstate/api/v1"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Builder provides struct for the NMState object containing connection to
// the cluster and the NMState definitions.
type Builder struct {
	// NMState definition. Used to create NMState object with minimum set of required elements.
	Definition *nmstateV1.NMState
	// Created NMState object on the cluster.
	Object *nmstateV1.NMState
	// API client to interact with the cluster.
	apiClient *clients.Settings
	// errorMsg is processed before NMState object is created.
	errorMsg string
}

// NewBuilder creates a new instance of nmstate Builder.
func NewBuilder(apiClient *clients.Settings, name string) *Builder {
	glog.V(100).Infof("Initializing new NMState structure with the name: %s", name)

	builder := Builder{
		apiClient: apiClient,
		Definition: &nmstateV1.NMState{
			ObjectMeta: metaV1.ObjectMeta{
				Name: name,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the NMState is empty")

		builder.errorMsg = "NMState 'name' cannot be empty"
	}

	return &builder
}

// Exists checks whether the given NMState exists.
func (builder *Builder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if NMState %s exists", builder.Definition.Name)

	var err error
	builder.Object, err = builder.Get()

	if err != nil {
		glog.V(100).Infof("Failed to collect NMState object due to %s", err.Error())
	}

	return err == nil || !k8serrors.IsNotFound(err)
}

// Get returns NMState object if found.
func (builder *Builder) Get() (*nmstateV1.NMState, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Collecting NMState object %s", builder.Definition.Name)

	nmstate := &nmstateV1.NMState{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{Name: builder.Definition.Name}, nmstate)

	if err != nil {
		glog.V(100).Infof("NMState object %s doesn't exist", builder.Definition.Name)

		return nil, err
	}

	return nmstate, err
}

// Create makes a NMState in the cluster and stores the created object in struct.
func (builder *Builder) Create() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the NMState %s", builder.Definition.Name)

	var err error
	if !builder.Exists() {
		err = builder.apiClient.Create(context.TODO(), builder.Definition)
		if err == nil {
			builder.Object = builder.Definition
		}
	}

	return builder, err
}

// Delete removes NMState object from a cluster.
func (builder *Builder) Delete() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting the NMState object %s", builder.Definition.Name)

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return builder, fmt.Errorf("can not delete NMState: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// Update renovates the existing NMState object with the NMState definition in builder.
func (builder *Builder) Update(force bool) (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating the NMState object", builder.Definition.Name)

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		if force {
			glog.V(100).Infof(
				"Failed to update the NMState object %s. "+
					"Note: Force flag set, executed delete/create methods instead", builder.Definition.Name)

			builder, err := builder.Delete()

			if err != nil {
				glog.V(100).Infof(
					"Failed to update the NMState object %s, "+
						"due to error in delete function", builder.Definition.Name)

				return nil, err
			}

			return builder.Create()
		}
	}

	return builder, err
}

// PullNMstate retrieves an existing NMState object from the cluster.
func PullNMstate(apiClient *clients.Settings, name string) (*Builder, error) {
	glog.V(100).Infof("Pulling NMState object name: %s", name)

	builder := Builder{
		apiClient: apiClient,
		Definition: &nmstateV1.NMState{
			ObjectMeta: metaV1.ObjectMeta{
				Name: name,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the NMState is empty")

		builder.errorMsg = "NMState 'name' cannot be empty"
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("NMState object %s doesn't exist", name)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *Builder) validate() (bool, error) {
	resourceCRD := "NMState"

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
