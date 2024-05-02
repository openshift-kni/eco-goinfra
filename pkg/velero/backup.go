package velero

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	veleroClient "github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackupBuilder provides a struct for backup object from the cluster and a backup definition.
type BackupBuilder struct {
	// Backup definition, used to create the backup object.
	Definition *velerov1.Backup
	// Created backup object.
	Object *velerov1.Backup
	// Used to store latest error message upon defining or mutating backup definition.
	errorMsg string
	// api client to interact with the cluster.
	apiClient veleroClient.Interface
}

// NewBackupBuilder creates a new instance of BackupBuilder.
func NewBackupBuilder(apiClient *clients.Settings, name, nsname string) *BackupBuilder {
	glog.V(100).Infof(
		"Initializing new backup structure with the following params: "+
			"name: %s, namespace: %s", name, nsname)

	builder := &BackupBuilder{
		apiClient: apiClient.VeleroClient,
		Definition: &velerov1.Backup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the backup is empty")

		builder.errorMsg = "backup name cannot be an empty string"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the backup is empty")

		builder.errorMsg = "backup namespace cannot be an empty string"
	}

	return builder
}

// PullBackup loads an existing backup into BackupBuilder struct.
func PullBackup(apiClient *clients.Settings, name, nsname string) (*BackupBuilder, error) {
	glog.V(100).Infof("Pulling existing backup name: %s under namespace: %s", name, nsname)

	builder := BackupBuilder{
		apiClient: apiClient.VeleroClient,
		Definition: &velerov1.Backup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		return nil, fmt.Errorf("backup name cannot be empty")
	}

	if nsname == "" {
		return nil, fmt.Errorf("backup namespace cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("backup object %s does not exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// WithStorageLocation adds a storage location to the backup.
func (builder *BackupBuilder) WithStorageLocation(location string) *BackupBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Adding storage location %s to backup %s in namespace %s",
		location, builder.Definition.Name, builder.Definition.Namespace)

	if location == "" {
		glog.V(100).Infof("Backup storage location is empty")

		builder.errorMsg = "backup storage location cannot be an empty string"
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

// WithIncludedNamespace adds the specified namespace for inclusion when performing a backup.
func (builder *BackupBuilder) WithIncludedNamespace(namespace string) *BackupBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Adding namespace %s to backup %s in namespace %s includedNamespaces field",
		namespace, builder.Definition.Name, builder.Definition.Namespace)

	if namespace == "" {
		glog.V(100).Infof("Backup includedNamespace is empty")

		builder.errorMsg = "backup includedNamespace cannot be an empty string"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.IncludedNamespaces = append(builder.Definition.Spec.IncludedNamespaces, namespace)

	return builder
}

// WithIncludedClusterScopedResource adds the specified cluster-scoped crd for inclusion when performing a backup.
func (builder *BackupBuilder) WithIncludedClusterScopedResource(crd string) *BackupBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Adding custom resource %s to backup %s in namespace %s includedClusterScopedResources field",
		crd, builder.Definition.Name, builder.Definition.Namespace)

	if crd == "" {
		glog.V(100).Infof("Backup includedClusterScopedResource is empty")

		builder.errorMsg = "backup includedClusterScopedResource cannot be an empty string"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.IncludedClusterScopedResources =
		append(builder.Definition.Spec.IncludedClusterScopedResources, crd)

	return builder
}

// WithIncludedNamespaceScopedResource adds the specified namespace-scoped crd for inclusion when performing a backup.
func (builder *BackupBuilder) WithIncludedNamespaceScopedResource(crd string) *BackupBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Adding custom resource %s to backup %s in namespace %s includedNamespaceScopedResources field",
		crd, builder.Definition.Name, builder.Definition.Namespace)

	if crd == "" {
		glog.V(100).Infof("Backup includedNamespaceScopedResource is empty")

		builder.errorMsg = "backup includedNamespaceScopedResource cannot be an empty string"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.IncludedNamespaceScopedResources =
		append(builder.Definition.Spec.IncludedNamespaceScopedResources, crd)

	return builder
}

// WithExcludedClusterScopedResource adds the specified cluster-scoped crd for exclusion when performing a backup.
func (builder *BackupBuilder) WithExcludedClusterScopedResource(crd string) *BackupBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Adding custom resource %s to backup %s in namespace %s excludedClusterScopedResources field",
		crd, builder.Definition.Name, builder.Definition.Namespace)

	if crd == "" {
		glog.V(100).Infof("Backup excludedClusterScopedResource is empty")

		builder.errorMsg = "backup excludedClusterScopedResource cannot be an empty string"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.ExcludedClusterScopedResources =
		append(builder.Definition.Spec.ExcludedClusterScopedResources, crd)

	return builder
}

// WithExcludedNamespaceScopedResources adds the specified namespace-scoped crd for exclusion when performing a backup.
func (builder *BackupBuilder) WithExcludedNamespaceScopedResources(crd string) *BackupBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Adding custom resource %s to backup %s in namespace %s excludedNamespaceScopedResources field",
		crd, builder.Definition.Name, builder.Definition.Namespace)

	if crd == "" {
		glog.V(100).Infof("Backup excludedNamespaceScopedResource is empty")

		builder.errorMsg = "backup excludedNamespaceScopedResource cannot be an empty string"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.ExcludedNamespaceScopedResources =
		append(builder.Definition.Spec.ExcludedNamespaceScopedResources, crd)

	return builder
}

// Exists checks whether the given backup exists.
func (builder *BackupBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if backup %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.apiClient.VeleroV1().Backups(builder.Definition.Namespace).Get(
		context.TODO(), builder.Definition.Name, metav1.GetOptions{})

	return err == nil || !k8serrors.IsNotFound(err)
}

// Create makes a backup according to the backup definition and stores the created object in the backup builder.
func (builder *BackupBuilder) Create() (*BackupBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating backup %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	if !builder.Exists() {
		builder.Object, err = builder.apiClient.VeleroV1().Backups(builder.Definition.Namespace).Create(
			context.TODO(), builder.Definition, metav1.CreateOptions{})
	}

	return builder, err
}

// Update renovates the existing backup object with the backup definition in builder.
func (builder *BackupBuilder) Update() (*BackupBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating backup %s in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.apiClient.VeleroV1().Backups(builder.Definition.Namespace).Update(
		context.TODO(), builder.Definition, metav1.UpdateOptions{})

	return builder, err
}

// Delete removes the backup object and resets the builder object.
func (builder *BackupBuilder) Delete() (*BackupBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting backup %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return builder, fmt.Errorf("backup cannot be deleted because it does not exist")
	}

	err := builder.apiClient.VeleroV1().Backups(builder.Definition.Namespace).Delete(
		context.TODO(), builder.Object.Name, metav1.DeleteOptions{})

	if err != nil {
		return builder, fmt.Errorf("can not delete backup: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *BackupBuilder) validate() (bool, error) {
	resourceCRD := "Backup"

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
