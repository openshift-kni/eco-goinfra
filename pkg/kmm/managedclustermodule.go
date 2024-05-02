package kmm

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	mcmV1Beta1 "github.com/rh-ecosystem-edge/kernel-module-management/api-hub/v1beta1"
	moduleV1Beta1 "github.com/rh-ecosystem-edge/kernel-module-management/api/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ManagedClusterModuleBuilder provides struct for the managedclustermodule object containing connection
// to cluster and the managedclustermodule definitions.
type ManagedClusterModuleBuilder struct {
	Definition *mcmV1Beta1.ManagedClusterModule
	Object     *mcmV1Beta1.ManagedClusterModule
	errorMsg   string
	apiClient  *clients.Settings
}

// ManagedClusterModuleAdditionalOptions additional options for managedclustermodule object.
type ManagedClusterModuleAdditionalOptions func(builder *ManagedClusterModuleBuilder) (
	*ManagedClusterModuleBuilder, error)

// NewManagedClusterModuleBuilder creates a new instance of ManagedClusterModuleBuilder.
func NewManagedClusterModuleBuilder(apiClient *clients.Settings, name, nsname string) *ManagedClusterModuleBuilder {
	glog.V(100).Infof(
		"Initializing new ManagedClusterModule structure with following params: %s, %s", name, nsname)

	builder := ManagedClusterModuleBuilder{
		apiClient: apiClient,
		Definition: &mcmV1Beta1.ManagedClusterModule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the ManagedClusterModule is empty")

		builder.errorMsg = "ManagedClusterModule 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the ManagedClusterModule is empty")

		builder.errorMsg = "ManagedClusterModule 'nsname' cannot be empty"
	}

	return &builder
}

// WithModuleSpec sets the ModuleSpec.
func (builder *ManagedClusterModuleBuilder) WithModuleSpec(
	moduleSpec moduleV1Beta1.ModuleSpec) *ManagedClusterModuleBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	builder.Definition.Spec.ModuleSpec = moduleSpec

	return builder
}

// WithSpokeNamespace sets the namespace where the module will be deployed on the spoke.
func (builder *ManagedClusterModuleBuilder) WithSpokeNamespace(
	spokeNamespace string) *ManagedClusterModuleBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	if spokeNamespace == "" {
		builder.errorMsg = "invalid 'spokeNamespace' argument cannot be nil"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.SpokeNamespace = spokeNamespace

	return builder
}

// WithSelector sets the selector for the managedclustermodule object.
func (builder *ManagedClusterModuleBuilder) WithSelector(
	selector map[string]string) *ManagedClusterModuleBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	if len(selector) == 0 {
		builder.errorMsg = "invalid 'selector' argument cannot be empty"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.Selector = selector

	return builder
}

// WithOptions creates ManagedClusterModule with generic mutation options.
func (builder *ManagedClusterModuleBuilder) WithOptions(
	options ...ManagedClusterModuleAdditionalOptions) *ManagedClusterModuleBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting ManagedClusterModule additional options")

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

// PullManagedClusterModule pulls existing module from cluster.
func PullManagedClusterModule(apiClient *clients.Settings, name, nsname string) (*ManagedClusterModuleBuilder, error) {
	glog.V(100).Infof("Pulling existing module name %s under namespace %s from cluster", name, nsname)

	builder := ManagedClusterModuleBuilder{
		apiClient: apiClient,
		Definition: &mcmV1Beta1.ManagedClusterModule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the managedclustermodule is empty")

		builder.errorMsg = "managedclustermodule 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the managedclustermodule is empty")

		builder.errorMsg = "managedclustermodule 'namespace' cannot be empty"
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("managedclustermodule object %s does not exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Create builds managedclustermodule in the cluster and stores object in struct.
func (builder *ManagedClusterModuleBuilder) Create() (*ManagedClusterModuleBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating managedclustermodule %s in namespace %s",
		builder.Definition.Name,
		builder.Definition.Namespace)

	var err error
	if !builder.Exists() {
		err = builder.apiClient.Create(context.TODO(), builder.Definition)
		if err == nil {
			builder.Object = builder.Definition
		}
	}

	return builder, err
}

// Update modifies an existing managedclustermodule on the cluster.
func (builder *ManagedClusterModuleBuilder) Update() (*ManagedClusterModuleBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating managedclustermodule %s in namespace %s",
		builder.Definition.Name,
		builder.Definition.Namespace)

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err == nil {
		builder.Object = builder.Definition
	}

	return builder, err
}

// Exists checks whether the given managedclustermodule exists.
func (builder *ManagedClusterModuleBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if managedclustermodule %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Delete removes the managedclustermodule.
func (builder *ManagedClusterModuleBuilder) Delete() (*ManagedClusterModuleBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting managedclustermodule %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return builder, fmt.Errorf("managedclustermodule cannot be deleted because it does not exist")
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return builder, err
	}

	builder.Object = nil

	return builder, err
}

// Get fetches the defined managedclustermodule from the cluster.
func (builder *ManagedClusterModuleBuilder) Get() (*mcmV1Beta1.ManagedClusterModule, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting managedclustermodule %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	mcm := &mcmV1Beta1.ManagedClusterModule{}

	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, mcm)

	if err != nil {
		return nil, err
	}

	return mcm, err
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *ManagedClusterModuleBuilder) validate() (bool, error) {
	resourceCRD := "ManagedClusterModule"

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
