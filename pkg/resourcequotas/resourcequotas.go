package resourcequotas

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	corev1Typed "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Builder provides struct for resource quotas containing connection to the cluster.
type Builder struct {
	// Resource Quota definition
	Definition *corev1.ResourceQuota

	// Created resource quota object
	Object *corev1.ResourceQuota

	// Used in functions that define or mutate deployment definition. errorMsg is processed before the deployment
	// object is created.
	errorMsg  string
	apiClient corev1Typed.CoreV1Interface
}

// NewBuilder creates a new resource quota builder.
func NewBuilder(apiClient *clients.Settings, name, nsname string) *Builder {
	glog.V(100).Infof("Initializing new resource quota structure with the following params: "+
		"name=%s, namespace=%s", name, nsname)

	if apiClient == nil {
		glog.V(100).Info("API client is nil")

		return nil
	}

	builder := &Builder{
		apiClient: apiClient.CoreV1Interface,
		Definition: &corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Info("Resource Quota name is empty")

		builder.errorMsg = "resource quota 'name' cannot be empty"

		return builder
	}

	if nsname == "" {
		glog.V(100).Info("Resource Quota namespace is empty")

		builder.errorMsg = "resource quota 'namespace' cannot be empty"

		return builder
	}

	return builder
}

// Pull retrieves the resource quota from the cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	if apiClient == nil {
		glog.V(100).Info("apiClient is nil")

		return nil, fmt.Errorf("apiClient is nil")
	}

	glog.V(100).Infof("Pulling resource quota with the following params: name=%s, namespace=%s", name, nsname)

	builder := &Builder{
		apiClient: apiClient.CoreV1Interface,
		Definition: &corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Info("Resource Quota name is empty")

		return nil, fmt.Errorf("resource quota 'name' cannot be empty")
	}

	if nsname == "" {
		glog.V(100).Info("Resource Quota namespace is empty")

		return nil, fmt.Errorf("resource quota 'namespace' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("resource quota %s does not exist in namespace %s",
			name, nsname)
	}

	builder.Definition = builder.Object

	return builder, nil
}

// WithQuotaSpec sets the resource quota spec.
func (builder *Builder) WithQuotaSpec(quotaSpec corev1.ResourceQuotaSpec) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting resource quota spec for %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec = quotaSpec

	return builder
}

// Update renovates the existing resource quota in the cluster.
func (builder *Builder) Update(force bool) (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating resource quota %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return builder, fmt.Errorf("resource quota %s does not exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	_, err := builder.apiClient.ResourceQuotas(
		builder.Definition.Namespace).Update(context.TODO(),
		builder.Definition, metav1.UpdateOptions{})

	if err != nil {
		if force {
			glog.V(100).Infof(
				msg.FailToUpdateNotification("resource quota", builder.Definition.Name, builder.Definition.Namespace))

			err := builder.Delete()

			if err != nil {
				glog.V(100).Infof(msg.FailToUpdateError("resource quota",
					builder.Definition.Name, builder.Definition.Namespace))

				return nil, err
			}

			return builder.Create()
		}

		return nil, err
	}

	builder.Object = builder.Definition

	return builder, err
}

// Exists checks if the resource quota exists in the cluster.
func (builder *Builder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if resource quota %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error

	builder.Object, err = builder.apiClient.ResourceQuotas(
		builder.Definition.Namespace).Get(context.TODO(),
		builder.Definition.Name, metav1.GetOptions{})

	return err == nil || !k8serrors.IsNotFound(err)
}

// Create creates the resource quota in the cluster.
func (builder *Builder) Create() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Creating resource quota %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	if !builder.Exists() {
		builder.Object, err = builder.apiClient.ResourceQuotas(builder.Definition.Namespace).
			Create(context.TODO(), builder.Definition, metav1.CreateOptions{})
	}

	return builder, err
}

// Delete deletes the resource quota from the cluster.
func (builder *Builder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting resource quota %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		glog.V(100).Infof("Resource quota %s cannot be deleted because it does not exist",
			builder.Definition.Name)

		builder.Object = nil

		return nil
	}

	err := builder.apiClient.ResourceQuotas(builder.Definition.Namespace).Delete(context.TODO(),
		builder.Definition.Name, metav1.DeleteOptions{})

	if err != nil {
		return err
	}

	builder.Object = nil

	return nil
}

// GetGVR returns resource quotas's GroupVersionResource which could be used for Clean function.
func GetGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "resourcequotas",
	}
}

func (builder *Builder) validate() (bool, error) {
	resourceCRD := "ResourceQuota"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Definition == nil {
		glog.V(100).Infof("The %s is undefined", resourceCRD)

		return false, fmt.Errorf("%s", msg.UndefinedCrdObjectErrString(resourceCRD))
	}

	if builder.apiClient == nil {
		glog.V(100).Infof("The %s API client is uninitialized", resourceCRD)

		return false, fmt.Errorf("%s builder cannot have nil apiClient", resourceCRD)
	}

	if builder.errorMsg != "" {
		glog.V(100).Infof("The %s builder has error message: %s",
			resourceCRD, builder.errorMsg)

		return false, fmt.Errorf("%s", builder.errorMsg)
	}

	return true, nil
}
