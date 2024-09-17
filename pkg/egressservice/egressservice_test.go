package egressservice

import (
	"testing"

	egresssvcv1 "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/egressservice/v1"
	"github.com/stretchr/testify/assert"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
)

var (
	egressTestSvcName      = "demo-egress-svc"
	egressTestSvcNamespace = "demo-ns"
	testSchemes            = []clients.SchemeAttacher{
		egresssvcv1.AddToScheme,
	}
)

func TestNewEgressServiceBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		sourceIPBy    string
		expectedError string
	}{
		{
			name:          "",
			namespace:     egressTestSvcNamespace,
			sourceIPBy:    "LoadBalancerIP",
			expectedError: "The name parameter of the EgressService is empty",
		},
		{
			name:          egressTestSvcName,
			namespace:     "",
			sourceIPBy:    "LoadBalancerIP",
			expectedError: "The namespace of the EgressService is empty",
		},
		{
			name:          egressTestSvcName,
			namespace:     egressTestSvcNamespace,
			sourceIPBy:    "LoadBalancerIP",
			expectedError: "",
		},
		{
			name:          egressTestSvcName,
			namespace:     egressTestSvcNamespace,
			sourceIPBy:    "Network",
			expectedError: "",
		},
		{
			name:          egressTestSvcName,
			namespace:     egressTestSvcNamespace,
			sourceIPBy:    "",
			expectedError: "Invalid sourceIPBy parameter for the EgressService",
		},
		{
			name:          egressTestSvcName,
			namespace:     egressTestSvcNamespace,
			sourceIPBy:    "DemoByIp",
			expectedError: "Invalid sourceIPBy parameter for the EgressService",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})

		testEgressServiceBuilder := NewEgressServiceBuilder(testSettings,
			testCase.namespace, testCase.name, testCase.sourceIPBy)

		assert.NotNil(t, testEgressServiceBuilder.Definition)

		t.Logf("Testing sourceIPBy: %q", testCase.sourceIPBy)
		t.Logf("Definition:\n%#v\n", testEgressServiceBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testEgressServiceBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testEgressServiceBuilder.Definition.Namespace)
			assert.Equal(t, "", testEgressServiceBuilder.errorMsg)
			assert.Equal(t, testCase.sourceIPBy, string(testEgressServiceBuilder.Definition.Spec.SourceIPBy))
			assert.Equal(t, "", testEgressServiceBuilder.Definition.Spec.Network)
		} else {
			assert.Equal(t, testCase.expectedError, testEgressServiceBuilder.errorMsg)
		}
	}
}

func TestWithNodeLabelSelector(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		sourceIPBy    string
		nodeSelector  map[string]string
		expectedError string
	}{
		{
			name:          egressTestSvcName,
			namespace:     egressTestSvcNamespace,
			sourceIPBy:    "LoadBalancerIP",
			expectedError: "",
			nodeSelector:  map[string]string{"egress-svc": "true"},
		},
		{
			name:          egressTestSvcName,
			namespace:     egressTestSvcNamespace,
			sourceIPBy:    "LoadBalancerIP",
			expectedError: "",
			nodeSelector:  map[string]string{"egress-svc": "true", "prod": ""},
		},
		{
			name:          egressTestSvcName,
			namespace:     egressTestSvcNamespace,
			sourceIPBy:    "LoadBalancerIP",
			expectedError: "cannot accept empty map as nodeSelector",
			nodeSelector:  map[string]string{},
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})

		testEgressServiceBuilder := NewEgressServiceBuilder(testSettings,
			testCase.namespace, testCase.name, testCase.sourceIPBy)

		assert.NotNil(t, testEgressServiceBuilder.Definition)

		t.Logf("Testing NodeSelector: %q", testCase.nodeSelector)

		testEgressServiceBuilder = testEgressServiceBuilder.WithNodeLabelSelector(testCase.nodeSelector)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testEgressServiceBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testEgressServiceBuilder.Definition.Namespace)
			assert.Equal(t, "", testEgressServiceBuilder.errorMsg)

			t.Logf("Definition:\n%#v\n", testEgressServiceBuilder.Definition.Spec.NodeSelector)

			for key, val := range testCase.nodeSelector {
				t.Logf("Processing: %q -> %q", key, val)
				assert.Equal(t,
					testEgressServiceBuilder.Definition.Spec.NodeSelector.MatchLabels[key],
					val)
			}
		} else {
			t.Logf("Error clause. Expected error: %q", testCase.expectedError)
			t.Logf("NodeSelector Definition:\n%#v\n", testEgressServiceBuilder.Definition.Spec.NodeSelector)
			t.Logf("Builder Definition:\n%#v\n", testEgressServiceBuilder)
			assert.Equal(t, testCase.expectedError, testEgressServiceBuilder.errorMsg)
		}

	}
}

func TestWithVRFNetwork(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		sourceIPBy    string
		vrfNetwork    string
		expectedError string
	}{
		{
			name:          egressTestSvcName,
			namespace:     egressTestSvcNamespace,
			sourceIPBy:    "LoadBalancerIP",
			vrfNetwork:    "",
			expectedError: "Cannot use emtpy VRF network",
		},
		{
			name:          egressTestSvcName,
			namespace:     egressTestSvcNamespace,
			sourceIPBy:    "LoadBalancerIP",
			vrfNetwork:    "1001",
			expectedError: "",
		},
		{
			name:          egressTestSvcName,
			namespace:     egressTestSvcNamespace,
			sourceIPBy:    "LoadBalancerIP",
			vrfNetwork:    "vrfName",
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})

		testEgressServiceBuilder := NewEgressServiceBuilder(testSettings,
			testCase.namespace, testCase.name, testCase.sourceIPBy)

		assert.NotNil(t, testEgressServiceBuilder.Definition)

		t.Logf("Testing VRF Network: %q", testCase.vrfNetwork)

		testEgressServiceBuilder = testEgressServiceBuilder.WithVRFNetwork(testCase.vrfNetwork)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testEgressServiceBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testEgressServiceBuilder.Definition.Namespace)
			assert.Equal(t, "", testEgressServiceBuilder.errorMsg)
			assert.Equal(t, testCase.vrfNetwork, testEgressServiceBuilder.Definition.Spec.Network)
		} else {
			assert.Equal(t, testCase.expectedError, testEgressServiceBuilder.errorMsg)
		}
	}
}
