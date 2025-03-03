package replicaset

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

// Builder provides struct for replicaset object containing connection to the cluster and the replicaset definitions.
type Builder struct {
	// Replicaset definition. Used to create a replicaset object.
	Definition *appsv1.ReplicaSet
	// Created replicaset object.
	Object *appsv1.ReplicaSet
	// Used in functions that define or mutate replicaset definition. errorMsg is processed before the replicaset
	// object is created.
	errorMsg  string
	apiClient *clients.Settings
}

// AdditionalOptions additional options for replicaset object.
type AdditionalOptions func(builder *Builder) (*Builder, error)

var retryInterval = time.Second * 3

// NewBuilder creates a new instance of Builder.
func NewBuilder(
	apiClient *clients.Settings,
	name, nsname string,
	labels map[string]string,
	containerSpec []corev1.Container) *Builder {
	glog.V(100).Infof(
		"Initializing new replicaset structure with the following params: "+
			"name: %s, namespace: %s, containerSpec %v",
		name, nsname, containerSpec)

	if apiClient == nil {
		glog.V(100).Infof("replicaset 'apiClient' cannot be empty")

		return nil
	}

	builder := Builder{
		apiClient: apiClient,
		Definition: &appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
				Labels:    labels,
			},
			Spec: appsv1.ReplicaSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: containerSpec,
					},
				},
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the replicaset is empty")

		builder.errorMsg = "replicaset 'name' cannot be empty"

		return &builder
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the replicaset is empty")

		builder.errorMsg = "replicaset 'nsname' cannot be empty"

		return &builder
	}

	if len(labels) == 0 {
		glog.V(100).Infof("The labels of the replicaset is empty")

		builder.errorMsg = "replicaset 'labels' cannot be empty"

		return &builder
	}

	if len(containerSpec) == 0 {
		glog.V(100).Infof("The containerSpec of the replicaset is empty")

		builder.errorMsg = "replicaset 'containerSpec' cannot be empty"

		return &builder
	}

	return &builder
}

// Pull loads an existing replicaset into the Builder struct.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	glog.V(100).Infof("Pulling existing replicaset name:%s under namespace:%s", name, nsname)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")

		return nil, fmt.Errorf("replicaset 'apiClient' cannot be empty")
	}

	builder := Builder{
		apiClient: apiClient,
		Definition: &appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		return nil, fmt.Errorf("replicaset 'name' cannot be empty")
	}

	if nsname == "" {
		return nil, fmt.Errorf("replicaset 'nsname' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("replicaset object %s does not exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// WithLabel applies label to replicaset's definition.
func (builder *Builder) WithLabel(labels map[string]string) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(fmt.Sprintf("Defining replicaset's labels to %v", labels))

	if len(labels) == 0 {
		glog.V(100).Infof("The 'labels' of the replicaset is empty")

		builder.errorMsg = "can not apply empty labels"

		return builder
	}

	for labelKey := range labels {
		if labelKey == "" {
			glog.V(100).Infof("The 'labels' labelKey cannot be empty")

			builder.errorMsg = "can not apply labels with an empty labelKey value"

			return builder
		}
	}

	builder.Definition.Labels = labels

	return builder
}

// WithNodeSelector applies nodeSelector to the replicaset definition.
func (builder *Builder) WithNodeSelector(nodeSelector map[string]string) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Applying nodeSelector %s to replicaset %s in namespace %s",
		nodeSelector, builder.Definition.Name, builder.Definition.Namespace)

	if len(nodeSelector) == 0 {
		glog.V(100).Infof("The 'nodeSelector' of the replicaset is empty")

		builder.errorMsg = "can not apply empty nodeSelector"

		return builder
	}

	for key := range nodeSelector {
		if key == "" {
			glog.V(100).Infof("The 'nodeSelector' key value cannot be empty")

			builder.errorMsg = "can not apply a nodeSelector with an empty key value"

			return builder
		}
	}

	builder.Definition.Spec.Template.Spec.NodeSelector = nodeSelector

	return builder
}

// WithVolume defines Volume of replicaset under ContainerTemplateSpec.
func (builder *Builder) WithVolume(rsVolume corev1.Volume) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	if rsVolume.Name == "" {
		glog.V(100).Infof("The Volume name parameter is empty")

		builder.errorMsg = "volume name parameter is empty"

		return builder
	}

	glog.V(100).Infof("Adding volume %s for replicaset %s container template in namespace %s",
		rsVolume.Name, builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Template.Spec.Volumes = append(
		builder.Definition.Spec.Template.Spec.Volumes,
		rsVolume)

	return builder
}

// WithAdditionalContainerSpecs appends a list of container specs to the replicaset definition.
func (builder *Builder) WithAdditionalContainerSpecs(specs []corev1.Container) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Appending a list of container specs %v to replicaset %s in namespace %s",
		specs, builder.Definition.Name, builder.Definition.Namespace)

	if len(specs) == 0 {
		glog.V(100).Infof("The container specs are empty")

		builder.errorMsg = "cannot accept empty list as container specs"

		return builder
	}

	builder.Definition.Spec.Template.Spec.Containers = append(builder.Definition.Spec.Template.Spec.Containers, specs...)

	return builder
}

