package sriov

import (
	"context"
	"fmt"

	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
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
	apiClient runtimeClient.Client
	errorMsg  string
}

// NewOperatorConfigBuilder creates new instance of OperatorConfigBuilder.
func NewOperatorConfigBuilder(apiClient *clients.Settings, nsname string) *OperatorConfigBuilder {
	glog.V(100).Infof(
		"Initializing new OperatorConfigBuilder structure with the following params: namespace: %s", nsname)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient cannot be nil")

		return nil
	}

	err := apiClient.AttachScheme(srIovV1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add sriovv1 scheme to client schemes")

		return nil
	}

	builder := &OperatorConfigBuilder{
		apiClient: apiClient.Client,
		Definition: &srIovV1.SriovOperatorConfig{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      sriovOperatorConfigName,
				Namespace: nsname,
			},
		},
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the SriovOperatorConfig is empty")

		builder.errorMsg = "SriovOperatorConfig 'nsname' is empty"
	}

	return builder
}

// Create generates SriovOperatorConfig in a cluster and stores the created object in struct.
func (builder *OperatorConfigBuilder) Create() (*OperatorConfigBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the SriovOperatorConfig in namespace %s", builder.Definition.Namespace)

	if !builder.Exists() {
		err := builder.apiClient.Create(context.TODO(), builder.Definition)

		if err != nil {
			glog.V(100).Infof("Failed to create the SriovOperatorConfig")

			return nil, err
		}
	}

	builder.Object = builder.Definition

	return builder, nil
}

// PullOperatorConfig loads an existing SriovOperatorConfig into OperatorConfigBuilder struct.
func PullOperatorConfig(apiClient *clients.Settings, nsname string) (*OperatorConfigBuilder, error) {
	glog.V(100).Infof("Pulling existing default SriovOperatorConfig: %s", sriovOperatorConfigName)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")

		return nil, fmt.Errorf("SriovOperatorConfig 'apiClient' cannot be empty")
	}

	err := apiClient.AttachScheme(srIovV1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add sriovv1 scheme to client schemes")

		return nil, err
	}

	builder := OperatorConfigBuilder{
		apiClient: apiClient.Client,
		Definition: &srIovV1.SriovOperatorConfig{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      sriovOperatorConfigName,
				Namespace: nsname,
			},
		},
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the SriovOperatorConfig is empty")

		return nil, fmt.Errorf("SriovOperatorConfig 'nsname' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("SriovOperatorConfig object %s does not exist in namespace %s",
			sriovOperatorConfigName, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Get returns CatalogSource object if found.
func (builder *OperatorConfigBuilder) Get() (*srIovV1.SriovOperatorConfig, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Collecting SriovOperatorConfig object %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	operatorConfig := &srIovV1.SriovOperatorConfig{}
	err := builder.apiClient.Get(context.TODO(),
		runtimeClient.ObjectKey{Name: builder.Definition.Name, Namespace: builder.Definition.Namespace},
		operatorConfig)

	if err != nil {
		glog.V(100).Infof("SriovOperatorConfig object %s does not exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)

		return nil, err
	}

	return operatorConfig, nil
}

// Exists checks whether the given SriovOperatorConfig exists.
func (builder *OperatorConfigBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof(
		"Checking if SriovOperatorConfig %s exists", builder.Definition.Name)

	var err error
	builder.Object, err = builder.Get()

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

	builder.Definition.Spec.EnableInjector = enable

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

	builder.Definition.Spec.EnableOperatorWebhook = enable

	return builder
}

// WithConfigDaemonNodeSelector configures configDaemonNodeSelector in the SriovOperatorConfig.
func (builder *OperatorConfigBuilder) WithConfigDaemonNodeSelector(
	configDaemonNodeSelector map[string]string) *OperatorConfigBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Configuring configDaemonNodeSelector %s in SriovOperatorConfig object %s in namespace %s",
		configDaemonNodeSelector, builder.Definition.Name, builder.Definition.Namespace,
	)

	if len(configDaemonNodeSelector) == 0 {
		glog.V(100).Infof("The 'configDaemonNodeSelector' of the SriovOperatorConfig is empty")

		builder.errorMsg = "can not apply empty configDaemonNodeSelector"

		return builder
	}

	for selectorKey := range configDaemonNodeSelector {
		if selectorKey == "" {
			glog.V(100).Infof("The 'configDaemonNodeSelector' selectorKey cannot be empty")

			builder.errorMsg = "can not apply configDaemonNodeSelector with an empty selectorKey value"

			return builder
		}
	}

	builder.Definition.Spec.ConfigDaemonNodeSelector = configDaemonNodeSelector

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

	if err == nil {
		builder.Object = builder.Definition
	}

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

		return false, fmt.Errorf("%s builder cannot have nil apiClient", resourceCRD)
	}

	if builder.errorMsg != "" {
		glog.V(100).Infof("The %s builder has error message: %s", resourceCRD, builder.errorMsg)

		return false, fmt.Errorf(builder.errorMsg)
	}

	return true, nil
}
