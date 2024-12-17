package clusterlogging

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	eskv1 "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ElasticsearchBuilder provides struct for the elasticsearch object.
type ElasticsearchBuilder struct {
	// Elasticsearch definition. Used to create elasticsearch object with minimum set of required elements.
	Definition *eskv1.Elasticsearch
	// Created elasticsearch object on the cluster.
	Object *eskv1.Elasticsearch
	// api client to interact with the cluster.
	apiClient goclient.Client
	// errorMsg is processed before elasticsearch object is created.
	errorMsg string
}

// NewElasticsearchBuilder method creates new instance of builder.
func NewElasticsearchBuilder(
	apiClient *clients.Settings, name, nsname string) *ElasticsearchBuilder {
	glog.V(100).Infof("Initializing new elasticsearch structure with the following params: name: %s, namespace: %s",
		name, nsname)

	if apiClient == nil {
		glog.V(100).Infof("elasticsearch 'apiClient' cannot be empty")

		return nil
	}

	err := apiClient.AttachScheme(eskv1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add eskv1 scheme to client schemes")

		return nil
	}

	builder := &ElasticsearchBuilder{
		apiClient: apiClient.Client,
		Definition: &eskv1.Elasticsearch{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the elasticsearch is empty")

		builder.errorMsg = "elasticsearch 'name' cannot be empty"

		return builder
	}

	if nsname == "" {
		glog.V(100).Infof("The nsname of the elasticsearch is empty")

		builder.errorMsg = "elasticsearch 'nsname' cannot be empty"

		return builder
	}

	return builder
}

// PullElasticsearch retrieves an existing elasticsearch object from the cluster.
func PullElasticsearch(apiClient *clients.Settings, name, nsname string) (*ElasticsearchBuilder, error) {
	glog.V(100).Infof(
		"Pulling elasticsearch object name:%s in namespace: %s", name, nsname)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")

		return nil, fmt.Errorf("elasticsearch 'apiClient' cannot be empty")
	}

	err := apiClient.AttachScheme(eskv1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add eskv1 scheme to client schemes")

		return nil, err
	}

	builder := &ElasticsearchBuilder{
		apiClient: apiClient.Client,
		Definition: &eskv1.Elasticsearch{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the elasticsearch is empty")

		return nil, fmt.Errorf("elasticsearch 'name' cannot be empty")
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the elasticsearch is empty")

		return nil, fmt.Errorf("elasticsearch 'nsname' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("elasticsearch object %s does not exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return builder, nil
}

// Get returns elasticsearch object if found.
func (builder *ElasticsearchBuilder) Get() (*eskv1.Elasticsearch, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting elasticsearch %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	elasticsearchObj := &eskv1.Elasticsearch{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, elasticsearchObj)

	if err != nil {
		return nil, err
	}

	return elasticsearchObj, err
}

// Create makes a elasticsearch in the cluster and stores the created object in struct.
func (builder *ElasticsearchBuilder) Create() (*ElasticsearchBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the elasticsearch %s in namespace %s",
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

// Delete removes elasticsearch from a cluster.
func (builder *ElasticsearchBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting the elasticsearch %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return fmt.Errorf("elasticsearch cannot be deleted because it does not exist")
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return fmt.Errorf("can not delete elasticsearch: %w", err)
	}

	builder.Object = nil

	return nil
}

// Exists checks whether the given elasticsearch exists.
func (builder *ElasticsearchBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if elasticsearch %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Update renovates the existing elasticsearch object with elasticsearch definition in builder.
func (builder *ElasticsearchBuilder) Update() (*ElasticsearchBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Info("Updating elasticsearch %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		glog.V(100).Infof(
			msg.FailToUpdateError("elasticsearch", builder.Definition.Name, builder.Definition.Namespace))

		return nil, err
	}

	if err == nil {
		builder.Object = builder.Definition
	}

	return builder, err
}

// WithManagementState sets the elasticsearch operator's management state.
func (builder *ElasticsearchBuilder) WithManagementState(
	expectedManagementState eskv1.ManagementState) *ElasticsearchBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Setting elasticsearch %s in namespace %s with the ManagementState: %v",
		builder.Definition.Name, builder.Definition.Namespace, expectedManagementState)

	builder.Definition.Spec.ManagementState = expectedManagementState

	return builder
}

// GetManagementState fetches elasticsearch ManagementState.
func (builder *ElasticsearchBuilder) GetManagementState() (*eskv1.ManagementState, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting elasticsearch ManagementState configuration")

	if !builder.Exists() {
		return nil, fmt.Errorf("elasticsearch object does not exist")
	}

	return &builder.Object.Spec.ManagementState, nil
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *ElasticsearchBuilder) validate() (bool, error) {
	resourceCRD := "Elasticsearch"

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
