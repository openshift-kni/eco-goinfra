package clusteroperator

import (
	"context"
	"time"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fmt"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	v1 "github.com/openshift/api/config/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	fiveScds time.Duration = 5 * time.Second
	isTrue                 = "True"
)

// ClusterOperatorBuilder provides struct for ClusterOperator object.
type ClusterOperatorBuilder struct {
	// ClusterOperator definition. Used to create a clusteroperator object.
	Definition *v1.ClusterOperator
	// Created clusteroperator object.
	Object *v1.ClusterOperator
	// apiClient opens api connection to the cluster.
	apiClient *clients.Settings
	// Used in functions that define or mutate ClusterOperator definition. errorMsg is processed before the
	// ClusterOperator object is created.
	errorMsg string
}

// Exists checks whether the given ClusterOperator exists.
func (builder *ClusterOperatorBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if clusteroperator %s exists", builder.Definition.Name)

	_, err := builder.apiClient.ClusterOperators().Get(
		context.Background(),
		builder.Definition.Name,
		metaV1.GetOptions{})

	return err == nil || !k8serrors.IsNotFound(err)
}

// IsAvailable check if the ClusterOperator is Available.
func (builder *ClusterOperatorBuilder) IsAvailable() bool {
	glog.V(100).Infof("Verify %s clusteroperator availability", builder.Definition.Name)

	if !builder.Exists() {
		return false
	}

	for _, condition := range builder.Object.Status.Conditions {
		if condition.Type == "Available" {
			return condition.Status == isTrue
		}
	}

	return false
}

// IsDegraded check if the ClusterOperator is Degraded.
func (builder *ClusterOperatorBuilder) IsDegraded() bool {
	glog.V(100).Infof("Verify %s clusteroperator is degraded", builder.Definition.Name)

	if !builder.Exists() {
		return false
	}

	for _, condition := range builder.Object.Status.Conditions {
		if condition.Type == "Degraded" {
			return condition.Status == isTrue
		}
	}

	return false
}

// IsProgressing check if the ClusterOperator is Degraded.
func (builder *ClusterOperatorBuilder) IsProgressing() bool {
	glog.V(100).Infof("Verify %s clusteroperator is progressing", builder.Definition.Name)

	if !builder.Exists() {
		return false
	}

	for _, condition := range builder.Object.Status.Conditions {
		if condition.Type == "Progressing" {
			return condition.Status == isTrue
		}
	}

	return false
}

// WaitUntilAvailable waits for timeout duration or until ClusterOperator is Available.
func (builder *ClusterOperatorBuilder) WaitUntilAvailable(timeout time.Duration) error {
	return builder.WaitUntilInStatus("Available", timeout)
}

// WaitUntilProgressing waits for timeout duration or until ClusterOperator is Progressing.
func (builder *ClusterOperatorBuilder) WaitUntilProgressing(timeout time.Duration) error {
	return builder.WaitUntilInStatus("Progressing", timeout)
}

// WaitUntilInStatus waits for timeout duration or until ClusterOperator gets to a specific status.
func (builder *ClusterOperatorBuilder) WaitUntilInStatus(
	status v1.ClusterStatusConditionType, timeout time.Duration) error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	return wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		_, err := builder.apiClient.ClusterOperators().Get(
			context.Background(),
			builder.Definition.Name,
			metaV1.GetOptions{})

		if err != nil {
			return false, nil
		}

		for _, condition := range builder.Object.Status.Conditions {
			if condition.Type == status {
				return condition.Status == "True", nil
			}
		}

		return false, err
	})
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *ClusterOperatorBuilder) validate() (bool, error) {
	resourceCRD := "ClusterOperator"

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
