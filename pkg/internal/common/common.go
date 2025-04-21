// Package common provides the Builder interface and functions that consume it.
//
// In my approach, I decided that the purpose of the interface should only be internal. In practice, I don't think
// there's much value for consumers of eco-goinfra to write code that abstracts over different types of builders. Its
// value is more for us to write implementations of common methods using attributes about builders. This also solves the
// problem of not every function having the same signature throughout the repo (such as Update, where only some take a
// force parameter).
//
// For new resources, I think going with an embedded struct like Trey suggested is a good idea so we can ensure better
// standardization. It's easy enough to make the embedded struct generic over the resource types and would allow us to
// have methods already implemented. I originally went with this approach, but the conflicting signatures would be an
// issue for providing CRUD methods for existing resources.
//
// I decided then to use an approach similar to Brandon's where the common package also includes the common methods that
// take the builder as an interface. This would allow us to have very standard signatures with the ability to mix and
// match methods just by wrapping them in individual packages. We could piecemeal replace implementations and add any
// necessary changes to ensure API compatibility.
//
//nolint:revive,staticcheck // annoying comment warning, just for PoC
package common

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// objectPointer is a type constraint that requires the type to be a pointer to a T and implements the
// runtimeclient.Object interface. This is necessary to give access to the concrete type itself while also ensuring its
// pointer implements the interface.
type objectPointer[T any] interface {
	*T
	runtimeclient.Object
}

// Even though we only use ST in the interface methods, we need T to be available to functions so that we can create new
// structs. new(T) returns *T but new(ST) returns *ST which does not satisfy runtimeclient.Object.
type Builder[T any, ST objectPointer[T]] interface {
	GetDefinition() ST
	SetDefinition(ST)

	GetObject() ST
	SetObject(ST)

	GetErrorMessage() string
	SetErrorMessage(string)

	// We will also need setters of client and kind to be able to implement generic versions of new builder and
	// pull.
	GetClient() runtimeclient.Client
	// Possibly better named GetGroupVersionKind() to be more explicit.
	GetKind() schema.GroupVersionKind
}

func Get[T any, ST objectPointer[T]](builder Builder[T, ST]) (*T, error) {
	if err := Validate(builder); err != nil {
		return nil, err
	}

	// This logic should probably be handled in a separate common logging library.
	namespace := builder.GetDefinition().GetNamespace()
	if namespace == "" {
		glog.V(100).Infof("Getting %s %s", builder.GetKind().Kind, builder.GetDefinition().GetName())
	} else {
		glog.V(100).Infof("Getting %s %s in namespace %s",
			builder.GetKind().Kind, builder.GetDefinition().GetName(), namespace)
	}

	// *T is not automatically inferred to be convertible to runtimeclient.Object, so we must use ST explicitly.
	var object ST = new(T)
	err := builder.GetClient().Get(context.TODO(), runtimeclient.ObjectKeyFromObject(builder.GetDefinition()), object)

	if err != nil {
		return nil, err
	}

	return object, nil
}

// Validate checks if the builder is valid. This is defined as being non-nil, having a non-nil client and definition,
// and not having an error message.
func Validate[T any, ST objectPointer[T]](builder Builder[T, ST]) error {
	// Since we take builder as an interface, we may inadvertently receive an interface with a nil concrete type,
	// requiring reflect.
	if builder == nil || reflect.ValueOf(builder).IsNil() {
		glog.V(100).Infof("The  builder is uninitialized")

		// It may be worth considering better error handling while we're overhauling, such as using custom types
		// and then exposing something like errors.IsBuilderNil() functions.
		return fmt.Errorf("error: received nil builder")
	}

	// Can find a more elegant solution, but we have to check that the builder is not nil before we can get the
	// kind.
	resourceCRD := builder.GetKind().Kind

	if builder.GetDefinition() == nil {
		glog.V(100).Infof("The %s is undefined", resourceCRD)

		return errors.New(msg.UndefinedCrdObjectErrString(resourceCRD))
	}

	// Checking against nil ensures that the interface itself is not nil. However, to verify that the interface does
	// not contain a concrete type that is nil, we must use reflect.
	client := builder.GetClient()
	if client == nil || reflect.ValueOf(client).IsNil() {
		glog.V(100).Infof("The %s builder apiClient is nil", resourceCRD)

		return fmt.Errorf("%s builder cannot have nil apiClient", resourceCRD)
	}

	if builder.GetErrorMessage() != "" {
		glog.V(100).Infof("The %s builder has error message %s", resourceCRD, builder.GetErrorMessage())

		return errors.New(builder.GetErrorMessage())
	}

	return nil
}
