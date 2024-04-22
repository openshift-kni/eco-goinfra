package nodesconfig

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	configv1 "github.com/openshift/api/config/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Builder provides a struct for nodesConfig object from the cluster and a nodesConfig definition.
type Builder struct {
	// nodesConfig definition, used to create the nodesConfig object.
	Definition *configv1.Node
	// Created nodesConfig object.
	Object *configv1.Node
	// api client to interact with the cluster.
	apiClient goclient.Client
	// Used in functions that define or mutate clusterOperator definition. errorMsg is processed before the
	// ClusterOperator object is created.
	errorMsg string
}

// Pull retrieves an existing nodesConfig object from the cluster.
func Pull(apiClient *clients.Settings, nodesConfigObjName string) (*Builder, error) {
	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")

		return nil, fmt.Errorf("nodesConfig Config 'apiClient' cannot be empty")
	}

	if nodesConfigObjName == "" {
		glog.V(100).Infof("The name of the nodesConfig is empty")

		return nil, fmt.Errorf("nodesConfig 'nodesConfigObjName' cannot be empty")
	}

	glog.V(100).Infof(
		"Pulling nodesConfig object name: %s", nodesConfigObjName)

	builder := Builder{
		apiClient: apiClient.Client,
		Definition: &configv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodesConfigObjName,
			},
		},
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("nodesConfig object %s doesn't exist", nodesConfigObjName)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Get fetches existing nodesConfig from cluster.
func (builder *Builder) Get() (*configv1.Node, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting existing nodesConfig with name %s from cluster", builder.Definition.Name)

	nodesConfig := &configv1.Node{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name: builder.Definition.Name,
	}, nodesConfig)

	if err != nil {
		glog.V(100).Infof("A nodesConfig object %s doesn't exist", builder.Definition.Name)

		return nil, err
	}

	return nodesConfig, nil
}

// Exists checks whether the given nodesConfig exists.
func (builder *Builder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if nodesConfig %s exists", builder.Definition.Name)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Update renovates the nodesConfig in the cluster and stores the created object in struct.
func (builder *Builder) Update() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating the nodesConfig %s", builder.Definition.Name)

	if !builder.Exists() {
		return nil, fmt.Errorf("nodesConfig object %s doesn't exist", builder.Definition.Name)
	}

	err := builder.apiClient.Update(context.TODO(), builder.Definition)
	if err == nil {
		builder.Object = builder.Definition
	}

	return builder, err
}

// GetCGroupMode fetches nodesConfig cgroupMode.
func (builder *Builder) GetCGroupMode() (configv1.CgroupMode, error) {
	if valid, err := builder.validate(); !valid {
		return "", err
	}

	glog.V(100).Infof("Getting nodesConfig ManagementState configuration")

	if !builder.Exists() {
		return "", fmt.Errorf("imageRegistry object doesn't exist")
	}

	return builder.Object.Spec.CgroupMode, nil
}

// WithCGroupMode sets the nodesConfig operator's cgroupMode.
func (builder *Builder) WithCGroupMode(expectedCGroupMode configv1.CgroupMode) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Setting nodesConfig %s with ManagementState: %v",
		builder.Definition.Name, expectedCGroupMode)

	builder.Definition.Spec.CgroupMode = expectedCGroupMode

	return builder
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *Builder) validate() (bool, error) {
	resourceCRD := "Nodes.Config"

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
