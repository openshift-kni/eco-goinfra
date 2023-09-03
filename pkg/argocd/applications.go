package argocd

import (
	"context"
	"fmt"

	argocd "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ApplicationBuilder provides a struct for an application object from the cluster and a definition.
type ApplicationBuilder struct {
	// application Definition, used to create the application object.
	Definition *argocd.Application
	// created application object.
	Object *argocd.Application
	// api client to interact with the cluster.
	apiClient *clients.Settings
	// used to store latest error message upon defining or mutating application definition.
	errorMsg string
}

// Pull pulls existing application into ApplicationBuilder struct.
func PullApplication(apiClient *clients.Settings, name, nsname string) (*ApplicationBuilder, error) {
	glog.V(100).Infof("Pulling existing Application name %s under namespace %s from cluster", name, nsname)

	builder := ApplicationBuilder{
		apiClient: apiClient,
		Definition: &argocd.Application{
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
		return nil, fmt.Errorf("application object %s doesn't exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Exists checks whether the given argocd application exists.
func (builder *ApplicationBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if argocd app %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Get returns argocd application object if found.
func (builder *ApplicationBuilder) Get() (*argocd.Application, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting argocd app %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	argocd := &argocd.Application{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, argocd)

	if err != nil {
		return nil, err
	}

	return argocd, err
}

// Update renovates the existing argocd application object with the argocd application definition in builder.
func (builder *ApplicationBuilder) Update(force bool) (*ApplicationBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating the argocd application object", builder.Definition.Name)

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		if force {
			glog.V(100).Infof(
				"Failed to update the argocd application object %s. "+
					"Note: Force flag set, executed delete/create methods instead", builder.Definition.Name)

			builder, err := builder.Delete()

			if err != nil {
				glog.V(100).Infof(
					"Failed to update the argocd application object %s, "+
						"due to error in delete function", builder.Definition.Name)

				return nil, err
			}

			return builder.Create()
		}
	}

	return builder, err
}

// Delete removes the argocd application object from a cluster.
func (builder *ApplicationBuilder) Delete() (*ApplicationBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting the argocd application object %s", builder.Definition.Name)

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return builder, fmt.Errorf("can not delete argocd application: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// Create makes an argocd application in the cluster and stores the created object in a struct.
func (builder *ApplicationBuilder) Create() (*ApplicationBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating argocd application %s", builder.Definition.Name)

	var err error
	if !builder.Exists() {
		err = builder.apiClient.Create(context.TODO(), builder.Definition)
		if err == nil {
			builder.Object = builder.Definition
		}
	}

	return builder, err
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
