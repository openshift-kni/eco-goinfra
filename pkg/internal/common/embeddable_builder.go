package common

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// EmbeddableBuilder is a struct implementing the Builder interface that can be embedded in existing builder structs. It
// provides the basic fields, methods, and methods for CRUD functions that builders need.
type EmbeddableBuilder[O any, SO objectPointer[O]] struct {
	// Definition is the desired form of the resource.
	Definition SO
	// Object is the last pulled form of the resource.
	Object SO
	// errorMessage is the error message stored in the builder.
	errorMessage string
	// apiClient is the client used for connecting with the K8s cluster.
	apiClient runtimeclient.Client
	// gvk is the GVK of the resource the builder represents. It is set by [SetGVK] and then returned by all
	// subsequent [GetGVK] calls.
	gvk schema.GroupVersionKind
}

// GetDefinition returns the desired form of the resource. In this case, it is equivalent to accessing the Definition
// field directly.
func (b *EmbeddableBuilder[O, SO]) GetDefinition() SO {
	return b.Definition
}

// SetDefinition updates the desired form of the resource. In this case, it is equivalent to updating the Definition
// field directly.
func (b *EmbeddableBuilder[O, SO]) SetDefinition(definition SO) {
	b.Definition = definition
}

// GetObject returns the last pulled form of the resource. In this case, it is equivalent to accessing the Object field
// directly.
func (b *EmbeddableBuilder[O, SO]) GetObject() SO {
	return b.Object
}

// SetObject updates the last pulled form of the resource. In this case, it is equivalent to updating the Object field
// directly. End users should not call this method directly since the object is automatically updated when the resource
// is pulled from the cluster.
func (b *EmbeddableBuilder[O, SO]) SetObject(object SO) {
	b.Object = object
}

// GetErrorMessage returns the error message stored in the builder. In this case, it is equivalent to accessing the
// errorMessage field directly. This method exists solely for internal use and should not be relied upon by end users.
// Instead, users should rely on error messages returned from the CRUD+ methods.
func (b *EmbeddableBuilder[O, SO]) GetErrorMessage() string {
	return b.errorMessage
}

// SetErrorMessage updates the error message stored in the builder. In this case, it is equivalent to updating the
// errorMessage field directly. This method exists solely for internal use and should not be relied upon by end users.
// Error messages are stored automatically during creation of the builder and should not be updated manually.
func (b *EmbeddableBuilder[O, SO]) SetErrorMessage(errorMessage string) {
	b.errorMessage = errorMessage
}

// GetClient returns the client used for connecting with the K8s cluster.
func (b *EmbeddableBuilder[O, SO]) GetClient() runtimeclient.Client {
	return b.apiClient
}

// SetClient updates the client used for connecting with the K8s cluster. This method exists solely for internal use and
// therefore does not perform the usual validation for the client. Setting it after creation also invalidates the
// object.
func (b *EmbeddableBuilder[O, SO]) SetClient(apiClient runtimeclient.Client) {
	b.apiClient = apiClient
}

// GetGVK returns the GVK for the resource the builder represents, even if the builder is zero-valued. However, the
// EmbeddableBuilder is unable to return the proper GVK for the embedding builder with its zero value. Therefore,
// embedders of EmbeddableBuilder should provide their own implementation of this method to handle builder
// initialization properly.
func (b *EmbeddableBuilder[O, SO]) GetGVK() schema.GroupVersionKind {
	return b.gvk
}

// SetGVK updates the GVK for the resource the builder represents. This method exists solely for internal use and should
// not be used nor overridden by embedders. This method allows for functions which create builders to set the GVK that
// the EmbeddableBuilder will return based on the resource-specific GVK.
func (b *EmbeddableBuilder[O, SO]) SetGVK(gvk schema.GroupVersionKind) {
	b.gvk = gvk
}

// Get pulls the resource from the cluster and returns it. It does not modify the builder.
func (b *EmbeddableBuilder[O, SO]) Get() (SO, error) {
	return Get(b)
}

// Exists checks whether the resource exists on the cluster. If the resource does exist, the builder's object is updated
// with the resource and this returns true. If the builder is invalid, or the resource cannot be retrieved, this returns
// false without modifying the builder.
func (b *EmbeddableBuilder[O, SO]) Exists() bool {
	return Exists(b)
}
