package networkpolicy

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewIngressRuleBuilder(t *testing.T) {
	builder := NewIngressRuleBuilder()

	assert.NotNil(t, builder)
	assert.NotNil(t, builder.definition)
}

func TestIngressWithPortAndProtocol(t *testing.T) {
	builder := NewIngressRuleBuilder()

	builder.WithPortAndProtocol(80, "TCP")

	assert.Len(t, builder.definition.Ports, 1)
	assert.Equal(t, int(builder.definition.Ports[0].Port.IntVal), 80)

	builder.WithPortAndProtocol(0, "TCP")

	assert.Equal(t, builder.errorMsg, "port number can not be 0")
}

func TestIngressWithOptions(t *testing.T) {
	testCases := []struct {
		testOptions   []IngressAdditionalOptions
		expectedError string
	}{
		{
			testOptions: []IngressAdditionalOptions{
				func(builder *IngressRuleBuilder) (*IngressRuleBuilder, error) {
					return builder, errors.New("this is an error")
				},
			},
			expectedError: "this is an error",
		},
		{
			testOptions: []IngressAdditionalOptions{
				func(builder *IngressRuleBuilder) (*IngressRuleBuilder, error) {
					return builder, nil
				},
			},
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		builder := NewIngressRuleBuilder()

		builder.WithOptions(testCase.testOptions...)

		assert.Equal(t, testCase.expectedError, builder.errorMsg)
	}
}

func TestIngressWithPeerPodSelector(t *testing.T) {
	builder := NewIngressRuleBuilder()

	builder.WithPeerPodSelector(metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "nginx",
		},
	})

	assert.Len(t, builder.definition.From, 1)
	assert.Equal(t, builder.definition.From[0].PodSelector.MatchLabels["app"], "nginx")

	builder = NewIngressRuleBuilder()

	builder.errorMsg = "error"

	builder.WithPeerPodSelector(metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "nginx",
		},
	})

	assert.Len(t, builder.definition.From, 0)
}

func TestIngressWithCIDR(t *testing.T) {
	builder := NewIngressRuleBuilder()

	builder = builder.WithCIDR("192.168.1.1/24", nil)

	assert.Len(t, builder.definition.From, 1)
	assert.Equal(t, builder.definition.From[0].IPBlock.CIDR, "192.168.1.1/24")

	// Test invalid CIDR
	builder = NewIngressRuleBuilder()

	builder = builder.WithCIDR("192.55.55.55", nil)

	assert.Len(t, builder.definition.From, 0)
	assert.Equal(t, builder.errorMsg, "Invalid CIDR argument 192.55.55.55")

	// Test CIDR with except
	builder = NewIngressRuleBuilder()

	builder = builder.WithCIDR("192.168.1.1/24", []string{"192.168.1.1"})

	assert.Len(t, builder.definition.From, 1)
	assert.Equal(t, builder.definition.From[0].IPBlock.Except, []string{"192.168.1.1"})
}

func TestIngressWithPeerPodSelectorAndCIDR(t *testing.T) {
	builder := NewIngressRuleBuilder()

	builder.WithPeerPodSelectorAndCIDR(metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "nginx",
		},
	}, "192.168.1.1/24", nil)

	assert.Len(t, builder.definition.From, 2)
	assert.Equal(t, builder.definition.From[0].PodSelector.MatchLabels["app"], "nginx")
	assert.Equal(t, builder.definition.From[1].IPBlock.CIDR, "192.168.1.1/24")
}

func TestIngressGetIngressRuleCfg(t *testing.T) {
	builder := NewIngressRuleBuilder()

	cfg, err := builder.GetIngressRuleCfg()
	assert.Nil(t, err)
	assert.NotNil(t, cfg)

	builder.errorMsg = "error"

	cfg, err = builder.GetIngressRuleCfg()
	assert.NotNil(t, err)
	assert.Nil(t, cfg)
}
