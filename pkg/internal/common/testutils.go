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
	apiClient    runtimeclient.Client
	definition   *corev1.Namespace
	object       *corev1.Namespace
	errorMessage string
	gvk          schema.GroupVersionKind
}

// Compile-time check to ensure mockClusterScopedBuilder implements Builder interface.
var _ Builder[corev1.Namespace, *corev1.Namespace] = (*mockClusterScopedBuilder)(nil)

func buildValidMockClusterScopedBuilder(client runtimeclient.Client) *mockClusterScopedBuilder {
	return &mockClusterScopedBuilder{
		apiClient:    client,
		definition:   buildDummyClusterScopedResource(),
		object:       buildDummyClusterScopedResource(),
		errorMessage: "",
	}
}

func buildInvalidMockClusterScopedBuilder(client runtimeclient.Client) *mockClusterScopedBuilder {
	return &mockClusterScopedBuilder{
		apiClient:    client,
		definition:   buildDummyClusterScopedResource(),
		object:       buildDummyClusterScopedResource(),
		errorMessage: defaultErrorMessage,
	}
}

func (builder *mockClusterScopedBuilder) GetDefinition() *corev1.Namespace {
	return builder.definition
}

func (builder *mockClusterScopedBuilder) SetDefinition(definition *corev1.Namespace) {
	builder.definition = definition
}

func (builder *mockClusterScopedBuilder) GetObject() *corev1.Namespace {
	return builder.object
}

func (builder *mockClusterScopedBuilder) SetObject(object *corev1.Namespace) {
	builder.object = object
}

func (builder *mockClusterScopedBuilder) GetErrorMessage() string {
	return builder.errorMessage
}

func (builder *mockClusterScopedBuilder) SetErrorMessage(errorMessage string) {
	builder.errorMessage = errorMessage
}

func (builder *mockClusterScopedBuilder) GetClient() runtimeclient.Client {
	return builder.apiClient
}

func (builder *mockClusterScopedBuilder) SetClient(client runtimeclient.Client) {
	builder.apiClient = client
}

func (builder *mockClusterScopedBuilder) GetGVK() schema.GroupVersionKind {
	return clusterScopedGVK
}

func (builder *mockClusterScopedBuilder) SetGVK(gvk schema.GroupVersionKind) {
	builder.gvk = gvk
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
	apiClient    runtimeclient.Client
	definition   *corev1.ConfigMap
	object       *corev1.ConfigMap
	errorMessage string
	gvk          schema.GroupVersionKind
}

// Compile-time check to ensure mockNamespacedBuilder implements Builder interface.
var _ Builder[corev1.ConfigMap, *corev1.ConfigMap] = (*mockNamespacedBuilder)(nil)

func (builder *mockNamespacedBuilder) GetDefinition() *corev1.ConfigMap {
	return builder.definition
}

func (builder *mockNamespacedBuilder) SetDefinition(definition *corev1.ConfigMap) {
	builder.definition = definition
}

func (builder *mockNamespacedBuilder) GetObject() *corev1.ConfigMap {
	return builder.object
}

func (builder *mockNamespacedBuilder) SetObject(object *corev1.ConfigMap) {
	builder.object = object
}

func (builder *mockNamespacedBuilder) GetErrorMessage() string {
	return builder.errorMessage
}

func (builder *mockNamespacedBuilder) SetErrorMessage(errorMessage string) {
	builder.errorMessage = errorMessage
}

func (builder *mockNamespacedBuilder) GetClient() runtimeclient.Client {
	return builder.apiClient
}

func (builder *mockNamespacedBuilder) SetClient(client runtimeclient.Client) {
	builder.apiClient = client
}

func (builder *mockNamespacedBuilder) GetGVK() schema.GroupVersionKind {
	return namespacedGVK
}

func (builder *mockNamespacedBuilder) SetGVK(gvk schema.GroupVersionKind) {
	builder.gvk = gvk
}
