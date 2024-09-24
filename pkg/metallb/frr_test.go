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
)

var (
	frrTestSchemes                = []clients.SchemeAttacher{frrtypes.AddToScheme}
	defaultFrrConfigurationName   = "frrconfiguration"
	defaultFrrConfigurationNsName = "test-namespace"
)

func TestNewFrrConfigurationBuilder(t *testing.T) {
	generateFrrConfiguration := NewFrrConfigurationBuilder

	testCases := []struct {
		name          string
		namespace     string
		client        bool
		expectedError string
	}{
		{
			name:          defaultFrrConfigurationName,
			namespace:     defaultFrrConfigurationNsName,
			client:        true,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     defaultFrrConfigurationNsName,
			client:        true,
			expectedError: "frrConfiguration 'name' cannot be empty",
		},
		{
			name:          defaultFrrConfigurationName,
			namespace:     "",
			client:        true,
			expectedError: "frrConfiguration 'nsname' cannot be empty",
		},
		{
			name:          defaultFrrConfigurationName,
			namespace:     defaultFrrConfigurationNsName,
			client:        false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testFRRConfigurationBuilder := generateFrrConfiguration(testSettings, testCase.name, testCase.namespace)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, testFRRConfigurationBuilder.errorMsg)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.name, testFRRConfigurationBuilder.Definition.Name)
			}
		} else {
			assert.Nil(t, testFRRConfigurationBuilder)
		}
	}
}

func TestFrrConfigurationCreate(t *testing.T) {
	testCases := []struct {
		testFrrConfiguration *FrrConfigurationBuilder
		expectedError        error
	}{
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			expectedError:        nil,
		},
		{
			testFrrConfiguration: buildInValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			expectedError:        fmt.Errorf("frrConfiguration 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		testFrrConfigBuilder, err := testCase.testFrrConfiguration.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testFrrConfigBuilder.Definition.Name, testFrrConfigBuilder.Object.Name)
		}
	}
}

func TestFrrConfigurationDelete(t *testing.T) {
	testCases := []struct {
		testFrrConfiguration *FrrConfigurationBuilder
		expectedError        error
	}{
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			expectedError:        nil,
		},
		{
			testFrrConfiguration: buildInValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			expectedError:        fmt.Errorf("frrConfiguration 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testFrrConfiguration.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testFrrConfiguration.Object)
		}
	}
}

func TestFrrConfigurationGet(t *testing.T) {
	testCases := []struct {
		testFrrConfiguration *FrrConfigurationBuilder
		expectedError        string
	}{
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			expectedError:        "",
		},
		{
			testFrrConfiguration: buildInValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			expectedError:        "frrConfiguration 'name' cannot be empty",
		},
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:        "frrconfigurations.frrk8s.metallb.io \"frrconfiguration\" not found",
		},
	}

	for _, testCase := range testCases {
		frrConfig, err := testCase.testFrrConfiguration.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, frrConfig.Name, testCase.testFrrConfiguration.Definition.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestFrrConfigurationExist(t *testing.T) {
	testCases := []struct {
		testFrrConfiguration *FrrConfigurationBuilder
		expectedStatus       bool
	}{
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			expectedStatus:       true,
		},
		{
			testFrrConfiguration: buildInValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			expectedStatus:       false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testFrrConfiguration.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestFrrConfigurationWithBGPRouter(t *testing.T) {
	testCases := []struct {
		testFrrConfiguration *FrrConfigurationBuilder
		localASN             uint32
		expectedError        string
	}{
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			localASN:             64500,
			expectedError:        "",
		},
	}

	for _, testCase := range testCases {
		frrConfigurationBuilder := testCase.testFrrConfiguration.WithBGPRouter(testCase.localASN)
		assert.Equal(t, testCase.expectedError, frrConfigurationBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.localASN, frrConfigurationBuilder.Definition.Spec.BGP.Routers[0].ASN)
		}
	}
}

func TestFrrConfigurationBGPNeighbor(t *testing.T) {
	testCases := []struct {
		testFrrConfiguration *FrrConfigurationBuilder
		bgpPeer              string
		expectedError        string
	}{
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			bgpPeer:              "10.46.71.131",
			expectedError:        "",
		},
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			bgpPeer:              "10.46.71.",
			expectedError:        "frrConfiguration 'peerIP' of the BGPPeer contains invalid ip address",
		},
	}

	for _, testCase := range testCases {
		frrConfigurationBuilder := testCase.testFrrConfiguration.WithBGPRouter(64500).
			WithBGPNeighbor(testCase.bgpPeer, 64500, 0)
		assert.Equal(t, testCase.expectedError, frrConfigurationBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.bgpPeer, frrConfigurationBuilder.Definition.Spec.BGP.Routers[0].
				Neighbors[0].Address)
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
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			bgpPassword:          "bgpPassword",
			expectedError:        "",
		},
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			bgpPassword:          "",
			expectedError:        "the bgpPassword  is an empty string",
		},
	}

	for _, testCase := range testCases {
		frrConfigurationBuilder := testCase.testFrrConfiguration.WithBGPRouter(64550).
			WithBGPNeighbor("10.46.73.131", 64500, 0).
			WithBGPPassword(testCase.bgpPassword, 0, 0)
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
		holdTime             *metav1.Duration
		expectedError        string
	}{
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			holdTime: &metav1.Duration{
				Duration: 90 * time.Second,
			},
		},
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			holdTime: &metav1.Duration{
				Duration: 0 * time.Minute,
			},
			expectedError: "frrConfiguration 'holdtime' value is not valid",
		},
	}

	for _, testCase := range testCases {
		frrConfigurationBuilder := testCase.testFrrConfiguration.WithBGPRouter(64550).
			WithBGPNeighbor("10.46.73.131", 64500, 0).WithHoldTime(*testCase.holdTime, 0, 0)
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
		keepalive            *metav1.Duration
		expectedError        string
	}{
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			keepalive: &metav1.Duration{
				Duration: 30 * time.Second,
			},
			expectedError: "",
		},
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			keepalive: &metav1.Duration{
				Duration: 0 * time.Second,
			},
			expectedError: "frrConfiguration 'keepAlive' value is not valid",
		},
	}

	for _, testCase := range testCases {
		frrConfigurationBuilder := testCase.testFrrConfiguration.WithBGPRouter(64550).
			WithBGPNeighbor("10.46.73.131", 64500, 0).WithKeepalive(*testCase.keepalive, 0, 0)
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
		connectTime          *metav1.Duration
		expectedError        string
	}{
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			connectTime: &metav1.Duration{
				Duration: 10 * time.Second,
			},
			expectedError: "",
		},
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			connectTime: &metav1.Duration{
				Duration: 0 * time.Second,
			},
			expectedError: "frrConfiguration 'connectTime' value is not valid",
		},
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			connectTime: &metav1.Duration{
				Duration: 65555 * time.Second,
			},
			expectedError: "frrConfiguration 'connectTime' value is not valid",
		},
	}

	for _, testCase := range testCases {
		frrConfigurationBuilder := testCase.testFrrConfiguration.WithBGPRouter(64550).
			WithBGPNeighbor("10.46.73.131", 64500, 0).
			WithConnectTime(*testCase.connectTime, 0, 0)
		assert.Equal(t, testCase.expectedError, frrConfigurationBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.connectTime,
				frrConfigurationBuilder.Definition.Spec.BGP.Routers[0].Neighbors[0].ConnectTime)
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
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			ebgpMultiHop:         true,
			expectedError:        "",
		},
	}

	for _, testCase := range testCases {
		frrConfigurationBuilder := testCase.testFrrConfiguration.WithBGPRouter(64550).
			WithBGPNeighbor("10.46.73.131", 64500, 0).WithEBGPMultiHop(0, 0)
		assert.Equal(t, testCase.expectedError, frrConfigurationBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.ebgpMultiHop,
				frrConfigurationBuilder.Definition.Spec.BGP.Routers[0].Neighbors[0].EBGPMultiHop)
		}
	}
}

