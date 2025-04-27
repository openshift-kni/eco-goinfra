//nolint:revive // annoying comment warning, just for PoC
package common

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type EmbeddableBuilder[T any, ST objectPointer[T]] struct {
	// Definition is the local version of the object that any changes will be applied to.
	Definition ST
	// Object is the object as it appears on the cluster.
	Object ST
	// errorMessage stores any error messages from modifying the builder.
	errorMessage string
	// apiClient is the connection to the cluster.
	apiClient runtimeclient.Client
}

// assertImplementsBuilder allows us to check at compile time that the EmbeddableBuilder struct implements the Builder
// interface for any type parameters T and ST.
func assertImplementsBuilder[T any, ST objectPointer[T]](builder *EmbeddableBuilder[T, ST]) struct{} {
	var _ Builder[T, ST] = builder

	return struct{}{}
}

// So that the compiler does not complain about the unused function, we create this dummy call. Really we should use a
// mock object but for the sake of the PoC we use Namespace.
var _ = assertImplementsBuilder((*EmbeddableBuilder[corev1.Namespace, *corev1.Namespace])(nil))

// These methods are mostly for ensuring that Builder is implemented and would probably have some sort of validation in
// a real implementation.

// GetDefinition returns the Definition field of the EmbeddableBuilder struct.
//
//nolint:ireturn // ST should be a pointer to a struct
func (builder *EmbeddableBuilder[T, ST]) GetDefinition() ST {
	return builder.Definition
}

// SetDefinition sets the Definition field of the EmbeddableBuilder struct.
func (builder *EmbeddableBuilder[T, ST]) SetDefinition(definition ST) {
	builder.Definition = definition
}

// GetObject returns the Object field of the EmbeddableBuilder struct.
//
//nolint:ireturn // ST should be a pointer to a struct
func (builder *EmbeddableBuilder[T, ST]) GetObject() ST {
	return builder.Object
}

// SetObject sets the Object field of the EmbeddableBuilder struct.
func (builder *EmbeddableBuilder[T, ST]) SetObject(object ST) {
	builder.Object = object
}

// GetErrorMessage returns the errorMessage field of the EmbeddableBuilder struct.
func (builder *EmbeddableBuilder[T, ST]) GetErrorMessage() string {
	return builder.errorMessage
}

// SetErrorMessage sets the errorMessage field of the EmbeddableBuilder struct.
func (builder *EmbeddableBuilder[T, ST]) SetErrorMessage(errorMessage string) {
	builder.errorMessage = errorMessage
}

// GetClient returns the Client field of the EmbeddableBuilder struct.
//
//nolint:ireturn
func (builder *EmbeddableBuilder[T, ST]) GetClient() runtimeclient.Client {
	return builder.apiClient
}

// SetClient sets the Client field of the EmbeddableBuilder struct.
func (builder *EmbeddableBuilder[T, ST]) SetClient(client runtimeclient.Client) {
	builder.apiClient = client
}

// GetKind is meant to be shadowed by the actual implementation of the builder so that the zero value of the builder can
// return a GVK.
func (builder *EmbeddableBuilder[T, ST]) GetKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{}
}

// Get is an example of a CRUD operation that uses the EmbeddableBuilder struct. Since it uses the common Validate
// function there is no need to create a similar method for EmbeddableBuilder.
func (builder *EmbeddableBuilder[T, ST]) Get() (*T, error) {
	return Get(builder)
}

// Exists checks whether the given object exists in the cluster.
func (builder *EmbeddableBuilder[T, ST]) Exists() bool {
	return Exists(builder)
}
