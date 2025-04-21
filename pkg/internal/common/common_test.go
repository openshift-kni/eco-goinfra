package common

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestValidate(t *testing.T) {
	testCases := []struct {
		name            string
		builderNil      bool
		definitionNil   bool
		apiClientNil    bool
		builderErrorMsg string
		expectedError   error
	}{
		{
			name:            "valid builder",
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   nil,
		},
		{
			name:            "nil builder",
			builderNil:      true,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("error: received nil builder"),
		},
		{
			name:            "nil definition",
			builderNil:      false,
			definitionNil:   true,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("can not redefine the undefined Namespace"),
		},
		{
			name:            "nil apiClient",
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    true,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("Namespace builder cannot have nil apiClient"),
		},
		{
			name:            "error message set",
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "test error",
			expectedError:   fmt.Errorf("test error"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := newMockBuilder()
			if testCase.builderNil {
				builder = nil
			}

			if testCase.definitionNil {
				builder.definition = nil
			}

			if testCase.apiClientNil {
				builder.client = nil
			}

			if testCase.builderErrorMsg != "" {
				builder.errorMessage = testCase.builderErrorMsg
			}

			err := Validate(builder)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

var _ Builder[corev1.Namespace, *corev1.Namespace] = &mockBuilder{}

// I just used the namespace since I didn't feel like creating a mock object for the PoC.
type mockBuilder struct {
	client       runtimeclient.Client
	definition   *corev1.Namespace
	object       *corev1.Namespace
	errorMessage string
	kind         schema.GroupVersionKind
}

func newMockBuilder() *mockBuilder {
	return &mockBuilder{
		client:       clients.GetTestClients(clients.TestClientParams{}),
		definition:   &corev1.Namespace{},
		object:       &corev1.Namespace{},
		errorMessage: "",
		kind:         corev1.SchemeGroupVersion.WithKind("Namespace"),
	}
}

func (builder *mockBuilder) GetDefinition() *corev1.Namespace {
	return builder.definition
}

func (builder *mockBuilder) SetDefinition(definition *corev1.Namespace) {
	builder.definition = definition
}

func (builder *mockBuilder) GetObject() *corev1.Namespace {
	return builder.object
}

func (builder *mockBuilder) SetObject(object *corev1.Namespace) {
	builder.object = object
}

func (builder *mockBuilder) GetErrorMessage() string {
	return builder.errorMessage
}

func (builder *mockBuilder) SetErrorMessage(errorMessage string) {
	builder.errorMessage = errorMessage
}

//nolint:ireturn
func (builder *mockBuilder) GetClient() runtimeclient.Client {
	return builder.client
}

// GetKind returns the GroupVersionKind of the underlying object.
func (builder *mockBuilder) GetKind() schema.GroupVersionKind {
	return builder.kind
}
