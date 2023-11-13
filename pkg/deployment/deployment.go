package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	multus "gopkg.in/k8snetworkplumbingwg/multus-cni.v4/pkg/types"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	v1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

// Builder provides struct for deployment object containing connection to the cluster and the deployment definitions.
type Builder struct {
	// Deployment definition. Used to create the deployment object.
	Definition *v1.Deployment
	// Created deployment object
	Object *v1.Deployment
	// Used in functions that define or mutate deployment definition. errorMsg is processed before the deployment
	// object is created.
	errorMsg  string
	apiClient *clients.Settings
}

// AdditionalOptions additional options for deployment object.
type AdditionalOptions func(builder *Builder) (*Builder, error)

// NewBuilder creates a new instance of Builder.
func NewBuilder(
	apiClient *clients.Settings, name, nsname string, labels map[string]string, containerSpec *coreV1.Container) *Builder {
	glog.V(100).Infof(
		"Initializing new deployment structure with the following params: "+
			"name: %s, namespace: %s, labels: %s, containerSpec %v",
		name, nsname, labels, containerSpec)

	builder := Builder{
		apiClient: apiClient,
		Definition: &v1.Deployment{
			Spec: v1.DeploymentSpec{
				Selector: &metaV1.LabelSelector{
					MatchLabels: labels,
				},
				Template: coreV1.PodTemplateSpec{
					ObjectMeta: metaV1.ObjectMeta{
						Labels: labels,
					},
				},
			},
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	builder.WithAdditionalContainerSpecs([]coreV1.Container{*containerSpec})

	if name == "" {
		glog.V(100).Infof("The name of the deployment is empty")

		builder.errorMsg = "deployment 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the deployment is empty")

		builder.errorMsg = "deployment 'namespace' cannot be empty"
	}

	if len(labels) == 0 {
		glog.V(100).Infof("There are no labels for the deployment")

		builder.errorMsg = "deployment 'labels' cannot be empty"
	}

	return &builder
}

// Pull loads an existing deployment into Builder struct.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	glog.V(100).Infof("Pulling existing deployment name: %s under namespace: %s", name, nsname)

	builder := Builder{
		apiClient: apiClient,
		Definition: &v1.Deployment{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		builder.errorMsg = "deployment 'name' cannot be empty"
	}

	if nsname == "" {
		builder.errorMsg = "deployment 'namespace' cannot be empty"
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("deployment oject %s doesn't exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// WithNodeSelector applies a nodeSelector to the deployment definition.
func (builder *Builder) WithNodeSelector(selector map[string]string) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Applying nodeSelector %s to deployment %s in namespace %s",
		selector, builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Template.Spec.NodeSelector = selector

	return builder
}

// WithReplicas sets the desired number of replicas in the deployment definition.
func (builder *Builder) WithReplicas(replicas int32) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting %d replicas in deployment %s in namespace %s",
		replicas, builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Replicas = &replicas

	return builder
}

// WithAdditionalContainerSpecs appends a list of container specs to the deployment definition.
func (builder *Builder) WithAdditionalContainerSpecs(specs []coreV1.Container) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Appending a list of container specs %v to deployment %s in namespace %s",
		specs, builder.Definition.Name, builder.Definition.Namespace)

	if len(specs) == 0 {
		glog.V(100).Infof("The container specs are empty")

		builder.errorMsg = "cannot accept empty list as container specs"
	}

	if builder.errorMsg != "" {
		return builder
	}

	if builder.Definition.Spec.Template.Spec.Containers == nil {
		builder.Definition.Spec.Template.Spec.Containers = specs
	} else {
		builder.Definition.Spec.Template.Spec.Containers = append(builder.Definition.Spec.Template.Spec.Containers, specs...)
	}

	return builder
}

// WithSecondaryNetwork applies Multus secondary network configuration on deployment definition.
func (builder *Builder) WithSecondaryNetwork(networks []*multus.NetworkSelectionElement) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Applying secondary networks %v to deployment %s", networks, builder.Definition.Name)

	if len(networks) == 0 {
		builder.errorMsg = "can not apply empty networks list"
	}

	netAnnotation, err := json.Marshal(networks)

	if err != nil {
		builder.errorMsg = fmt.Sprintf("error to unmarshal networks annotation due to: %s", err.Error())
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.Template.ObjectMeta.Annotations = map[string]string{
		"k8s.v1.cni.cncf.io/networks": string(netAnnotation)}

	return builder
}

