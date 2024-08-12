package nrop

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	nropv1 "github.com/openshift-kni/numaresources-operator/api/numaresourcesoperator/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Builder provides a struct for NUMAResourcesOperator object from the cluster and a NUMAResourcesOperator definition.
type Builder struct {
	// NUMAResourcesOperator definition, used to create the NUMAResourcesOperator object.
	Definition *nropv1.NUMAResourcesOperator
	// Created NUMAResourcesOperator object.
	Object *nropv1.NUMAResourcesOperator
	// Used to store latest error message upon defining or mutating NUMAResourcesOperator definition.
	errorMsg string
	// api client to interact with the cluster.
	apiClient goclient.Client
}

// NewBuilder creates a new instance of NUMAResourcesOperator.
func NewBuilder(
	apiClient *clients.Settings, name string) *Builder {
	glog.V(100).Infof(
		"Initializing new NUMAResourcesOperator structure with the following name: %s", name)

	if apiClient == nil {
		glog.V(100).Infof("NUMAResourcesOperator 'apiClient' cannot be empty")

		return nil
	}

	err := apiClient.AttachScheme(nropv1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add nrop v1 scheme to client schemes")

		return nil
	}

	builder := &Builder{
		apiClient: apiClient.Client,
		Definition: &nropv1.NUMAResourcesOperator{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the NUMAResourcesOperator is empty")

		builder.errorMsg = "NUMAResourcesOperator 'name' cannot be empty"

		return builder
	}

	return builder
}

// Pull pulls existing NUMAResourcesOperator from cluster.
func Pull(apiClient *clients.Settings, name string) (*Builder, error) {
	glog.V(100).Infof("Pulling existing NUMAResourcesOperator %s from the cluster", name)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")

		return nil, fmt.Errorf("NUMAResourcesOperator 'apiClient' cannot be empty")
	}

	err := apiClient.AttachScheme(nropv1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add nrop v1 scheme to client schemes")

		return nil, err
	}

	builder := Builder{
		apiClient: apiClient.Client,
		Definition: &nropv1.NUMAResourcesOperator{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the NUMAResourcesOperator is empty")

		return nil, fmt.Errorf("NUMAResourcesOperator 'name' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("NUMAResourcesOperator object %s does not exist", name)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Get fetches the defined NUMAResourcesOperator from the cluster.
func (builder *Builder) Get() (*nropv1.NUMAResourcesOperator, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting NUMAResourcesOperator %s", builder.Definition.Name)

	nropObj := &nropv1.NUMAResourcesOperator{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name: builder.Definition.Name,
	}, nropObj)

	if err != nil {
		return nil, err
	}

	return nropObj, nil
}

// Create makes a NUMAResourcesOperator in the cluster and stores the created object in struct.
func (builder *Builder) Create() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the NUMAResourcesOperator %s", builder.Definition.Name)

	var err error
	if !builder.Exists() {
		err = builder.apiClient.Create(context.TODO(), builder.Definition)
		if err == nil {
			builder.Object = builder.Definition
		}
	}

	return builder, err
}

// Delete removes NUMAResourcesOperator from a cluster.
func (builder *Builder) Delete() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting the NUMAResourcesOperator %s", builder.Definition.Name)

	if !builder.Exists() {
		glog.V(100).Infof("NUMAResourcesOperator %s cannot be deleted because it does not exist",
			builder.Definition.Name)

		builder.Object = nil

		return builder, nil
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return builder, fmt.Errorf("can not delete NUMAResourcesOperator: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// Exists checks whether the given NUMAResourcesOperator exists.
func (builder *Builder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if NUMAResourcesOperator %s exists", builder.Definition.Name)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Update renovates the existing NUMAResourcesOperator object with NUMAResourcesOperator definition in builder.
func (builder *Builder) Update() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating NUMAResourcesOperator %s", builder.Definition.Name)

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		glog.V(100).Infof(
			msg.FailToUpdateError("NUMAResourcesOperator", builder.Definition.Name))

		return nil, err
	}

	builder.Object = builder.Definition

	return builder, nil
}

// WithMCPSelector sets the NUMAResourcesOperator operator's mcpSelector.
func (builder *Builder) WithMCPSelector(config nropv1.NodeGroupConfig, mcpSelector metav1.LabelSelector) *Builder {
	glog.V(100).Infof(
		"Adding machineConfigPoolSelector to the NUMAResourcesOperator %s; machineConfigPoolSelector %v",
		builder.Definition.Name, mcpSelector)

	if valid, _ := builder.validate(); !valid {
		return builder
	}

	nodeGroup := nropv1.NodeGroup{
		Config:                    &config,
		MachineConfigPoolSelector: &mcpSelector,
	}

	if len(mcpSelector.MatchLabels) == 0 && len(mcpSelector.MatchExpressions) == 0 {
		glog.V(100).Infof("There are no labels for the machineConfigPoolSelector")

		builder.errorMsg = "NUMAResourcesOperator 'machineConfigPoolSelector' cannot be empty"

		return builder
	}

	builder.Definition.Spec.NodeGroups = append(builder.Definition.Spec.NodeGroups, nodeGroup)

	return builder
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *Builder) validate() (bool, error) {
	resourceCRD := "NUMAResourcesOperator"

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
