package metallb

import (
	"fmt"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/metallb/mlbtypesv1beta2"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	bgpPeerGVK = schema.GroupVersionKind{
		Group:   APIGroup,
		Version: APIVersionv2,
		Kind:    bpgPeerKind,
	}
	defaultBGPPeerName   = "default-bgp-peer"
	defaultBGPPeerNsName = "test-namespace"
)

func TestNewBPGPeerBuilder(t *testing.T) {
	generateBPGPeer := NewBPGPeerBuilder

	testCases := []struct {
		name          string
		namespace     string
		peerIP        string
		asn           uint32
		remoteAsn     uint32
		expectedError string
	}{
		{
			name:          "bgppeer",
			namespace:     "test-namespace",
			peerIP:        "192.168.1.1",
			asn:           5001,
			remoteAsn:     5002,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     "test-namespace",
			peerIP:        "192.168.1.1",
			asn:           5001,
			remoteAsn:     5002,
			expectedError: "BGPPeer 'name' cannot be empty",
		},
		{
			name:          "bgppeer",
			namespace:     "",
			peerIP:        "192.168.1.1",
			asn:           5001,
			remoteAsn:     5002,
			expectedError: "BGPPeer 'nsname' cannot be empty",
		},
		{
			name:          "bgppeer",
			namespace:     "test-namespace",
			peerIP:        "",
			asn:           5001,
			remoteAsn:     5002,
			expectedError: "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address",
		},
		{
			name:          "bgppeer",
			namespace:     "test-namespace",
			peerIP:        "192.168.1.1000",
			asn:           5001,
			remoteAsn:     5002,
			expectedError: "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{
			GVK: []schema.GroupVersionKind{bgpPeerGVK},
		})
		testBGPPeerBuilder := generateBPGPeer(
			testSettings, testCase.name, testCase.namespace, testCase.peerIP, testCase.asn, testCase.remoteAsn)
		assert.Equal(t, testCase.expectedError, testBGPPeerBuilder.errorMsg)
		assert.NotNil(t, testBGPPeerBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testBGPPeerBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testBGPPeerBuilder.Definition.Namespace)
		}
	}
}

func TestBGPPeerWithDynamicASN(t *testing.T) {
	testCases := []struct {
		testBGPPeer   *BGPPeerBuilder
		dynamicASN    string
		expectedError string
	}{
		{
			testBGPPeer:   buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			dynamicASN:    "internal",
			expectedError: "",
		},
		{
			testBGPPeer:   buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			dynamicASN:    "external",
			expectedError: "",
		},
		{
			testBGPPeer:   buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			dynamicASN:    "ebgp",
			expectedError: "bgpPeer 'dynamicASN' must be either internal or external",
		},
	}

	for _, testCase := range testCases {
		bgpPeerBuilder := testCase.testBGPPeer.WithDynamicASN(mlbtypesv1beta2.DynamicASNMode(testCase.dynamicASN))
		assert.Equal(t, testCase.expectedError, bgpPeerBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, mlbtypesv1beta2.DynamicASNMode(testCase.dynamicASN), bgpPeerBuilder.Definition.Spec.DynamicASN)
		}
	}
}

func TestBGPPeerWithRouterID(t *testing.T) {
	testCases := []struct {
		testBGPPeer   *BGPPeerBuilder
		routerID      string
		expectedError string
	}{
		{
			testBGPPeer: buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			routerID:    "1.1.1.1",
		},
		{
			testBGPPeer:   buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			routerID:      "test",
			expectedError: "the routerID of the BGPPeer contains invalid ip address test",
		},
		{
			testBGPPeer:   buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			routerID:      "",
			expectedError: "the routerID of the BGPPeer contains invalid ip address ",
		},
		{
			testBGPPeer:   buildInValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			routerID:      "1.1.1.1",
			expectedError: "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address",
		},
	}

	for _, testCase := range testCases {
		bgpPeerBuilder := testCase.testBGPPeer.WithRouterID(testCase.routerID)
		assert.Equal(t, testCase.expectedError, bgpPeerBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.routerID, bgpPeerBuilder.Definition.Spec.RouterID)
		}
	}
}

