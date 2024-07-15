package metallb

import (
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/schemes/metallb/frrtypes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	frrConfigurationGVK = schema.GroupVersionKind{
		Group:   APIGroup,
		Version: APIVersion,
		Kind:    frrConfigurationKind,
	}
	defaultFrrConfigurationName   = "default-frrconfiguration"
	defaultFrrConfigurationNsName = "test-namespace"
)

func TestNewFrrConfigurationBuilder(t *testing.T) {
	generateFrrConfiguration := NewFrrConfigurationBuilder

	testCases := []struct {
		name          string
		namespace     string
		peerIP        string
		localAsn      uint32
		remoteAsn     uint32
		IPPrefix      string
		expectedError string
	}{
		{
			name:          "frrconfiguration",
			namespace:     "test-namespace",
			peerIP:        "192.168.1.1",
			localAsn:      5001,
			remoteAsn:     5002,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     "test-namespace",
			peerIP:        "192.168.1.1",
			localAsn:      5001,
			remoteAsn:     5002,
			expectedError: "FrrConfiguration 'name' cannot be empty",
		},
		{
			name:          "frrconfiguration",
			namespace:     "",
			peerIP:        "192.168.1.1",
			localAsn:      5001,
			remoteAsn:     5002,
			expectedError: "FrrConfiguration 'nsname' cannot be empty",
		},
		{
			name:          "frrconfiguration",
			namespace:     "test-namespace",
			peerIP:        "",
			localAsn:      5001,
			remoteAsn:     5002,
			expectedError: "FrrConfiguration 'peerIP' of the BGPPeer contains invalid ip address",
		},
		{
			name:          "frrconfiguration",
			namespace:     "test-namespace",
			peerIP:        "192.168.1.1000",
			localAsn:      5001,
			remoteAsn:     5002,
			expectedError: "FrrConfiguration 'peerIP' of the BGPPeer contains invalid ip address",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{
			GVK: []schema.GroupVersionKind{frrConfigurationGVK},
		})
		testFRRConfigurationBuilder := generateFrrConfiguration(
			testSettings, testCase.name, testCase.namespace, testCase.peerIP, testCase.localAsn, testCase.remoteAsn)
		assert.Equal(t, testCase.expectedError, testFRRConfigurationBuilder.errorMsg)
		assert.NotNil(t, testFRRConfigurationBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testFRRConfigurationBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testFRRConfigurationBuilder.Definition.Namespace)
		}
	}
}

func TestFrrConfigurationWithBGPPassword(t *testing.T) {
	testCases := []struct {
		testFrrConfiguration *FrrConfigurationBuilder
		bgpPassword          string
		expectedError        string
	}{
		{
			testFrrConfiguration: buildValidConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			bgpPassword:          "bgpPassword",
		},
		{
			testFrrConfiguration: buildInValidFrrConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			bgpPassword:          "test",
			expectedError:        "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address",
		},
		{
			testFrrConfiguration: buildValidConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			bgpPassword:          "",
			expectedError:        "password can not be empty string",
		},
	}

	for _, testCase := range testCases {
		frrConfigurationBuilder := testCase.testFrrConfiguration.WithBGPPassword(testCase.bgpPassword)
		assert.Equal(t, testCase.expectedError, frrConfigurationBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.bgpPassword,
				frrConfigurationBuilder.Definition.Spec.BGP.Routers[0].Neighbors[0].Password)
		}
	}
}

func TestFrrConfigurationWithHoldTime(t *testing.T) {
	testCases := []struct {
		testFrrConfiguration *FrrConfigurationBuilder
		holdTime             metav1.Duration
		expectedError        string
	}{
		{
			testFrrConfiguration: buildValidConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			holdTime: metav1.Duration{
				Duration: 90 * time.Second,
			},
		},
		{
			testFrrConfiguration: buildInValidFrrConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			holdTime: metav1.Duration{
				Duration: 0 * time.Minute,
			},
			expectedError: "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address",
		},
	}

	for _, testCase := range testCases {
		frrConfigurationBuilder := testCase.testFrrConfiguration.WithHoldTime(testCase.holdTime)
		assert.Equal(t, testCase.expectedError, frrConfigurationBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.holdTime,
				frrConfigurationBuilder.Definition.Spec.BGP.Routers[0].Neighbors[0].HoldTime)
		}
	}
}

