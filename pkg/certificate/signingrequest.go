package certificate

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	certificatesv1 "k8s.io/api/certificates/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// SigningRequestBuilder provides a struct for CertificateSigningRequest resource containing a connection to the cluster
// and the CertificateSigningRequest definition.
type SigningRequestBuilder struct {
	// SigningRequest definition, used to create the signing request object.
	Definition *certificatesv1.CertificateSigningRequest
	// Created signing request object on cluster.
	Object *certificatesv1.CertificateSigningRequest
	// apiClient to interact with the cluster.
	apiClient runtimeclient.Client
}

// PullSigningRequest loads an existing signing request into SigningRequestBuilder struct.
func PullSigningRequest(apiClient *clients.Settings, name string) (*SigningRequestBuilder, error) {
	glog.V(100).Infof("Pulling existing CertificateSigningRequest with name %s", name)

	if apiClient == nil {
		glog.V(100).Infof("CertificateSigningRequest apiClient cannot be nil")

		return nil, fmt.Errorf("certificateSigniingRequest apiClient cannot be nil")
	}

	err := apiClient.AttachScheme(certificatesv1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add certificates v1 scheme to client schemes")

		return nil, err
	}

	builder := &SigningRequestBuilder{
		apiClient: apiClient.Client,
		Definition: &certificatesv1.CertificateSigningRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the CertificateSigningRequest is empty")

		return nil, fmt.Errorf("certificateSigningRequest 'name' cannot be empty")
	}

	if !builder.Exists() {
		glog.V(100).Infof("CertificateSigningRequest %s does not exist", name)

		return nil, fmt.Errorf("certificateSigningRequest %s does not exist", name)
	}

	builder.Definition = builder.Object

	return builder, nil
}

// Get returns the CertificateSigningRequest object if found.
func (builder *SigningRequestBuilder) Get() (*certificatesv1.CertificateSigningRequest, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Collecting CertificateSigningRequest object %s", builder.Definition.Name)

	signingRequest := &certificatesv1.CertificateSigningRequest{}
	err := builder.apiClient.Get(context.TODO(), runtimeclient.ObjectKey{
		Name: builder.Definition.Name,
	}, signingRequest)

	if err != nil {
		glog.V(100).Infof("Failed to get CertificateSigningRequest object %s: %v", builder.Definition.Name, err)

		return nil, err
	}

	return signingRequest, nil
}

// Exists checks whether the given CertificateSigningRequest object exists.
func (builder *SigningRequestBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if CertificateSigningRequest %s exists", builder.Definition.Name)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Create creates a new CertificateSigningRequest object if it does not exist.
func (builder *SigningRequestBuilder) Create() (*SigningRequestBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating CertificateSigningRequest %s", builder.Definition.Name)

	if builder.Exists() {
		return builder, nil
	}

	err := builder.apiClient.Create(context.TODO(), builder.Definition)
	if err != nil {
		return builder, err
	}

	builder.Object = builder.Definition

	return builder, nil
}

// Delete removes a CertificateSigningRequest object from the cluster if it exists.
func (builder *SigningRequestBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting CertificateSigningRequest %s", builder.Definition.Name)

	if !builder.Exists() {
		glog.V(100).Infof("CertificateSigningRequest %s does not exist", builder.Definition.Name)

		builder.Object = nil

		return nil
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)
	if err != nil {
		return err
	}

	builder.Object = nil

	return nil
}

func (builder *SigningRequestBuilder) validate() (bool, error) {
	resourceCRD := "certificateSigningRequest"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Definition == nil {
		glog.V(100).Infof("The %s is undefined", resourceCRD)

		return false, fmt.Errorf("%s", msg.UndefinedCrdObjectErrString(resourceCRD))
	}

	if builder.apiClient == nil {
		glog.V(100).Infof("The %s builder apiclient is nil", resourceCRD)

		return false, fmt.Errorf("%s builder cannot have nil apiClient", resourceCRD)
	}

	return true, nil
}
