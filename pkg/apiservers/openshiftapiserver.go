package apiservers

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	operatorV1 "github.com/openshift/api/operator/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var openshiftAPIServerObjName = "cluster"

// OpenshiftAPIServerBuilder provides struct for openshiftAPIServer object.
type OpenshiftAPIServerBuilder struct {
	// OpenshiftAPIServer definition. Used to create an openshiftAPIServer object.
	Definition *operatorV1.OpenShiftAPIServer
	// Created openshiftAPIServer object.
	Object *operatorV1.OpenShiftAPIServer
	// apiClient opens api connection to the cluster.
	apiClient goclient.Client
	// Used in functions that define or mutate openshiftAPIServer definition. errorMsg is processed before the
	// OpenshiftApiServer object is created.
	errorMsg string
}

// PullOpenshiftAPIServer pulls existing openshiftApiServer from the cluster.
func PullOpenshiftAPIServer(apiClient *clients.Settings) (*OpenshiftAPIServerBuilder, error) {
	glog.V(100).Infof("Pulling existing openshiftApiServer from cluster")

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")

		return nil, fmt.Errorf("openshiftApiServer 'apiClient' cannot be empty")
	}

	builder := OpenshiftAPIServerBuilder{
		apiClient: apiClient.Client,
		Definition: &operatorV1.OpenShiftAPIServer{
			ObjectMeta: metav1.ObjectMeta{
				Name: openshiftAPIServerObjName,
			},
		},
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("openshiftAPIServer object %s does not exist", openshiftAPIServerObjName)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Exists checks whether the given openshiftAPIServer exists.
func (builder *OpenshiftAPIServerBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	var err error
	builder.Object, err = builder.Get()

	if err != nil {
		glog.V(100).Infof("Failed to collect openshiftAPIServer object due to %s", err.Error())
	}

	return err == nil || !k8serrors.IsNotFound(err)
}

// Get returns openshiftAPIServer object if found.
func (builder *OpenshiftAPIServerBuilder) Get() (*operatorV1.OpenShiftAPIServer, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	openshiftAPIServer := &operatorV1.OpenShiftAPIServer{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name: builder.Definition.Name,
	}, openshiftAPIServer)

	if err != nil {
		glog.V(100).Infof("openshiftAPIServer object does not exist")

		return nil, err
	}

	return openshiftAPIServer, err
}

// GetCondition get specific openshiftAPIServer condition and message if presented.
func (builder *OpenshiftAPIServerBuilder) GetCondition(conditionType string) (
	*operatorV1.ConditionStatus, string, error) {
	if valid, err := builder.validate(); !valid {
		return nil, "", err
	}

	glog.V(100).Infof("Get %s openshiftAPIServer %s condition", builder.Definition.Name, conditionType)

	if conditionType == "" {
		return nil, "", fmt.Errorf("openshiftAPIServer 'conditionType' cannot be empty")
	}

	if !builder.Exists() {
		return nil, "", fmt.Errorf("%s openshiftAPIServer not found", builder.Definition.Name)
	}

	openshiftAPIServer, err := builder.Get()

	if err != nil {
		return nil, "", err
	}

	for _, condition := range openshiftAPIServer.Status.Conditions {
		if condition.Type == conditionType {
			return &condition.Status, condition.Reason, nil
		}
	}

	return nil, "", fmt.Errorf("the %s openshiftAPIServer %s condition not found",
		builder.Definition.Name, conditionType)
}

// WaitUntilConditionTrue waits for timeout duration or until openshiftAPIServer gets to a specific status.
func (builder *OpenshiftAPIServerBuilder) WaitUntilConditionTrue(
	conditionType string, timeout time.Duration) error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	if conditionType == "" {
		return fmt.Errorf("openshiftAPIServer 'conditionType' cannot be empty")
	}

	if !builder.Exists() {
		return fmt.Errorf("%s openshiftAPIServer not found", builder.Definition.Name)
	}

	var errMsg error

	err := wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			builder.Object, errMsg = builder.Get()

			if errMsg != nil {
				return false, nil
			}

			for _, condition := range builder.Object.Status.Conditions {
				if condition.Type == conditionType {
					if condition.Status == "True" {
						return true, nil
					}

					errMsg = fmt.Errorf("the %s condition did not reach True state yet", conditionType)

					return false, nil
				}
			}

			errMsg = fmt.Errorf("the %s condition not found exists", conditionType)

			return false, nil
		})

	if err != nil {
		return fmt.Errorf("%w: %w", errMsg, err)
	}

	return nil
}

// WaitAllPodsAtTheLatestGeneration waits for timeout duration or until openshiftAPIServer
// pods will reach the latest generation.
func (builder *OpenshiftAPIServerBuilder) WaitAllPodsAtTheLatestGeneration(timeout time.Duration) error {
	conditionType := "APIServerDeploymentProgressing"
	verificationStr := "AsExpected"

	err := builder.WaitUntilConditionTrue(conditionType, timeout)

	if err != nil {
		return err
	}

	err = wait.PollUntilContextTimeout(
		context.TODO(),
		time.Second,
		timeout,
		true,
		func(ctx context.Context) (bool, error) {
			var err error

			_, reasonMsg, err := builder.GetCondition(conditionType)

			if err != nil {
				return false, nil
			}

			glog.V(100).Infof("Found reason message: %s", reasonMsg)

			if reasonMsg != verificationStr {
				return false, nil
			}

			return true, nil
		})

	if err != nil {
		return err
	}

	return nil
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *OpenshiftAPIServerBuilder) validate() (bool, error) {
	resourceCRD := "OpenshiftAPIServer"

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

	if builder.errorMsg != "" {
		glog.V(100).Infof("The %s builder has error message: %s", resourceCRD, builder.errorMsg)

		return false, fmt.Errorf(builder.errorMsg)
	}

	return true, nil
}
