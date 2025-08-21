package egressservice

import (
	"fmt"
	"testing"

	egresssvcv1 "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/egressservice/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
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
			expectedError: "the name parameter of the EgressService is empty",
		},
		{
			name:          egressTestSvcName,
			namespace:     "",
			sourceIPBy:    "LoadBalancerIP",
			expectedError: "the namespace of the EgressService is empty",
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
			expectedError: "invalid sourceIPBy parameter for the EgressService",
		},
		{
			name:          egressTestSvcName,
			namespace:     egressTestSvcNamespace,
			sourceIPBy:    "DemoByIp",
			expectedError: "invalid sourceIPBy parameter for the EgressService",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})

		testEgressServiceBuilder := NewEgressServiceBuilder(testSettings,
			testCase.name, testCase.namespace, testCase.sourceIPBy)

		assert.NotNil(t, testEgressServiceBuilder.Definition)

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
			testCase.name, testCase.namespace, testCase.sourceIPBy)

		assert.NotNil(t, testEgressServiceBuilder.Definition)

		testEgressServiceBuilder = testEgressServiceBuilder.WithNodeLabelSelector(testCase.nodeSelector)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testEgressServiceBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testEgressServiceBuilder.Definition.Namespace)
			assert.Equal(t, "", testEgressServiceBuilder.errorMsg)

			assert.Equal(t,
				testCase.nodeSelector,
				testEgressServiceBuilder.Definition.Spec.NodeSelector.MatchLabels)
		} else {
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
			expectedError: "cannot use empty VRF network",
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
			testCase.name, testCase.namespace, testCase.sourceIPBy)

		assert.NotNil(t, testEgressServiceBuilder.Definition)

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
			expectedError:       fmt.Errorf("egressService's name cannot be empty"),
		},
		{
			name:                egressTestSvcName,
			nsname:              "",
			sourceIPBy:          "LoadBalancerIP",
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("egressService's namespace cannot be empty"),
		},
		{
			name:                egressTestSvcName,
			nsname:              egressTestSvcNamespace,
			sourceIPBy:          "LoadBalancerIP",
			addToRuntimeObjects: false,
			client:              false,
			expectedError:       fmt.Errorf("egressService's 'apiClient' cannot be empty"),
		},
		{
			name:                egressTestSvcName,
			nsname:              egressTestSvcNamespace,
			sourceIPBy:          "LoadBalancerIP",
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf("egressService object %q does not exist in namespace %q",
				egressTestSvcName, egressTestSvcNamespace),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testEgressService := buildDummyEgressService(testCase.name, testCase.nsname, testCase.sourceIPBy)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testEgressService)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testEgressServiceBuilder, err := Pull(testSettings, testCase.name, testCase.nsname)

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
		expectedError       error
	}{
		{
			name:                egressTestSvcName,
			nsname:              egressTestSvcNamespace,
			sourceIPBy:          "LoadBalancerIP",
			addToRuntimeObjects: true,
			expectedError:       nil,
		},
		{
			name:                egressTestSvcName,
			nsname:              egressTestSvcNamespace,
			sourceIPBy:          "LoadBalancerIP",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("the name parameter of the EgressService is empty"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects           []runtime.Object
			testSettings             *clients.Settings
			testEgressServiceBuilder *EgressServiceBuilder
		)

		testEgressService := buildDummyEgressService(testCase.name, testCase.nsname, testCase.sourceIPBy)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testEgressService)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: testSchemes,
		})

		if testCase.expectedError != nil {
			testEgressServiceBuilder = buildInvalidDummyEgressServiceBuilder(testSettings)
		} else {
			testEgressServiceBuilder = buildDummyEgressServiceBuilder(
				testSettings, testCase.name, testCase.nsname, testCase.sourceIPBy)
		}

		testEgressService, err := testEgressServiceBuilder.Get()

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
			assert.Empty(t, testEgressService)
		} else {
			assert.Equal(t, testCase.name, testEgressService.ObjectMeta.Name)
			assert.Equal(t, testCase.nsname, testEgressService.ObjectMeta.Namespace)
			assert.Equal(t, testCase.sourceIPBy, string(testEgressService.Spec.SourceIPBy))
		}
	}
}

