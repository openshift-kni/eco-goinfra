package machine

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListWorkerMachineSets returns a slice of SetBuilder objects in a namespace on a cluster.
func ListWorkerMachineSets(
	apiClient *clients.Settings,
	namespace string,
	workerLabel string) ([]*SetBuilder, error) {
	if namespace == "" {
		glog.V(100).Infof("machineSet 'namespace' parameter can not be empty")

		return nil, fmt.Errorf("failed to list MachineSets, 'namespace' parameter is empty")
	}

	if workerLabel == "" {
		glog.V(100).Infof("machineSet 'workerLabel' parameter can not be empty")

		return nil, fmt.Errorf("failed to list MachineSets, 'workerLabel' parameter is empty")
	}

	machineSetList, err := apiClient.MachineSets(namespace).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		glog.V(100).Infof("Failed to list MachineSets in the namespace %s due to %s",
			namespace, err.Error())

		return nil, err
	}

	var machineSetObjects []*SetBuilder

	for _, runningMachineSet := range machineSetList.Items {
		copiedMachineSet := runningMachineSet
		SetBuilder := &SetBuilder{
			apiClient:  apiClient,
			Object:     &copiedMachineSet,
			Definition: &copiedMachineSet,
		}

		if val, ok := SetBuilder.Definition.Spec.Template.ObjectMeta.Labels[workerLabel]; ok && val == "worker" {
			machineSetObjects = append(machineSetObjects, SetBuilder)
		}
	}

	return machineSetObjects, nil
}
