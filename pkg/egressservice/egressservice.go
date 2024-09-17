package egressservice

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	"k8s.io/apimachinery/pkg/util/wait"

	egresssvcv1 "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/egressservice/v1"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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

func Pull(apiClient *clients.Settings, name, nsname string) (*EgressServiceBuilder, error) {
	glog.V(100).Infof("Pulling existing EgressService %q in namespace %q from cluster",
		name, nsname)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")

		return nil, fmt.Errorf("EgressService's 'apiClient' cannot be empty")
	}

	err := apiClient.AttachScheme(egresssvcv1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add EgressService scheme to client schemes")

		return nil, err
	}

	if name == "" {
		glog.V(100).Infof("EgressService's name cannot be empty")

		return nil, fmt.Errorf("EgressService's name cannot be empty")
	}

	if nsname == "" {
		glog.V(100).Infof("EgressService's namespace cannot be empty")

		return nil, fmt.Errorf("EgressService's namespace cannot be empty")
	}

	builder := &EgressServiceBuilder{
		apiClient: apiClient.Client,
		Definition: &egresssvcv1.EgressService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("EgressService object %q does not exist in namespace %q",
			name, nsname)
	}

	builder.Definition = builder.Object

	return builder, nil
}

// Exists checks whether the given EgressService exists.
func (builder *EgressServiceBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if EgressService %q exists in namespace %q",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Get fetches the EgressService from the cluster.
func (builder *EgressServiceBuilder) Get() (*egresssvcv1.EgressService, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting EgressService %q in namespace %q",
		builder.Definition.Name, builder.Definition.Namespace)

	egrSvc := &egresssvcv1.EgressService{}

	err := wait.PollUntilContextTimeout(
		context.TODO(), 5*time.Second, 1*time.Minute, true, func(ctx context.Context) (bool, error) {

			err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
				Name:      builder.Definition.Name,
				Namespace: builder.Definition.Namespace,
			}, egrSvc)

			if err != nil && k8serrors.IsNotFound(err) {
				glog.V(100).Infof("EgressService not found: %v", err)

				return false, err
			}

			if err != nil && !k8serrors.IsNotFound(err) {
				glog.V(100).Infof("Error retrieving EgressService: %v", err)

				return false, nil
			}

			return err == nil && egrSvc != nil, err
		})

	return egrSvc, err
}

// Create makes a EgressService in the cluster and stores the created object in struct.
func (builder *EgressServiceBuilder) Create() (*EgressServiceBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the EgressServcice %q in namespace %q",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error

	if !builder.Exists() {
		err = wait.PollUntilContextTimeout(
			context.TODO(), 5*time.Second, 1*time.Minute, true, func(ctx context.Context) (bool, error) {
				err = builder.apiClient.Create(context.TODO(), builder.Definition)

				if err == nil {
					glog.V(100).Infof("Created EgressServcice %q in namespace %q",
						builder.Definition.Name, builder.Definition.Namespace)

					builder.Object = builder.Definition

					return true, nil
				}

				glog.V(100).Infof("Error creating EgressServcice: %v", err)

				return false, nil
			})
	}

	return builder, err
}

// Delete removes EgressService from a cluster.
func (builder *EgressServiceBuilder) Delete() (*EgressServiceBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting EgressService %q in namespace %q",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		glog.V(100).Infof("EgressService %q in namespace %q does not exist",
			builder.Definition.Name, builder.Definition.Namespace)

		builder.Object = nil

		return builder, nil
	}

	err := wait.PollUntilContextTimeout(
		context.TODO(), 5*time.Second, 1*time.Minute, true, func(ctx context.Context) (bool, error) {

			err := builder.apiClient.Delete(context.TODO(), builder.Definition)

			if err != nil {
				glog.V(100).Infof("Error deleting EgressService: %v", err)

				return false, nil
			}

			return true, err
		})

	if err != nil {
		glog.V(100).Infof("Error deleting EgressService: %v", err)

		return builder, fmt.Errorf("failed to delete EgressService due to %v", err)
	}

	builder.Object = nil

	return builder, nil
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