// Exists checks whether the given replicaset exists.
func (builder *Builder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if replicaset %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.apiClient.ReplicaSets(builder.Definition.Namespace).Get(
		context.TODO(), builder.Definition.Name, metav1.GetOptions{})

	return err == nil || !k8serrors.IsNotFound(err)
}

// Create builds replicaset in the cluster and stores the created object in struct.
func (builder *Builder) Create() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating replicaset %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	if !builder.Exists() {
		builder.Object, err = builder.apiClient.ReplicaSets(builder.Definition.Namespace).Create(
			context.TODO(), builder.Definition, metav1.CreateOptions{})
	}

	return builder, err
}

// Update renovates the existing replicaset object with replicaset definition in builder.
func (builder *Builder) Update() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating replicaset %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.apiClient.ReplicaSets(builder.Definition.Namespace).Update(
		context.TODO(), builder.Definition, metav1.UpdateOptions{})

	return builder, err
}

// Delete removes the replicaset.
func (builder *Builder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting replicaset %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		glog.V(100).Infof("ReplicaSet %s in namespaces %s does not exist",
			builder.Definition.Name, builder.Definition.Namespace)

		builder.Object = nil

		return nil
	}

	err := builder.apiClient.ReplicaSets(builder.Definition.Namespace).Delete(
		context.TODO(), builder.Definition.Name, metav1.DeleteOptions{})

	if err != nil {
		return err
	}

	builder.Object = nil

	return err
}

// CreateAndWaitUntilReady creates a replicaset in the cluster and waits until the replicaset is available.
func (builder *Builder) CreateAndWaitUntilReady(timeout time.Duration) (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating replicaset %s in namespace %s and waiting for the defined period "+
		"until it is ready",
		builder.Definition.Name, builder.Definition.Namespace)

	_, err := builder.Create()
	if err != nil {
		glog.V(100).Infof("Failed to create replicaset. Error is: '%s'", err.Error())

		return nil, err
	}

	// Polls every retryInterval to determine if replicaset is available.
	err = wait.PollUntilContextTimeout(
		context.TODO(), retryInterval, timeout, true, func(ctx context.Context) (bool, error) {
			builder.Object, err = builder.apiClient.ReplicaSets(builder.Definition.Namespace).Get(
				context.TODO(), builder.Definition.Name, metav1.GetOptions{})

			if err != nil {
				return false, nil
			}

			if builder.Object.Status.ReadyReplicas == builder.Object.Status.Replicas {
				return true, nil
			}

			return false, err
		})

	if err == nil {
		return builder, nil
	}

	return nil, err
}

// DeleteAndWait deletes a replicaset and waits until it is removed from the cluster.
func (builder *Builder) DeleteAndWait(timeout time.Duration) error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting replicaset %s in namespace %s and waiting for the defined period "+
		"until it is removed",
		builder.Definition.Name, builder.Definition.Namespace)

	if err := builder.Delete(); err != nil {
		return err
	}

	// Polls the replicaset every retryInterval until it is removed.
	return wait.PollUntilContextTimeout(
		context.TODO(), retryInterval, timeout, true, func(ctx context.Context) (bool, error) {
			_, err := builder.apiClient.ReplicaSets(builder.Definition.Namespace).Get(
				context.TODO(), builder.Definition.Name, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				return true, nil
			}

			return false, nil
		})
}

// IsReady waits for the replicaset to reach expected number of pods in Ready state.
func (builder *Builder) IsReady(timeout time.Duration) bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Running periodic check until replicaset %s in namespace %s is ready or "+
		"timeout %s exceeded", builder.Definition.Name, builder.Definition.Namespace, timeout.String())

	// Polls every retryInterval to determine if replicaset is available.
	err := wait.PollUntilContextTimeout(
		context.TODO(), retryInterval, timeout, true, func(ctx context.Context) (bool, error) {
			if !builder.Exists() {
				return false, fmt.Errorf("replicaset %s is not present on cluster", builder.Object.Name)
			}

			var err error
			builder.Object, err = builder.apiClient.ReplicaSets(builder.Definition.Namespace).Get(
				context.TODO(), builder.Definition.Name, metav1.GetOptions{})

			if err != nil {
				glog.V(100).Infof("Failed to get replicaset from cluster. Error is: '%s'", err.Error())

				return false, nil
			}

			if builder.Object.Status.ReadyReplicas == builder.Object.Status.Replicas {
				return true, nil
			}

			return false, err
		})

	return err == nil
}

// GetGVR returns the GroupVersionResource for replicaset.
func GetGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "replicasets"}
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *Builder) validate() (bool, error) {
	resourceCRD := "ReplicaSet"

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
