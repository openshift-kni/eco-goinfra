package clusteroperator

import (
	"context"
	"time"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// List returns clusterOperators inventory.
func List(apiClient *clients.Settings) ([]*ClusterOperatorBuilder, error) {
	glog.V(100).Info("Listing all clusterOperators")

	coList, err := apiClient.ClusterOperators().List(context.Background(), metaV1.ListOptions{})

	if err != nil {
		glog.V(100).Infof("Failed to list clusterOperators due to %s", err.Error())

		return nil, err
	}

	var coObjects []*ClusterOperatorBuilder

	for _, clusterOperator := range coList.Items {
		copiedCo := clusterOperator
		coBuilder := &ClusterOperatorBuilder{
			apiClient:  apiClient,
			Object:     &copiedCo,
			Definition: &copiedCo,
		}

		coObjects = append(coObjects, coBuilder)
	}

	return coObjects, nil
}

// WaitForAllClusteroperatorsAvailable check that all clusterOperators are in available state.
func WaitForAllClusteroperatorsAvailable(apiClient *clients.Settings, timeout time.Duration) (bool, error) {
	glog.V(100).Info("Waiting for all clusterOperators to be in available state")

	coList, err := List(apiClient)
	if err != nil {
		glog.V(100).Infof("Failed to list all clusterOperators due to %s", err.Error())

		return false, err
	}

	// Wait 5 secs in each iteration before condition function () returns true or errors or times out
	// after availableDuration
	err = wait.PollImmediate(fiveScds, timeout, func() (bool, error) {

		// iterate through the clusterOperators in the list.
		for _, clusteroperator := range coList {
			if !clusteroperator.IsAvailable() {
				glog.V(100).Infof("The %s clusterOperator is not available",
					clusteroperator.Object.Name)

				return false, nil
			}
		}

		return true, nil
	})

	if err == nil {
		glog.V(100).Infof("All clusterOperators were found available during timeout: %v",
			timeout)

		return true, nil
	}

	// Here err is "timed out waiting for the condition"
	glog.V(100).Infof("Not all clusterOperators were found available during timeout: %v",
		timeout)

	return false, err
}

// WaitForAllClusteroperatorsStopProgressing check that all pods in namespace that match options are in running state.
func WaitForAllClusteroperatorsStopProgressing(apiClient *clients.Settings, timeout time.Duration) (bool, error) {
	glog.V(100).Infof("Waiting for all clusteroperators stop processing")

	coList, err := List(apiClient)
	if err != nil {
		glog.V(100).Infof("Failed to list all clusterOperators due to %s", err.Error())

		return false, err
	}

	// Wait 5 secs in each iteration before condition function () returns true or errors or times out
	// after availableDuration
	err = wait.PollImmediate(fiveScds, timeout, func() (bool, error) {

		// iterate through the clusterOperators in the list.
		for _, clusteroperator := range coList {
			if clusteroperator.IsProgressing() {
				glog.V(100).Infof("The %s clusterOperator is still progressing",
					clusteroperator.Object.Name)

				return false, nil
			}
		}

		return true, nil
	})

	if err == nil {
		glog.V(100).Infof("All clusterOperators stopped progressing during timeout: %v",
			timeout)

		return true, nil
	}

	// Here err is "timed out waiting for the condition"
	glog.V(100).Infof("Not all clusterOperators stopped progressing during timeout: %v",
		timeout)

	return false, err
}
