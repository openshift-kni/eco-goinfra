package olm

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	pkgManifestV1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PackageManifestBuilder provides a struct for PackageManifest object from the cluster
// and a PackageManifest definition.
type PackageManifestBuilder struct {
	// PackageManifest definition. Used to create
	// PackageManifest object with minimum set of required elements.
	Definition *pkgManifestV1.PackageManifest
	// Created PackageManifest object on the cluster.
	Object *pkgManifestV1.PackageManifest
	// api client to interact with the cluster.
	apiClient *clients.Settings
	// errorMsg is processed before PackageManifest object is created.
	errorMsg string
}

// PullPackageManifest loads an existing PackageManifest into Builder struct.
func PullPackageManifest(apiClient *clients.Settings, name, nsname string) (*PackageManifestBuilder, error) {
	glog.V(100).Infof("Pulling existing PackageManifest name %s in namespace %s", name, nsname)

	builder := &PackageManifestBuilder{
		apiClient: apiClient,
		Definition: &pkgManifestV1.PackageManifest{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The Name of the PackageManifest is empty")

		builder.errorMsg = "PackageManifest 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The Namespace of the PackageManifest is empty")

		builder.errorMsg = "PackageManifest 'nsname' cannot be empty"
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("PackageManifest object %s doesn't exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return builder, nil
}

// PullPackageManifestByCatalog loads an existing PackageManifest from specified catalog into Builder struct.
func PullPackageManifestByCatalog(apiClient *clients.Settings, name, nsname,
	catalog string) (*PackageManifestBuilder, error) {
	glog.V(100).Infof("Pulling existing PackageManifest name %s in namespace %s and from catalog %s",
		name, nsname, catalog)

	packageManifests, err := ListPackageManifest(apiClient, nsname, metaV1.ListOptions{
		LabelSelector: fmt.Sprintf("catalog=%s", catalog),
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	})

	if err != nil {
		glog.V(100).Infof("Failed to list PackageManifests with name %s in namespace %s from catalog"+
			" %s due to %s", name, nsname, catalog, err.Error())

		return nil, err
	}

	if len(packageManifests) == 0 {
		glog.V(100).Infof("The list of matching PackageManifests is empty")

		return nil, fmt.Errorf("no matching PackageManifests were found")
	}

	if len(packageManifests) > 1 {
		glog.V(100).Infof("More than one matching PackageManifests were found")

		return nil, fmt.Errorf("more than one matching PackageManifests were found")
	}

	return packageManifests[0], nil
}

// Exists checks whether the given PackageManifest exists.
func (builder *PackageManifestBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof(
		"Checking if PackageManifest %s exists", builder.Definition.Name)

	var err error
	builder.Object, err = builder.apiClient.PackageManifestInterface.PackageManifests(
		builder.Definition.Namespace).Get(context.Background(), builder.Definition.Name, metaV1.GetOptions{})

	return err == nil || !k8serrors.IsNotFound(err)
}

// Delete removes a PackageManifest.
func (builder *PackageManifestBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting PackageManifest %s in namespace %s", builder.Definition.Name,
		builder.Definition.Namespace)

	if !builder.Exists() {
		return nil
	}

	err := builder.apiClient.PackageManifestInterface.PackageManifests(builder.Definition.Namespace).Delete(
		context.TODO(), builder.Object.Name, metaV1.DeleteOptions{})

	if err != nil {
		return err
	}

	builder.Object = nil

	return err
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *PackageManifestBuilder) validate() (bool, error) {
	resourceCRD := "PackageManifest"

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
