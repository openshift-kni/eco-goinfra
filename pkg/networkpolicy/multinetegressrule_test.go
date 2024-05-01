package networkpolicy

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewEgressRuleBuilder(t *testing.T) {
	builder := NewEgressRuleBuilder()

	assert.NotNil(t, builder)
	assert.NotNil(t, builder.definition)
}

func TestEgressWithPortAndProtocol(t *testing.T) {
	builder := NewEgressRuleBuilder()

	builder.WithPortAndProtocol(80, "TCP")

	assert.Len(t, builder.definition.Ports, 1)
	assert.Equal(t, int(builder.definition.Ports[0].Port.IntVal), 80)

	builder.WithPortAndProtocol(0, "TCP")

	assert.Equal(t, builder.errorMsg, "port number can not be 0")
}

func TestEgressWithOptions(t *testing.T) {
	testCases := []struct {
		testOptions   []EgressAdditionalOptions
		expectedError string
	}{
		{
			testOptions: []EgressAdditionalOptions{
				func(builder *EgressRuleBuilder) (*EgressRuleBuilder, error) {
					return builder, errors.New("this is an error")
				},
			},
			expectedError: "this is an error",
		},
		{
			testOptions: []EgressAdditionalOptions{
				func(builder *EgressRuleBuilder) (*EgressRuleBuilder, error) {
					return builder, nil
				},
			},
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		builder := NewEgressRuleBuilder()

		builder.WithOptions(testCase.testOptions...)

		assert.Equal(t, testCase.expectedError, builder.errorMsg)
	}
}

func TestEgressWithPeerPodSelector(t *testing.T) {
	builder := NewEgressRuleBuilder()

	builder.WithPeerPodSelector(v1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "nginx",
		},
	})

	assert.Len(t, builder.definition.To, 1)
	assert.Equal(t, builder.definition.To[0].PodSelector.MatchLabels["app"], "nginx")

	builder.errorMsg = "error"

	builder.WithPeerPodSelector(v1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "nginx",
		},
	})

	assert.Len(t, builder.definition.To, 1)
}

func TestEgressWithPeerPodSelectorAndCIDR(t *testing.T) {
	builder := NewEgressRuleBuilder()

	builder.WithPeerPodSelectorAndCIDR(v1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "nginx",
		},
	}, "192.168.0.1/24", nil)

	assert.Len(t, builder.definition.To, 1)
	assert.Equal(t, builder.definition.To[0].PodSelector.MatchLabels["app"], "nginx")
	assert.Equal(t, builder.definition.To[0].IPBlock.CIDR, "192.168.0.1/24")

	// Test invalid CIDR
	builder.WithPeerPodSelectorAndCIDR(v1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "nginx",
		},
	}, "192.55.55.55", nil)

	assert.Equal(t, builder.errorMsg, "Invalid CIDR argument 192.55.55.55")

	// Test with exception
	builder.WithPeerPodSelectorAndCIDR(v1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "nginx",
		},
	}, "192.168.1.1/24", []string{"192.168.1.1"})

	assert.Equal(t, builder.definition.To[0].IPBlock.Except[0], "192.168.1.1")
}

func TestEgressGetEgressRuleCfg(t *testing.T) {
	builder := NewEgressRuleBuilder()

	cfg, err := builder.GetEgressRuleCfg()
	assert.Nil(t, err)
	assert.NotNil(t, cfg)

	builder.errorMsg = "error"

	cfg, err = builder.GetEgressRuleCfg()
	assert.NotNil(t, err)
	assert.Nil(t, cfg)
}
