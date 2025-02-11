package networkpolicy

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestEgressWithProtocol(t *testing.T) {
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
		builder := NewEgressRuleBuilder().WithProtocol(testCase.protocol)
		assert.Equal(t, testCase.expectedError, builder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.protocol, *builder.definition.Ports[0].Protocol)
		}
	}
}

func TestEgressWithPort(t *testing.T) {
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
		builder := NewEgressRuleBuilder().WithPort(testCase.port)
		assert.Equal(t, testCase.expectedError, builder.errorMsg)
	}
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

	builder.WithPeerPodSelector(metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "nginx",
		},
	})

	assert.Len(t, builder.definition.To, 1)
	assert.Equal(t, builder.definition.To[0].PodSelector.MatchLabels["app"], "nginx")

	builder = NewEgressRuleBuilder()

	//nolint:goconst
	builder.errorMsg = "error"

	builder.WithPeerPodSelector(metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "nginx",
		},
	})

	assert.Len(t, builder.definition.To, 0)
}

func TestEgressWithPeerNamespaceSelector(t *testing.T) {
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
		builder := NewEgressRuleBuilder().WithPeerNamespaceSelector(testCase.namespaceSelector)
		assert.Equal(t, testCase.expectedError, builder.errorMsg)
		assert.Equal(t, &testCase.namespaceSelector, builder.definition.To[0].NamespaceSelector)
	}
}

func TestEgressWithCIDR(t *testing.T) {
	testCases := []struct {
		cidr           string
		except         string
		expectedLength int
	}{
		{
			cidr:           "192.168.1.1/24",
			expectedLength: 1,
		},
		{
			cidr:           "192.168.1.1",
			expectedLength: 0,
		},
		{
			cidr:           "192.168.1.1/24",
			except:         "192.168.1.10/32",
			expectedLength: 1,
		},
	}
	for _, testCase := range testCases {
		if len(testCase.except) != 0 {
			builder := NewEgressRuleBuilder().WithCIDR(testCase.cidr, []string{testCase.except})
			assert.Equal(t, testCase.expectedLength, len(builder.definition.To))
			assert.Equal(t, testCase.except, builder.definition.To[0].IPBlock.Except[0])

			if len(builder.definition.To) != 0 {
				assert.Equal(t, testCase.cidr, builder.definition.To[0].IPBlock.CIDR)
			}
		} else {
			builder := NewEgressRuleBuilder().WithCIDR(testCase.cidr)
			assert.Equal(t, testCase.expectedLength, len(builder.definition.To))

			if len(builder.definition.To) != 0 {
				assert.Equal(t, testCase.cidr, builder.definition.To[0].IPBlock.CIDR)
			}
		}
	}
}

func TestEgressWithPeerPodSelectorAndCIDR(t *testing.T) {
	builder := NewEgressRuleBuilder()

	builder.WithPeerPodSelectorAndCIDR(metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "nginx",
		},
	}, "192.168.0.1/24", nil)

	assert.Len(t, builder.definition.To, 1)
	assert.Equal(t, builder.definition.To[0].PodSelector.MatchLabels["app"], "nginx")
	assert.Equal(t, builder.definition.To[0].IPBlock.CIDR, "192.168.0.1/24")

	builder = NewEgressRuleBuilder()

	// Test invalid CIDR
	builder.WithPeerPodSelectorAndCIDR(metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "nginx",
		},
	}, "192.55.55.55", nil)

	assert.Equal(t, builder.errorMsg, "Invalid CIDR argument 192.55.55.55")

	builder = NewEgressRuleBuilder()

	// Test with exception
	builder.WithPeerPodSelectorAndCIDR(metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "nginx",
		},
	}, "192.168.1.1/24", []string{"192.168.1.1"})

	assert.Equal(t, builder.definition.To[0].IPBlock.Except[0], "192.168.1.1")
}

func TestEgressWithPeerPodAndNamespaceSelector(t *testing.T) {
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
		builder := NewEgressRuleBuilder().WithPeerPodAndNamespaceSelector(testCase.podSelector, testCase.namespaceSelector)
		assert.Equal(t, testCase.expectedError, builder.errorMsg)
		assert.Equal(t, &testCase.podSelector, builder.definition.To[0].PodSelector)
		assert.Equal(t, &testCase.namespaceSelector, builder.definition.To[0].NamespaceSelector)
	}
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
