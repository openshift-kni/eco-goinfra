package ocm

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	placementrulev1 "open-cluster-management.io/multicloud-operators-subscription/pkg/apis/apps/placementrule/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// PlacementRuleBuilder provides struct for the PlacementRule object containing connection to
// the cluster and the PlacementRule definitions.
type PlacementRuleBuilder struct {
	// PlacementRule Definition, used to create the PlacementRule object.
	Definition *placementrulev1.PlacementRule
	// created PlacementRule object.
	Object *placementrulev1.PlacementRule
	// api client to interact with the cluster.
	apiClient runtimeclient.Client
	// used to store latest error message upon defining or mutating PlacementRule definition.
	errorMsg string
}

// NewPlacementRuleBuilder creates a new instance of PlacementRuleBuilder.
func NewPlacementRuleBuilder(apiClient *clients.Settings, name, nsname string) *PlacementRuleBuilder {
	glog.V(100).Infof(
		"Initializing new placement rule structure with the following params: name: %s, nsname: %s",
		name, nsname)

	if apiClient == nil {
		glog.V(100).Info("The apiClient of the PlacementRule is nil")

		return nil
	}

	err := apiClient.AttachScheme(placementrulev1.AddToScheme)
	if err != nil {
		glog.V(100).Info("Failed to add PlacementRule scheme to client schemes")

		return nil
	}

	builder := PlacementRuleBuilder{
		apiClient: apiClient.Client,
		Definition: &placementrulev1.PlacementRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Info("The name of the PlacementRule is empty")

		builder.errorMsg = "placementrule's 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Info("The namespace of the PlacementRule is empty")

		builder.errorMsg = "placementrule's 'nsname' cannot be empty"
	}

	return &builder
}

// PullPlacementRule pulls existing placementrule into Builder struct.
func PullPlacementRule(apiClient *clients.Settings, name, nsname string) (*PlacementRuleBuilder, error) {
	glog.V(100).Infof("Pulling existing placementrule name %s under namespace %s from cluster", name, nsname)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")

		return nil, fmt.Errorf("placementrule's 'apiClient' cannot be empty")
	}

	err := apiClient.AttachScheme(placementrulev1.AddToScheme)
	if err != nil {
		glog.V(100).Info("Failed to add PlacementRule scheme to client schemes")

		return nil, err
	}

	builder := PlacementRuleBuilder{
		apiClient: apiClient.Client,
		Definition: &placementrulev1.PlacementRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the placementrule is empty")

		return nil, fmt.Errorf("placementrule's 'name' cannot be empty")
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the placementrule is empty")

		return nil, fmt.Errorf("placementrule's 'namespace' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("placementrule object %s does not exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Exists checks whether the given placementrule exists.
func (builder *PlacementRuleBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if placementrule %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Get returns a placementrule object if found.
func (builder *PlacementRuleBuilder) Get() (*placementrulev1.PlacementRule, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting placementrule %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	placementRule := &placementrulev1.PlacementRule{}

	err := builder.apiClient.Get(context.TODO(), runtimeclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, placementRule)

	if err != nil {
		return nil, err
	}

	return placementRule, err
}

// Create makes a placementrule in the cluster and stores the created object in struct.
func (builder *PlacementRuleBuilder) Create() (*PlacementRuleBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the placementrule %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	if !builder.Exists() {
		err = builder.apiClient.Create(context.TODO(), builder.Definition)
		if err == nil {
			builder.Object = builder.Definition
		}
	}

	return builder, err
}

// Delete removes a placementrule from a cluster.
func (builder *PlacementRuleBuilder) Delete() (*PlacementRuleBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting the placementrule %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		glog.V(100).Infof("placementrule %s cannot be deleted because it does not exist",
			builder.Definition.Name)

		builder.Object = nil

		return builder, nil
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)
	if err != nil {
		return builder, fmt.Errorf("cannot delete placementrule: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// Update renovates the existing placementrule object with the placementrule definition in builder.
func (builder *PlacementRuleBuilder) Update(force bool) (*PlacementRuleBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	if !builder.Exists() {
		glog.V(100).Infof(
			"PlacementRule %s does not exist in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

		return nil, fmt.Errorf("cannot update non-existent placementrule")
	}

	glog.V(100).Infof("Updating the placementrule object: %s in namespace: %s",
		builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.ResourceVersion = builder.Object.ResourceVersion
	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		if force {
			glog.V(100).Infof(
				msg.FailToUpdateNotification("placementrule", builder.Definition.Name, builder.Definition.Namespace))

			builder, err := builder.Delete()
			builder.Definition.ResourceVersion = ""

			if err != nil {
				glog.V(100).Infof(msg.FailToUpdateError("placementrule", builder.Definition.Name, builder.Definition.Namespace))

				return nil, err
			}

			return builder.Create()
		}
	}

	builder.Object = builder.Definition

	return builder, err
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *PlacementRuleBuilder) validate() (bool, error) {
	resourceCRD := "placementRule"

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
