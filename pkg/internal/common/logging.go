package common

import "github.com/golang/glog"

// This file exists to abstract the logging for the common package. This is for a few reasons:
//  - It reduces the clutter caused by having to check namespaced/cluster-scoped inside the actual methods.
//  - It allows us to decouple the logging from the actual methods, making it much easier to change the specifics
//    without going through and updating each method. Especially since the goal is to improve the logging in the future.
//
// In this initial phase, these functions are relatively thin wrappers around the existing glog calls. This is subject
// to change as we improve the logging.

// --- Initialization Logs

// logNewNamespacedBuilderInitializing prints the log for when initializing a new namespaced builder.
func logNewNamespacedBuilderInitializing(kind, name, namespace string) {
	glog.V(100).Infof("Initializing new %s builder with the following params: name=%s, nsname=%s",
		kind, name, namespace)
}

// logPullNamespacedBuilderPulling prints the log for when pulling a namespaced builder.
func logPullNamespacedBuilderPulling(kind, name, namespace string) {
	glog.V(100).Infof("Pulling %s builder with the following params: name=%s, nsname=%s",
		kind, name, namespace)
}

// logNewClusterScopedBuilderInitializing prints the log for when initializing a new cluster-scoped builder.
func logNewClusterScopedBuilderInitializing(kind, name string) {
	glog.V(100).Infof("Initializing new %s builder with the following params: name=%s", kind, name)
}

// logPullClusterScopedBuilderPulling prints the log for when pulling a cluster-scoped builder.
func logPullClusterScopedBuilderPulling(kind, name string) {
	glog.V(100).Infof("Pulling %s builder with the following params: name=%s", kind, name)
}

// --- Initialization Errors

// logAPIClientNil prints the log for when the apiClient is nil.
func logAPIClientNil(kind string) {
	glog.V(100).Infof("The apiClient for a new %s builder cannot be nil", kind)
}

// logSchemedFailedToAttach prints the log for when the scheme fails to attach.
func logSchemedFailedToAttach(kind string, err error) {
	glog.V(100).Infof("Failed to attach scheme for a new %s builder: %v", kind, err)
}

// logBuilderNameEmpty prints the log for when the builder's name is empty.
func logBuilderNameEmpty(kind string) {
	glog.V(100).Infof("The name for a new %s builder cannot be empty", kind)
}

// logBuilderNamespaceEmpty prints the log for when the builder's namespace is empty.
func logBuilderNamespaceEmpty(kind string) {
	glog.V(100).Infof("The namespace for a new %s builder cannot be empty", kind)
}

// --- Method Call Logs

// logBuilderGet prints the log for using the builder's Get method. It assumes the builder has already been validated.
func logBuilderGet[O any, SO objectPointer[O]](builder Builder[O, SO]) {
	kind := builder.GetGVK().Kind
	name := builder.GetDefinition().GetName()
	namespace := builder.GetDefinition().GetNamespace()

	if namespace == "" {
		glog.V(100).Infof("Getting %s builder %s", kind, name)
	} else {
		glog.V(100).Infof("Getting %s builder %s in namespace %s", kind, name, namespace)
	}
}

// logBuilderExists prints the log for using the builder's Exists method. It assumes the builder has already been
// validated.
func logBuilderExists[O any, SO objectPointer[O]](builder Builder[O, SO]) {
	kind := builder.GetGVK().Kind
	name := builder.GetDefinition().GetName()
	namespace := builder.GetDefinition().GetNamespace()

	if namespace == "" {
		glog.V(100).Infof("Checking if %s builder %s exists", kind, name)
	} else {
		glog.V(100).Infof("Checking if %s builder %s in namespace %s exists", kind, name, namespace)
	}
}

// --- Not Found Errors

// logBuilderNotFound prints the log for when a builder could not be found.
func logBuilderNotFound(kind string) {
	glog.V(100).Infof("The %s builder could not be found", kind)
}

// logBuilderNotFoundWithError prints the log for when a builder could not be found with an error.
func logBuilderNotFoundWithError(kind string, err error) {
	glog.V(100).Infof("The %s builder could not be found: %v", kind, err)
}

// --- Validation Errors

// logBuilderUninitialized prints the log for when the builder is uninitialized.
func logBuilderUninitialized() {
	glog.V(100).Infof("The builder is uninitialized")
}

// logBuilderUndefined prints the log for when the builder is undefined.
func logBuilderUndefined(kind string) {
	glog.V(100).Infof("The %s builder is undefined", kind)
}

// logBuilderAPIClientNil prints the log for when the builder's apiClient is nil.
func logBuilderAPIClientNil(kind string) {
	glog.V(100).Infof("The %s builder apiClient is nil", kind)
}

// logBuilderErrorMessage prints the log for when the builder has an error message.
func logBuilderErrorMessage(kind, errorMessage string) {
	glog.V(100).Infof("The %s builder has error message %s", kind, errorMessage)
}