func TestFrrConfigurationWithKeepalive(t *testing.T) {
	testCases := []struct {
		testFrrConfiguration *FrrConfigurationBuilder
		keepalive            metav1.Duration
		expectedError        string
	}{
		{
			testFrrConfiguration: buildValidConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			keepalive: metav1.Duration{
				Duration: 30 * time.Second,
			},
		},
		{
			testFrrConfiguration: buildInValidFrrConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			keepalive: metav1.Duration{
				Duration: 0 * time.Second,
			},
			expectedError: "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address",
		},
	}

	for _, testCase := range testCases {
		frrConfigurationBuilder := testCase.testFrrConfiguration.WithKeepalive(testCase.keepalive)
		assert.Equal(t, testCase.expectedError, frrConfigurationBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.keepalive,
				frrConfigurationBuilder.Definition.Spec.BGP.Routers[0].Neighbors[0].KeepaliveTime)
		}
	}
}

func TestFrrConfigurationWithConnectTime(t *testing.T) {
	testCases := []struct {
		testFrrConfiguration *FrrConfigurationBuilder
		connectTime          metav1.Duration
		expectedError        string
	}{
		{
			testFrrConfiguration: buildValidConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			connectTime: metav1.Duration{
				Duration: 120 * time.Second,
			},
		},
		{
			testFrrConfiguration: buildInValidFrrConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			connectTime: metav1.Duration{
				Duration: 0 * time.Second,
			},
			expectedError: "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address",
		},
	}

	for _, testCase := range testCases {
		frrConfigurationBuilder := testCase.testFrrConfiguration.WithHoldTime(testCase.connectTime)
		assert.Equal(t, testCase.expectedError, frrConfigurationBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.connectTime,
				frrConfigurationBuilder.Definition.Spec.BGP.Routers[0].Neighbors[0].HoldTime)
		}
	}
}

func TestFrrConfigurationWithEBGPMultiHop(t *testing.T) {
	testCases := []struct {
		testFrrConfiguration *FrrConfigurationBuilder
		ebgpMultiHop         bool
		expectedError        string
	}{
		{
			testFrrConfiguration: buildValidConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			ebgpMultiHop:         true,
		},
		{
			testFrrConfiguration: buildInValidFrrConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			expectedError:        "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address",
		},
	}

	for _, testCase := range testCases {
		frrConfigurationBuilder := testCase.testFrrConfiguration.WithEBGPMultiHop(testCase.ebgpMultiHop)
		assert.Equal(t, testCase.expectedError, frrConfigurationBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.ebgpMultiHop,
				frrConfigurationBuilder.Definition.Spec.BGP.Routers[0].Neighbors[0].HoldTime)
		}
	}
}

func TestFrrConfigurationWithToReceiveModeAll(t *testing.T) {
	testCases := []struct {
		testFrrConfiguration *FrrConfigurationBuilder
		mode                 string
		expectedError        string
	}{
		{
			testFrrConfiguration: buildValidConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			mode:                 "all",
		},
		{
			testFrrConfiguration: buildValidConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			mode:                 "allowAll",
			expectedError:        "toReceive allowed mode invalid value can only be set to all or filtered",
		},
	}

	for _, testCase := range testCases {
		frrConfigurationBuilder := testCase.testFrrConfiguration.WithToReceiveModeAll(testCase.mode)
		assert.Equal(t, testCase.expectedError, frrConfigurationBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.mode,
				frrConfigurationBuilder.Definition.Spec.BGP.Routers[0].Neighbors[0].ToReceive.Allowed.Mode)
		}
	}
}

