package machine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// MachineSetBuilder provides a struct for MachineSet object from the cluster and a MachineSet definition.
type MachineSetBuilder struct {
	// MachineSetBuilder definition. Used to create
	// MachineSet object with minimum set of required elements.
	Definition *machinev1beta1.MachineSet
	// Created MachineSetBuilder object on the cluster.
	Object *machinev1beta1.MachineSet
	// api client to interact with the cluster.
	apiClient *clients.Settings
	// errorMsg is processed before MachineSetBuilder object is created.
	errorMsg string
}

// NewMachineSetBuilderFromCopy returns an MachineSetBuilder struct from a copied MachineSet.
func NewMachineSetBuilderFromCopy(
	apiClient *clients.Settings,
	nsName,
	publicCloud,
	instanceType string,
	replicas int32) *MachineSetBuilder {
	glog.V(100).Infof(
		"Initializing new MachineSetBuilder structure from copied MachineSet with the following params: "+
			"namespace: %s, publicCloud: %s, instanceType: %s and replicas: %v",
		nsName, publicCloud, instanceType, replicas)

	newMachineSet, err := CreateNewWorkerMachineSetFromCopy(apiClient, nsName, publicCloud, instanceType, replicas)

	builder := MachineSetBuilder{
		apiClient:  apiClient,
		Definition: newMachineSet,
	}

	if err != nil {
		glog.V(100).Infof(
			"Error initializing MachineSet from copy: %s", err.Error())

		builder.errorMsg = fmt.Sprintf("Error initializing MachineSet from copy: %s",
			err.Error())
	}

	if nsName == "" {
		glog.V(100).Infof("The Namespace of the MachineSet is empty")

		builder.errorMsg = "MachineSet 'nseName' cannot be empty"
	}

	if instanceType == "" {
		glog.V(100).Infof("The instanceType of the MachineSet is empty")

		builder.errorMsg = "MachineSet 'instanceType' cannot be empty"
	}

	if replicas == 0 {
		glog.V(100).Infof("The replicas of the MachineSet is zero")

		builder.errorMsg = "MachineSet 'replicas' cannot be zero"
	}

	if builder.Definition == nil {
		glog.V(100).Infof("The MachineSet object definition is nil")

		builder.errorMsg = "MachineSet 'Object.Definition' is nil"
	}

	return &builder
}

