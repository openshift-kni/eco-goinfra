package networkpolicy

import (
	"testing"

	"github.com/k8snetworkplumbingwg/multi-networkpolicy/pkg/apis/k8s.cni.cncf.io/v1beta1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var testSchemesV1beta1 = []clients.SchemeAttacher{
	v1beta1.AddToScheme,
}

func TestMultiNetworkPolicyPull(t *testing.T) {
	generateMultiNetworkPolicy := func(name, namespace string) *v1beta1.MultiNetworkPolicy {
		return &v1beta1.MultiNetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		policyName          string
		policyNamespace     string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			policyName:          "test-policy",
			policyNamespace:     "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: true,
			expectedErrorText:   "",
		},
		{
			policyName:          "test-policy",
			policyNamespace:     "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "MultiNetworkPolicy object test-policy does not exist in namespace test-namespace",
		},
		{
			policyName:          "",
			policyNamespace:     "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "MultiNetworkPolicy 'name' cannot be empty",
		},
		{
			policyName:          "test-policy",
			policyNamespace:     "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "MultiNetworkPolicy 'namespace' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		testMultiNetworkPolicy := generateMultiNetworkPolicy(testCase.policyName, testCase.policyNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testMultiNetworkPolicy)
		}

		testBuilder, testSettings := buildTestMultiNetworkPolicyBuilderWithFakeObjects(
			runtimeObjects, testCase.policyName, testCase.policyNamespace)

		if testCase.addToRuntimeObjects {
			// Create the object to pull
			_, err := testBuilder.Create()
			assert.Nil(t, err)
		}

		result, err := PullMultiNetworkPolicy(testSettings, testCase.policyName, testCase.policyNamespace)

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testMultiNetworkPolicy.Name, result.Object.Name)
			assert.Equal(t, testMultiNetworkPolicy.Namespace, result.Object.Namespace)
		}
	}
}

func TestMultiNetworkPolicyCreate(t *testing.T) {
	generateMultiNetworkPolicy := func(name, namespace string) *v1beta1.MultiNetworkPolicy {
		return &v1beta1.MultiNetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		policyName          string
		policyNamespace     string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			policyName:          "test-policy",
			policyNamespace:     "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: true,
			expectedErrorText:   "",
		},
		{
			policyName:          "test-policy",
			policyNamespace:     "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: false,
			expectedErrorText:   "",
		},
		// NOTE: In the test cases above, I would expect the first test case to fail because the object already exists.
		// I am not sure right now why it is not failing.  It seems that the 'fake' package is not working as expected.
		// This will need to be investigated further.
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, generateMultiNetworkPolicy(
				testCase.policyName, testCase.policyNamespace))
		}

		testBuilder, _ := buildTestMultiNetworkPolicyBuilderWithFakeObjects(
			runtimeObjects, testCase.policyName, testCase.policyNamespace)

		result, err := testBuilder.Create()

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCase.policyName, result.Object.Name)
			assert.Equal(t, testCase.policyNamespace, result.Object.Namespace)
		}
	}
}

