package sriov

import (
	"context"
	"fmt"
	"testing"
	"time"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultNetName         = "sriovnet"
	defaultNetNsName       = "testnamespace"
	defaultNetTargetNsName = "targetns"
	defaultNetResName      = "resname"
)

var (
	testSchemes = []clients.SchemeAttacher{
		srIovV1.AddToScheme,
	}
)

//nolint:funlen
func TestPullNetwork(t *testing.T) {
	generateNetwork := func(name, namespace string) *srIovV1.SriovNetwork {
		return &srIovV1.SriovNetwork{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: srIovV1.SriovNetworkSpec{},
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
			expectedErrorText:   "sriovnetwork object test2 does not exist in namespace test-namespace",
			client:              true,
		},
		{
			networkName:         "",
			networkNamespace:    "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "sriovnetwork 'name' cannot be empty",
			client:              true,
		},
		{
			networkName:         "test3",
			networkNamespace:    "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "sriovnetwork 'namespace' cannot be empty",
			client:              true,
		},
		{
			networkName:         "test3",
			networkNamespace:    "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "sriovnetwork 'apiClient' cannot be empty",
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testNetwork := generateNetwork(testCase.networkName, testCase.networkNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testNetwork)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		// Test the Pull method
		builderResult, err := PullNetwork(testSettings, testNetwork.Name, testNetwork.Namespace)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			// Check the error message
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testNetwork.Name, builderResult.Object.Name)
			assert.Equal(t, testNetwork.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestNewNetworkBuilder(t *testing.T) {
	generateNetworkBuilder := NewNetworkBuilder

	testCases := []struct {
		networkName       string
		networkNamespace  string
		targetNs          string
		resName           string
		expectedErrorText string
		client            bool
	}{
		{
			networkName:      "test1",
			networkNamespace: "test-namespace",
			targetNs:         "target-namspace",
			resName:          "sriovNetwork",
			client:           true,
		},
		{
			networkName:       "",
			networkNamespace:  "test-namespace",
			targetNs:          "target-namespace",
			resName:           "sriovNetwork",
			expectedErrorText: "SrIovNetwork 'name' cannot be empty",
			client:            true,
		},
		{
			networkName:       "sriovnetworktest",
			networkNamespace:  "",
			targetNs:          "target-namespace",
			resName:           "sriovNetwork",
			expectedErrorText: "SrIovNetwork 'nsname' cannot be empty",
			client:            true,
		},
		{
			networkName:       "sriovnetworktest",
			networkNamespace:  "test-namespace",
			targetNs:          "",
			resName:           "sriovNetwork",
			expectedErrorText: "SrIovNetwork 'targetNsname' cannot be empty",
			client:            true,
		},
		{
			networkName:       "sriovnetworktest",
			networkNamespace:  "test-namespace",
			targetNs:          "target-namespace",
			resName:           "",
			expectedErrorText: "SrIovNetwork 'resName' cannot be empty",
			client:            true,
		},
	}
	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testNetworkStructure := generateNetworkBuilder(
			testSettings, testCase.networkName, testCase.networkNamespace, testCase.targetNs, testCase.resName)
		assert.NotNil(t, testNetworkStructure)

		if len(testCase.expectedErrorText) > 0 {
			assert.Equal(t, testNetworkStructure.errorMsg, testCase.expectedErrorText)
		}
	}
}

func TestWithLogLevel(t *testing.T) {
	testCases := []struct {
		loglevel          string
		expectedErrorText string
	}{
		{
			loglevel:          "panic",
			expectedErrorText: "",
		},
		{
			loglevel:          "error",
			expectedErrorText: "",
		},
		{
			loglevel:          "warning",
			expectedErrorText: "",
		},
		{
			loglevel:          "info",
			expectedErrorText: "",
		},
		{
			loglevel:          "debug",
			expectedErrorText: "",
		},
		{
			loglevel:          "",
			expectedErrorText: "",
		},
		{
			loglevel: "invalid",
			expectedErrorText: "invalid logLevel value, allowed logLevel values are:" +
				" panic, error, warning, info, debug or empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyObject()
		netBuilder := buildValidSriovNetworkTestBuilder(testSettings).WithLogLevel(testCase.loglevel)
		assert.Equal(t, netBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, netBuilder.Definition.Spec.LogLevel, testCase.loglevel)
		}
	}
}

func TestWithVlan(t *testing.T) {
	testCases := []struct {
		vlanID            uint16
		expectedErrorText string
	}{
		{
			vlanID:            100,
			expectedErrorText: "",
		},
		{
			vlanID:            9000,
			expectedErrorText: "invalid vlanID, allowed vlanID values are between 0-4094",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyObject()
		netBuilder := buildValidSriovNetworkTestBuilder(testSettings).WithVLAN(testCase.vlanID)
		assert.Equal(t, netBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, netBuilder.Definition.Spec.Vlan, int(testCase.vlanID))
		}
	}
}

func TestWithVlanProto(t *testing.T) {
	testCases := []struct {
		vlanProtocol      string
		expectedErrorText string
	}{
		{
			vlanProtocol:      "802.1q",
			expectedErrorText: "",
		},
		{
			vlanProtocol:      "802.1Q",
			expectedErrorText: "",
		},
		{
			vlanProtocol:      "802.1ad",
			expectedErrorText: "",
		},
		{
			vlanProtocol:      "802.1AD",
			expectedErrorText: "",
		},
		{
			vlanProtocol:      "802.1",
			expectedErrorText: "invalid 'vlanProtocol' parameters",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyObject()
		netBuilder := buildValidSriovNetworkTestBuilder(testSettings).WithVlanProto(testCase.vlanProtocol)
		assert.Equal(t, netBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, netBuilder.Definition.Spec.VlanProto, testCase.vlanProtocol)
		}
	}
}

func TestWithSpoof(t *testing.T) {
	testCases := []struct {
		spoof             bool
		expectedSpoofFlag string
	}{
		{
			spoof:             true,
			expectedSpoofFlag: "on",
		},
		{
			spoof:             false,
			expectedSpoofFlag: "off",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyObject()
		netBuilder := buildValidSriovNetworkTestBuilder(testSettings).WithSpoof(testCase.spoof)
		assert.Equal(t, netBuilder.Definition.Spec.SpoofChk, testCase.expectedSpoofFlag)
	}
}

func TestWithMetaPluginAllMultiFlag(t *testing.T) {
	testCases := []struct {
		allMulti             bool
		expectedSpoofSetting string
	}{
		{
			allMulti:             true,
			expectedSpoofSetting: `{ "type": "tuning", "allmulti": true }`,
		},
		{
			allMulti:             false,
			expectedSpoofSetting: `{ "type": "tuning", "allmulti": false }`,
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyObject()
		netBuilder := buildValidSriovNetworkTestBuilder(testSettings).WithMetaPluginAllMultiFlag(testCase.allMulti)
		assert.Equal(t, netBuilder.Definition.Spec.MetaPluginsConfig, testCase.expectedSpoofSetting)
	}
}

func TestWithLinkState(t *testing.T) {
	testCases := []struct {
		linkState           string
		expectedErrorOutput string
	}{
		{
			linkState: "enable",
		},
		{
			linkState: "disable",
		},
		{
			linkState: "auto",
		},
		{
			linkState:           "invalid",
			expectedErrorOutput: "invalid 'linkState' parameters",
		},
		{
			linkState:           "",
			expectedErrorOutput: "invalid 'linkState' parameters",
		},
	}
	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyObject()
		netBuilder := buildValidSriovNetworkTestBuilder(testSettings).WithLinkState(testCase.linkState)
		assert.Equal(t, netBuilder.errorMsg, testCase.expectedErrorOutput)

		if len(testCase.expectedErrorOutput) == 0 {
			assert.Equal(t, netBuilder.Definition.Spec.LinkState, testCase.linkState)
		}
	}
}

func TestWithMaxTxRate(t *testing.T) {
	testCases := []struct {
		maxTxRage uint16
	}{
		{
			maxTxRage: 0,
		},
		{
			maxTxRage: 100,
		},
		{
			maxTxRage: 10000,
		},
	}
	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyObject()
		netBuilder := buildValidSriovNetworkTestBuilder(testSettings).WithMaxTxRate(testCase.maxTxRage)
		assert.Equal(t, uint16(*netBuilder.Definition.Spec.MaxTxRate), testCase.maxTxRage)
	}
}

func TestWithMinTxRate(t *testing.T) {
	testCases := []struct {
		minTxRage uint16
	}{
		{
			minTxRage: 0,
		},
		{
			minTxRage: 100,
		},
		{
			minTxRage: 10000,
		},
	}
	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyObject()
		netBuilder := buildValidSriovNetworkTestBuilder(testSettings).WithMinTxRate(testCase.minTxRage)
		assert.Equal(t, uint16(*netBuilder.Definition.Spec.MaxTxRate), testCase.minTxRage)
	}
}

func TestWithTrustFlag(t *testing.T) {
	testCases := []struct {
		trustFlag         bool
		expectedTrustFlag string
	}{
		{
			trustFlag:         true,
			expectedTrustFlag: "on",
		},
		{
			trustFlag:         false,
			expectedTrustFlag: "off",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyObject()
		netBuilder := buildValidSriovNetworkTestBuilder(testSettings).WithSpoof(testCase.trustFlag)
		assert.Equal(t, netBuilder.Definition.Spec.SpoofChk, testCase.expectedTrustFlag)
	}
}

func TestWithVlanQoS(t *testing.T) {
	testCases := []struct {
		vlanQoS          uint16
		expectedErrorMsg string
	}{
		{
			vlanQoS:          0,
			expectedErrorMsg: "",
		},
		{
			vlanQoS:          8,
			expectedErrorMsg: "Invalid QoS class. Supported vlan QoS class values are between 0...7",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyObject()
		netBuilder := buildValidSriovNetworkTestBuilder(testSettings).WithVlanQoS(testCase.vlanQoS)

		assert.Equal(t, netBuilder.errorMsg, testCase.expectedErrorMsg)

		if testCase.expectedErrorMsg == "" {
			assert.Equal(t, uint16(netBuilder.Definition.Spec.VlanQoS), testCase.vlanQoS)
		}
	}
}

func TestWithIPAddressSupport(t *testing.T) {
	testSettings := buildTestClientWithDummyObject()
	netBuilder := buildValidSriovNetworkTestBuilder(testSettings).WithIPAddressSupport()
	assert.Equal(t, netBuilder.Definition.Spec.Capabilities, `{ "ips": true }`)
}

func TestWithMacAddressSupport(t *testing.T) {
	testSettings := buildTestClientWithDummyObject()
	netBuilder := buildValidSriovNetworkTestBuilder(testSettings).WithMacAddressSupport()
	assert.Equal(t, netBuilder.Definition.Spec.Capabilities, `{ "mac": true }`)
}

func TestWithStaticIpam(t *testing.T) {
	testSettings := buildTestClientWithDummyObject()
	netBuilder := buildValidSriovNetworkTestBuilder(testSettings).WithStaticIpam()
	assert.Equal(t, netBuilder.Definition.Spec.IPAM, `{ "type": "static" }`)
}

func TestWithOptions(t *testing.T) {
	testSettings := buildTestClientWithDummyObject()
	testBuilder := buildValidSriovNetworkTestBuilder(testSettings).WithOptions(
		func(builder *NetworkBuilder) (*NetworkBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidSriovNetworkTestBuilder(testSettings).WithOptions(
		func(builder *NetworkBuilder) (*NetworkBuilder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestCreate(t *testing.T) {
	testCases := []struct {
		testNetwork   *NetworkBuilder
		expectedError error
	}{
		{
			testNetwork:   buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testNetwork:   buildInvalidSrIovNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("SrIovNetwork 'resName' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		netBuilder, err := testCase.testNetwork.Create()
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.Equal(t, netBuilder.Definition, netBuilder.Object)
		}
	}
}

func TestDelete(t *testing.T) {
	testCases := []struct {
		testNetwork   *NetworkBuilder
		expectedError error
	}{
		{
			testNetwork:   buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testNetwork:   buildInvalidSrIovNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("SrIovNetwork 'resName' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testNetwork.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testNetwork.Object)
		}
	}
}

func TestNetworkDeleteAndWait(t *testing.T) {
	testCases := []struct {
		testNetwork   *NetworkBuilder
		expectedError error
	}{
		{
			testNetwork:   buildValidSriovNetworkTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testNetwork:   buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testNetwork:   buildInvalidSrIovNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("SrIovNetwork 'resName' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testNetwork.DeleteAndWait(1 * time.Second)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testNetwork.Object)
		}
	}
}

func TestNetworkWaitUntilDeleted(t *testing.T) {
	testCases := []struct {
		testNetwork   *NetworkBuilder
		expectedError error
	}{
		{
			testNetwork:   buildValidSriovNetworkTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testNetwork:   buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedError: context.DeadlineExceeded,
		},
		{
			testNetwork:   buildInvalidSrIovNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("SrIovNetwork 'resName' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testNetwork.WaitUntilDeleted(1 * time.Second)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testNetwork.Object)
		}
	}
}

func TestExist(t *testing.T) {
	testCases := []struct {
		testNetwork    *NetworkBuilder
		expectedStatus bool
	}{
		{
			testNetwork:    buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testNetwork:    buildInvalidSrIovNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testNetwork.Exists()
		assert.Equal(t, testCase.expectedStatus, exists)
	}
}

func TestGetSriovNetworksGVR(t *testing.T) {
	assert.Equal(t, GetSriovNetworksGVR(),
		schema.GroupVersionResource{
			Group: "sriovnetwork.openshift.io", Version: "v1", Resource: "sriovnetworks",
		})
}

func TestUpdate(t *testing.T) {
	testCases := []struct {
		testNetwork   *NetworkBuilder
		expectedError error
	}{
		{
			testNetwork:   buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
	}
	for _, testCase := range testCases {
		assert.Equal(t, "", testCase.testNetwork.Definition.Spec.IPAM)
		assert.Nil(t, nil, testCase.testNetwork.Object)
		testCase.testNetwork.WithStaticIpam()
		testCase.testNetwork.Definition.ObjectMeta.ResourceVersion = "999"
		netBuilder, err := testCase.testNetwork.Update(false)
		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, `{ "type": "static" }`, testCase.testNetwork.Object.Spec.IPAM)
		assert.Equal(t, netBuilder.Definition, netBuilder.Object)
	}
}

// buildValidTestBuilder returns a valid Builder for testing purposes.
func buildValidSriovNetworkTestBuilder(apiClient *clients.Settings) *NetworkBuilder {
	return NewNetworkBuilder(
		apiClient, defaultNetName, defaultNetNsName, defaultNetTargetNsName, defaultNetResName)
}

// buildValidTestBuilder returns a valid Builder for testing purposes.
func buildInvalidSrIovNetworkTestBuilder(apiClient *clients.Settings) *NetworkBuilder {
	return NewNetworkBuilder(
		apiClient, defaultNetName, defaultNetNsName, defaultNetTargetNsName, "")
}

func buildTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummySrIovNetworkObject(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummySrIovNetworkObject() []runtime.Object {
	return append([]runtime.Object{}, &srIovV1.SriovNetwork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultNetName,
			Namespace: defaultNetNsName,
		},
		Spec: srIovV1.SriovNetworkSpec{
			ResourceName:     defaultNetResName,
			NetworkNamespace: defaultNetTargetNsName,
		},
	})
}
