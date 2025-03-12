package oran

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	provisioningv1alpha1 "github.com/openshift-kni/oran-o2ims/api/provisioning/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ProvisioningRequestBuilder provides a struct to inferface with ProvisioningRequest resources on a specific cluster.
type ProvisioningRequestBuilder struct {
	// Definition of the ProvisioningRequest used to create the resource.
	Definition *provisioningv1alpha1.ProvisioningRequest
	// Object of the ProvisioningRequest as it is on the cluster.
	Object *provisioningv1alpha1.ProvisioningRequest
	// apiClient used to interact with the cluster.
	apiClient runtimeclient.Client
	// errorMsg used to store latest error message from functions that do not return errors.
	errorMsg string
}

// NewPRBuilder creates a new instance of a ProvisioningRequest builder.
func NewPRBuilder(
	apiClient *clients.Settings, name, templateName, templateVersion string) *ProvisioningRequestBuilder {
	glog.V(100).Infof(
		"Initializing new ProvisioningRequest structure with the following params: "+
			"name: %s, templateName: %s, templateVersion: %s",
		name, templateName, templateVersion)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient of the ProvisioningRequest is nil")

		return nil
	}

	err := apiClient.AttachScheme(provisioningv1alpha1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add provisioning v1alpha1 scheme to client schemes: %v", err)

		return nil
	}

	builder := &ProvisioningRequestBuilder{
		apiClient: apiClient.Client,
		Definition: &provisioningv1alpha1.ProvisioningRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: provisioningv1alpha1.ProvisioningRequestSpec{
				TemplateName:    templateName,
				TemplateVersion: templateVersion,
			},
		},
	}

	if err := uuid.Validate(name); err != nil {
		glog.V(100).Infof("The name of the ProvisioningRequest is not a valid UUID: %v", err)

		builder.errorMsg = "provisioningRequest 'name' must be a valid UUID"

		return builder
	}

	if templateName == "" {
		glog.V(100).Info("The template name of the ProvisioningRequest is empty")

		builder.errorMsg = "provisioningRequest 'templateName' cannot be empty"

		return builder
	}

	if templateVersion == "" {
		glog.V(100).Info("The template version of the ProvisioningRequest is empty")

		builder.errorMsg = "provisioningRequest 'templateVersion' cannot be empty"

		return builder
	}

	return builder
}

// WithTemplateParameter sets key to value in the TemplateParameters field.
func (builder *ProvisioningRequestBuilder) WithTemplateParameter(key string, value any) *ProvisioningRequestBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting ProvisioningRequest TemplateParameter %s to %v", key, value)

	if key == "" {
		glog.V(100).Info("ProvisioningRequest TemplateParameter key is empty")

		builder.errorMsg = "provisioningRequest TemplateParameter 'key' cannot be empty"

		return builder
	}

	templateParameters, err := builder.GetTemplateParameters()
	if err != nil {
		glog.V(100).Infof("Failed to unmarshal ProvisioningRequest TemplateParameters: %v", err)

		builder.errorMsg = fmt.Sprintf("failed to unmarshal TemplateParameters: %v", err)

		return builder
	}

	templateParameters[key] = value
	builder = builder.WithTemplateParameters(templateParameters)

	return builder
}

// GetTemplateParameters unmarshals the raw JSON stored in the TemplateParameters in the Definition, returning an empty
// rather than nil map if TemplateParameters is empty.
func (builder *ProvisioningRequestBuilder) GetTemplateParameters() (map[string]any, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting the TemplateParameters map for ProvisioningRequest %s", builder.Definition.Name)

	templateParameters := make(map[string]any)

	if len(builder.Definition.Spec.TemplateParameters.Raw) == 0 {
		return templateParameters, nil
	}

	err := json.Unmarshal(builder.Definition.Spec.TemplateParameters.Raw, &templateParameters)
	if err != nil {
		return nil, err
	}

	return templateParameters, nil
}

