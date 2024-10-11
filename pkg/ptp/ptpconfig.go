package ptp

import (
	"context"
	"fmt"

	goclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	ptpv1 "github.com/openshift/ptp-operator/api/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PtpConfigBuilder provides a struct for the PtpConfig resource containing a connection to the cluster and the
// PtpConfig definition.
type PtpConfigBuilder struct {
	// Definition of the PtpConfig used to create the object.
	Definition *ptpv1.PtpConfig
	// Object of the PtpConfig as it is on the cluster.
	Object    *ptpv1.PtpConfig
	apiClient goclient.Client
	errorMsg  string
}

// NewPtpConfigBuilder creates a new instance of a PtpConfig builder.
func NewPtpConfigBuilder(apiClient *clients.Settings, name, nsname string) *PtpConfigBuilder {
	glog.V(100).Infof("Initializing new PtpConfig structure with the following params: name: %s, nsname: %s", name, nsname)

	if apiClient == nil {
		glog.V(100).Info("The apiClient of the PtpConfig is nil")

		return nil
	}

	err := apiClient.AttachScheme(ptpv1.AddToScheme)
	if err != nil {
		glog.V(100).Info("Failed to add ptp v1 scheme to client schemes")

		return nil
	}

	builder := &PtpConfigBuilder{
		apiClient: apiClient.Client,
		Definition: &ptpv1.PtpConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Info("The name of the PtpConfig is empty")

		builder.errorMsg = "ptpConfig 'name' cannot be empty"

		return builder
	}

	if nsname == "" {
		glog.V(100).Info("The namespace of the PtpConfig is empty")

		builder.errorMsg = "ptpConfig 'nsname' cannot be empty"

		return builder
	}

	return builder
}

// PullPtpConfig pulls an existing PtpConfig into a Builder struct.
func PullPtpConfig(apiClient *clients.Settings, name, nsname string) (*PtpConfigBuilder, error) {
	glog.V(100).Infof("Pulling existing PtpConfig %s under namespace %s from cluster", name, nsname)

	if apiClient == nil {
		glog.V(100).Info("The apiClient is empty")

		return nil, fmt.Errorf("ptpConfig 'apiClient' cannot be nil")
	}

	err := apiClient.AttachScheme(ptpv1.AddToScheme)
	if err != nil {
		glog.V(100).Info("Failed to add PtpConfig scheme to client schemes")

		return nil, err
	}

	builder := &PtpConfigBuilder{
		apiClient: apiClient.Client,
		Definition: &ptpv1.PtpConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Info("The name of the PtpConfig is empty")

		return nil, fmt.Errorf("ptpConfig 'name' cannot be empty")
	}

	if nsname == "" {
		glog.V(100).Info("The namespace of the PtpConfig is empty")

		return nil, fmt.Errorf("ptpConfig 'nsname' cannot be empty")
	}

	if !builder.Exists() {
		glog.V(100).Info("The PtpConfig %s does not exist in namespace %s", name, nsname)

		return nil, fmt.Errorf("ptpConfig object %s does not exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return builder, nil
}

// Get returns the PtpConfig object if found.
func (builder *PtpConfigBuilder) Get() (*ptpv1.PtpConfig, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof(
		"Getting PtpConfig object %s in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

	ptpConfig := &ptpv1.PtpConfig{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, ptpConfig)

	if err != nil {
		glog.V(100).Infof(
			"PtpConfig object %s does not exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)

		return nil, err
	}

	return ptpConfig, nil
}

// Exists checks whether the given PtpConfig exists on the cluster.
func (builder *PtpConfigBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof(
		"Checking if PtpConfig %s exists in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Create makes a PtpConfig on the cluster if it does not already exist.
func (builder *PtpConfigBuilder) Create() (*PtpConfigBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof(
		"Creating PtpConfig %s in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

	if builder.Exists() {
		return builder, nil
	}

	err := builder.apiClient.Create(context.TODO(), builder.Definition)
	if err != nil {
		return nil, err
	}

	builder.Object = builder.Definition

	return builder, nil
}

// Update changes the existing PtpConfig resource on the cluster, falling back to deleting and recreating if the update
// fails when force is set.
func (builder *PtpConfigBuilder) Update(force bool) (*PtpConfigBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof(
		"Updating PtpConfig %s in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		glog.V(100).Infof(
			"PtpConfig %s does not exist in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

		return nil, fmt.Errorf("cannot update non-existent ptpConfig")
	}

	builder.Definition.ResourceVersion = builder.Object.ResourceVersion

	err := builder.apiClient.Update(context.TODO(), builder.Definition)
	if err != nil {
		if force {
			glog.V(100).Infof(msg.FailToUpdateNotification("ptpConfig", builder.Definition.Name))

			err := builder.Delete()
			builder.Definition.ResourceVersion = ""

			if err != nil {
				glog.V(100).Infof(msg.FailToUpdateError("ptpConfig", builder.Definition.Name))

				return nil, err
			}

			return builder.Create()
		}

		return nil, err
	}

	builder.Object = builder.Definition

	return builder, nil
}

// Delete removes a PtpConfig from the cluster if it exists.
func (builder *PtpConfigBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof(
		"Deleting PtpConfig %s in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		glog.V(100).Infof(
			"PtpConfig %s in namespace %s does not exist",
			builder.Definition.Name, builder.Definition.Namespace)

		builder.Object = nil

		return nil
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Object)
	if err != nil {
		return err
	}

	builder.Object = nil

	return nil
}

// validate checks that the builder, definition, and apiClient are properly initialized and there is no errorMsg.
func (builder *PtpConfigBuilder) validate() (bool, error) {
	resourceCRD := "ptpConfig"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Definition == nil {
		glog.V(100).Infof("The %s is uninitialized", resourceCRD)

		return false, fmt.Errorf(msg.UndefinedCrdObjectErrString(resourceCRD))
	}

	if builder.apiClient == nil {
		glog.V(100).Infof("The %s builder apiClient is nil", resourceCRD)

		return false, fmt.Errorf("%s builder cannot have nil apiClient", resourceCRD)
	}

	if builder.errorMsg != "" {
		glog.V(100).Infof("The %s builder has error message %s", resourceCRD, builder.errorMsg)

		return false, fmt.Errorf(builder.errorMsg)
	}

	return true, nil
}
