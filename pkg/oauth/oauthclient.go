package oauth

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	oauthv1 "github.com/openshift/api/oauth/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type OAuthClientBuilder struct {
	// oauthclient definition, used to create the oauthclient object.
	Definition *oauthv1.OAuthClient
	// Created oauthclient object.
	Object *oauthv1.OAuthClient
	// api clients to interact with the cluster.
	apiClient goclient.Client
}

// Pull loads an existing OAuthClientBuilder into Builder struct.
func Pull(apiClient *clients.Settings, name string) (*OAuthClientBuilder, error) {
	glog.V(100).Infof("Pulling existing OAuthClientBuilder %s", name)

	fmt.Println(" I am here")

	builder := OAuthClientBuilder{
		apiClient: apiClient.Client,
		Definition: &oauthv1.OAuthClient{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	fmt.Println(" I am here2 ")

	if name == "" {
		glog.V(100).Infof("The OAuthClientBuilder name is empty")

		return nil, fmt.Errorf("OAuthClientBuilder name cannot be empty")
	}

	fmt.Println(" I am here3 ")

	if !builder.Exists() {
		return nil, fmt.Errorf("OAuthClientBuilder object %s not found", name)
	}

	fmt.Println(" I am here4 ")

	builder.Definition = builder.Object

	fmt.Println(" I am here5 ")

	return &builder, nil
}

// Get fetches existing OAuthClientBuilder from cluster.
func (builder *OAuthClientBuilder) Get() (*oauthv1.OAuthClient, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Fetching existing OAuthClientBuilder with name %s from cluster", builder.Definition.Name)

	oauthClient := &oauthv1.OAuthClient{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name: builder.Definition.Name,
	}, oauthClient)

	if err != nil {
		return nil, err
	}

	return oauthClient, nil
}

func (builder *OAuthClientBuilder) Create() (*OAuthClientBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the OAuthClient %s",
		builder.Definition.Name)

	var err error
	if !builder.Exists() {
		err = builder.apiClient.Create(context.TODO(), builder.Definition)

		if err == nil {
			builder.Object = builder.Definition
		}
	}

	return builder, err

}

// Exists checks whether the given OAuthClientBuilder exists.
func (builder *OAuthClientBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if OAuthClientBuilder %s exists", builder.Definition.Name)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Update renovates a Builder in the cluster and stores the created object in struct.
func (builder *OAuthClientBuilder) Update() (*OAuthClientBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating the OAuthClientBuilder %s", builder.Definition.Name)

	if !builder.Exists() {
		return nil, fmt.Errorf("OAuthClientBuilder object %s does not exist", builder.Definition.Name)
	}

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		return nil, fmt.Errorf("cannot update OAuthClientBuilder: %w", err)
	}

	return builder, nil
}

// Delete removes a OAuthClientBuilder.
func (builder *OAuthClientBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting the OAuthClientBuilder %s", builder.Definition.Name)

	if !builder.Exists() {
		return nil
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return fmt.Errorf("cannot delete OAuthClientBuilder: %w", err)
	}

	builder.Object = nil

	return nil
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *OAuthClientBuilder) validate() (bool, error) {
	resourceCRD := "OAuthClient"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Definition == nil {
		glog.V(100).Infof("The %s is undefined", resourceCRD)

		return false, fmt.Errorf(msg.UndefinedCrdObjectErrString(resourceCRD))
	}

	if builder.apiClient == nil {
		glog.V(100).Infof("The %s builder apiClient is nil", resourceCRD)

		return false, fmt.Errorf("%s builder cannot have nil apiClient", resourceCRD)
	}

	return true, nil
}