func TestBGPPeerWithBFDProfile(t *testing.T) {
	testCases := []struct {
		testBGPPeer   *BGPPeerBuilder
		bdfProfile    string
		expectedError string
	}{
		{
			testBGPPeer: buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			bdfProfile:  "testprofile",
		},
		{
			testBGPPeer:   buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			bdfProfile:    "",
			expectedError: "The bfdProfile is empty string",
		},
		{
			testBGPPeer:   buildInValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			bdfProfile:    "",
			expectedError: "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address",
		},
	}

	for _, testCase := range testCases {
		bgpPeerBuilder := testCase.testBGPPeer.WithBFDProfile(testCase.bdfProfile)
		assert.Equal(t, testCase.expectedError, bgpPeerBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.bdfProfile, bgpPeerBuilder.Definition.Spec.BFDProfile)
		}
	}
}

func TestBGPPeerWithSRCAddress(t *testing.T) {
	testCases := []struct {
		testBGPPeer   *BGPPeerBuilder
		srcAddress    string
		expectedError string
	}{
		{
			testBGPPeer: buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			srcAddress:  "1.1.1.1",
		},
		{
			testBGPPeer:   buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			srcAddress:    "",
			expectedError: "the srcAddress of the BGPPeer contains invalid ip address ",
		},
		{
			testBGPPeer:   buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			srcAddress:    "abc",
			expectedError: "the srcAddress of the BGPPeer contains invalid ip address abc",
		},
		{
			testBGPPeer:   buildInValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			srcAddress:    "1.1.1.1",
			expectedError: "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address",
		},
	}

	for _, testCase := range testCases {
		bgpPeerBuilder := testCase.testBGPPeer.WithSRCAddress(testCase.srcAddress)
		assert.Equal(t, testCase.expectedError, bgpPeerBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.srcAddress, bgpPeerBuilder.Definition.Spec.SrcAddress)
		}
	}
}

func TestBGPPeerWithPort(t *testing.T) {
	testCases := []struct {
		testBGPPeer   *BGPPeerBuilder
		port          uint16
		expectedError string
	}{
		{
			testBGPPeer: buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			port:        10,
		},
		{
			testBGPPeer:   buildInValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			port:          0,
			expectedError: "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address",
		},
	}

	for _, testCase := range testCases {
		bgpPeerBuilder := testCase.testBGPPeer.WithPort(testCase.port)
		assert.Equal(t, testCase.expectedError, bgpPeerBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.port, bgpPeerBuilder.Definition.Spec.Port)
		}
	}
}

func TestBGPPeerWithHoldTime(t *testing.T) {
	testCases := []struct {
		testBGPPeer   *BGPPeerBuilder
		holdTime      metav1.Duration
		expectedError string
	}{
		{
			testBGPPeer: buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			holdTime: metav1.Duration{
				Duration: 5 * time.Minute,
			},
		},
		{
			testBGPPeer: buildInValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			holdTime: metav1.Duration{
				Duration: 5 * time.Minute,
			},
			expectedError: "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address",
		},
	}

	for _, testCase := range testCases {
		bgpPeerBuilder := testCase.testBGPPeer.WithHoldTime(testCase.holdTime)
		assert.Equal(t, testCase.expectedError, bgpPeerBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.holdTime, *bgpPeerBuilder.Definition.Spec.HoldTime)
		}
	}
}

func TestBGPPeerWithKeepalive(t *testing.T) {
	testCases := []struct {
		testBGPPeer   *BGPPeerBuilder
		keepalive     metav1.Duration
		expectedError string
	}{
		{
			testBGPPeer: buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			keepalive: metav1.Duration{
				Duration: 5 * time.Minute,
			},
		},
		{
			testBGPPeer: buildInValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			keepalive: metav1.Duration{
				Duration: 5 * time.Minute,
			},
			expectedError: "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address",
		},
	}

	for _, testCase := range testCases {
		bgpPeerBuilder := testCase.testBGPPeer.WithKeepalive(testCase.keepalive)
		assert.Equal(t, testCase.expectedError, bgpPeerBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.keepalive, *bgpPeerBuilder.Definition.Spec.KeepaliveTime)
		}
	}
}

