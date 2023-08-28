package argocd

import (
	"context"
	"fmt"
	arocd "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ApplicationBuilder provides struct ...
type ApplicationBuilder struct {
	Definition *arocd.Application
	Object     *arocd.Application
	apiClient  *clients.Settings
	errorMsg   string
}

// Pull pulls existing ...
func Pull(apiClient *clients.Settings, name, nsname string) (*ApplicationBuilder, error) {
	glog.V(100).Infof("Pulling existing Application name %s under namespace %s from cluster", name, nsname)

	builder := ApplicationBuilder{
		apiClient: apiClient,
		Definition: &arocd.Application{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the Application is empty")

		builder.errorMsg = "Application 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the Application is empty")

		builder.errorMsg = "Application 'namespace' cannot be empty"
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("Application object %s doesn't exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Create makes a bmh in the cluster and stores the created object in struct.
func (builder *ApplicationBuilder) Create() (*ApplicationBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the Application %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	if !builder.Exists() {
		builder.Object, err = builder.apiClient.Applications(builder.Definition.Namespace).
			Create(context.TODO(), builder.Definition, metaV1.CreateOptions{})
	}

	return builder, err
}

// Delete removes bmh from a cluster.
func (builder *ApplicationBuilder) Delete() (*ApplicationBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting the Application %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return builder, fmt.Errorf("bmh cannot be deleted because it does not exist")
	}

	err := builder.apiClient.Applications(builder.Definition.Namespace).
		Delete(context.TODO(), builder.Definition.Name, metaV1.DeleteOptions{})

	if err != nil {
		return builder, fmt.Errorf("can not delete bmh: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// Exists checks whether the given deployment exists.
func (builder *ApplicationBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if Application %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.apiClient.Applications(builder.Definition.Namespace).Get(
		context.Background(), builder.Definition.Name, metaV1.GetOptions{})

	return err == nil || !k8serrors.IsNotFound(err)
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *ApplicationBuilder) validate() (bool, error) {
	resourceCRD := "Application"

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
