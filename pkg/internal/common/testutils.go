package common

import (
	"errors"
	"fmt"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultName         = "test-resource"
	defaultNamespace    = "test-namespace"
	defaultErrorMessage = "test-error-message"
)

var (
	clusterScopedGVK = corev1.SchemeGroupVersion.WithKind("Namespace")
	namespacedGVK    = corev1.SchemeGroupVersion.WithKind("ConfigMap")
)

var errInvalidBuilder = errors.New(defaultErrorMessage)

var (
	// errSchemeAttachment is the error returned by testFailingSchemeAttacher.
	errSchemeAttachment                              = fmt.Errorf("scheme attachment failed")
	testSchemeAttacher        clients.SchemeAttacher = corev1.AddToScheme
	testFailingSchemeAttacher clients.SchemeAttacher = func(scheme *runtime.Scheme) error {
		return errSchemeAttachment
	}
)

// buildDummyClusterScopedResource creates a dummy cluster-scoped resource for testing. In this case, it is a Namespace,
// although the specific resource type is intentionally unimportant for the purpose of testing.
func buildDummyClusterScopedResource() *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultName,
		},
	}
}

// mockClusterScopedBuilder implements the Builder interface for testing using a cluster-scoped resource.
type mockClusterScopedBuilder struct {
	EmbeddableBuilder[corev1.Namespace, *corev1.Namespace]
}

// Compile-time check to ensure mockClusterScopedBuilder implements Builder interface.
var _ Builder[corev1.Namespace, *corev1.Namespace] = (*mockClusterScopedBuilder)(nil)

func buildValidMockClusterScopedBuilder(client runtimeclient.Client) *mockClusterScopedBuilder {
	return &mockClusterScopedBuilder{
		EmbeddableBuilder: EmbeddableBuilder[corev1.Namespace, *corev1.Namespace]{
			apiClient:    client,
			Definition:   buildDummyClusterScopedResource(),
			Object:       buildDummyClusterScopedResource(),
			errorMessage: "",
		},
	}
}

func buildInvalidMockClusterScopedBuilder(client runtimeclient.Client) *mockClusterScopedBuilder {
	return &mockClusterScopedBuilder{
		EmbeddableBuilder: EmbeddableBuilder[corev1.Namespace, *corev1.Namespace]{
			apiClient:    client,
			Definition:   buildDummyClusterScopedResource(),
			Object:       buildDummyClusterScopedResource(),
			errorMessage: defaultErrorMessage,
		},
	}
}

func (builder *mockClusterScopedBuilder) GetGVK() schema.GroupVersionKind {
	return clusterScopedGVK
}

// buildDummyNamespacedResource creates a dummy cluster-scoped resource for testing. In this case, it is a ConfigMap,
// although the specific resource type is intentionally unimportant for the purpose of testing.
func buildDummyNamespacedResource(name, namespace string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

// mockNamespacedBuilder implements the Builder interface for testing using a namespaced resource.
type mockNamespacedBuilder struct {
	EmbeddableBuilder[corev1.ConfigMap, *corev1.ConfigMap]
}

// Compile-time check to ensure mockNamespacedBuilder implements Builder interface.
var _ Builder[corev1.ConfigMap, *corev1.ConfigMap] = (*mockNamespacedBuilder)(nil)

func (builder *mockNamespacedBuilder) GetGVK() schema.GroupVersionKind {
	return namespacedGVK
}
