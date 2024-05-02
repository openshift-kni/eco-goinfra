package events

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	k8sv1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1Typed "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Builder provides struct for Event object which contains connection to cluster.
type Builder struct {
	// Dynamically discovered Event object.
	Object *k8sv1.Event
	// apiClient opens api connection to the cluster.
	apiClient corev1Typed.EventInterface
	// errorMsg used in discovery function before sending api request to cluster.
	errorMsg string
}

// Pull pulls existing Event from cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")

		return nil, fmt.Errorf("apiClient cannot be nil")
	}

	glog.V(100).Infof("Pulling existing Event name %s under namespace %s from cluster", name, nsname)

	builder := &Builder{
		apiClient: apiClient.Events(nsname),
		Object: &k8sv1.Event{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the Event is empty")

		return nil, fmt.Errorf("event 'name' cannot be empty")
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the Event is empty")

		return nil, fmt.Errorf("event 'nsname' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("event object %s does not exist in namespace %s", name, nsname)
	}

	return builder, nil
}

// Exists checks whether the given Event exists.
func (builder *Builder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if Event %s exists", builder.Object.Name)

	var err error
	builder.Object, err = builder.apiClient.Get(context.TODO(),
		builder.Object.Name, metaV1.GetOptions{})

	return err == nil || !k8serrors.IsNotFound(err)
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *Builder) validate() (bool, error) {
	resourceCRD := "Event"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
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
