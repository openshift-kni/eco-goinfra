package egressservice

import (
	"fmt"
	"testing"

	egresssvcv1 "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/egressservice/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"k8s.io/apimachinery/pkg/runtime"
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

func TestPull(t *testing.T) {
	testCases := []struct {
		name                string
		nsname              string
		sourceIPBy          string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                egressTestSvcName,
			nsname:              egressTestSvcNamespace,
			sourceIPBy:          "LoadBalancerIP",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			nsname:              egressTestSvcNamespace,
			sourceIPBy:          "LoadBalancerIP",
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("EgressService's name cannot be empty"),
		},
		{
			name:                egressTestSvcName,
			nsname:              "",
			sourceIPBy:          "LoadBalancerIP",
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("EgressService's namespace cannot be empty"),
		},
		{
			name:                egressTestSvcName,
			nsname:              egressTestSvcNamespace,
			sourceIPBy:          "LoadBalancerIP",
			addToRuntimeObjects: false,
			client:              false,
			expectedError:       fmt.Errorf("EgressService's 'apiClient' cannot be empty"),
		},
		{
			name:                egressTestSvcName,
			nsname:              egressTestSvcNamespace,
			sourceIPBy:          "LoadBalancerIP",
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf("EgressService object %q does not exist in namespace %q",
				egressTestSvcName, egressTestSvcNamespace),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testEgressService := buildDummyEgressService(testCase.name, testCase.nsname, testCase.sourceIPBy)
		t.Logf("Generated EgressService:\n%#v\n", testEgressService)

		if testCase.addToRuntimeObjects {
			t.Logf("Adding EgressService to runtime objects\n")
			runtimeObjects = append(runtimeObjects, testEgressService)
		}

		if testCase.client {
			t.Logf("Instantiating test Client\n")
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		t.Logf("Pulling EgressService\n")

		testEgressServiceBuilder, err := Pull(testSettings, testCase.name, testCase.nsname)

		t.Logf("Pulled Builder:\n%#v\n", testEgressServiceBuilder)

		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testEgressServiceBuilder.Definition.Name)
			assert.Equal(t, testCase.nsname, testEgressServiceBuilder.Definition.Namespace)
			assert.Equal(t, testCase.sourceIPBy, string(testEgressServiceBuilder.Definition.Spec.SourceIPBy))
		}
	}
}

func TestGet(t *testing.T) {
	testCases := []struct {
		name                string
		nsname              string
		sourceIPBy          string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                egressTestSvcName,
			nsname:              egressTestSvcNamespace,
			sourceIPBy:          "LoadBalancerIP",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                egressTestSvcName,
			nsname:              egressTestSvcNamespace,
			sourceIPBy:          "LoadBalancerIP",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("egressservices.k8s.ovn.org %q not found", egressTestSvcName),
		},
		{
			name:                egressTestSvcName,
			nsname:              egressTestSvcNamespace,
			sourceIPBy:          "LoadBalancerIP",
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("egressservices.k8s.ovn.org %q not found", egressTestSvcName),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		var svcName string

		if testCase.expectedError != nil {
			svcName = "hello"
		} else {
			svcName = testCase.name
		}

		t.Logf("Starting GET test\n")

		testEgressService := buildDummyEgressService(svcName, testCase.nsname, testCase.sourceIPBy)
		t.Logf("\tGenerated EgressService:\n%#v\n", testEgressService)

		if testCase.addToRuntimeObjects {
			t.Logf("\tAdding EgressService to runtime objects\n")
			runtimeObjects = append(runtimeObjects, testEgressService)
		}

		if testCase.client {
			t.Logf("\tInstantiating test Client\n")
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testEgressServiceBuilder := buildDummyEgressServiceBuilder(
			testSettings, testCase.name, testCase.nsname, testCase.sourceIPBy)

		t.Logf("\tRetrieving EgressService\n")

		testEgressService, err := testEgressServiceBuilder.Get()

		t.Logf("Retrieved EgressService:\n%#v\n", testEgressService)

		if testCase.expectedError != nil {
			t.Logf("\tChecking error handling\n")
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
			assert.Empty(t, testEgressService)
		} else {
			t.Logf("\tChecking EgressService's properties\n")
			assert.Equal(t, testCase.name, testEgressService.ObjectMeta.Name)
			assert.Equal(t, testCase.nsname, testEgressService.ObjectMeta.Namespace)
			assert.Equal(t, testCase.sourceIPBy, string(testEgressService.Spec.SourceIPBy))
		}
	}
}

// buildDummyEgressService generates EgressService definition.
func buildDummyEgressService(name, nsname, sourceIPBy string) *egresssvcv1.EgressService {
	return &egresssvcv1.EgressService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
		Spec: egresssvcv1.EgressServiceSpec{
			SourceIPBy: egresssvcv1.SourceIPMode(sourceIPBy),
		},
	}
}

func buildDummyEgressServiceBuilder(apiClient *clients.Settings, name, nsname, sourceIPBy string) *EgressServiceBuilder {
	return NewEgressServiceBuilder(apiClient, nsname, name, sourceIPBy)
}