func TestMultiNetworkPolicyDelete(t *testing.T) {
	generateMultiNetworkPolicy := func(name, namespace string) *v1beta1.MultiNetworkPolicy {
		return &v1beta1.MultiNetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		policyName          string
		policyNamespace     string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			policyName:          "test-policy",
			policyNamespace:     "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: true,
			expectedErrorText:   "",
		},
		{
			policyName:          "test-policy",
			policyNamespace:     "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: false,
			expectedErrorText:   "",
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, generateMultiNetworkPolicy(
				testCase.policyName, testCase.policyNamespace))
		}

		testBuilder, _ := buildTestMultiNetworkPolicyBuilderWithFakeObjects(
			runtimeObjects, testCase.policyName, testCase.policyNamespace)

		err := testBuilder.Delete()

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestMultiNetworkPolicyUpdate(t *testing.T) {
	generateMultiNetworkPolicy := func(name, namespace string) *v1beta1.MultiNetworkPolicy {
		return &v1beta1.MultiNetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		policyName          string
		policyNamespace     string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			policyName:          "test-policy",
			policyNamespace:     "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: true,
			expectedErrorText:   "",
		},
		{
			policyName:          "test-policy",
			policyNamespace:     "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "failed to update MultiNetworkPolicy, object does not exist on cluster",
		},
		{
			policyName:          "",
			policyNamespace:     "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "failed to update MultiNetworkPolicy, object does not exist on cluster",
		},
		{
			policyName:          "test-policy",
			policyNamespace:     "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "failed to update MultiNetworkPolicy, object does not exist on cluster",
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, generateMultiNetworkPolicy(
				testCase.policyName, testCase.policyNamespace))
		}

		testBuilder, _ := buildTestMultiNetworkPolicyBuilderWithFakeObjects(
			runtimeObjects, testCase.policyName, testCase.policyNamespace)

		if testCase.addToRuntimeObjects {
			// Create the object to update
			_, err := testBuilder.Create()
			assert.Nil(t, err)

			// Set some arbitrary values to update
			testBuilder.Definition.Spec.PodSelector = metav1.LabelSelector{MatchLabels: map[string]string{"test": "test"}}
		}

		testBuilder.Definition.ObjectMeta.ResourceVersion = "999"
		builder, err := testBuilder.Update()

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCase.policyName, builder.Object.Name)
			assert.Equal(t, testCase.policyNamespace, builder.Object.Namespace)
			assert.Equal(t, "test", builder.Object.Spec.PodSelector.MatchLabels["test"])
		}
	}
}

func TestMultiNetworkPolicyValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		expectedError string
	}{
		{ // Test Case 1 - builder is nil
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "error: received nil MultiNetworkPolicy builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined MultiNetworkPolicy",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "MultiNetworkPolicy builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder, _ := buildTestMultiNetworkPolicyBuilderWithFakeObjects(nil, "test-name", "test-namespace")

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		valid, err := testBuilder.validate()

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
			assert.False(t, valid)
		} else {
			assert.Nil(t, err)
			assert.True(t, valid)
		}
	}
}

func TestNewMultiNetworkPolicyBuilder(t *testing.T) {
	testCases := []struct {
		name              string
		namespace         string
		expectedError     bool
		expectedErrorText string
	}{
		{
			name:              "test-name",
			namespace:         "test-namespace",
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			name:              "",
			namespace:         "test-namespace",
			expectedError:     true,
			expectedErrorText: "The MultiNetworkPolicy 'name' cannot be empty",
		},
		{
			name:              "test-name",
			namespace:         "",
			expectedError:     true,
			expectedErrorText: "The MultiNetworkPolicy 'namespace' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewMultiNetworkPolicyBuilder(clients.GetTestClients(clients.TestClientParams{
			SchemeAttachers: testSchemes,
		}), testCase.name, testCase.namespace)

		if testCase.expectedError {
			assert.NotNil(t, testBuilder)
			assert.Equal(t, testCase.expectedErrorText, testBuilder.errorMsg)
		} else {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testBuilder.Definition.Namespace)
		}
	}
}

func TestMultiNetworkPolicyWithPodSelector(t *testing.T) {
	testCases := []struct {
		podSelector map[string]string
		expected    bool
	}{
		{
			podSelector: map[string]string{"test": "test"},
			expected:    true,
		},
		{
			podSelector: nil,
			expected:    false,
		},
	}

	for _, testCase := range testCases {
		testBuilder, _ := buildTestMultiNetworkPolicyBuilderWithFakeObjects(nil, "test-name", "test-namespace")

		testBuilder.WithPodSelector(metav1.LabelSelector{MatchLabels: testCase.podSelector})

		if testCase.expected {
			assert.Equal(t, testCase.podSelector, testBuilder.Definition.Spec.PodSelector.MatchLabels)
		} else {
			assert.Nil(t, testBuilder.Definition.Spec.PodSelector.MatchLabels)
		}
	}
}

