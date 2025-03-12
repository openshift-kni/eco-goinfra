package imagestream

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	imagev1 "github.com/openshift/api/image/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Builder provides a struct for imageStream object from the cluster and a imageStream definition.
type Builder struct {
	// imageStream definition, used to create the imageStream object.
	Definition *imagev1.ImageStream
	// Created imageStream object.
	Object *imagev1.ImageStream
	// api client to interact with the cluster.
	apiClient goclient.Client
	// Used in functions that define or mutate clusterOperator definition. errorMsg is processed before the
	// ClusterOperator object is created.
	errorMsg string
}

// Pull retrieves an existing imageStream object from the cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	glog.V(100).Infof(
		"Pulling imageStream object name %s from namespace %s", name, nsname)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")

		return nil, fmt.Errorf("imageStream 'apiClient' cannot be empty")
	}

	err := apiClient.AttachScheme(imagev1.AddToScheme)
	if err != nil {
		glog.V(100).Info("Failed to add imageStream v1 scheme to client schemes")

		return nil, err
	}

	builder := Builder{
		apiClient: apiClient.Client,
		Definition: &imagev1.ImageStream{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
			Spec: imagev1.ImageStreamSpec{},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the imageStream is empty")

		return nil, fmt.Errorf("imageStream 'name' cannot be empty")
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the imageStream is empty")

		return nil, fmt.Errorf("imageStream 'nsname' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("imageStream object %s does not exist in namespace %s",
			name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Get fetches existing imageStream from cluster.
func (builder *Builder) Get() (*imagev1.ImageStream, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting existing imageStream with name %s in namespace %s from cluster",
		builder.Definition.Name, builder.Definition.Namespace)

	imageStreamObj := &imagev1.ImageStream{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, imageStreamObj)

	if err != nil {
		glog.V(100).Infof("imageStream object %s does not exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)

		return nil, err
	}

	return imageStreamObj, nil
}

// Exists checks whether the given imageStream exists.
func (builder *Builder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if imageStream %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// GetDockerImage fetches imageStream DockerImage value.
func (builder *Builder) GetDockerImage(imageTag string) (string, error) {
	if valid, err := builder.validate(); !valid {
		return "", err
	}

	glog.V(100).Infof("Getting imageStream DockerImage value")

	if imageTag == "" {
		glog.V(100).Infof("The imageTag of the imageStream is empty")

		return "", fmt.Errorf("imageStream 'imageTag' cannot be empty")
	}

	if !builder.Exists() {
		return "", fmt.Errorf("imageStream object does not exist")
	}

	if len(builder.Object.Spec.Tags) == 0 {
		return "", fmt.Errorf("imageStream object %s in namespace %s has no tags",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	for _, tag := range builder.Object.Spec.Tags {
		if tag.From == nil {
			return "", fmt.Errorf("imageStream object %s in namespace %s has no DockerImage value",
				builder.Definition.Name, builder.Definition.Namespace)
		}

		if tag.Name == "" {
			return "", fmt.Errorf("imageStream object %s in namespace %s has no DockerImage tag value",
				builder.Definition.Name, builder.Definition.Namespace)
		}

		if tag.Name == imageTag {
			return tag.From.Name, nil
		}
	}

	return "", fmt.Errorf("image tag %s not found for imageStream object %s in namespace %s",
		imageTag, builder.Definition.Name, builder.Definition.Namespace)
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *Builder) validate() (bool, error) {
	resourceCRD := "ImageStream"

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

	if builder.errorMsg != "" {
		glog.V(100).Infof("The %s builder has error message: %s", resourceCRD, builder.errorMsg)

		return false, fmt.Errorf("%s", builder.errorMsg)
	}

	return true, nil
}
