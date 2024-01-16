package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var validPVCModesMap = map[string]string{
	"ReadWriteOnce":    "ReadWriteOnce",
	"ReadOnlyMany":     "ReadOnlyMany",
	"ReadWriteMany":    "ReadWriteMany",
	"ReadWriteOncePod": "ReadWriteOncePod",
}

// PVCBuilder provides struct for persistentvolumeclaim object containing connection
// to the cluster and the persistentvolumeclaim definitions.
type PVCBuilder struct {
	// PersistentVolumeClaim definition. Used to create a persistentvolumeclaim object
	Definition *v1.PersistentVolumeClaim
	// Created persistentvolumeclaim object
	Object *v1.PersistentVolumeClaim

	errorMsg  string
	apiClient *clients.Settings
}

// NewPVCBuilder creates a new structure for persistentvolumeclaim.
func NewPVCBuilder(apiClient *clients.Settings, name, nsname string) *PVCBuilder {
	glog.V(100).Infof("Creating PersistentVolumeClaim %s in namespace %s",
		name, nsname)

	builder := PVCBuilder{
		Definition: &v1.PersistentVolumeClaim{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
			Spec: v1.PersistentVolumeClaimSpec{},
		},
	}

	builder.apiClient = apiClient

	if name == "" {
		glog.V(100).Infof("PVC name is empty")

		builder.errorMsg = "PVC name is empty"
	}

	if nsname == "" {
		glog.V(100).Infof("PVC namespace is empty")

		builder.errorMsg = "PVC namespace is empty"
	}

	return &builder
}

// WithPVCAccessMode configure access mode for the PV.
func (builder *PVCBuilder) WithPVCAccessMode(accessMode string) (*PVCBuilder, error) {
	glog.V(100).Infof("Set PVC accessMode: %s", accessMode)

	if accessMode == "" {
		glog.V(100).Infof("Empty accessMode for PVC %s", builder.Definition.Name)
		builder.errorMsg = "Empty accessMode for PVC requested"

		return builder, fmt.Errorf(builder.errorMsg)
	}

	if !validatePVCAccessMode(accessMode) {
		glog.V(100).Infof("Invalid accessMode for PVC %s", accessMode)
		builder.errorMsg = fmt.Sprintf("Invalid accessMode for PVC %s", accessMode)

		return builder, fmt.Errorf(builder.errorMsg)
	}

	if builder.Definition.Spec.AccessModes != nil {
		builder.Definition.Spec.AccessModes = append(builder.Definition.Spec.AccessModes,
			v1.PersistentVolumeAccessMode(accessMode))
	} else {
		builder.Definition.Spec.AccessModes = []v1.PersistentVolumeAccessMode{v1.PersistentVolumeAccessMode(accessMode)}
	}

	return builder, nil
}

// validatePVCAccessMode validates if requested mode is valid for PVC.
func validatePVCAccessMode(accessMode string) bool {
	glog.V(100).Info("Validating accessMode %s", accessMode)

	_, ok := validPVCModesMap[accessMode]

	return ok
}

// WithPVCCapacity configures the minimum resources the volume should have.
func (builder *PVCBuilder) WithPVCCapacity(capacity string) (*PVCBuilder, error) {
	if capacity == "" {
		glog.V(100).Infof("Capacity of the PersistentVolumeClaim is empty")

		builder.errorMsg = "Capacity of the PersistentVolumeClaim is empty"

		return builder, fmt.Errorf(builder.errorMsg)
	}

	defer func() (*PVCBuilder, error) {
		if r := recover(); r != nil {
			glog.V(100).Infof("Failed to parse %v", capacity)
			builder.errorMsg = fmt.Sprintf("Failed to parse: %v", capacity)

			return builder, fmt.Errorf(fmt.Sprintf("Failed to parse: %v", capacity))
		}

		return builder, nil
	}() //nolint:errcheck

	capMap := make(map[v1.ResourceName]resource.Quantity)
	capMap[v1.ResourceStorage] = resource.MustParse(capacity)

	builder.Definition.Spec.Resources = v1.ResourceRequirements{Requests: capMap}

	return builder, nil
}

// WithStorageClass configures storageClass required by the claim.
func (builder *PVCBuilder) WithStorageClass(storageClass string) (*PVCBuilder, error) {
	glog.V(100).Infof("Set storage class %s for the PersistentVolumeClaim", storageClass)

	if storageClass == "" {
		glog.V(100).Infof("Empty storageClass requested for the PersistentVolumeClaim", storageClass)

		builder.errorMsg = fmt.Sprintf("Empty storageClass requested for the PersistentVolumeClaim %s",
			builder.Definition.Name)

		return builder, fmt.Errorf(builder.errorMsg)
	}

	builder.Definition.Spec.StorageClassName = &storageClass

	return builder, nil
}