func TestBGPPeerWithNodeSelector(t *testing.T) {
	testCases := []struct {
		testBGPPeer   *BGPPeerBuilder
		nodeSelector  map[string]string
		expectedError string
	}{
		{
			testBGPPeer:  buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			nodeSelector: map[string]string{"test": "test1"},
		},
		{
			testBGPPeer:   buildInValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			nodeSelector:  map[string]string{"test": "test1"},
			expectedError: "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address",
		},
		{
			testBGPPeer:   buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			nodeSelector:  map[string]string{},
			expectedError: "BGPPeer 'nodeSelector' cannot be empty map",
		},
	}

	for _, testCase := range testCases {
		bgpPeerBuilder := testCase.testBGPPeer.WithNodeSelector(testCase.nodeSelector)
		assert.Equal(t, testCase.expectedError, bgpPeerBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, metav1.LabelSelector{MatchLabels: testCase.nodeSelector},
				bgpPeerBuilder.Definition.Spec.NodeSelectors[0])
		}
	}
}

func TestBGPPeerWithPassword(t *testing.T) {
	testCases := []struct {
		testBGPPeer   *BGPPeerBuilder
		password      string
		expectedError string
	}{
		{
			testBGPPeer: buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			password:    "test",
		},
		{
			testBGPPeer:   buildInValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			password:      "test",
			expectedError: "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address",
		},
		{
			testBGPPeer:   buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			password:      "",
			expectedError: "password can not be empty string",
		},
	}

	for _, testCase := range testCases {
		bgpPeerBuilder := testCase.testBGPPeer.WithPassword(testCase.password)
		assert.Equal(t, testCase.expectedError, bgpPeerBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.password, bgpPeerBuilder.Definition.Spec.Password)
		}
	}
}

func TestBGPPeerWithEBGPMultiHop(t *testing.T) {
	testCases := []struct {
		testBGPPeer   *BGPPeerBuilder
		ebgpMultiHop  bool
		expectedError string
	}{
		{
			testBGPPeer:  buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			ebgpMultiHop: false,
		},
		{
			testBGPPeer:  buildValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			ebgpMultiHop: true,
		},
		{
			testBGPPeer:   buildInValidBGPPeerBuilder(buildBGPPeerTestClientWithDummyObject()),
			ebgpMultiHop:  true,
			expectedError: "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address",
		},
	}

	for _, testCase := range testCases {
		bgpPeerBuilder := testCase.testBGPPeer.WithEBGPMultiHop(testCase.ebgpMultiHop)
		assert.Equal(t, testCase.expectedError, bgpPeerBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.ebgpMultiHop, bgpPeerBuilder.Definition.Spec.EBGPMultiHop)
		}
	}
}

func TestBGPPeerWithOptions(t *testing.T) {
	testSettings := buildBGPPeerTestClientWithDummyObject()
	testBuilder := buildValidBGPPeerBuilder(testSettings).WithOptions(
		func(builder *BGPPeerBuilder) (*BGPPeerBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidBGPPeerBuilder(testSettings).WithOptions(
		func(builder *BGPPeerBuilder) (*BGPPeerBuilder, error) {
			return builder, fmt.Errorf("error")
		})

	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestBGPPeerGVR(t *testing.T) {
	assert.Equal(t, GetBGPPeerGVR(),
		schema.GroupVersionResource{
			Group: APIGroup, Version: APIVersionv2, Resource: "bgppeers",
		})
}

func buildValidBGPPeerBuilder(apiClient *clients.Settings) *BGPPeerBuilder {
	return NewBPGPeerBuilder(apiClient, defaultBGPPeerName, defaultBGPPeerNsName, "192.168.1.1", 1000, 2000)
}

func buildInValidBGPPeerBuilder(apiClient *clients.Settings) *BGPPeerBuilder {
	return NewBPGPeerBuilder(apiClient, defaultBGPPeerName, defaultBGPPeerNsName, "", 1000, 2000)
}

func buildBGPPeerTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		// Work around. Dynamic client and Unstructured does not support unit.
		K8sMockObjects: buildDummyBFDProfile(),
		GVK:            []schema.GroupVersionKind{bgpPeerGVK},
	})
}
