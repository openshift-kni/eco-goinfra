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
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
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

	GetClient() runtimeclient.Client
	SetClient(runtimeclient.Client)

	// GetKind is expected to return an appropriate GVK even for a zero-valued builder. Possibly better named
	// GetGroupVersionKind() to be more explicit.
	GetKind() schema.GroupVersionKind
}

// Similar to the objectPointer type constraint, builderPointer is a type constraint that requires the type be a pointer
// to B that implements the Builder interface.
type builderPointer[B, T any, ST objectPointer[T]] interface {
	*B
	Builder[T, ST]
}

// NewNamespacedBuilder is an admittedly pretty cursed way to create a new builder but is completely generic.
//
//nolint:ireturn // SB should be a pointer to a struct in practice
func NewNamespacedBuilder[T any, B any, ST objectPointer[T], SB builderPointer[B, T, ST]](
	apiClient runtimeclient.Client, schemeAttacher clients.SchemeAttacher, name, nsname string) SB {
	var builder SB = new(B)

	glog.V(100).Infof("Initializing new %s builder with the following params: name: %s, nsname: %s",
		builder.GetKind().Kind, name, nsname)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is nil")

		return nil
	}

	err := schemeAttacher(apiClient.Scheme())
	if err != nil {
		glog.V(100).Infof("Failed to attach scheme for %s: %v", builder.GetKind().Kind, err)

		return nil
	}

	builder.SetClient(apiClient)
	builder.SetDefinition(new(T))
	builder.GetDefinition().SetName(name)
	builder.GetDefinition().SetNamespace(nsname)

	if name == "" {
		glog.V(100).Infof("The name is empty")

		builder.SetErrorMessage(fmt.Sprintf("name for %s cannot be empty", builder.GetKind().Kind))
	}

	if nsname == "" {
		glog.V(100).Infof("The nsname is empty")

		builder.SetErrorMessage(fmt.Sprintf("nsname for %s cannot be empty", builder.GetKind().Kind))
	}

	return builder
}

//nolint:ireturn // SB should be a pointer to a struct in practice
func PullNamespacedBuilder[T any, B any, ST objectPointer[T], SB builderPointer[B, T, ST]](
	apiClient runtimeclient.Client, schemeAttacher clients.SchemeAttacher, name, nsname string) (SB, error) {
	var builder SB = new(B)

	glog.V(100).Infof("Pulling existing %s builder %s in namespace %s", builder.GetKind().Kind, name, nsname)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is nil")

		return nil, fmt.Errorf("apiClient cannot be nil")
	}

	err := schemeAttacher(apiClient.Scheme())
	if err != nil {
		glog.V(100).Infof("Failed to attach scheme for %s: %v", builder.GetKind().Kind, err)

		return nil, err
	}

	builder.SetClient(apiClient)
	builder.SetDefinition(new(T))
	builder.GetDefinition().SetName(name)
	builder.GetDefinition().SetNamespace(nsname)

	if name == "" {
		glog.V(100).Infof("The name is empty")

		return nil, fmt.Errorf("name for %s cannot be empty", builder.GetKind().Kind)
	}

	if nsname == "" {
		glog.V(100).Infof("The nsname is empty")

		return nil, fmt.Errorf("nsname for %s cannot be empty", builder.GetKind().Kind)
	}

	if !Exists(builder) {
		glog.V(100).Infof("The %s %s does not exist in namespace %s", builder.GetKind().Kind, name, nsname)

		return nil, fmt.Errorf("%s %s does not exist in namespace %s", builder.GetKind().Kind, name, nsname)
	}

	builder.SetDefinition(builder.GetObject())

	return builder, nil
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

	var object ST = new(T)
	err := builder.GetClient().Get(context.TODO(), runtimeclient.ObjectKeyFromObject(builder.GetDefinition()), object)

	if err != nil {
		return nil, err
	}

	return object, nil
}

func Exists[T any, ST objectPointer[T]](builder Builder[T, ST]) bool {
	if err := Validate(builder); err != nil {
		return false
	}

	namespace := builder.GetDefinition().GetNamespace()
	if namespace == "" {
		glog.V(100).Infof("Checking if %s %s exists", builder.GetKind().Kind, builder.GetDefinition().GetName())
	} else {
		glog.V(100).Infof("Checking if %s %s exists in namespace %s",
			builder.GetKind().Kind, builder.GetDefinition().GetName(), namespace)
	}

	object, err := Get(builder)

	if err != nil {
		if namespace == "" {
			glog.V(100).Infof("Failed to get %s %s: %v", builder.GetKind().Kind, builder.GetDefinition().GetName(), err)
		} else {
			glog.V(100).Infof("Failed to get %s %s in namespace %s: %v",
				builder.GetKind().Kind, builder.GetDefinition().GetName(), namespace, err)
		}

		return false
	}

	builder.SetObject(object)

	return true
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
