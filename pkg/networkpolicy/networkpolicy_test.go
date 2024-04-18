package networkpolicy

import (
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestNetworkPolicyPull(t *testing.T) {
	generateNetworkPolicy := func(name, namespace string) *netv1.NetworkPolicy {
		return &netv1.NetworkPolicy{
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
			expectedErrorText:   "networkPolicy object test-policy doesn't exist in namespace test-namespace",
		},
		{
			policyName:          "",
			policyNamespace:     "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "networkPolicy 'name' cannot be empty",
		},
		{
			policyName:          "test-policy",
			policyNamespace:     "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "networkPolicy 'namespace' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testNetworkPolicy := generateNetworkPolicy(testCase.policyName, testCase.policyNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testNetworkPolicy)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		result, err := Pull(testSettings, testCase.policyName, testCase.policyNamespace)

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testNetworkPolicy.Name, result.Object.Name)
			assert.Equal(t, testNetworkPolicy.Namespace, result.Object.Namespace)
		}
	}
}

func TestNetworkPolicyWithNamespaceIngressRule(t *testing.T) {
	testCases := []struct {
		testNamespaceIngressLabels map[string]string
		testPodIngressLabels       map[string]string
		expectedError              bool
		expectedErrorText          string
	}{
		{ // Test Case 1 - empty labels
			testNamespaceIngressLabels: map[string]string{},
			testPodIngressLabels:       map[string]string{},
			expectedError:              true,
			expectedErrorText:          "Both namespaceIngressMatchLabels and podIngressMatchLabels parameters are empty maps",
		},
		{ // Test Case 2 - empty namespace labels
			testNamespaceIngressLabels: map[string]string{},
			testPodIngressLabels:       map[string]string{"test": "test"},
			expectedError:              false,
			expectedErrorText:          "",
		},
		{ // Test Case 3 - empty pod labels
			testNamespaceIngressLabels: map[string]string{"test": "test"},
			testPodIngressLabels:       map[string]string{},
			expectedError:              false,
			expectedErrorText:          "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildTestBuilderWithFakeObjects(nil, "test-name", "test-namespace")

		result := testBuilder.WithNamespaceIngressRule(testCase.testNamespaceIngressLabels, testCase.testPodIngressLabels)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			// Assert that the IngressRule was added to the NetworkPolicyBuilder
			assert.NotNil(t, result)
			assert.NotNil(t, result.Definition.Spec.Ingress)

			// Assert that the IngressRule has the correct labels
			for _, ingressRule := range result.Definition.Spec.Ingress {
				if len(testCase.testNamespaceIngressLabels) != 0 {
					assert.NotNil(t, ingressRule.From[0].NamespaceSelector)
					assert.Equal(t, testCase.testNamespaceIngressLabels, ingressRule.From[0].NamespaceSelector.MatchLabels)
				}

				if len(testCase.testPodIngressLabels) != 0 {
					assert.NotNil(t, ingressRule.From[0].PodSelector)
					assert.Equal(t, testCase.testPodIngressLabels, ingressRule.From[0].PodSelector.MatchLabels)
				}
			}
		}
	}
}

func TestNetworkPolicyWithPolicyType(t *testing.T) {
	testCases := []struct {
		testPolicyType    netv1.PolicyType
		expectedError     bool
		expectedErrorText string
	}{
		{ // Test Case 1 - empty policy type
			testPolicyType:    "",
			expectedError:     true,
			expectedErrorText: "The policyType is an empty string",
		},
		{ // Test Case 2 - valid policy type
			testPolicyType:    netv1.PolicyTypeIngress,
			expectedError:     false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildTestBuilderWithFakeObjects(nil, "test-name", "test-namespace")

		result := testBuilder.WithPolicyType(testCase.testPolicyType)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			// Assert that the PolicyType was added to the NetworkPolicyBuilder
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testPolicyType, result.Definition.Spec.PolicyTypes[0])
		}
	}
}

func TestNetworkPolicyWithPodSelector(t *testing.T) {
	testCases := []struct {
		testPodSelector     map[string]string
		expectedError       bool
		expectedErrorText   string
		expectedPodSelector map[string]string
	}{
		{ // Test Case 1 - empty pod selector
			testPodSelector:     map[string]string{},
			expectedError:       true,
			expectedErrorText:   "The podSelector is an empty string",
			expectedPodSelector: nil,
		},
		{ // Test Case 2 - valid pod selector
			testPodSelector:     map[string]string{"test": "test"},
			expectedError:       false,
			expectedErrorText:   "",
			expectedPodSelector: map[string]string{"test": "test"},
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildTestBuilderWithFakeObjects(nil, "test-name", "test-namespace")

		result := testBuilder.WithPodSelector(testCase.testPodSelector)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			// Assert that the PodSelector was added to the NetworkPolicyBuilder
			assert.NotNil(t, result)
			assert.Equal(t, testCase.expectedPodSelector, result.Definition.Spec.PodSelector.MatchLabels)
		}
	}
}

