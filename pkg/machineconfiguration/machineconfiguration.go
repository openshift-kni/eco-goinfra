package machineconfiguration

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	machineconfigurationv1 "github.com/openshift/api/operator/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// APIGroup represents metallb api group.
const (
	APIGroup   = "operator.openshift.io"
	APIVersion = "v1"
)

// MCBuilder provides struct for a MachineConfiguration object.
type MCBuilder struct {
	Definition *machineconfigurationv1.MachineConfiguration
	Object     *machineconfigurationv1.MachineConfiguration
	apiClient  runtimeClient.Client
	errorMsg   string
}

// NewMachineConfigurationBuilder creates a new instance of machineconfiguration.
func NewMachineConfigurationBuilder(apiClient *clients.Settings, name string) *MCBuilder {

	err := apiClient.AttachScheme(machineconfigurationv1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add metallb scheme to client schemes")

		return nil
	}

	return &MCBuilder{
		apiClient: apiClient,
		Definition: &machineconfigurationv1.MachineConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}
}

// Get retrieves existing machineconfiguration from cluster if found.
func (builder *MCBuilder) Get() (*machineconfigurationv1.MachineConfiguration, error) {
	glog.V(100).Infof("Pulling existing machineSet name")

	machineConfig := &machineconfigurationv1.MachineConfiguration{}
	err := builder.apiClient.Get(context.TODO(),
		runtimeClient.ObjectKey{Name: builder.Definition.Name}, machineConfig)

	return machineConfig, err
}

// Create makes a Machineconfiguration in the cluster and stores the created object in struct.
func (builder *MCBuilder) Create() (*MCBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the Machineconfiguration %s",
		builder.Definition.Name,
	)

	var err error
	if !builder.Exists() {
		err = builder.apiClient.Create(context.TODO(), builder.Definition)

		if err != nil {
			glog.V(100).Infof("Failed to create Machineconfiguration")

			return nil, err
		}

		builder.Object = builder.Definition

		if err != nil {
			return nil, err
		}
	}

	return builder, err
}

// Pull pulls existing machineconfiguration from cluster.
func Pull(apiClient *clients.Settings, name string) (b *MCBuilder, err error) {
	glog.V(100).Infof("Pulling existing machineSet name %s ", name)

	err = apiClient.AttachScheme(machineconfigurationv1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add machineconfiguration scheme to client schemes")

		return nil, err
	}

	builder := MCBuilder{
		apiClient: apiClient,
		Definition: &machineconfigurationv1.MachineConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	builder.Definition = builder.Object

	return &MCBuilder{}, nil
}

// GetMachineConfigurationGVR returns machineconfiguration's GroupVersionResource which could be used for Clean function.
func GetMachineConfigurationGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: "machinesets",
	}
}

// Update renovates the existing BGPPeer object with the BGPPeer definition in builder.
func (builder *MCBuilder) Update(force bool) (*MCBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating the machineconfiguration object %s",
		builder.Definition.Name,
	)

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		if force {
			glog.V(100).Infof(
				msg.FailToUpdateNotification("Machineconfiguration", builder.Definition.Name))

			builder, err := builder.Delete()

			if err != nil {
				glog.V(100).Infof(
					msg.FailToUpdateError("Machineconfiguration", builder.Definition.Name))

				return nil, err
			}

			return builder.Create()
		}
	}

	return builder, err
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *MCBuilder) validate() (bool, error) {
	resourceCRD := "BGPPeer"

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

// Exists checks whether the given machineconfiguration exists.
func (builder *MCBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof(
		"Checking if BGPPeer %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Delete removes Machineconfiguration object from a cluster.
func (builder *MCBuilder) Delete() (*MCBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting the BGPPeer object %s",
		builder.Definition.Name,
	)

	if !builder.Exists() {
		glog.V(100).Infof("Machineconfiguration object %s does not exist",
			builder.Definition.Name)

		builder.Object = nil

		return builder, nil
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return builder, fmt.Errorf("can not delete machineconfiguration: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// WithNodeDisruptionPolicy defines the nodeDisruptionPolicy on the cluster nodes.
func (builder *MCBuilder) WithNodeDisruptionPolicy(name, path string) *MCBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Creating machineconfiguration with this NodeDisruptionPolicy: %s",
		builder.Definition.Name)

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.NodeDisruptionPolicy = machineconfigurationv1.NodeDisruptionPolicyConfig{
		Files: []machineconfigurationv1.NodeDisruptionPolicySpecFile{
			{
				Actions: []machineconfigurationv1.NodeDisruptionPolicySpecAction{
					{
						Type: "Restart",
						Restart: &machineconfigurationv1.RestartService{
							ServiceName: machineconfigurationv1.NodeDisruptionPolicyServiceName(name),
						},
					},
				},
				Path: path,
			},
		},
		Units: []machineconfigurationv1.NodeDisruptionPolicySpecUnit{
			{
				Actions: []machineconfigurationv1.NodeDisruptionPolicySpecAction{
					{
						Type: "Reload",
						Reload: &machineconfigurationv1.ReloadService{
							ServiceName: machineconfigurationv1.NodeDisruptionPolicyServiceName(name),
						},
					},
					{
						Type: "DaemonReload",
					},
				},
				Name: machineconfigurationv1.NodeDisruptionPolicyServiceName(name),
			},
		},
	}

	return builder
}
