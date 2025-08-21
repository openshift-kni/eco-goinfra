package sriov

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/msg"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	sriovOperatorConfigName = "default"
)

// OperatorConfigBuilder provides a struct for SriovOperatorConfig object from the cluster and
// a SriovOperatorConfig definition.
type OperatorConfigBuilder struct {
	// SriovOperatorConfig definition, used to create the SriovOperatorConfig object.
	Definition *srIovV1.SriovOperatorConfig
	// Created SriovOperatorConfig object.
	Object *srIovV1.SriovOperatorConfig
	// api client to interact with the cluster.
	apiClient *clients.Settings
	errorMsg  string
}

// PullOperatorConfig loads an existing SriovOperatorConfig into OperatorConfigBuilder struct.
func PullOperatorConfig(apiClient *clients.Settings, nsname string) (*OperatorConfigBuilder, error) {
	glog.V(100).Infof("Pulling existing default SriovOperatorConfig: %s", sriovOperatorConfigName)

	builder := OperatorConfigBuilder{
		apiClient: apiClient,
		Definition: &srIovV1.SriovOperatorConfig{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      sriovOperatorConfigName,
				Namespace: nsname,
			},
		},
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the SriovOperatorConfig is empty")

		builder.errorMsg = "SriovOperatorConfig 'nsname' cannot be empty"
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("SriovOperatorConfig object %s doesn't exist", sriovOperatorConfigName)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Exists checks whether the given SriovOperatorConfig exists.
func (builder *OperatorConfigBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof(
		"Checking if SriovOperatorConfig %s exists", builder.Definition.Name)

	var err error
	builder.Object, err = builder.apiClient.SriovOperatorConfigs(builder.Definition.Namespace).
		Get(context.TODO(), sriovOperatorConfigName, metaV1.GetOptions{})

	return err == nil || !k8serrors.IsNotFound(err)
}

// WithInjector configures enableInjector in the SriovOperatorConfig.
func (builder *OperatorConfigBuilder) WithInjector(enable bool) *OperatorConfigBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Configuring enableInjector %t to SriovOperatorConfig object %s",
		enable, builder.Definition.Name,
	)

	builder.Definition.Spec.EnableInjector = &enable

	return builder
}

// WithOperatorWebhook configures enableOperatorWebhook in the SriovOperatorConfig.
func (builder *OperatorConfigBuilder) WithOperatorWebhook(enable bool) *OperatorConfigBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Configuring WithOperatorWebhook %t to SriovOperatorConfig object %s",
		enable, builder.Definition.Name,
	)

	builder.Definition.Spec.EnableOperatorWebhook = &enable

	return builder
}

// Update renovates the existing SriovOperatorConfig object with the new definition in builder.
func (builder *OperatorConfigBuilder) Update() (*OperatorConfigBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating the SriovOperatorConfig object %s",
		builder.Definition.Name,
	)

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	return builder, err
}

// Delete removes SriovOperatorConfig object from a cluster.
func (builder *OperatorConfigBuilder) Delete() (*OperatorConfigBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting the SriovOperatorConfig object %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace,
	)

	if !builder.Exists() {
		return builder, fmt.Errorf("SriovOperatorConfig cannot be deleted because it does not exist")
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)
	if err != nil {
		return builder, fmt.Errorf("can not delete SriovOperatorConfig: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *OperatorConfigBuilder) validate() (bool, error) {
	resourceCRD := "SriovOperatorConfig"

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

		return false, fmt.Errorf(fmt.Sprintf("%s builder cannot have nil apiClient", resourceCRD))
	}

	return true, nil
}
