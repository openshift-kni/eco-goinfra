package sriov

import (
	"fmt"
	"testing"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultPolicyName         = "sriovnet"
	defaultPolicyNsName       = "testnamespace"
	defaultPolicyResName      = "resname"
	defaultPolicyVFNum        = 1
	defaultPolicyNICs         = []string{"eth1"}
	defaultPolicyNodeSelector = map[string]string{"node": "selector"}
)

//nolint:funlen
func TestPullPolicy(t *testing.T) {
	generatePolicy := func(name, namespace string) *srIovV1.SriovNetworkNodePolicy {
		return &srIovV1.SriovNetworkNodePolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: srIovV1.SriovNetworkNodePolicySpec{},
		}
	}

	testCases := []struct {
		networkName         string
		networkNamespace    string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
		client              bool
	}{
		{
			networkName:         "test1",
			networkNamespace:    "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			networkName:         "test2",
			networkNamespace:    "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "sriovnetworknodepolicy object test2 does not exist in namespace test-namespace",
			client:              true,
		},
		{
			networkName:         "",
			networkNamespace:    "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "sriovnetworknodepolicy 'name' cannot be empty",
			client:              true,
		},
		{
			networkName:         "test3",
			networkNamespace:    "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "sriovnetworknodepolicy 'namespace' cannot be empty",
			client:              true,
		},
		{
			networkName:         "test3",
			networkNamespace:    "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "sriovnetworknodepolicy 'apiClient' cannot be empty",
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testPolicy := generatePolicy(testCase.networkName, testCase.networkNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPolicy)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		// Test the Pull method
		builderResult, err := PullPolicy(testSettings, testPolicy.Name, testPolicy.Namespace)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			// Check the error message
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testPolicy.Name, builderResult.Object.Name)
			assert.Equal(t, testPolicy.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestNewPolicyBuilder(t *testing.T) {
	generatePolicyBuilder := NewPolicyBuilder

	testCases := []struct {
		policyName        string
		policyNamespace   string
		resName           string
		vfsNumber         int
		nicNames          []string
		nodeSelector      map[string]string
		expectedErrorText string
		client            bool
	}{
		{
			policyName:        "test1",
			policyNamespace:   "test-namespace",
			resName:           "sriovPolicy",
			vfsNumber:         1,
			nodeSelector:      map[string]string{"node": "selector"},
			nicNames:          []string{"eth0"},
			expectedErrorText: "",
		},
		{
			policyName:        "",
			policyNamespace:   "test-namespace",
			resName:           "sriovPolicy",
			vfsNumber:         1,
			nodeSelector:      map[string]string{"node": "selector"},
			nicNames:          []string{"eth0"},
			expectedErrorText: "SriovNetworkNodePolicy 'name' cannot be empty",
		},
		{
			policyName:        "test1",
			policyNamespace:   "",
			resName:           "sriovPolicy",
			vfsNumber:         1,
			nodeSelector:      map[string]string{"node": "selector"},
			nicNames:          []string{"eth0"},
			expectedErrorText: "SriovNetworkNodePolicy 'nsname' cannot be empty",
		},
		{
			policyName:        "test1",
			policyNamespace:   "test-namespace",
			resName:           "",
			vfsNumber:         1,
			nodeSelector:      map[string]string{"node": "selector"},
			nicNames:          []string{"eth0"},
			expectedErrorText: "SriovNetworkNodePolicy 'resName' cannot be empty",
		},
		{
			policyName:        "test1",
			policyNamespace:   "test-namespace",
			resName:           "sriovPolicy",
			vfsNumber:         -1,
			nodeSelector:      map[string]string{"node": "selector"},
			nicNames:          []string{"eth0"},
			expectedErrorText: "SriovNetworkNodePolicy 'vfsNumber' cannot be zero of negative",
		},
		{
			policyName:        "test1",
			policyNamespace:   "test-namespace",
			resName:           "sriovPolicy",
			vfsNumber:         1,
			nodeSelector:      map[string]string{},
			nicNames:          []string{"eth0"},
			expectedErrorText: "SriovNetworkNodePolicy 'nodeSelector' cannot be empty map",
		},
		{
			policyName:        "test1",
			policyNamespace:   "test-namespace",
			resName:           "sriovPolicy",
			vfsNumber:         1,
			nodeSelector:      map[string]string{"node": "selector"},
			nicNames:          []string{},
			expectedErrorText: "SriovNetworkNodePolicy 'nicNames' cannot be empty list",
		},
	}
	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testPolicyStructure := generatePolicyBuilder(
			testSettings,
			testCase.policyName,
			testCase.policyNamespace,
			testCase.resName,
			testCase.vfsNumber,
			testCase.nicNames,
			testCase.nodeSelector)
		assert.NotNil(t, testPolicyStructure)
		assert.Equal(t, testPolicyStructure.errorMsg, testCase.expectedErrorText)
	}
}

func TestPolicyWithDevType(t *testing.T) {
	testCases := []struct {
		devType           string
		expectedErrorText string
	}{
		{
			devType:           "vfio-pci",
			expectedErrorText: "",
		},
		{
			devType:           "netdevice",
			expectedErrorText: "",
		},
		{
			devType:           "",
			expectedErrorText: "invalid device type, allowed devType values are: vfio-pci or netdevice",
		},
		{
			devType:           "invalid",
			expectedErrorText: "invalid device type, allowed devType values are: vfio-pci or netdevice",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyPolicyObject()
		netBuilder := buildValidSriovPolicyTestBuilder(testSettings).WithDevType(testCase.devType)
		assert.Equal(t, netBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, netBuilder.Definition.Spec.DeviceType, testCase.devType)
		}
	}
}

func TestPolicyVFRange(t *testing.T) {
	testCases := []struct {
		firstVF           int
		lastVF            int
		expectedErrorText string
	}{
		{
			firstVF:           0,
			lastVF:            63,
			expectedErrorText: "",
		},
		{
			firstVF:           0,
			lastVF:            65,
			expectedErrorText: "lastVF can not be greater than 63",
		},
		{
			firstVF:           65,
			lastVF:            1,
			expectedErrorText: "firstPF argument can not be greater than lastPF",
		},
		{
			firstVF:           -2,
			lastVF:            0,
			expectedErrorText: "firstPF or lastVF can not be less than 0",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyPolicyObject()
		netBuilder := buildValidSriovPolicyTestBuilder(testSettings).WithVFRange(testCase.firstVF, testCase.lastVF)
		assert.Equal(t, netBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, fmt.Sprintf("%s#%d-%d", "eth1", testCase.firstVF, testCase.lastVF),
				netBuilder.Definition.Spec.NicSelector.PfNames[0])
		}
	}
}

func TestPolicyWithMTU(t *testing.T) {
	testCases := []struct {
		mtu               int
		expectedErrorText string
	}{
		{
			mtu:               1500,
			expectedErrorText: "",
		},
		{
			mtu:               0,
			expectedErrorText: "invalid mtu size 0 allowed mtu should be in range 1...9192",
		},
		{
			mtu:               -1,
			expectedErrorText: "invalid mtu size -1 allowed mtu should be in range 1...9192",
		},
		{
			mtu:               20000,
			expectedErrorText: "invalid mtu size 20000 allowed mtu should be in range 1...9192",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyPolicyObject()
		netBuilder := buildValidSriovPolicyTestBuilder(testSettings).WithMTU(testCase.mtu)
		assert.Equal(t, netBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.mtu, netBuilder.Definition.Spec.Mtu)
		}
	}
}

func TestPolicyWithRDMA(t *testing.T) {
	testCases := []struct {
		rdma              bool
		expectedErrorText string
	}{
		{
			rdma: true,
		},
		{
			rdma: false,
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyPolicyObject()
		netBuilder := buildValidSriovPolicyTestBuilder(testSettings).WithRDMA(testCase.rdma)
		assert.Equal(t, netBuilder.errorMsg, testCase.expectedErrorText)
		assert.Equal(t, testCase.rdma, netBuilder.Definition.Spec.IsRdma)
	}
}

func TestPolicyVhostNet(t *testing.T) {
	testCases := []struct {
		vhostNet          bool
		expectedErrorText string
	}{
		{
			vhostNet: true,
		},
		{
			vhostNet: false,
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyPolicyObject()
		netBuilder := buildValidSriovPolicyTestBuilder(testSettings).WithVhostNet(testCase.vhostNet)
		assert.Equal(t, netBuilder.errorMsg, testCase.expectedErrorText)
		assert.Equal(t, testCase.vhostNet, netBuilder.Definition.Spec.NeedVhostNet)
	}
}

func TestPolicyWithExternallyManaged(t *testing.T) {
	testCases := []struct {
		externallyManaged bool
		expectedErrorText string
	}{
		{
			externallyManaged: true,
		},
		{
			externallyManaged: false,
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyPolicyObject()
		netBuilder := buildValidSriovPolicyTestBuilder(testSettings).WithExternallyManaged(testCase.externallyManaged)
		assert.Equal(t, netBuilder.errorMsg, testCase.expectedErrorText)
		assert.Equal(t, testCase.externallyManaged, netBuilder.Definition.Spec.ExternallyManaged)
	}
}

func TestPolicyWithOptions(t *testing.T) {
	testSettings := buildTestClientWithDummyObject()
	testBuilder := buildValidSriovPolicyTestBuilder(testSettings).WithOptions(
		func(builder *PolicyBuilder) (*PolicyBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidSriovPolicyTestBuilder(testSettings).WithOptions(
		func(builder *PolicyBuilder) (*PolicyBuilder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestPolicyCreate(t *testing.T) {
	testCases := []struct {
		testPolicy    *PolicyBuilder
		expectedError error
	}{
		{
			testPolicy:    buildValidSriovPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError: nil,
		},
		{
			testPolicy:    buildInvalidSriovPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError: fmt.Errorf("SriovNetworkNodePolicy 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		netBuilder, err := testCase.testPolicy.Create()
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.Equal(t, netBuilder.Definition, netBuilder.Object)
		}
	}
}

func TestPolicyDelete(t *testing.T) {
	testCases := []struct {
		testPolicy    *PolicyBuilder
		expectedError error
	}{
		{
			testPolicy:    buildValidSriovPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError: nil,
		},
		{
			testPolicy:    buildInvalidSriovPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError: fmt.Errorf("SriovNetworkNodePolicy 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testPolicy.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testPolicy.Object)
		}
	}
}

func TestPolicyExist(t *testing.T) {
	testCases := []struct {
		testPolicy     *PolicyBuilder
		expectedStatus bool
	}{
		{
			testPolicy:     buildValidSriovPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedStatus: true,
		},
		{
			testPolicy:     buildInvalidSriovPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testPolicy.Exists()
		assert.Equal(t, testCase.expectedStatus, exists)
	}
}

// buildValidSriovPolicyTestBuilder returns a valid PolicyBuilder for testing purposes.
func buildValidSriovPolicyTestBuilder(apiClient *clients.Settings) *PolicyBuilder {
	return NewPolicyBuilder(
		apiClient,
		defaultPolicyName,
		defaultPolicyNsName,
		defaultPolicyResName,
		defaultPolicyVFNum,
		defaultPolicyNICs,
		defaultPolicyNodeSelector)
}

// buildInvalidSriovPolicyTestBuilder returns an invalid PolicyBuilder for testing purposes.
func buildInvalidSriovPolicyTestBuilder(apiClient *clients.Settings) *PolicyBuilder {
	return NewPolicyBuilder(
		apiClient,
		defaultPolicyName,
		"",
		defaultPolicyResName,
		defaultPolicyVFNum,
		defaultPolicyNICs,
		defaultPolicyNodeSelector)
}

func buildTestClientWithDummyPolicyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummySrIovPolicyObject(),
	})
}

func buildDummySrIovPolicyObject() []runtime.Object {
	return append([]runtime.Object{}, buildDummySrIovPolicy(defaultNetName, defaultNetNsName))
}

func buildDummySrIovPolicy(name, namespace string) *srIovV1.SriovNetworkNodePolicy {
	return &srIovV1.SriovNetworkNodePolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: srIovV1.SriovNetworkNodePolicySpec{
			ResourceName: defaultNetResName,
			NodeSelector: defaultPolicyNodeSelector,
			NumVfs:       defaultPolicyVFNum,
			NicSelector: srIovV1.SriovNetworkNicSelector{
				PfNames: defaultPolicyNICs,
			},
			Priority: 1,
		},
	}
}
