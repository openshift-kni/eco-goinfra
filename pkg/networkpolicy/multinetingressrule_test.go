package networkpolicy

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
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

func TestIngressWithProtocol(t *testing.T) {
	testCases := []struct {
		protocol      corev1.Protocol
		expectedError string
	}{
		{protocol: corev1.ProtocolTCP, expectedError: ""},
		{protocol: corev1.ProtocolUDP, expectedError: ""},
		{protocol: corev1.ProtocolSCTP, expectedError: ""},
		{protocol: "dummy", expectedError: "invalid protocol argument. Allowed protocols: TCP, UDP & SCTP"},
	}
	for _, testCase := range testCases {
		builder := NewIngressRuleBuilder().WithProtocol(testCase.protocol)
		assert.Equal(t, testCase.expectedError, builder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.protocol, *builder.definition.Ports[0].Protocol)
		}
	}
}

func TestIngressWithPort(t *testing.T) {
	testCases := []struct {
		port          uint16
		expectedError string
	}{
		{
			port:          5001,
			expectedError: "",
		},
		{
			port:          0,
			expectedError: "port number cannot be 0",
		},
	}
	for _, testCase := range testCases {
		builder := NewIngressRuleBuilder().WithPort(testCase.port)
		assert.Equal(t, testCase.expectedError, builder.errorMsg)
	}
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

func TestIngressWithPeerNamespaceSelector(t *testing.T) {
	testCases := []struct {
		namespaceSelector metav1.LabelSelector
		expectedError     string
	}{
		{
			namespaceSelector: metav1.LabelSelector{MatchLabels: map[string]string{"app": "nginx"}},
			expectedError:     "",
		},
		{
			namespaceSelector: metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "app",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"nginx"},
				},
			}},
			expectedError: "",
		},
	}
	for _, testCase := range testCases {
		builder := NewIngressRuleBuilder().WithPeerNamespaceSelector(testCase.namespaceSelector)
		assert.Equal(t, testCase.expectedError, builder.errorMsg)
		assert.Equal(t, &testCase.namespaceSelector, builder.definition.From[0].NamespaceSelector)
	}
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

func TestIngressWithPeerPodAndNamespaceSelector(t *testing.T) {
	testCases := []struct {
		podSelector       metav1.LabelSelector
		namespaceSelector metav1.LabelSelector
		expectedError     string
	}{
		{
			podSelector: metav1.LabelSelector{MatchLabels: map[string]string{"app": "nginx"}},
			namespaceSelector: metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{
				Key:      "app",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"nginx"},
			}}},
			expectedError: "",
		},
	}
	for _, testCase := range testCases {
		builder := NewIngressRuleBuilder().WithPeerPodAndNamespaceSelector(testCase.podSelector, testCase.namespaceSelector)
		assert.Equal(t, testCase.expectedError, builder.errorMsg)
		assert.Equal(t, &testCase.podSelector, builder.definition.From[0].PodSelector)
		assert.Equal(t, &testCase.namespaceSelector, builder.definition.From[0].NamespaceSelector)
	}
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
