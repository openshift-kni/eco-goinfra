package console

import (
	"context"
	"fmt"

	"k8s.io/utils/strings/slices"

	goclient "sigs.k8s.io/controller-runtime/pkg/client"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	operatorv1 "github.com/openshift/api/operator/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConsoleOperatorBuilder provides a struct for consoleOperator object from the cluster and a console definition.
type ConsoleOperatorBuilder struct {
	// ConsoleOperator definition, used to create the pod object.
	Definition *operatorv1.Console
	// Created consoleOperator object.
	Object *operatorv1.Console
	// api client to interact with the cluster.
	apiClient goclient.Client
	// errorMsg is processed before consoleOperator object is created.
	errorMsg string
}

// PullConsoleOperator loads an existing consoleOperator into the ConsoleOperatorBuilder struct.
func PullConsoleOperator(apiClient *clients.Settings, consoleOperatorName string) (*ConsoleOperatorBuilder, error) {
	glog.V(100).Infof("Pulling cluster consoleOperator %s", consoleOperatorName)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")

		return nil, fmt.Errorf("consoleOperator 'apiClient' cannot be empty")
	}

	builder := ConsoleOperatorBuilder{
		apiClient: apiClient.Client,
		Definition: &operatorv1.Console{
			ObjectMeta: metav1.ObjectMeta{
				Name: consoleOperatorName,
			},
		},
	}

	if consoleOperatorName == "" {
		glog.V(100).Info("The consoleOperatorName of the consoleOperator is empty")

		return nil, fmt.Errorf("the consoleOperator 'consoleOperatorName' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("the consoleOperator object %s does not exist", consoleOperatorName)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Get fetches existing consoleOperator from cluster.
func (builder *ConsoleOperatorBuilder) Get() (*operatorv1.Console, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting existing consoleOperator with name %s from cluster", builder.Definition.Name)

	consoleOperator := &operatorv1.Console{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name: builder.Definition.Name,
	}, consoleOperator)

	if err != nil {
		glog.V(100).Infof("Failed to get consoleOperator object %s from cluster due to: %v",
			builder.Definition.Name, err)

		return nil, err
	}

	return consoleOperator, nil
}

// Exists checks whether the given consoleOperator exists.
func (builder *ConsoleOperatorBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if consoleOperator %s exists", builder.Definition.Name)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Update renovates the existing cluster consoleOperator object with cluster consoleOperator definition in builder.
func (builder *ConsoleOperatorBuilder) Update() (*ConsoleOperatorBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Info("Updating cluster consoleOperator %s", builder.Definition.Name)

	err := builder.apiClient.Update(context.TODO(), builder.Definition)
	if err == nil {
		builder.Object = builder.Definition
	}

	return builder, err
}

// GetPlugins fetches consoleOperator plugins list.
func (builder *ConsoleOperatorBuilder) GetPlugins() (*[]string, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting consoleOperator plugins list configuration")

	if !builder.Exists() {
		return nil, fmt.Errorf("consoleOperator %s object does not exist", builder.Definition.Name)
	}

	return &builder.Object.Spec.Plugins, nil
}

// WithPlugins adds to the consoleOperator operator's new plugins.
func (builder *ConsoleOperatorBuilder) WithPlugins(newPluginsList []string, redefine bool) *ConsoleOperatorBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting consoleOperator %s with new plugins: %v",
		builder.Definition.Name, newPluginsList)

	if len(newPluginsList) == 0 {
		glog.V(100).Infof("the newPluginsList can not be empty")

		builder.errorMsg = "the newPluginsList can not be empty"

		return builder
	}

	if builder.Definition.Spec.Plugins == nil {
		glog.V(100).Infof("Plugins are nil. Initializing one")

		builder.Definition.Spec.Plugins = []string{}
	}

	if redefine {
		glog.V(100).Infof("Redefining existing plugins list with %v", newPluginsList)

		builder.Definition.Spec.Plugins = newPluginsList
	} else {
		glog.V(100).Infof("Existing plugins list will not be redefined")

		for _, newPlugin := range newPluginsList {
			if slices.Contains(builder.Definition.Spec.Plugins, newPlugin) {
				glog.V(100).Infof("the newPlugin %s was already defined in Plugins list %v",
					newPlugin, builder.Definition.Spec.Plugins)
			} else {
				builder.Definition.Spec.Plugins = append(builder.Definition.Spec.Plugins, newPlugin)
			}
		}
	}

	return builder
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *ConsoleOperatorBuilder) validate() (bool, error) {
	resourceCRD := "Console.Operator"

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
