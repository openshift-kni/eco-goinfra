package storage

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PVBuilder provides struct for persistentvolume object containing connection
// to the cluster and the persistentvolume definitions.
type PVBuilder struct {
	// PersistentVolume definition. Used to create a persistentvolume object
	Definition *corev1.PersistentVolume
	// Created persistentvolume object
	Object *corev1.PersistentVolume

	apiClient *clients.Settings
}

// PullPersistentVolume gets an existing PersistentVolume from the cluster.
func PullPersistentVolume(apiClient *clients.Settings, persistentVolume string) (*PVBuilder, error) {
	glog.V(100).Infof("Pulling existing PersistentVolume object: %s", persistentVolume)

	builder := PVBuilder{
		apiClient: apiClient,
		Definition: &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name: persistentVolume,
			},
		},
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("PersistentVolume object %s does not exist", persistentVolume)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Exists checks whether the given PersistentVolume exists.
func (builder *PVBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if PersistentVolume %s exists", builder.Definition.Name)

	var err error
	builder.Object, err = builder.apiClient.PersistentVolumes().Get(
		context.TODO(), builder.Definition.Name, metav1.GetOptions{})

	return err == nil || !k8serrors.IsNotFound(err)
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *PVBuilder) validate() (bool, error) {
	resourceCRD := "PersistentVolume"

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

	return true, nil
}
