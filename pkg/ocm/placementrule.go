package ocm

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/internal/common"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	"k8s.io/apimachinery/pkg/runtime/schema"
	placementrulev1 "open-cluster-management.io/multicloud-operators-subscription/pkg/apis/apps/placementrule/v1"
)

// PlacementRuleBuilder provides struct for the PlacementRule object containing connection to
// the cluster and the PlacementRule definitions.
type PlacementRuleBuilder struct {
	common.EmbeddableBuilder[placementrulev1.PlacementRule, *placementrulev1.PlacementRule]
}

// NewPlacementRuleBuilder creates a new instance of PlacementRuleBuilder.
func NewPlacementRuleBuilder(apiClient *clients.Settings, name, nsname string) *PlacementRuleBuilder {
	if apiClient == nil {
		return nil
	}

	return common.NewNamespacedBuilder[placementrulev1.PlacementRule, PlacementRuleBuilder](
		apiClient.Client, placementrulev1.AddToScheme, name, nsname)
}

// PullPlacementRule pulls existing placementrule into Builder struct.
func PullPlacementRule(apiClient *clients.Settings, name, nsname string) (*PlacementRuleBuilder, error) {
	if apiClient == nil {
		return nil, fmt.Errorf("apiClient cannot be nil")
	}

	return common.PullNamespacedBuilder[placementrulev1.PlacementRule, PlacementRuleBuilder](
		apiClient.Client, placementrulev1.AddToScheme, name, nsname)
}

// Create makes a placementrule in the cluster and stores the created object in struct.
func (builder *PlacementRuleBuilder) Create() (*PlacementRuleBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	glog.V(100).Infof("Creating the placementrule %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if builder.Exists() {
		return builder, nil
	}

	err := builder.GetClient().Create(context.TODO(), builder.Definition)
	if err != nil {
		return builder, err
	}

	builder.Object = builder.Definition

	return builder, nil
}

// Delete removes a placementrule from a cluster.
func (builder *PlacementRuleBuilder) Delete() (*PlacementRuleBuilder, error) {
	if err := common.Validate(builder); err != nil {
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

	err := builder.GetClient().Delete(context.TODO(), builder.Definition)
	if err != nil {
		return builder, fmt.Errorf("cannot delete placementrule: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// Update renovates the existing placementrule object with the placementrule definition in builder.
func (builder *PlacementRuleBuilder) Update(force bool) (*PlacementRuleBuilder, error) {
	if err := common.Validate(builder); err != nil {
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
	err := builder.GetClient().Update(context.TODO(), builder.Definition)

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

	return builder, nil
}

// GetKind returns the GVK of the PlacementRule object. It may be called on a nil builder or a zero value of the
// builder.
func (builder *PlacementRuleBuilder) GetKind() schema.GroupVersionKind {
	return placementrulev1.SchemeGroupVersion.WithKind("PlacementRule")
}