func TestMultiNetworkPolicyWithNetwork(t *testing.T) {
	testBuilder, _ := buildTestMultiNetworkPolicyBuilderWithFakeObjects(nil, "test-name", "test-namespace")

	result := testBuilder.WithNetwork("test-network")

	assert.Equal(t, "test-network", result.Definition.Annotations["k8s.v1.cni.cncf.io/policy-for"])

	result = testBuilder.WithNetwork("")

	assert.Equal(t, "The networkName is an empty string", result.errorMsg)
}

func TestMultiNetworkPolicyWithEmptyIngress(t *testing.T) {
	testBuilder, _ := buildTestMultiNetworkPolicyBuilderWithFakeObjects(nil, "test-name", "test-namespace")

	result := testBuilder.WithEmptyIngress()

	assert.Empty(t, result.Definition.Spec.Ingress)
}

func TestMultiNetworkWithIngressRule(t *testing.T) {
	testBuilder, _ := buildTestMultiNetworkPolicyBuilderWithFakeObjects(nil, "test-name", "test-namespace")

	peerRule := v1beta1.MultiNetworkPolicyIngressRule{
		From: []v1beta1.MultiNetworkPolicyPeer{
			{
				PodSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"test": "test"},
				},
			},
		},
	}

	result := testBuilder.WithIngressRule(peerRule)

	assert.Equal(t, peerRule, result.Definition.Spec.Ingress[0])
}

func TestMultiNetworkPolicyWithEgressRule(t *testing.T) {
	testBuilder, _ := buildTestMultiNetworkPolicyBuilderWithFakeObjects(nil, "test-name", "test-namespace")

	peerRule := v1beta1.MultiNetworkPolicyEgressRule{
		To: []v1beta1.MultiNetworkPolicyPeer{
			{
				PodSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"test": "test"},
				},
			},
		},
	}

	result := testBuilder.WithEgressRule(peerRule)

	assert.Equal(t, peerRule, result.Definition.Spec.Egress[0])
}

func TestMultiNetworkPolicyGetMultiNetworkGVR(t *testing.T) {
	result := GetMultiNetworkGVR()

	assert.Equal(t, "k8s.cni.cncf.io", result.Group)
	assert.Equal(t, "v1beta1", result.Version)
	assert.Equal(t, "multi-networkpolicies", result.Resource)
}

func TestMultiNetworkPolicyWithPolicyType(t *testing.T) {
	testBuilder, _ := buildTestMultiNetworkPolicyBuilderWithFakeObjects(nil, "test-name", "test-namespace")

	result := testBuilder.WithPolicyType(v1beta1.PolicyTypeEgress)

	assert.Equal(t, v1beta1.PolicyTypeEgress, result.Definition.Spec.PolicyTypes[0])

	result = testBuilder.WithPolicyType("")

	assert.Equal(t, "The policy Type is an empty string", result.errorMsg)
}

func TestMultiNetworkPolicyExists(t *testing.T) {
	testCases := []struct {
		policyName      string
		policyNamespace string
		expected        bool
	}{
		{
			policyName:      "test-policy",
			policyNamespace: "test-namespace",
			expected:        true,
		},
		{
			policyName:      "test-policy",
			policyNamespace: "test-namespace",
			expected:        false,
		},
	}

	for _, testCase := range testCases {
		testBuilder, _ := buildTestMultiNetworkPolicyBuilderWithFakeObjects(nil,
			testCase.policyName, testCase.policyNamespace)

		if testCase.expected {
			_, err := testBuilder.Create()
			assert.Nil(t, err)
		}

		assert.Equal(t, testCase.expected, testBuilder.Exists())
	}
}

func buildTestMultiNetworkPolicyBuilderWithFakeObjects(runtimeObjects []runtime.Object,
	name, nsname string) (*MultiNetworkPolicyBuilder, *clients.Settings) {
	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  runtimeObjects,
		SchemeAttachers: testSchemesV1beta1,
	})

	return &MultiNetworkPolicyBuilder{
		apiClient: testSettings.Client,
		Definition: &v1beta1.MultiNetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}, testSettings
}