// WithTemplateParameters marshals the provided map into JSON and stores it in the builder Definition. Nil maps are
// converted to empty maps before marshaling.
func (builder *ProvisioningRequestBuilder) WithTemplateParameters(
	templateParameters map[string]any) *ProvisioningRequestBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting the TemplateParameters map for ProvisioningRequest %s", builder.Definition.Name)

	if templateParameters == nil {
		templateParameters = make(map[string]any)
	}

	marshaled, err := json.Marshal(templateParameters)
	if err != nil {
		glog.V(100).Infof("Failed to marshal TemplateParameters for ProvisioningRequest %s: %v", builder.Definition.Name, err)

		builder.errorMsg = fmt.Sprintf("failed to marshal TemplateParameters: %v", err)

		return builder
	}

	builder.Definition.Spec.TemplateParameters = runtime.RawExtension{Raw: marshaled}

	return builder
}

// PullPR pulls an existing ProvisioningRequest into a Builder struct.
func PullPR(apiClient *clients.Settings, name string) (*ProvisioningRequestBuilder, error) {
	glog.V(100).Infof("Pulling existing ProvisioningRequest %s from cluster", name)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient of the ProvisioningRequest is nil")

		return nil, fmt.Errorf("provisioningRequest 'apiClient' cannot be nil")
	}

	err := apiClient.AttachScheme(provisioningv1alpha1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add provisioning v1alpha1 scheme to client schemes: %v", err)

		return nil, err
	}

	builder := &ProvisioningRequestBuilder{
		apiClient: apiClient.Client,
		Definition: &provisioningv1alpha1.ProvisioningRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	if name == "" {
		glog.V(100).Info("The name of the ProvisioningRequest is empty")

		return nil, fmt.Errorf("provisioningRequest 'name' cannot be empty")
	}

	if !builder.Exists() {
		glog.V(100).Info("The ProvisioningRequest %s does not exist", name)

		return nil, fmt.Errorf("provisioningRequest object %s does not exist", name)
	}

	builder.Definition = builder.Object

	return builder, nil
}

// Get returns the ProvisioningRequest object if found.
func (builder *ProvisioningRequestBuilder) Get() (*provisioningv1alpha1.ProvisioningRequest, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof(
		"Getting ProvisioningRequest object %s", builder.Definition.Name)

	provisioningRequest := &provisioningv1alpha1.ProvisioningRequest{}
	err := builder.apiClient.Get(context.TODO(), runtimeclient.ObjectKey{
		Name: builder.Definition.Name,
	}, provisioningRequest)

	if err != nil {
		glog.V(100).Infof("Failed to get ProvisioningRequest object %s: %v", builder.Definition.Name, err)

		return nil, err
	}

	return provisioningRequest, nil
}

// Exists checks whether the given ProvisioningRequest exists on the cluster.
func (builder *ProvisioningRequestBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if ProvisioningRequest %s exists", builder.Definition.Name)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Create makes a ProvisioningRequest on the cluster if it does not already exist.
func (builder *ProvisioningRequestBuilder) Create() (*ProvisioningRequestBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Creating ProvisioningRequest %s", builder.Definition.Name)

	if builder.Exists() {
		return builder, nil
	}

	err := builder.apiClient.Create(context.TODO(), builder.Definition)
	if err != nil {
		return nil, err
	}

	builder.Object = builder.Definition

	return builder, nil
}

// Update changes the existing ProvisioningRequest resource on the cluster. Since deleting a ProvisioningRequest is a
// non-trivial operation and corresponds to deleting a cluster, there is no option to fall back to deleting and
// recreating the ProvisioningRequest.
func (builder *ProvisioningRequestBuilder) Update() (*ProvisioningRequestBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Updating ProvisioningRequest %s", builder.Definition.Name)

	if !builder.Exists() {
		glog.V(100).Infof("ProvisioningRequest %s does not exist", builder.Definition.Name)

		return nil, fmt.Errorf("cannot update non-existent provisioningRequest")
	}

	builder.Definition.ResourceVersion = builder.Object.ResourceVersion

	err := builder.apiClient.Update(context.TODO(), builder.Definition)
	if err != nil {
		return nil, err
	}

	builder.Object = builder.Definition

	return builder, nil
}

// Delete removes a ProvisioningRequest from the cluster if it exists.
func (builder *ProvisioningRequestBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting ProvisioningRequest %s", builder.Definition.Name)

	if !builder.Exists() {
		glog.V(100).Infof("ProvisioningRequest %s does not exist", builder.Definition.Name)

		builder.Object = nil

		return nil
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Object)
	if err != nil {
		return err
	}

	builder.Object = nil

	return nil
}

// DeleteAndWait deletes the ProvisioningRequest then waits up to timeout until the ProvisioningRequest no longer
// exists.
func (builder *ProvisioningRequestBuilder) DeleteAndWait(timeout time.Duration) error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting ProvisioningRequest %s and waiting up to %s until it is deleted",
		builder.Definition.Name, timeout)

	err := builder.Delete()
	if err != nil {
		return err
	}

	return wait.PollUntilContextTimeout(
		context.TODO(), 3*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			return !builder.Exists(), nil
		})
}

