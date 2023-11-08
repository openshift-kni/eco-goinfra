package cgu

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/cluster-group-upgrades-operator/api/v1alpha1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// CguBuilder provides struct for the cgu object containing connection to
// the cluster and the cgu definitions.
type CguBuilder struct {
	// cgu Definition, used to create the cgu object.
	Definition *v1alpha1.ClusterGroupUpgrade
	// created cgu object.
	Object *v1alpha1.ClusterGroupUpgrade
	// api client to interact with the cluster.
	apiClient *clients.Settings
	// used to store latest error message upon defining or mutating application definition.
	errorMsg string
}

// Pull pulls existing cgu into CguBuilder struct.
func Pull(apiClient *clients.Settings, name, nsname string) (*CguBuilder, error) {
	glog.V(100).Infof("Pulling existing cgu name %s under namespace %s from cluster", name, nsname)

	builder := CguBuilder{
		apiClient: apiClient,
		Definition: &v1alpha1.ClusterGroupUpgrade{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the cgu is empty")

		builder.errorMsg = "cgu's 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the cgu is empty")

		builder.errorMsg = "cgu's 'namespace' cannot be empty"
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("cgu object %s doesn't exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Exists checks whether the given cgu exists.
func (builder *CguBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if cgu %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Get returns a cgu object if found.
func (builder *CguBuilder) Get() (*v1alpha1.ClusterGroupUpgrade, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting cgu %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	cgu := &v1alpha1.ClusterGroupUpgrade{}

	err := builder.apiClient.Get(context.TODO(), runtimeclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, cgu)

	if err != nil {
		return nil, err
	}

	return cgu, err
}

// Create makes a cgu in the cluster and stores the created object in struct.
func (builder *CguBuilder) Create() (*CguBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the cgu %s in namespace %s",
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

// Delete removes a cgu from a cluster.
func (builder *CguBuilder) Delete() (*CguBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting the cgu %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return builder, fmt.Errorf("cgu cannot be deleted because it does not exist")
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return builder, fmt.Errorf("can not delete cgu: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// Update renovates the existing cgu object with the cgu definition in builder.
func (builder *CguBuilder) Update(force bool) (*CguBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating the cgu object", builder.Definition.Name)

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		if force {
			glog.V(100).Infof(
				"Failed to update the cgu object %s. "+
					"Note: Force flag set, executed delete/create methods instead", builder.Definition.Name)

			builder, err := builder.Delete()

			if err != nil {
				glog.V(100).Infof(
					"Failed to update the cgu object %s, "+
						"due to error in delete function", builder.Definition.Name)

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
func (builder *CguBuilder) validate() (bool, error) {
	resourceCRD := "cgu"

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
