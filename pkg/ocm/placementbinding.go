package ocm

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	policiesv1 "open-cluster-management.io/governance-policy-propagator/api/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// PlacementBindingBuilder type definition.
type PlacementBindingBuilder struct {
	// placementBinding Definition, used to create the placementBinding object.
	Definition *policiesv1.PlacementBinding
	// created placementBinding object.
	Object *policiesv1.PlacementBinding
	// api client to interact with the cluster.
	apiClient *clients.Settings
	// used to store latest error message upon defining or mutating placementBinding definition.
	errorMsg string
}

// NewPlacementBindingBuilder creates a new instance of PlacementBindingBuilder.
func NewPlacementBindingBuilder(
	apiClient *clients.Settings,
	name,
	nsname string,
	placementRef policiesv1.PlacementSubject,
	subject policiesv1.Subject) *PlacementBindingBuilder {
	glog.V(100).Infof(
		"Initializing new placement binding structure with the following params: name: %s, nsname: %s",
		name, nsname)

	builder := PlacementBindingBuilder{
		apiClient: apiClient,
		Definition: &policiesv1.PlacementBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
			PlacementRef: placementRef,
			Subjects:     []policiesv1.Subject{subject},
		},
	}

	if name == "" {
		builder.errorMsg = "placementBinding's 'name' cannot be empty"
	}

	if nsname == "" {
		builder.errorMsg = "placementBinding's 'nsname' cannot be empty"
	}

	if placementRefErr := validatePlacementRef(placementRef); placementRefErr != "" {
		builder.errorMsg = placementRefErr
	}

	if subjectErr := validateSubject(subject); subjectErr != "" {
		builder.errorMsg = subjectErr
	}

	return &builder
}

// PullPlacementBinding pulls existing placementBinding into Builder struct.
func PullPlacementBinding(apiClient *clients.Settings, name, nsname string) (*PlacementBindingBuilder, error) {
	glog.V(100).Infof("Pulling existing placementBinding name %s under namespace %s from cluster", name, nsname)

	if apiClient == nil {
		glog.V(100).Info("The apiClient is empty")

		return nil, fmt.Errorf("placementBinding's 'apiClient' cannot be empty")
	}

	builder := PlacementBindingBuilder{
		apiClient: apiClient,
		Definition: &policiesv1.PlacementBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the placementBinding is empty")

		return nil, fmt.Errorf("placementBinding's 'name' cannot be empty")
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the placementBinding is empty")

		return nil, fmt.Errorf("placementBinding's 'namespace' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("placementBinding object %s does not exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Exists checks whether the given placementBinding exists.
func (builder *PlacementBindingBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if placementBinding %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Get returns a placementBinding object if found.
func (builder *PlacementBindingBuilder) Get() (*policiesv1.PlacementBinding, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting placementBinding %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	placementBinding := &policiesv1.PlacementBinding{}

	err := builder.apiClient.Get(context.TODO(), runtimeclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, placementBinding)

	if err != nil {
		return nil, err
	}

	return placementBinding, err
}

// Create makes a placementBinding in the cluster and stores the created object in struct.
func (builder *PlacementBindingBuilder) Create() (*PlacementBindingBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the placementBinding %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	if !builder.Exists() {
		err = builder.apiClient.Create(context.TODO(), builder.Definition)
		if err == nil {
			builder.Object = builder.Definition
		}
	}

	return builder, err
}

// Delete removes a placementBinding from a cluster.
func (builder *PlacementBindingBuilder) Delete() (*PlacementBindingBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting the placementBinding %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return builder, fmt.Errorf("placementBinding cannot be deleted because it does not exist")
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return builder, fmt.Errorf("can not delete placementBinding: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// Update renovates the existing placementBinding object with the placementBinding definition in builder.
func (builder *PlacementBindingBuilder) Update(force bool) (*PlacementBindingBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating the placementBinding object: %s in namespace: %s",
		builder.Definition.Name, builder.Definition.Namespace)

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		if force {
			glog.V(100).Infof(
				"Failed to update the placementBinding object %s. "+
					"Note: Force flag set, executed delete/create methods instead", builder.Definition.Name)

			builder, err := builder.Delete()

			if err != nil {
				glog.V(100).Infof(
					"Failed to update the placementBinding object %s, "+
						"due to error in delete function", builder.Definition.Name)

				return nil, err
			}

			return builder.Create()
		}
	}

	if err == nil {
		builder.Object = builder.Definition
	}

	return builder, err
}

// WithAdditionalSubject appends a subject to the subjects list in the PlacementBinding definition.
func (builder *PlacementBindingBuilder) WithAdditionalSubject(subject policiesv1.Subject) *PlacementBindingBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Adding Subject %s to PlacementBinding %s", subject.Name, builder.Definition.Name)

	if err := validateSubject(subject); err != "" {
		builder.errorMsg = err

		return builder
	}

	builder.Definition.Subjects = append(builder.Definition.Subjects, subject)

	return builder
}

// validatePlacementRef validates all the fields of the PlacementRef and returns an errorMsg based on the validation.
// The errorMsg will be empty for valid Subjects.
func validatePlacementRef(placementRef policiesv1.PlacementSubject) string {
	apiGroup := placementRef.APIGroup
	if apiGroup != "apps.open-cluster-management.io" && apiGroup != "cluster.open-cluster-management.io" {
		glog.V(100).Info("The APIGroup of the PlacementRef of the PlacementBinding is invalid")

		return "placementBinding's 'PlacementRef.APIGroup' must be a valid option"
	}

	kind := placementRef.Kind
	if kind != "PlacementRule" && kind != "Placement" {
		glog.V(100).Info("The Kind of the PlacementRef of the PlacementBinding is invalid")

		return "placementBinding's 'PlacementRef.Kind' must be a valid option"
	}

	if placementRef.Name == "" {
		glog.V(100).Info("The Name of the PlacementRef of the PlacementBinding is empty")

		return "placementBinding's 'PlacementRef.Name' cannot be empty"
	}

	return ""
}

// validateSubject validates the fields of the Subject and returns an errorMsg based on the validation. The errorMsg
// will be empty for valid Subjects.
func validateSubject(subject policiesv1.Subject) string {
	if subject.APIGroup != "policy.open-cluster-management.io" {
		glog.V(100).Info("The APIGroup of the PlacementBinding subject is invalid")

		return "placementBinding's 'Subject.APIGroup' must be 'policy.open-cluster-management.io'"
	}

	if subject.Kind != "Policy" && subject.Kind != "PolicySet" {
		glog.V(100).Info("The Kind of the subject of the PlacementBinding is invalid")

		return "placementBinding's 'Subject.Kind' must be a valid option"
	}

	if subject.Name == "" {
		glog.V(100).Info("The Name of the subject of the PlacementBinding is empty")

		return "placementBinding's 'Subject.Name' cannot be empty"
	}

	return ""
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *PlacementBindingBuilder) validate() (bool, error) {
	resourceCRD := "PlacementBinding"

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