// WaitForCondition waits up to the provided timeout for a condition matching expected. It checks only the Type, Status,
// Reason, and Message fields. For the message, it matches if the message contains the expected. Zero fields in the
// expected condition are ignored.
func (builder *ProvisioningRequestBuilder) WaitForCondition(
	expected metav1.Condition, timeout time.Duration) (*ProvisioningRequestBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Waiting up to %s until ProvisioningRequest %s has condition %v",
		timeout, builder.Definition.Name, expected)

	if !builder.Exists() {
		glog.V(100).Infof("ProvisioningRequest %s does not exist", builder.Definition.Name)

		return nil, fmt.Errorf("cannot wait for non-existent ProvisioningRequest")
	}

	err := wait.PollUntilContextTimeout(
		context.TODO(), 3*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			var err error
			builder.Object, err = builder.Get()

			if err != nil {
				glog.V(100).Infof("Failed to get ProvisioningRequest %s: %v", builder.Definition.Name, err)

				return false, nil
			}

			builder.Definition = builder.Object

			for _, condition := range builder.Object.Status.Conditions {
				if expected.Type != "" && condition.Type != expected.Type {
					continue
				}

				if expected.Status != "" && condition.Status != expected.Status {
					continue
				}

				if expected.Reason != "" && condition.Reason != expected.Reason {
					continue
				}

				if expected.Message != "" && !strings.Contains(condition.Message, expected.Message) {
					continue
				}

				return true, nil
			}

			return false, nil
		})

	if err != nil {
		return nil, err
	}

	return builder, nil
}

// WaitUntilFulfilled waits up to timeout until the ProvisioningRequest is in the fulfilled phase. Changes to the
// template will not be accepted until it is fulfilled.
func (builder *ProvisioningRequestBuilder) WaitUntilFulfilled(
	timeout time.Duration) (*ProvisioningRequestBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Waiting up to %s until ProvisioningRequest %s is fulfilled", timeout, builder.Definition.Name)

	if !builder.Exists() {
		glog.V(100).Infof("ProvisioningRequest %s does not exist", builder.Definition.Name)

		return nil, fmt.Errorf("cannot wait for non-existent ProvisioningRequest")
	}

	err := wait.PollUntilContextTimeout(
		context.TODO(), 3*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			var err error
			builder.Object, err = builder.Get()

			if err != nil {
				glog.V(100).Infof("Failed to get ProvisioningRequest %s: %v", builder.Definition.Name, err)

				return false, nil
			}

			builder.Definition = builder.Object
			fulfilled := builder.Definition.Status.ProvisioningStatus.ProvisioningPhase == provisioningv1alpha1.StateFulfilled

			return fulfilled, nil
		})

	if err != nil {
		return nil, err
	}

	return builder, nil
}

// validate checks that the builder, definition, and apiClient are properly initialized and there is no errorMsg.
func (builder *ProvisioningRequestBuilder) validate() (bool, error) {
	resourceCRD := "provisioningRequest"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Definition == nil {
		glog.V(100).Infof("The %s is uninitialized", resourceCRD)

		return false, fmt.Errorf("%s", msg.UndefinedCrdObjectErrString(resourceCRD))
	}

	if builder.apiClient == nil {
		glog.V(100).Infof("The %s builder apiClient is nil", resourceCRD)

		return false, fmt.Errorf("%s builder cannot have nil apiClient", resourceCRD)
	}

	if builder.errorMsg != "" {
		glog.V(100).Infof("The %s builder has error message %s", resourceCRD, builder.errorMsg)

		return false, fmt.Errorf("%s", builder.errorMsg)
	}

	return true, nil
}