// Create generates a PVC in cluster and stores the created object in struct.
func (builder *PVCBuilder) Create() (*PVCBuilder, error) {
	if valid, _ := builder.validate(); !valid {
		return builder, fmt.Errorf("invalid builder")
	}

	glog.V(100).Infof("Creating persistentVolumeClaim %s", builder.Definition.Name)

	var err error
	if !builder.Exists() {
		builder.Object, err = builder.apiClient.PersistentVolumeClaims(builder.Definition.Namespace).Create(
			context.TODO(), builder.Definition, metaV1.CreateOptions{})
	}

	if err != nil {
		glog.V(100).Infof("Error creating persistentVolumeClaim %s - %v", builder.Definition.Name, err)

		builder.errorMsg = fmt.Sprintf("failed to create PVC: %v", err)

		return builder, err
	}

	return builder, nil
}

// Delete removes PVC from cluster.
func (builder *PVCBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		glog.V(100).Infof("PersistentVolumeClaim %s in %s namespace is invalid: %v",
			builder.Definition.Name, builder.Definition.Namespace, err)

		return err
	}

	glog.V(100).Infof("Delete PersistentVolumeClaim %s from %s namespace",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		glog.V(100).Infof("PersistentVolumeClaim %s not found in %s namespace",
			builder.Definition.Name, builder.Definition.Namespace)

		return nil
	}

	err := builder.apiClient.PersistentVolumeClaims(builder.Definition.Namespace).Delete(
		context.TODO(), builder.Definition.Name, metaV1.DeleteOptions{})

	if err != nil {
		glog.V(100).Infof("Failed to delete PersistentVolumeClaim %s from %s namespace",
			builder.Definition.Name, builder.Definition.Namespace)
		glog.V(100).Infof("PersistenteVolumeClaim deletion error: %v", err)

		return err
	}

	glog.V(100).Infof("Deleted PersistentVolumeClaim %s from %s namespace",
		builder.Definition.Name, builder.Definition.Namespace)

	builder.Object = nil

	return err
}

// DeleteAndWait deletes PersistentVolumeClaim and waits until it's removed from the cluster.
func (builder *PVCBuilder) DeleteAndWait(timeout time.Duration) error {
	if valid, err := builder.validate(); !valid {
		glog.V(100).Infof("PersistentVolumeClaim %s in %s namespace is invalid: %v",
			builder.Definition.Name, builder.Definition.Namespace, err)

		return err
	}

	glog.V(100).Infof("Deleting PersistenVolumeClaim %s from %s namespace and waiting for the removal to complete",
		builder.Definition.Name, builder.Definition.Namespace)

	if err := builder.Delete(); err != nil {
		glog.V(100).Infof("Failed to delete PersistentVolumeClaim %s from %s namespace",
			builder.Definition.Name, builder.Definition.Namespace)
		glog.V(100).Infof("PersistenteVolumeClaim deletion error: %v", err)

		return err
	}

	return wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			_, err := builder.apiClient.PersistentVolumeClaims(builder.Definition.Namespace).Get(
				context.Background(), builder.Definition.Name, metaV1.GetOptions{})
			if k8serrors.IsNotFound(err) {

				return true, nil
			}

			return false, nil
		})
}

// PullPersistentVolumeClaim gets an existing PersistentVolumeClaim
// from the cluster.
func PullPersistentVolumeClaim(
	apiClient *clients.Settings, persistentVolumeClaim string, nsname string) (
	*PVCBuilder, error) {
	glog.V(100).Infof("Pulling existing PersistentVolumeClaim object: %s from namespace %s",
		persistentVolumeClaim, nsname)

	builder := PVCBuilder{
		apiClient: apiClient,
		Definition: &v1.PersistentVolumeClaim{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      persistentVolumeClaim,
				Namespace: nsname,
			},
		},
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("PersistentVolumeClaim object %s doesn't exist in namespace %s",
			persistentVolumeClaim, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Exists checks whether the given PersistentVolumeClaim exists.
func (builder *PVCBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if PersistentVolumeClaim %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.apiClient.PersistentVolumeClaims(builder.Definition.Namespace).Get(
		context.Background(), builder.Definition.Name, metaV1.GetOptions{})

	return err == nil || !k8serrors.IsNotFound(err)
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *PVCBuilder) validate() (bool, error) {
	resourceCRD := "PersistentVolumeClaim"

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
