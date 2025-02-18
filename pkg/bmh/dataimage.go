package bmh

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// DataImageBuilder provides struct for the dataimage object containing connection to
// the cluster and the dataimage definitions.
type DataImageBuilder struct {
	Definition *bmhv1alpha1.DataImage
	Object     *bmhv1alpha1.DataImage
	apiClient  goclient.Client
	errorMsg   string
}

// PullDataImage retrieves an existing DataImage resource from the cluster.
func PullDataImage(apiClient *clients.Settings, name, nsname string) (*DataImageBuilder, error) {
	glog.V(100).Infof("Pulling existing dataimage name %s under namespace %s from cluster", name, nsname)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")

		return nil, fmt.Errorf("dataimage 'apiClient' cannot be empty")
	}

	err := apiClient.AttachScheme(bmhv1alpha1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add bmhv1alpha1 scheme to client schemes")

		return nil, err
	}

	builder := &DataImageBuilder{
		apiClient: apiClient.Client,
		Definition: &bmhv1alpha1.DataImage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the dataimage is empty")

		return nil, fmt.Errorf("dataimage 'name' cannot be empty")
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the dataimage is empty")

		return nil, fmt.Errorf("dataimage 'namespace' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("dataimage object %s does not exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return builder, nil
}

// Delete removes the dataimage from the cluster.
func (builder *DataImageBuilder) Delete() (*DataImageBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting the dataimage %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		glog.V(100).Infof("dataimage %s namespace: %s cannot be deleted because it does not exist",
			builder.Definition.Name, builder.Definition.Namespace)

		builder.Object = nil

		return builder, nil
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return builder, fmt.Errorf("cannot delete dataimage: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// Get returns dataimage object if found.
func (builder *DataImageBuilder) Get() (*bmhv1alpha1.DataImage, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting dataimage %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	dataimage := &bmhv1alpha1.DataImage{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, dataimage)

	if err != nil {
		return nil, err
	}

	return dataimage, nil
}

// Exists checks whether the given dataimage exists.
func (builder *DataImageBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if dataimage %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *DataImageBuilder) validate() (bool, error) {
	resourceCRD := "dataimage"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Definition == nil {
		glog.V(100).Infof("The %s is undefined", resourceCRD)

		return false, fmt.Errorf(msg.UndefinedCrdObjectErrString(resourceCRD))
	}

	if builder.apiClient == nil {
		glog.V(100).Infof("The %s builder apiclient is nil", resourceCRD)

		return false, fmt.Errorf("%s builder cannot have nil apiClient", resourceCRD)
	}

	if builder.errorMsg != "" {
		glog.V(100).Infof("The %s builder has error message: %s", resourceCRD, builder.errorMsg)

		return false, fmt.Errorf(builder.errorMsg)
	}

	return true, nil
}