// WithHugePages sets hugePages on all containers inside the deployment.
func (builder *Builder) WithHugePages() *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Applying hugePages configuration to all containers in deployment: %s",
		builder.Definition.Name)

	if builder.Definition.Spec.Template.Spec.Volumes != nil {
		builder.Definition.Spec.Template.Spec.Volumes = append(builder.Definition.Spec.Template.Spec.Volumes, coreV1.Volume{
			Name: "hugepages", VolumeSource: coreV1.VolumeSource{
				EmptyDir: &coreV1.EmptyDirVolumeSource{Medium: "HugePages"}}})
	} else {
		builder.Definition.Spec.Template.Spec.Volumes = []coreV1.Volume{
			{Name: "hugepages", VolumeSource: coreV1.VolumeSource{
				EmptyDir: &coreV1.EmptyDirVolumeSource{Medium: "HugePages"}},
			},
		}
	}

	for idx := range builder.Definition.Spec.Template.Spec.Containers {
		if builder.Definition.Spec.Template.Spec.Containers[idx].VolumeMounts != nil {
			builder.Definition.Spec.Template.Spec.Containers[idx].VolumeMounts = append(
				builder.Definition.Spec.Template.Spec.Containers[idx].VolumeMounts,
				coreV1.VolumeMount{Name: "hugepages", MountPath: "/mnt/huge"})
		} else {
			builder.Definition.Spec.Template.Spec.Containers[idx].VolumeMounts = []coreV1.VolumeMount{{
				Name:      "hugepages",
				MountPath: "/mnt/huge",
			},
			}
		}
	}

	return builder
}

// WithSecurityContext sets SecurityContext on deployment definition.
func (builder *Builder) WithSecurityContext(securityContext *coreV1.PodSecurityContext) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Applying SecurityContext configuration on deployment %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if securityContext == nil {
		glog.V(100).Infof("The 'securityContext' of the deployment is empty")

		builder.errorMsg = "'securityContext' parameter is empty"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.Template.Spec.SecurityContext = securityContext

	return builder
}

// WithLabel applies label to deployment's definition.
func (builder *Builder) WithLabel(labelKey, labelValue string) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(fmt.Sprintf("Defining deployment's label to %s:%s", labelKey, labelValue))

	if labelKey == "" {
		glog.V(100).Infof("The 'labelKey' of the deployment is empty")

		builder.errorMsg = "can not apply empty labelKey"
	}

	if builder.errorMsg != "" {
		return builder
	}

	if builder.Definition.Spec.Template.Labels == nil {
		builder.Definition.Spec.Template.Labels = map[string]string{labelKey: labelValue}

		return builder
	}

	builder.Definition.Spec.Template.Labels[labelKey] = labelValue

	return builder
}

// WithServiceAccountName sets the ServiceAccountName on deployment definition.
func (builder *Builder) WithServiceAccountName(serviceAccountName string) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting ServiceAccount %s on deployment %s in namespace %s",
		serviceAccountName, builder.Definition.Name, builder.Definition.Namespace)

	if serviceAccountName == "" {
		glog.V(100).Infof("The 'serviceAccount' of the deployment is empty")

		builder.errorMsg = "can not apply empty serviceAccount"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.Template.Spec.ServiceAccountName = serviceAccountName

	return builder
}

// WithVolume attaches given volume to the deployment.
func (builder *Builder) WithVolume(deployVolume coreV1.Volume) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	if deployVolume.Name == "" {
		glog.V(100).Infof("The volume's name cannot be empty")

		builder.errorMsg = "The volume's name cannot be empty"
	}

	if builder.errorMsg != "" {
		return builder
	}

	glog.V(100).Infof("Adding volume %s to deployment %s in namespace %s",
		deployVolume.Name, builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Template.Spec.Volumes = append(
		builder.Definition.Spec.Template.Spec.Volumes,
		deployVolume)

	return builder
}

// WithOptions creates deployment with generic mutation options.
func (builder *Builder) WithOptions(options ...AdditionalOptions) *Builder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting deployment additional options")

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

