package velero

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// RestoreBuilder provides a struct for restore object from the cluster and a restore definition.
type RestoreBuilder struct {
	// Restore definition, used to create the restore object.
	Definition *velerov1.Restore
	// Created restore object.
	Object *velerov1.Restore
	// Used to store latest error message upon defining or mutating restore definition.
	errorMsg string
	// api client to interact with the cluster.
	apiClient goclient.Client
}

// NewRestoreBuilder creates a new instance of RestoreBuilder.
func NewRestoreBuilder(apiClient *clients.Settings, name, nsname, backupName string) *RestoreBuilder {
	glog.V(100).Infof(
		"Initializing new restore structure with the following params: "+
			"name: %s, namespace: %s, restoreName: %s", name, nsname, backupName)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient cannot be nil")

		return nil
	}

	err := apiClient.AttachScheme(velerov1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add velero v1 scheme to client schemes")

		return nil
	}

	builder := &RestoreBuilder{
		apiClient: apiClient.Client,
		Definition: &velerov1.Restore{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
			Spec: velerov1.RestoreSpec{
				BackupName: backupName,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the restore is empty")

		builder.errorMsg = "restore name cannot be an empty string"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the restore is empty")

		builder.errorMsg = "restore namespace cannot be an empty string"
	}

	if backupName == "" {
		glog.V(100).Infof("The backupName of the restore is empty")

		builder.errorMsg = "restore backupName cannot be an empty string"
	}

	return builder
}

// PullRestore loads an existing restore into RestoreBuilder struct.
func PullRestore(apiClient *clients.Settings, name, nsname string) (*RestoreBuilder, error) {
	glog.V(100).Infof("Pulling existing restore name: %s under namespace: %s", name, nsname)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient cannot be nil")

		return nil, fmt.Errorf("the apiClient cannot be nil")
	}

	err := apiClient.AttachScheme(velerov1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add velero v1 scheme to client schemes")

		return nil, err
	}

	builder := RestoreBuilder{
		apiClient: apiClient.Client,
		Definition: &velerov1.Restore{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		return nil, fmt.Errorf("restore name cannot be empty")
	}

	if nsname == "" {
		return nil, fmt.Errorf("restore namespace cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("restore object %s does not exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// WithStorageLocation adds a storage location to the restore.
func (builder *RestoreBuilder) WithStorageLocation(location string) *RestoreBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Adding storage location %s to restore %s in namespace %s",
		location, builder.Definition.Name, builder.Definition.Namespace)

	if location == "" {
		glog.V(100).Infof("Backup storage location is empty")

		builder.errorMsg = "restore storage location cannot be an empty string"
	}

	if builder.errorMsg != "" {
		return builder
	}

	if builder.Definition.ObjectMeta.Labels == nil {
		builder.Definition.ObjectMeta.Labels = make(map[string]string)
	}

	builder.Definition.ObjectMeta.Labels["velero.io/storage-location"] = location

	return builder
}

// Get returns Backup object if found.
func (builder *RestoreBuilder) Get() (*velerov1.Restore, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Collecting Restore object %s", builder.Definition.Name)

	restore := &velerov1.Restore{}
	err := builder.apiClient.Get(
		context.TODO(),
		goclient.ObjectKey{Name: builder.Definition.Name, Namespace: builder.Definition.Namespace}, restore)

	if err != nil {
		glog.V(100).Infof("Restore object %s does not exist", builder.Definition.Name)

		return nil, err
	}

	return restore, err
}

// Exists checks whether the given restore object exists.
func (builder *RestoreBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if restore %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Create makes a restore according to the restore definition and stores the created object in the restore builder.
func (builder *RestoreBuilder) Create() (*RestoreBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating restore %s in namespace %s",
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

// Update renovates the existing restore object with the restore definition in builder.
func (builder *RestoreBuilder) Update() (*RestoreBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating restore %s in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err == nil {
		builder.Object = builder.Definition
	}

	return builder, err
}

// Delete removes the restore object and resets the builder object.
func (builder *RestoreBuilder) Delete() (*RestoreBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting restore %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		glog.V(100).Infof("Restore %s in namespace %s cannot be deleted"+
			" because it does not exist",
			builder.Definition.Name, builder.Definition.Namespace)

		builder.Object = nil

		return builder, nil
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return builder, fmt.Errorf("can not delete restore: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *RestoreBuilder) validate() (bool, error) {
	resourceCRD := "Restore"

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
