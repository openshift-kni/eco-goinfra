package argocd

import (
	"context"
	"fmt"

	argocdoperatorv1alpha1 "github.com/argoproj-labs/argocd-operator/api/v1alpha1"
	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Builder provides struct for the argocd object containing connection to
// the cluster and the argocd definitions.
type Builder struct {
	// argocd Definition, used to create the argocd object.
	Definition *argocdoperatorv1alpha1.ArgoCD
	// created argocd object.
	Object *argocdoperatorv1alpha1.ArgoCD
	// api client to interact with the cluster.
	apiClient goclient.Client
	// used to store latest error message upon defining the argocd definition.
	errorMsg string
}

// NewBuilder creates a new instance of Builder.
func NewBuilder(apiClient *clients.Settings, name, nsname string) *Builder {
	builder := Builder{
		apiClient: apiClient.Client,
		Definition: &argocdoperatorv1alpha1.ArgoCD{
			Spec: argocdoperatorv1alpha1.ArgoCDSpec{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the argocd is empty")

		builder.errorMsg = "argocd 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the argocd is empty")

		builder.errorMsg = "argocd 'nsname' cannot be empty"
	}

	return &builder
}

// Pull pulls existing argocd from cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	glog.V(100).Infof("Pulling existing argocd name %s under namespace %s from cluster", name, nsname)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")

		return nil, fmt.Errorf("argocd 'apiClient' cannot be empty")
	}

	builder := Builder{
		apiClient: apiClient.Client,
		Definition: &argocdoperatorv1alpha1.ArgoCD{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the argocd is empty")

		return nil, fmt.Errorf("argocd 'name' cannot be empty")
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the argocd is empty")

		return nil, fmt.Errorf("argocd 'namespace' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("argocd object %s does not exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Exists checks whether the given argocd exists.
func (builder *Builder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if argocd %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Get returns argocd object if found.
func (builder *Builder) Get() (*argocdoperatorv1alpha1.ArgoCD, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting argocd %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	argocd := &argocdoperatorv1alpha1.ArgoCD{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, argocd)

	if err != nil {
		return nil, err
	}

	return argocd, err
}

// Create makes an argocd in the cluster and stores the created object in struct.
func (builder *Builder) Create() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the argocd %s in namespace %s",
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

// Delete removes argocd from a cluster.
func (builder *Builder) Delete() (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting the argocd %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		builder.Object = nil

		glog.V(100).Infof("argocd %s in namespace %s cannot be deleted because it does not exist",
			builder.Definition.Name, builder.Definition.Namespace)

		return builder, nil
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return builder, fmt.Errorf("can not delete argocd: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// Update renovates the existing argocd object with the argocd definition in builder.
func (builder *Builder) Update(force bool) (*Builder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating the argocd object", builder.Definition.Name)

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		if force {
			glog.V(100).Infof(
				msg.FailToUpdateNotification("argocd", builder.Definition.Name))

			builder, err := builder.Delete()

			if err != nil {
				glog.V(100).Infof(
					msg.FailToUpdateError("argocd", builder.Definition.Name))

				return nil, err
			}

			return builder.Create()
		}
	}

	builder.Object = builder.Definition

	return builder, err
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *Builder) validate() (bool, error) {
	resourceCRD := "argocds"

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