// PullMachineSet loads an existing MachineSet into Builder struct.
func PullMachineSet(apiClient *clients.Settings, name, namespace string) (*MachineSetBuilder, error) {
	glog.V(100).Infof("Pulling existing machineSet name %s in namespace %s", name, namespace)

	builder := MachineSetBuilder{
		apiClient: apiClient,
		Definition: &machinev1beta1.MachineSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		},
	}

	if name == "" {
		builder.errorMsg = "MachineSet 'name' cannot be empty"
	}

	if namespace == "" {
		builder.errorMsg = "MachineSet 'namespace' cannot be empty"
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("MachineSet object %s doesn't exist in namespace %s", name, namespace)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// CreateNewWorkerMachineSetFromCopy returns a MachineSet object which was copied from an exiting one on cluster.
func CreateNewWorkerMachineSetFromCopy(
	apiClient *clients.Settings,
	namespace,
	publicCloud,
	instanceType string,
	replicas int32) (*machinev1beta1.MachineSet, error) {
	// currently only supporting AWS MachineSets
	if publicCloud != "aws" {
		return nil, fmt.Errorf("only AWS public cloud is currently supported for copied MachineSet")
	}

	workerMachineSetList, err := GetWorkerMachineSetList(apiClient, namespace)

	if err != nil {
		return nil, fmt.Errorf("could not get worker MachineSetList: %w", err)
	}

	// picking the first worker MachineSet in list
	baseMs := workerMachineSetList.Items[0]

	glog.V(100).Infof("Creating new MachineSet copy of first existing worker MachineSet: %v",
		baseMs.Name)

	copiedMachineSet := &machinev1beta1.MachineSet{
		ObjectMeta: *baseMs.ObjectMeta.DeepCopy(),
		Spec:       *baseMs.Spec.DeepCopy(),
	}

	glog.V(100).Infof("Renaming copied MachineSet to: %s", copiedMachineSet.ObjectMeta.Name)

	copiedMachineSet.ObjectMeta.Name = fmt.Sprintf("%v-%v", copiedMachineSet.Name,
		strings.ReplaceAll(instanceType, ".", "-"))

	glog.V(100).Infof("Updating copied MachineSet name in metadata, selector and template parameter ...")

	copiedMachineSet.ObjectMeta.UID = ""
	copiedMachineSet.ObjectMeta.ResourceVersion = ""

	// change spec labels
	copiedMachineSet.Spec.Selector.MatchLabels["machine.openshift.io/cluster-api-machineset"] =
		copiedMachineSet.ObjectMeta.Name
	copiedMachineSet.Spec.Template.ObjectMeta.Labels["machine.openshift.io/cluster-api-machineset"] =
		copiedMachineSet.ObjectMeta.Name

	glog.V(100).Infof("Updating copied MachineSet provider instanceType to: %s", instanceType)

	// currently only AWS public cloud supported when copying MachineSets
	err = ChangeAWSProviderInstanceType(copiedMachineSet, instanceType)
	if err != nil {
		return nil, fmt.Errorf("could not change the provider instanceType in copied MachineSet: %w", err)
	}

	glog.V(100).Infof("Updating copied MachineSet replicas value to: %v", replicas)
	copiedMachineSet.Spec.Replicas = &replicas

	return copiedMachineSet, nil
}

// Get returns MachineSet object if found.
func (builder *MachineSetBuilder) Get() (*machinev1beta1.MachineSet, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof(
		"Collecting MachineSet object %s", builder.Definition.Name)

	machineSet, err := builder.apiClient.MachineSets(builder.Definition.Namespace).Get(context.TODO(),
		builder.Definition.Name, metav1.GetOptions{})

	if err != nil {
		glog.V(100).Infof(
			"MachineSet object %s doesn't exist", builder.Definition.Name)

		return nil, err
	}

	return machineSet, err
}

// Exists checks whether the given MachineSet exists.
func (builder *MachineSetBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof(
		"Checking if MachineSet %s exists in namespace %s", builder.Definition.Name,
		builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	if err != nil {
		glog.V(100).Infof("Failed to collect MachineSet object due to %s", err.Error())
	}

	return err == nil || !k8serrors.IsNotFound(err)
}

// Create makes a MachineSet in cluster and stores the created object in struct.
func (builder *MachineSetBuilder) Create() (*MachineSetBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the MachineSet %s", builder.Definition.Name)

	var err error
	if !builder.Exists() {
		builder.Object, err = builder.apiClient.MachineSets(builder.Definition.Namespace).Create(
			context.TODO(), builder.Definition, metav1.CreateOptions{})
	}

	return builder, err
}

// Delete removes a MachineSet object from a cluster.
func (builder *MachineSetBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting the MachineSet object %s",
		builder.Definition.Name)

	if !builder.Exists() {
		return fmt.Errorf("MachineSet cannot be deleted because it does not exist")
	}

	err := builder.apiClient.MachineSets(builder.Object.Namespace).Delete(
		context.TODO(), builder.Object.Name, metav1.DeleteOptions{})

	if err != nil {
		return fmt.Errorf("cannot delete MachineSet: %w", err)
	}

	return err
}

// GetWorkerMachineSetList returns the list of worker MachineSets in a namespace on a cluster.
func GetWorkerMachineSetList(apiClient *clients.Settings, namespace string) (*machinev1beta1.MachineSetList, error) {
	list := &machinev1beta1.MachineSetList{
		Items: []machinev1beta1.MachineSet{},
	}

	resp, err := apiClient.MachineSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, ms := range resp.Items {
		if val, ok := ms.Spec.Template.ObjectMeta.Labels["machine.openshift.io/cluster-api-machine-role"]; ok &&
			val == "worker" {
			list.Items = append(list.Items, ms)
		}
	}

	return list, nil
}

// WaitForMachineSetReady waits until MachineSet first replica is Ready.
func WaitForMachineSetReady(apiClient *clients.Settings, namespace, machineSetName string, pollInterval,
	timeout time.Duration) error {
	return wait.PollImmediate(pollInterval, timeout, func() (bool, error) {
		machineSetPulled, err := PullMachineSet(apiClient, namespace, machineSetName)

		if err != nil {
			glog.V(100).Infof("MachineSet pull from cluster error: %s\n", err)

			return false, err
		}

		if machineSetPulled.Object.Status.ReadyReplicas == 1 {
			glog.V(100).Infof("MachineSet %s has now %v replicas in Ready state",
				machineSetPulled.Object.Name, machineSetPulled.Object.Status.ReadyReplicas)

			// this exits out of the wait.PollImmediate()
			return true, nil
		}

		glog.V(100).Infof("MachineSet %s has now %v replicas in Ready state",
			machineSetPulled.Object.Name, machineSetPulled.Object.Status.ReadyReplicas)

		return false, err
	})
}

// GetMapFromProviderSpec returns a map representation of the MachineSet ProviderSpec.Value element.
func GetMapFromProviderSpec(ms machinev1beta1.MachineSet) (map[string]interface{}, error) {
	providerSpecMap := make(map[string]interface{})

	// Value field is of type *runtime.RawExtension
	byteArray, err := json.Marshal(ms.Spec.Template.Spec.ProviderSpec.Value)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(byteArray, &providerSpecMap)
	if err != nil {
		return nil, err
	}

	return providerSpecMap, nil
}

// ChangeAWSProviderInstanceType updates the instanceType of an AWS MachineSet.
func ChangeAWSProviderInstanceType(machineSet *machinev1beta1.MachineSet, instanceType string) error {
	providerSpecMap, err := GetMapFromProviderSpec(*machineSet)
	if err != nil {
		return err
	}

	providerSpecMap["instanceType"] = instanceType
	byteArray, err := json.Marshal(providerSpecMap)

	if err != nil {
		return err
	}

	err = json.Unmarshal(byteArray, machineSet.Spec.Template.Spec.ProviderSpec.Value)

	if err != nil {
		return err
	}

	return nil
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *MachineSetBuilder) validate() (bool, error) {
	resourceCRD := "MachineSet"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Definition == nil {
		glog.V(100).Infof("The %s is undefined", resourceCRD)

		builder.errorMsg = msg.UndefinedCrdObjectErrString(resourceCRD)
	}

	if builder.apiClient == nil {
		glog.V(100).Infof("The %s builder apiClient is nil", resourceCRD)

		builder.errorMsg = fmt.Sprintf("%s builder cannot have nil apiClient", resourceCRD)
	}

	if builder.errorMsg != "" {
		glog.V(100).Infof("The %s builder has error message: %s", resourceCRD, builder.errorMsg)

		return false, fmt.Errorf(builder.errorMsg)
	}

	return true, nil
}
