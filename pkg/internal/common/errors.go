package common

import "fmt"

// This file serves a similar purpose to the logging file, but for errors. It provides consistency for the error
// messages we return and by decoupling the error messages from the actual methods, it makes it much easier to update
// errors in the future.
//
// Improved errors is also a good future goal so this lays the groundwork for however that may look.

// --- Initialization Errors

// getAPIClientNilError returns an error for when the apiClient is nil.
func getAPIClientNilError(kind string) error {
	return fmt.Errorf("apiClient for a new %s builder cannot be nil", kind)
}

// wrapSchemeAttacherError wraps the error from when the scheme attacher fails to attach.
func wrapSchemeAttacherError(kind string, err error) error {
	return fmt.Errorf("failed to attach scheme for a new %s builder: %w", kind, err)
}

// getBuilderNameEmptyError returns an error for when the builder's name is empty.
func getBuilderNameEmptyError(kind string) error {
	return fmt.Errorf("name for a new %s builder cannot be empty", kind)
}

// getBuilderNamespaceEmptyError returns an error for when the builder's namespace is empty.
func getBuilderNamespaceEmptyError(kind string) error {
	return fmt.Errorf("namespace for a new %s builder cannot be empty", kind)
}

// --- Not Found Errors

// getBuilderNotFoundError returns an error for when the builder could not be found.
func getBuilderNotFoundError(kind string) error {
	return fmt.Errorf("the %s builder could not be found", kind)
}

// wrapGetError wraps the error from when the Get method fails.
func wrapGetError[O any, SO objectPointer[O]](builder Builder[O, SO], err error) error {
	kind := builder.GetGVK().Kind
	name := builder.GetDefinition().GetName()
	namespace := builder.GetDefinition().GetNamespace()

	if namespace == "" {
		return fmt.Errorf("failed to get the %s builder %s: %w", kind, name, err)
	}

	return fmt.Errorf("failed to get the %s builder %s in namespace %s: %w", kind, name, namespace, err)
}

// --- Validation Errors

// getBuilderUninitializedError returns an error for when the builder is uninitialized.
func getBuilderUninitializedError() error {
	return fmt.Errorf("the builder is uninitialized")
}

// getBuilderDefinitionNilError returns an error for when the builder's definition is nil.
func getBuilderDefinitionNilError(kind string) error {
	return fmt.Errorf("the %s builder cannot have nil definition", kind)
}

// getBuilderAPIClientNilError returns an error for when the builder's apiClient is nil.
func getBuilderAPIClientNilError(kind string) error {
	return fmt.Errorf("the %s builder cannot have nil apiClient", kind)
}