func TestNetworkPolicyCreate(t *testing.T) {
	generateNetworkPolicy := func(name, namespace string) *netv1.NetworkPolicy {
		return &netv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		testName            string
		testNamespace       string
		addToRuntimeObjects bool
		expectedError       bool
		expectedErrorText   string
	}{
		{ // Test Case 1 - empty name
			testName:            "",
			testNamespace:       "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "The networkPolicy 'name' cannot be empty",
		},
		{ // Test Case 2 - empty namespace
			testName:            "test-name",
			testNamespace:       "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "The networkPolicy 'namespace' cannot be empty",
		},
		{ // Test Case 3 - valid name and namespace
			testName:            "test-name",
			testNamespace:       "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: false,
			expectedErrorText:   "",
		},
		{ // Test Case 4 - valid name and namespace with existing object
			testName:            "test-name",
			testNamespace:       "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: true,
			expectedErrorText:   "",
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, generateNetworkPolicy(testCase.testName, testCase.testNamespace))
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects, testCase.testName, testCase.testNamespace)

		result, err := testBuilder.Create()

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testName, result.Object.Name)
			assert.Equal(t, testCase.testNamespace, result.Object.Namespace)
		}
	}
}

func TestNetworkPolicyDelete(t *testing.T) {
	generateNetworkPolicy := func(name, namespace string) *netv1.NetworkPolicy {
		return &netv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		testName            string
		testNamespace       string
		addToRuntimeObjects bool
		expectedError       bool
		expectedErrorText   string
	}{
		{ // Test Case 1 - empty name
			testName:            "",
			testNamespace:       "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "The networkPolicy 'name' cannot be empty",
		},
		{ // Test Case 2 - empty namespace
			testName:            "test-name",
			testNamespace:       "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "The networkPolicy 'namespace' cannot be empty",
		},
		{ // Test Case 3 - valid name and namespace
			testName:            "test-name",
			testNamespace:       "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: false,
			expectedErrorText:   "",
		},
		{ // Test Case 4 - valid name and namespace with existing object
			testName:            "test-name",
			testNamespace:       "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: true,
			expectedErrorText:   "",
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, generateNetworkPolicy(testCase.testName, testCase.testNamespace))
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects, testCase.testName, testCase.testNamespace)

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

func TestNetworkPolicyUpdate(t *testing.T) {
	generateNetworkPolicy := func(name, namespace string) *netv1.NetworkPolicy {
		return &netv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		addToRuntimeObjects bool
		expectedError       bool
		expectedErrorText   string
	}{
		{ // Test Case 1 - valid update
			addToRuntimeObjects: true,
			expectedError:       false,
			expectedErrorText:   "",
		},
		{ // Test Case 2 - update with no existing object
			addToRuntimeObjects: false,
			expectedError:       true,
			expectedErrorText:   "networkpolicies.networking.k8s.io \"test-name\" not found",
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, generateNetworkPolicy("test-name", "test-namespace"))
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects, "test-name", "test-namespace")

		// Set some arbitrary values to update
		testBuilder.Definition.Labels = map[string]string{"test": "test"}

		builder, err := testBuilder.Update()

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testBuilder.Definition, builder.Definition)
			assert.Equal(t, testBuilder.Object.Labels, builder.Object.Labels)
		}
	}
}

func TestNetworkPolicyValidate(t *testing.T) {
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
			expectedError: "error: received nil NetworkPolicy builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined NetworkPolicy",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "NetworkPolicy builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildTestBuilderWithFakeObjects(nil, "test-name", "test-namespace")

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

func TestNewNetworkPolicyBuilder(t *testing.T) {
	testCases := []struct {
		testName          string
		testNamespace     string
		expectedError     bool
		expectedErrorText string
	}{
		{ // Test Case 1 - empty name
			testName:          "",
			testNamespace:     "test-namespace",
			expectedError:     true,
			expectedErrorText: "The networkPolicy 'name' cannot be empty",
		},
		{ // Test Case 2 - empty namespace
			testName:          "test-name",
			testNamespace:     "",
			expectedError:     true,
			expectedErrorText: "The networkPolicy 'namespace' cannot be empty",
		},
		{ // Test Case 3 - valid name and namespace
			testName:      "test-name",
			testNamespace: "test-namespace",
			expectedError: false,
		},
	}

	for _, testCase := range testCases {
		testNBP := NewNetworkPolicyBuilder(&clients.Settings{
			K8sClient:             nil,
			CoreV1Interface:       nil,
			AppsV1Interface:       nil,
			NetworkingV1Interface: nil,
		}, testCase.testName, testCase.testNamespace)

		if testCase.expectedError {
			assert.Equal(t, testNBP.errorMsg, testCase.expectedErrorText)
		} else {
			assert.NotNil(t, testNBP)
		}
	}
}

func buildTestBuilderWithFakeObjects(objects []runtime.Object, name, namespace string) *NetworkPolicyBuilder {
	fakeClient := k8sfake.NewSimpleClientset(objects...)

	return NewNetworkPolicyBuilder(&clients.Settings{
		K8sClient:             fakeClient,
		CoreV1Interface:       fakeClient.CoreV1(),
		AppsV1Interface:       fakeClient.AppsV1(),
		NetworkingV1Interface: fakeClient.NetworkingV1(),
	}, name, namespace)
}
