package servicemesh

import (
	"context"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	istioV1 "maistra.io/api/core/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// MemberRollBuilder provides a struct for serviceMeshMemberRoll object from the cluster and
// a serviceMeshMemberRoll definition.
type MemberRollBuilder struct {
	// serviceMeshMemberRoll definition, used to create the serviceMeshMemberRoll object.
	Definition *istioV1.ServiceMeshMemberRoll
	// Created serviceMeshMemberRoll object.
	Object *istioV1.ServiceMeshMemberRoll
	// Used in functions that define or mutate serviceMeshMemberRoll definition. errorMsg is processed
	// before the serviceMeshMemberRoll object is created.
	errorMsg string
	// api client to interact with the cluster.
	apiClient *clients.Settings
}

// NewMemberRollBuilder method creates new instance of builder.
func NewMemberRollBuilder(apiClient *clients.Settings, name, nsname string) *MemberRollBuilder {
	glog.V(100).Infof("Initializing new serviceMeshMemberRollBuilder structure with the following "+
		"params: name: %s, namespace: %s", name, nsname)

	builder := &MemberRollBuilder{
		apiClient: apiClient,
		Definition: &istioV1.ServiceMeshMemberRoll{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the serviceMeshMemberRoll is empty")

		builder.errorMsg = "The serviceMeshMemberRoll 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the serviceMeshMemberRoll is empty")

		builder.errorMsg = "The serviceMeshMemberRoll 'namespace' cannot be empty"
	}

	return builder
}

// PullMemberRole retrieves an existing serviceMeshMemberRoll object from the cluster.
func PullMemberRole(apiClient *clients.Settings, name, nsname string) (*MemberRollBuilder, error) {
	glog.V(100).Infof(
		"Pulling serviceMeshMemberRoll object name: %s in namespace: %s", name, nsname)

	builder := MemberRollBuilder{
		apiClient: apiClient,
		Definition: &istioV1.ServiceMeshMemberRoll{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the serviceMeshMemberRoll is empty")

		builder.errorMsg = "serviceMeshMemberRoll 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the serviceMeshMemberRoll is empty")

		builder.errorMsg = "serviceMeshMemberRoll 'nsname' cannot be empty"
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("serviceMeshMemberRoll object %s does not exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Get fetches existing serviceMeshMemberRoll from cluster.
func (builder *MemberRollBuilder) Get() (*istioV1.ServiceMeshMemberRoll, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Fetching existing serviceMeshMemberRoll with name %s under namespace %s from cluster",
		builder.Definition.Name, builder.Definition.Namespace)

	servicemeshmemberroll := &istioV1.ServiceMeshMemberRoll{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, servicemeshmemberroll)

	if err != nil {
		return nil, err
	}

	return servicemeshmemberroll, nil
}

// Create makes a serviceMeshMemberRoll in the cluster and stores the created object in struct.
func (builder *MemberRollBuilder) Create() (*MemberRollBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the serviceMeshMemberRoll %s in namespace %s",
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

// Delete removes serviceMeshMemberRoll from a cluster.
func (builder *MemberRollBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting the serviceMeshMemberRoll %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return fmt.Errorf("serviceMeshMemberRoll %s in namespace %s cannot be deleted because it does not exist",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return fmt.Errorf("can not delete serviceMeshMemberRoll %s in namespace %s due to %w",
			builder.Definition.Name, builder.Definition.Namespace, err)
	}

	builder.Object = nil

	return nil
}

// Update renovates the existing serviceMeshMemberRoll object with serviceMeshMemberRoll definition in builder.
func (builder *MemberRollBuilder) Update(force bool) (*MemberRollBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Info("Updating serviceMeshMemberRoll %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		if force {
			glog.V(100).Infof(
				msg.FailToUpdateNotification("serviceMeshMemberRoll", builder.Definition.Name, builder.Definition.Namespace))

			err := builder.Delete()

			if err != nil {
				glog.V(100).Infof(
					msg.FailToUpdateError("serviceMeshMemberRoll", builder.Definition.Name, builder.Definition.Namespace))

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

// Exists checks whether the given serviceMeshMemberRoll exists.
func (builder *MemberRollBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if serviceMeshMemberRoll %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// WithMembersList adds member list section to the ServiceMeshMemberRoleBuilder.
func (builder *MemberRollBuilder) WithMembersList(membersList []string) *MemberRollBuilder {
	glog.V(100).Infof("Adding member list %v section to the MemberRollBuilder", membersList)

	if len(membersList) == 0 {
		glog.V(100).Infof("Cannot add empty membersList to the memberRoll structure")

		builder.errorMsg = "can not modify memberRoll config with empty membersList"
	}

	if builder.errorMsg != "" {
		return builder
	}

	if builder.Definition.Spec.Members == nil {
		builder.Definition.Spec.Members = membersList
	} else {
		builder.Definition.Spec.Members = append(builder.Definition.Spec.Members, membersList...)
	}

	return builder
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *MemberRollBuilder) validate() (bool, error) {
	resourceCRD := "ServiceMeshMemberRoll"

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
