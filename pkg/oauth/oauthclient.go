package oauth

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	oauthv1 "github.com/openshift/api/oauth/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/msg"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// OAuthClientBuilder provides struct for the OAuthClient object containing connection to
// the cluster and the OAuthClient definitions.
type OAuthClientBuilder struct {
	// OAuthClient definition, used to create the OAuthClient object.
	Definition *oauthv1.OAuthClient
	// Created OAuthClient object.
	Object *oauthv1.OAuthClient
	// api clients to interact with the cluster.
	apiClient goclient.Client
}

// PullOAuthClient loads an existing OAuthClient into Builder struct.
func PullOAuthClient(apiClient *clients.Settings, name string) (*OAuthClientBuilder, error) {
	glog.V(100).Infof("Pulling existing OAuthClient %s", name)

	if apiClient == nil {
		return nil, fmt.Errorf("apiClient cannot be nil")
	}

	err := apiClient.AttachScheme(oauthv1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add oauth v1 scheme to client schemes")

		return nil, err
	}

	builder := OAuthClientBuilder{
		apiClient: apiClient.Client,
		Definition: &oauthv1.OAuthClient{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The OAuthClient 'name' is empty")

		return nil, fmt.Errorf("error: OAuthClient 'name' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("error: OAuthClient object %s not found", name)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Get returns OAuthClient if found.
func (builder *OAuthClientBuilder) Get() (*oauthv1.OAuthClient, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Fetching existing OAuthClient with name %s from cluster", builder.Definition.Name)

	oauthClient := &oauthv1.OAuthClient{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name: builder.Definition.Name,
	}, oauthClient)

	if err != nil {
		return nil, err
	}

	return oauthClient, nil
}

// Create constructs an OAuthClient object on the cluster from a builder.
func (builder *OAuthClientBuilder) Create() (*OAuthClientBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the OAuthClient %s", builder.Definition.Name)

	var err error
	if !builder.Exists() {
		err = builder.apiClient.Create(context.TODO(), builder.Definition)

		if err == nil {
			builder.Object = builder.Definition
		}
	}

	return builder, err
}

// Exists checks whether the given OAuthClient exists.
func (builder *OAuthClientBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if OAuthClient %s exists", builder.Definition.Name)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Update renovates a builder in the cluster and stores the created object in struct.
func (builder *OAuthClientBuilder) Update() (*OAuthClientBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating the OAuthClient %s", builder.Definition.Name)

	if !builder.Exists() {
		return nil, fmt.Errorf("error: OAuthClient object %s does not exist", builder.Definition.Name)
	}

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err == nil {
		builder.Object = builder.Definition
	}

	return builder, nil
}

// Delete removes a OAuthClient from the cluster.
func (builder *OAuthClientBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting the OAuthClient %s", builder.Definition.Name)

	if !builder.Exists() {
		glog.V(100).Infof("OAuthClient %s cannot be deleted"+
			" because it does not exist", builder.Definition.Name)

		builder.Object = nil

		return nil
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return fmt.Errorf("error: cannot delete OAuthClient: %w", err)
	}

	builder.Object = nil
	builder.Definition.ResourceVersion = ""
	builder.Definition.CreationTimestamp = metav1.Time{}

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

		return false, fmt.Errorf("%s", msg.UndefinedCrdObjectErrString(resourceCRD))
	}

	if builder.apiClient == nil {
		glog.V(100).Infof("The %s builder apiClient is nil", resourceCRD)

		return false, fmt.Errorf("error: %s builder cannot have nil apiClient", resourceCRD)
	}

	return true, nil
}
