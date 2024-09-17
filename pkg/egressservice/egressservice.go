package egressservice

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	// egresssvcv1 "github.com/kedacore/keda-olm-operator/apis/keda/v1alpha1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	egresssvcv1 "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/egressservice/v1"

	// k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// EgressServiceBuilder provides a struct for EgressService object.
type EgressServiceBuilder struct {
	// EgressService definition, used to create the EgressService object.
	Definition *egresssvcv1.EgressService
	// Created EgressService object.
	Object *egresssvcv1.EgressService
	// Used to store latest error message upon defining or mutating EgressService definition.
	errorMsg string
	// api client to interact with the cluster.
	apiClient goclient.Client
}

func NewEgressServiceBuilder(
	apiClient *clients.Settings, nsname, name, sourceIPBy string) *EgressServiceBuilder {

	glog.V(100).Infof(
		"Initializing new EgressService structure with the following params: "+
			"name: %s; namespace: %s; sourceIPBy: %s",
		name, nsname, sourceIPBy)

	if apiClient == nil {
		glog.V(100).Infof("EgressService 'apiClient' cannot be empty")

		return nil
	}

	err := apiClient.AttachScheme(egresssvcv1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add 'egressservice' scheme to client schemes")

		return nil
	}

	builder := &EgressServiceBuilder{
		apiClient: apiClient.Client,
		Definition: &egresssvcv1.EgressService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
			Spec: egresssvcv1.EgressServiceSpec{
				SourceIPBy: egresssvcv1.SourceIPMode(sourceIPBy),
			},
		},
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the EgressService is empty")

		builder.errorMsg = "The namespace of the EgressService is empty"
	}

	if name == "" {
		glog.V(100).Infof("The name parameter of the EgressService is empty")

		builder.errorMsg = "The name parameter of the EgressService is empty"
	}

	if _, err := validateSourceIPMode(sourceIPBy); err != nil {
		glog.V(100).Infof("Invalid sourceIPBy parameter for the EgressService")

		builder.errorMsg = "Invalid sourceIPBy parameter for the EgressService"
	}

	return builder
}

// WithNodeLabelSelector applies nodeSelector to the EgressService definition,
// which uses key:value pairs for nodes matching.
func (builder *EgressServiceBuilder) WithNodeLabelSelector(selector map[string]string) *EgressServiceBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Applying nodeSelector %s to EgressService %q in namespace %q",
		selector, builder.Definition.Name, builder.Definition.Namespace)

	if len(selector) == 0 {
		glog.V(100).Infof("The nodeselector is empty")

		builder.errorMsg = "cannot accept empty map as nodeSelector"

		return builder
	}

	builder.Definition.Spec.NodeSelector = metav1.LabelSelector{MatchLabels: selector}

	return builder
}

// WithVRFNetwork sets the network to be used for sending egress and corresponding ingress replies to.
func (builder *EgressServiceBuilder) WithVRFNetwork(vrfnet string) *EgressServiceBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting  VRF network %q to EgressService %q in namespace %q",
		vrfnet, builder.Definition.Name, builder.Definition.Namespace)

	if vrfnet == "" {
		glog.V(100).Infof("Cannot use emtpy VRF network")

		builder.errorMsg = "Cannot use emtpy VRF network"
	}

	builder.Definition.Spec.Network = vrfnet

	return builder
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *EgressServiceBuilder) validate() (bool, error) {
	resourceCRD := "EgressService"

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

func validateSourceIPMode(sourceIPMode string) (string, error) {
	glog.V(100).Infof("Validating requested sourceIPMode %q", sourceIPMode)

	switch {
	case strings.EqualFold(sourceIPMode, ""):
		glog.V(100).Infof("Invalid sourceIPMode %q", sourceIPMode)

		return "", fmt.Errorf("invalid empty sourceIPMode %q", sourceIPMode)
	case strings.EqualFold(sourceIPMode, string(egresssvcv1.SourceIPLoadBalancer)):
		glog.V(100).Infof("Valid sourceIPMode %q", sourceIPMode)

		return string(egresssvcv1.SourceIPLoadBalancer), nil
	case strings.EqualFold(sourceIPMode, string(egresssvcv1.SourceIPNetwork)):
		glog.V(100).Infof("Valid sourceIPMode %q", sourceIPMode)

		return string(egresssvcv1.SourceIPNetwork), nil
	default:
		glog.V(100).Infof("Invalid sourceIPMode %q", sourceIPMode)

		return "", fmt.Errorf("invalid sourceIPMode %s", sourceIPMode)
	}
}