func TestFrrConfigurationWithPort(t *testing.T) {
	var (
		portIDpass uint16 = 179
		portIDfail uint16 = 16385
	)

	testCases := []struct {
		testFrrConfiguration *FrrConfigurationBuilder
		port                 *uint16
		expectedError        string
	}{
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			port:                 &portIDpass,
			expectedError:        "",
		},
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			port:                 &portIDfail,
			expectedError:        "invalid port number: 16385",
		},
	}

	for _, testCase := range testCases {
		frrConfigurationBuilder := testCase.testFrrConfiguration.WithBGPRouter(64550).
			WithBGPNeighbor("10.46.73.131", 64500, 0).
			WithPort(*testCase.port, 0, 0)
		assert.Equal(t, testCase.expectedError, frrConfigurationBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.port,
				frrConfigurationBuilder.Definition.Spec.BGP.Routers[0].Neighbors[0].Port)
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
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			mode:                 "all",
			expectedError:        "",
		},
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			mode:                 "allowAll",
			expectedError:        "",
		},
	}

	for _, testCase := range testCases {
		frrConfigurationBuilder := testCase.testFrrConfiguration.WithBGPRouter(64550).
			WithBGPNeighbor("10.46.73.131", 64500, 0).WithToReceiveModeAll(0, 0)
		assert.Equal(t, testCase.expectedError, frrConfigurationBuilder.errorMsg)

		if testCase.expectedError == "all" {
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
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			prefix:               "192.168.1.0/24",
			expectedError:        "",
		},
		{
			testFrrConfiguration: buildValidFRRConfigurationBuilder(buildFrrConfigurationTestClientWithDummyObject()),
			prefix:               "192.168.1.",
			expectedError:        "the prefix 192.168.1. is not a valid CIDR",
		},
	}

	for _, testCase := range testCases {
		frrConfigurationBuilder := testCase.testFrrConfiguration.WithBGPRouter(64500).
			WithBGPNeighbor("10.46.73.131", 64500, 0).
			WithToReceiveModeFiltered([]string{testCase.prefix}, 0, 0)
		assert.Equal(t, testCase.expectedError, frrConfigurationBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.prefix,
				frrConfigurationBuilder.Definition.Spec.BGP.Routers[0].Neighbors[0].ToReceive.Allowed.Prefixes[0].Prefix)
		}
	}
}

func buildValidFRRConfigurationBuilder(apiClient *clients.Settings) *FrrConfigurationBuilder {
	return NewFrrConfigurationBuilder(apiClient, defaultFrrConfigurationName, defaultFrrConfigurationNsName)
}

func buildInValidFRRConfigurationBuilder(apiClient *clients.Settings) *FrrConfigurationBuilder {
	return NewFrrConfigurationBuilder(apiClient, "", defaultFrrConfigurationNsName)
}

func buildFrrConfigurationTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyFrrConfig(defaultFrrConfigurationName),
		},
		SchemeAttachers: frrTestSchemes,
	})
}

// buildDummyFrrConfig returns a FrrConfiguration with the provided name.
func buildDummyFrrConfig(name string) *frrtypes.FRRConfiguration {
	return &frrtypes.FRRConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: defaultFrrConfigurationNsName,
		},
	}
}
