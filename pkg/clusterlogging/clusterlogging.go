package clusterlogging

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	clov1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Builder provides struct for clusterLogging object.
type Builder struct {
	// ClusterLogging definition. Used to create clusterLogging object with minimum set of required elements.
	Definition *clov1.ClusterLogging
	// Created clusterLogging object on the cluster.
	Object *clov1.ClusterLogging
	// api client to interact with the cluster.
	apiClient *clients.Settings
	// errorMsg is processed before clusterLogging object is created.
	errorMsg string
}

// NewBuilder method creates new instance of builder.
func NewBuilder(
	apiClient *clients.Settings, name, nsname string) *Builder {
	glog.V(100).Infof("Initializing new clusterLogging structure with the following params: name: %s, namespace: %s",
		name, nsname)

	builder := &Builder{
		apiClient: apiClient,
		Definition: &clov1.ClusterLogging{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the clusterLogging is empty")

		builder.errorMsg = "The clusterLogging 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the clusterLogging is empty")

		builder.errorMsg = "The clusterLogging 'namespace' cannot be empty"
	}

	return builder
}

// Pull retrieves an existing clusterLogging object from the cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	glog.V(100).Infof(
		"Pulling clusterLogging object name:%s in namespace: %s", name, nsname)

	builder := Builder{
		apiClient: apiClient,
		Definition: &clov1.ClusterLogging{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the clusterLogging is empty")

		builder.errorMsg = "clusterLogging 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the clusterLogging is empty")

		builder.errorMsg = "clusterLogging 'nsname' cannot be empty"
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("clusterLogging object %s does not exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Get returns clusterLogging object if found.
func (builder *Builder) Get() (*clov1.ClusterLogging, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting clusterLogging %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	clusterLogging := &clov1.ClusterLogging{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, clusterLogging)

	if err != nil {
		return nil, err
	}

	return clusterLogging, err
}

// Create makes a clusterLogging in the cluster and stores the created object in struct.
func (builder *Builder) Create() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the clusterLogging %s in namespace %s",
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

// Delete removes clusterLogging from a cluster.
func (builder *Builder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting the clusterLogging %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return fmt.Errorf("clusterLogging cannot be deleted because it does not exist")
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return fmt.Errorf("can not delete clusterLogging: %w", err)
	}

	builder.Object = nil

	return nil
}

// Exists checks whether the given clusterLogging exists.
func (builder *Builder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if clusterLogging %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Update renovates the existing clusterLogging object with clusterLogging definition in builder.
func (builder *Builder) Update(force bool) (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Info("Updating clusterLogging %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		if force {
			glog.V(100).Infof(
				msg.FailToUpdateNotification("clusterLogging", builder.Definition.Name, builder.Definition.Namespace))

			err := builder.Delete()

			if err != nil {
				glog.V(100).Infof(
					msg.FailToUpdateError("clusterLogging", builder.Definition.Name, builder.Definition.Namespace))

				return nil, err
			}

			return builder.Create()
		}
	}

	if err == nil {
		builder.Object = builder.Definition
	}

	return builder, err
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *Builder) validate() (bool, error) {
	resourceCRD := "ClusterLogging"

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
