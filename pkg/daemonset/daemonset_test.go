package daemonset

import (
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

// buildValidTestBuilder returns a valid Builder for testing purposes.
func buildValidTestBuilder() *Builder {
	return NewBuilder(&clients.Settings{
		Client: nil,
	}, "test-name", "test-namespace", map[string]string{
		"test-key": "test-value",
	}, corev1.Container{
		Name: "test-container",
	})
}

func TestWithNodeSelector(t *testing.T) {
	testBuilder := buildValidTestBuilder()
	testBuilder.WithNodeSelector(map[string]string{
		"test-node-selector-key": "test-node-selector-value",
	})

	assert.Equal(t, "test-node-selector-value",
		testBuilder.Definition.Spec.Template.Spec.NodeSelector["test-node-selector-key"])
}

func TestWithAdditionalContainerSpecs(t *testing.T) {
	testBuilder := buildValidTestBuilder()
	testBuilder.WithAdditionalContainerSpecs([]corev1.Container{
		{
			Name: "test-additional-container",
		},
	})

	assert.Equal(t, "test-additional-container",
		testBuilder.Definition.Spec.Template.Spec.Containers[1].Name)
}

func TestWithOptions(t *testing.T) {
	testBuilder := buildValidTestBuilder()
	testBuilder.WithOptions(func(builder *Builder) (*Builder, error) {
		builder.Definition.Spec.Template.Spec.Containers[0].Name = "test-container-name"

		return builder, nil
	})

	assert.Equal(t, "test-container-name",
		testBuilder.Definition.Spec.Template.Spec.Containers[0].Name)
}
