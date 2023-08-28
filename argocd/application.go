package argocd

import (
	"context"
	"github.com/golang/glog"
	"log"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"

	"fmt"

	arocd "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Builder provides struct for the bmh object containing connection to
// the cluster and the bmh definitions.
type Builder struct {
	Definition *arocd.Application
	Object     *arocd.Application
	apiClient  *clients.Settings
	errorMsg   string
}

// AdditionalOptions additional options for bmh object.
type AdditionalOptions func(builder *Builder) (*Builder, error)

// Pull pulls existing baremetalhost from cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	glog.V(100).Infof("Pulling existing baremetalhost name %s under namespace %s from cluster", name, nsname)

	builder := Builder{
		apiClient: apiClient,
		Definition: &arocd.Application{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the baremetalhost is empty")

		builder.errorMsg = "baremetalhost 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the baremetalhost is empty")

		builder.errorMsg = "baremetalhost 'namespace' cannot be empty"
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("baremetalhost object %s doesn't exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Create makes a bmh in the cluster and stores the created object in struct.
func (builder *Builder) Create() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the baremetalhost %s in namespace %s",
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

// Delete removes bmh from a cluster.
func (builder *Builder) Delete() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting the baremetalhost %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return builder, fmt.Errorf("bmh cannot be deleted because it does not exist")
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return builder, fmt.Errorf("can not delete bmh: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// Exists checks whether the given bmh exists.
func (builder *Builder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if baremetalhost %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Get returns bmh object if found.
func (builder *Builder) Get() (*arocd.Application, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting baremetalhost %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	bmh := &arocd.Application{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, bmh)
	log.Print(err)
	log.Print("DEBUG")
	if err != nil {
		return nil, err
	}

	return bmh, err
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *Builder) validate() (bool, error) {
	resourceCRD := "BareMetalHost"

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