// Create generates a deployment in cluster and stores the created object in struct.
func (builder *Builder) Create() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating deployment %s in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

	var err error
	if !builder.Exists() {
		builder.Object, err = builder.apiClient.Deployments(builder.Definition.Namespace).Create(
			context.TODO(), builder.Definition, metaV1.CreateOptions{})
	}

	return builder, err
}

// Update renovates the existing deployment object with the deployment definition in builder.
func (builder *Builder) Update() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating deployment %s in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.apiClient.Deployments(builder.Definition.Namespace).Update(
		context.TODO(), builder.Definition, metaV1.UpdateOptions{})

	return builder, err
}

// Delete removes a deployment.
func (builder *Builder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting deployment %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return nil
	}

	err := builder.apiClient.Deployments(builder.Definition.Namespace).Delete(
		context.TODO(), builder.Object.Name, metaV1.DeleteOptions{})

	if err != nil {
		return err
	}

	builder.Object = nil

	return err
}

// CreateAndWaitUntilReady creates a deployment in the cluster and waits until the deployment is available.
func (builder *Builder) CreateAndWaitUntilReady(timeout time.Duration) (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating deployment %s in namespace %s and waiting for the defined period until it's ready",
		builder.Definition.Name, builder.Definition.Namespace)

	if _, err := builder.Create(); err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	if builder.IsReady(timeout) {
		return builder, nil
	}

	return nil, fmt.Errorf("deployment %s in namespace %s is not ready",
		builder.Definition.Name, builder.Definition.Namespace,
	)
}

// IsReady periodically checks if deployment is in ready status.
func (builder *Builder) IsReady(timeout time.Duration) bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Running periodic check until deployment %s in namespace %s is ready",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return false
	}

	err := wait.PollImmediate(time.Second, timeout, func() (bool, error) {

		var err error
		builder.Object, err = builder.apiClient.Deployments(builder.Definition.Namespace).Get(
			context.Background(), builder.Definition.Name, metaV1.GetOptions{})

		if err != nil {
			return false, err
		}

		if builder.Object.Status.ReadyReplicas > 0 && builder.Object.Status.Replicas == builder.Object.Status.ReadyReplicas {
			return true, nil
		}

		return false, nil
	})

	return err == nil
}

// DeleteAndWait deletes a deployment and waits until it is removed from the cluster.
func (builder *Builder) DeleteAndWait(timeout time.Duration) error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting deployment %s in namespace %s and waiting for the defined period until it's removed",
		builder.Definition.Name, builder.Definition.Namespace)

	if err := builder.Delete(); err != nil {
		return err
	}

	// Polls the deployment every second until it's removed.
	return wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		_, err := builder.apiClient.Deployments(builder.Definition.Namespace).Get(
			context.Background(), builder.Definition.Name, metaV1.GetOptions{})
		if k8serrors.IsNotFound(err) {

			return true, nil
		}

		return false, nil
	})
}

// Exists checks whether the given deployment exists.
func (builder *Builder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if deployment %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.apiClient.Deployments(builder.Definition.Namespace).Get(
		context.Background(), builder.Definition.Name, metaV1.GetOptions{})

	return err == nil || !k8serrors.IsNotFound(err)
}

// WaitUntilCondition waits for the duration of the defined timeout or until the
// deployment gets to a specific condition.
func (builder *Builder) WaitUntilCondition(condition v1.DeploymentConditionType, timeout time.Duration) error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Waiting for the defined period until deployment %s in namespace %s has condition %v",
		builder.Definition.Name, builder.Definition.Namespace, condition)

	if !builder.Exists() {
		return fmt.Errorf("cannot wait for deployment condition because it does not exist")
	}

	return wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		updateDeployment, err := builder.apiClient.Deployments(builder.Definition.Namespace).Get(
			context.Background(), builder.Definition.Name, metaV1.GetOptions{})
		if err != nil {
			return false, nil
		}

		for _, cond := range updateDeployment.Status.Conditions {
			if cond.Type == condition && cond.Status == coreV1.ConditionTrue {
				return true, nil
			}
		}

		return false, nil

	})
}

// GetGVR returns deployment's GroupVersionResource which could be used for Clean function.
func GetGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *Builder) validate() (bool, error) {
	resourceCRD := "ClusterDeployment"

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