func TestFrrConfigurationWithToReceiveModeFiltered(t *testing.T) {
	testCases := []struct {
		testFrrConfiguration *FrrConfigurationBuilder
		prefix               string
		expectedError        string
	}{
		{
			testFrrConfiguration: buildValidConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			prefix:               "192.168.100.0/24",
		},
		{
			testFrrConfiguration: buildInValidFrrConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			prefix:               "192.168.100.0/24",
			expectedError:        "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address",
		},
		{
			testFrrConfiguration: buildValidConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			prefix:               "",
			expectedError:        "Frrconfiguration allow to receive route 'prefix' is an invalid ip address",
		},
	}

	neighbor := FrrConfigurationBuilder{}.Definition.Spec.BGP.Routers[0].Neighbors

	for _, testCase := range testCases {
		frrConfigurationBuilder := testCase.testFrrConfiguration.WithToReceiveModeFiltered(neighbor[0],
			[]string{testCase.prefix})
		assert.Equal(t, testCase.expectedError, frrConfigurationBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.prefix,
				frrConfigurationBuilder.Definition.Spec.BGP.Routers[0].Neighbors[0].ToReceive.Allowed.Prefixes[0].Prefix)
		}
	}
}

func TestFrrConfigurationGVR(t *testing.T) {
	assert.Equal(t, GetFrrConfigurationGVR(),
		schema.GroupVersionResource{
			Group: APIGroup, Version: APIVersion, Resource: "frrconfigurations",
		})
}

func TestFrrConfigurationExist(t *testing.T) {
	testCases := []struct {
		testFrrConfiguration *Builder
		expectedStatus       bool
	}{
		{
			testFrrConfiguration: buildValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedStatus:       true,
		},
		{
			testFrrConfiguration: buildInValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedStatus:       false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testFrrConfiguration.Exists()
		assert.Equal(t, exist, testCase.expectedStatus)
	}
}

func TestFrrConfigurationGet(t *testing.T) {
	testCases := []struct {
		testFrrConfiguration *Builder
		expectedError        error
	}{
		{
			testFrrConfiguration: buildValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError:        nil,
		},
		{
			testFrrConfiguration: buildInValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError:        fmt.Errorf("FrrConfiguration 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		metalLb, err := testCase.testFrrConfiguration.Get()
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.Equal(t, metalLb, testCase.testFrrConfiguration.Definition)
		}
	}
}

func TestFrrConfigurationCreate(t *testing.T) {
	testCases := []struct {
		testFrrConfiguration *Builder
		expectedError        error
	}{
		{
			testFrrConfiguration: buildValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError:        nil,
		},
		{
			testFrrConfiguration: buildInValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError:        fmt.Errorf("FrrConfiguration 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		testFRRConfigurationBuilder, err := testCase.testFrrConfiguration.Create()
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.Equal(t, testFRRConfigurationBuilder.Definition, testFRRConfigurationBuilder.Object)
		}
	}
}

func TestFrrConfigurationDelete(t *testing.T) {
	testCases := []struct {
		testFrrConfiguration *Builder
		expectedError        error
	}{
		{
			testFrrConfiguration: buildValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError:        nil,
		},
		{
			testFrrConfiguration: buildInValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError:        fmt.Errorf("FrrConfiguration 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testFrrConfiguration.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testFrrConfiguration.Object)
		}
	}
}

func buildValidConfigurationBuilder(apiClient *clients.Settings) *FrrConfigurationBuilder {
	return NewFrrConfigurationBuilder(apiClient, defaultFrrConfigurationName, defaultFrrConfigurationNsName,
		"192.168.1.1", 1000, 2000)
}

func buildInValidFrrConfigurationBuilder(apiClient *clients.Settings) *FrrConfigurationBuilder {
	return NewFrrConfigurationBuilder(apiClient, defaultFrrConfigurationName, defaultFrrConfigurationNsName, "",
		1000, 2000)
}

func buildFrrConfigurationTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		// Work around. Dynamic client and Unstructured does not support unit.
		K8sMockObjects: buildDummyFRRProfile(),
		GVK:            []schema.GroupVersionKind{frrConfigurationGVK},
	})
}

func buildDummyFRRProfile() []runtime.Object {
	return append([]runtime.Object{}, &frrtypes.FRRConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "frrconfiguration",
			Namespace: "test-namespace",
		},
		Spec: frrtypes.FRRConfigurationSpec{},
	})
}
