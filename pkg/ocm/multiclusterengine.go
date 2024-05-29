package ocm

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	mceV1 "github.com/stolostron/backplane-operator/api/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// MultiClusterEngineBuilder provides struct for the MultiClusterEngine object containing connection to
// the cluster and the MultiClusterEngine definitions.
type MultiClusterEngineBuilder struct {
	Definition *mceV1.MultiClusterEngine
	Object     *mceV1.MultiClusterEngine
	errorMsg   string
	apiClient  runtimeClient.Client
}

// MultiClusterEngineAdditionalOptions additional options for MultiClusterEngine object.
type MultiClusterEngineAdditionalOptions func(builder *MultiClusterEngineBuilder) (*MultiClusterEngineBuilder, error)

// NewMultiClusterEngineBuilder creates a new instance of MultiClusterEngineBuilder.
func NewMultiClusterEngineBuilder(apiClient *clients.Settings, name string) *MultiClusterEngineBuilder {
	glog.V(100).Infof(
		`Initializing new MultiClusterEngine structure with the following params: name: %s`, name)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")
		return nil
	}

	builder := MultiClusterEngineBuilder{
		apiClient: apiClient.Client,
		Definition: &mceV1.MultiClusterEngine{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: mceV1.MultiClusterEngineSpec{},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the MultiClusterEngine is empty")

		builder.errorMsg = "multiclusterengine 'name' cannot be empty"
	}

	return &builder
}

// WithOptions creates MultiClusterEngine with generic mutation options.
func (builder *MultiClusterEngineBuilder) WithOptions(
	options ...MultiClusterEngineAdditionalOptions) *MultiClusterEngineBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting MultiClusterEngine additional options")

	for _, option := range options {
		if option != nil {
			builder, err := option(builder)

			if err != nil {
				glog.V(100).Infof("Error occurred in mutation function")

				builder.errorMsg = err.Error()

				return builder
			}
		}
	}

	return builder
}

// PullMultiClusterEngine loads an existing MultiClusterEngine into MultiClusterEngineBuilder struct.
func PullMultiClusterEngine(apiClient *clients.Settings, name string) (*MultiClusterEngineBuilder, error) {
	glog.V(100).Infof("Pulling existing MultiClusterEngine name: %s", name)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")
		return nil, fmt.Errorf("multiclusterengine 'apiclient' cannot be empty")
	}

	builder := MultiClusterEngineBuilder{
		apiClient: apiClient.Client,
		Definition: &mceV1.MultiClusterEngine{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	if name == "" {
		builder.errorMsg = "multiclusterengine 'name' cannot be empty"
		return nil, fmt.Errorf("multiclusterengine 'name' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("multiclusterengine object %s does not exist", name)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Get fetches the defined MultiClusterEngine from the cluster.
func (builder *MultiClusterEngineBuilder) Get() (*mceV1.MultiClusterEngine, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting MultiClusterEngine %s", builder.Definition.Name)

	MultiClusterEngine := &mceV1.MultiClusterEngine{}
	err := builder.apiClient.Get(context.TODO(), runtimeClient.ObjectKey{
		Name: builder.Definition.Name,
	}, MultiClusterEngine)

	if err != nil {
		return nil, err
	}

	return MultiClusterEngine, err
}

// Update modifies an existing MultiClusterEngine on the cluster.
func (builder *MultiClusterEngineBuilder) Update() (*MultiClusterEngineBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating MultiClusterEngine %s", builder.Definition.Name)

	err := builder.apiClient.Update(context.TODO(), builder.Definition)
	builder.Object = builder.Definition

	return builder, err
}

// Delete removes a MultiClusterEngine from the cluster.
func (builder *MultiClusterEngineBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting the MultiClusterEngine %s", builder.Definition.Name)

	if !builder.Exists() {
		return fmt.Errorf("multiclusterengine cannot be deleted because it does not exist")
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return fmt.Errorf("cannot delete multiclusterengine: %w", err)
	}

	builder.Object = nil
	builder.Definition.ResourceVersion = ""
	builder.Definition.CreationTimestamp = metav1.Time{}

	return nil
}

// Exists checks if the defined MultiClusterEngine has already been created.
func (builder *MultiClusterEngineBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if multiclusterengine %s exists", builder.Definition.Name)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *MultiClusterEngineBuilder) validate() (bool, error) {
	resourceCRD := "MultiClusterEngine"

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