func TestUpdate(t *testing.T) {
	testCases := []struct {
		name          string
		nsname        string
		sourceIPBy    string
		newVrfNetwork string
		nodeSelector  map[string]string
		expectedError error
	}{
		{
			name:          egressTestSvcName,
			nsname:        egressTestSvcNamespace,
			sourceIPBy:    "LoadBalancerIP",
			newVrfNetwork: "5678",
			nodeSelector:  map[string]string{},
			expectedError: nil,
		},
		{
			name:          egressTestSvcName,
			nsname:        egressTestSvcNamespace,
			sourceIPBy:    "LoadBalancerIP",
			newVrfNetwork: " ",
			nodeSelector:  map[string]string{},
			expectedError: fmt.Errorf("cannot use empty VRF network"),
		},
	}

	for _, testCase := range testCases {
		testEgressServiceBuilder, err := Pull(buildTestClientWithDummyEgressService(),
			testCase.name, testCase.nsname)

		assert.Nil(t, err)

		testEgressServiceBuilder = testEgressServiceBuilder.WithVRFNetwork(testCase.newVrfNetwork)

		testEgressServiceBuilder, err = testEgressServiceBuilder.Update()

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.Equal(t, testCase.name, testEgressServiceBuilder.Object.Name)
			assert.Equal(t, testCase.nsname, testEgressServiceBuilder.Object.Namespace)

			assert.Equal(t, testCase.newVrfNetwork, testEgressServiceBuilder.Object.Spec.Network)
			assert.Equal(t,
				testEgressServiceBuilder.Definition.Spec.Network,
				testEgressServiceBuilder.Object.Spec.Network)
		}
	}
}

func TestCreate(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		sourceIPBy    string
		expectedError error
	}{
		{
			name:          egressTestSvcName,
			namespace:     egressTestSvcNamespace,
			sourceIPBy:    "Network",
			expectedError: nil,
		},
		{
			name:          egressTestSvcName,
			namespace:     egressTestSvcNamespace,
			sourceIPBy:    "LoadBalancerIP",
			expectedError: nil,
		},
		{
			name:          egressTestSvcName,
			namespace:     egressTestSvcNamespace,
			sourceIPBy:    "fake",
			expectedError: fmt.Errorf("invalid sourceIPBy parameter for the EgressService"),
		},
	}

	for _, testCase := range testCases {
		testEgressServiceBuilder := buildDummyEgressServiceBuilder(buildTestClientWithCustomizedEgressService(
			testCase.name, testCase.namespace, testCase.sourceIPBy),
			testCase.name, testCase.namespace, testCase.sourceIPBy)

		createdEgressService, err := testEgressServiceBuilder.Create()

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.Equal(t, testCase.name, createdEgressService.Definition.Name)
			assert.Equal(t, testCase.namespace, createdEgressService.Definition.Namespace)
			assert.Equal(t, testCase.sourceIPBy, string(createdEgressService.Definition.Spec.SourceIPBy))
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

func buildDummyEgressServiceBuilder(
	apiClient *clients.Settings, name, nsname, sourceIPBy string) *EgressServiceBuilder {
	return NewEgressServiceBuilder(apiClient, name, nsname, sourceIPBy)
}

func buildInvalidDummyEgressServiceBuilder(
	apiClient *clients.Settings) *EgressServiceBuilder {
	return NewEgressServiceBuilder(apiClient, "", egressTestSvcName, "LoadBalancerIP")
}

// buildTestClientWithDummyPod returns a client with a dummy Pod.
func buildTestClientWithDummyEgressService() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyEgressService(egressTestSvcName, egressTestSvcNamespace, "LoadBalancerIP"),
		},
		SchemeAttachers: testSchemes,
	})
}

func buildTestClientWithCustomizedEgressService(name, nsname, sourceIPBy string) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyEgressService(name, nsname, sourceIPBy),
		},
		SchemeAttachers: testSchemes,
	})
}
