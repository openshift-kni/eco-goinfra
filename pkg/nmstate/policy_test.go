package nmstate

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/nmstate/kubernetes-nmstate/api/shared"
	nmstatev1 "github.com/nmstate/kubernetes-nmstate/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultPolicyName = "policyname"
)

func TestNewPolicyBuilder(t *testing.T) {
	testCases := []struct {
		name              string
		nodeSelector      map[string]string
		expectedErrorText string
		client            bool
	}{
		{
			name:              "test1",
			nodeSelector:      map[string]string{"test": "test1"},
			client:            true,
			expectedErrorText: "",
		},
		{
			name:              "",
			nodeSelector:      map[string]string{"test": "test1"},
			expectedErrorText: "nodeNetworkConfigurationPolicy 'name' cannot be empty",
			client:            true,
		},
		{
			name:              "test1",
			nodeSelector:      map[string]string{},
			expectedErrorText: "nodeNetworkConfigurationPolicy 'nodeSelector' cannot be empty map",
			client:            true,
		},
		{
			name:              "test1",
			nodeSelector:      map[string]string{"test": "test1"},
			expectedErrorText: "",
			client:            false,
		},
	}
	for _, testCase := range testCases {
		var (
			client *clients.Settings
		)

		if testCase.client {
			client = clients.GetTestClients(clients.TestClientParams{})
		}

		testPolicy := NewPolicyBuilder(client, testCase.name, testCase.nodeSelector)
		if testCase.client {
			assert.NotNil(t, testPolicy)
		}

		if len(testCase.expectedErrorText) > 0 {
			assert.Equal(t, testCase.expectedErrorText, testPolicy.errorMsg)
		}
	}
}

func TestPolicyGet(t *testing.T) {
	testCases := []struct {
		policyBuilder *PolicyBuilder
		expectedError error
	}{
		{
			policyBuilder: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError: nil,
		},
		{
			policyBuilder: buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError: fmt.Errorf("nodeNetworkConfigurationPolicy 'nodeSelector' cannot be empty map"),
		},
		{
			policyBuilder: buildValidPolicyTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("nodenetworkconfigurationpolicies.nmstate.io \"policyname\" not found"),
		},
	}

	for _, testCase := range testCases {
		policyBuilder, err := testCase.policyBuilder.Get()
		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedError, err)
		}

		if testCase.expectedError == nil {
			assert.Equal(t, policyBuilder.Name, testCase.policyBuilder.Definition.Name)
		}
	}
}

func TestPolicyExist(t *testing.T) {
	testCases := []struct {
		testPolicy     *PolicyBuilder
		expectedStatus bool
	}{
		{
			testPolicy:     buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedStatus: true,
		},
		{
			testPolicy:     buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedStatus: false,
		},
		{
			testPolicy:     buildValidPolicyTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.expectedStatus, testCase.testPolicy.Exists())
	}
}

func TestPolicyCreate(t *testing.T) {
	testCases := []struct {
		testPolicy    *PolicyBuilder
		expectedError error
	}{
		{
			testPolicy:    buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError: nil,
		},
		{
			testPolicy:    buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError: fmt.Errorf("nodeNetworkConfigurationPolicy 'nodeSelector' cannot be empty map"),
		},
		{
			testPolicy:    buildValidPolicyTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		nmStatePolicyBuilder, err := testCase.testPolicy.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, nmStatePolicyBuilder.Definition.Name, nmStatePolicyBuilder.Object.Name)
		}
	}
}

func TestPolicyDelete(t *testing.T) {
	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		expectedError     error
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nil,
		},
		{
			testNMStatePolicy: buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     fmt.Errorf("nodeNetworkConfigurationPolicy 'nodeSelector' cannot be empty map"),
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:     nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testNMStatePolicy.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testNMStatePolicy.Object)
		}
	}
}

func TestPolicyUpdate(t *testing.T) {
	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		expectedError     error
		forceFlag         bool
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nil,
			forceFlag:         false,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyObject()),
			expectedError:     nil,
			forceFlag:         true,
		},
	}
	for _, testCase := range testCases {
		assert.Equal(t, map[string]string{"test": "test"}, testCase.testNMStatePolicy.Definition.Spec.NodeSelector)
		assert.Nil(t, nil, testCase.testNMStatePolicy.Object)
		testCase.testNMStatePolicy.Definition.Spec.NodeSelector = map[string]string{"test2": "test2"}

		if !testCase.forceFlag {
			testCase.testNMStatePolicy.Definition.ObjectMeta.ResourceVersion = "999"
		}

		nmStatePolicyBuilder, err := testCase.testNMStatePolicy.Update(testCase.forceFlag)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, map[string]string{"test2": "test2"}, testCase.testNMStatePolicy.Object.Spec.NodeSelector)
			assert.Equal(t, nmStatePolicyBuilder.Definition, nmStatePolicyBuilder.Object)
		}
	}
}

func TestPolicyWithInterfaceAndVFs(t *testing.T) {
	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		expectedError     string
		sriovInterface    string
		numberOfVF        uint8
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			sriovInterface:    "ens1",
			numberOfVF:        10,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "The sriovInterface is empty string",
			sriovInterface:    "",
			numberOfVF:        10,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			sriovInterface:    "ens1",
			numberOfVF:        0,
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.WithInterfaceAndVFs(testCase.sriovInterface, testCase.numberOfVF)
		assert.Equal(t, testCase.expectedError, testPolicy.errorMsg)
		numberOfVFs := int(testCase.numberOfVF)
		desireState := &DesiredState{}

		if testCase.expectedError == "" {
			_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)
			assert.Equal(t, desireState, &DesiredState{
				Interfaces: []NetworkInterface{
					{
						Name:  testCase.sriovInterface,
						Type:  "ethernet",
						State: "up",
						Ethernet: Ethernet{
							Sriov: Sriov{
								TotalVfs: &numberOfVFs,
							},
						},
					},
				},
			})
		}
	}
}

func TestPolicyWithWithBondInterface(t *testing.T) {
	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		slavePorts        []string
		bondName          string
		mode              string
		expectedError     string
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			slavePorts:        []string{"ens1", "ens2"},
			bondName:          "bd1",
			mode:              "active-backup",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "The bondName is empty sting",
			slavePorts:        []string{"ens1", "ens2"},
			bondName:          "",
			mode:              "active-backup",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "invalid Bond mode parameter",
			slavePorts:        []string{"ens1", "ens2"},
			bondName:          "bd1",
			mode:              "",
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.WithBondInterface(testCase.slavePorts, testCase.bondName, testCase.mode)
		assert.Equal(t, testCase.expectedError, testPolicy.errorMsg)

		desireState := &DesiredState{}
		if testCase.expectedError == "" {
			_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)
			assert.Equal(t, desireState, &DesiredState{
				Interfaces: []NetworkInterface{
					{
						Name:  testCase.bondName,
						Type:  "bond",
						State: "up",
						LinkAggregation: LinkAggregation{
							Mode: testCase.mode,
							Port: testCase.slavePorts,
						},
					},
				},
			})
		}
	}
}

func TestPolicyWithVlanInterface(t *testing.T) {
	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		expectedError     string
		sriovInterface    string
		vlanID            uint16
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			sriovInterface:    "ens1",
			vlanID:            10,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "nodenetworkconfigurationpolicy 'baseInterface' cannot be empty",
			sriovInterface:    "",
			vlanID:            10,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "invalid vlanID, allowed vlanID values are between 0-4094",
			sriovInterface:    "ens1",
			vlanID:            4099,
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.WithVlanInterface(testCase.sriovInterface, testCase.vlanID)
		assert.Equal(t, testCase.expectedError, testPolicy.errorMsg)

		desireState := &DesiredState{}
		if testCase.expectedError == "" {
			_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)
			assert.Equal(t, desireState, &DesiredState{
				Interfaces: []NetworkInterface{
					{
						Name:  fmt.Sprintf("%s.%d", testCase.sriovInterface, testCase.vlanID),
						Type:  "vlan",
						State: "up",
						Vlan: Vlan{
							BaseIface: testCase.sriovInterface,
							ID:        int(testCase.vlanID),
						},
					},
				},
			})
		}
	}
}

func TestPolicyWithVlanInterfaceIP(t *testing.T) {
	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		expectedError     string
		sriovInterface    string
		vlanID            uint16
		ipv4              string
		ipv6              string
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			sriovInterface:    "ens1",
			vlanID:            10,
			ipv4:              "10.10.10.10",
			ipv6:              "2001:db8::68",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "nodenetworkconfigurationpolicy 'baseInterface' cannot be empty",
			sriovInterface:    "",
			vlanID:            10,
			ipv4:              "10.10.10.10",
			ipv6:              "2001:db8::68",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "invalid vlanID, allowed vlanID values are between 0-4094",
			sriovInterface:    "ens1",
			vlanID:            4099,
			ipv4:              "10.10.10.10",
			ipv6:              "2001:db8::68",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "vlanInterfaceIP 'ipv4Addresses' is an invalid ipv4 address",
			sriovInterface:    "ens1",
			vlanID:            10,
			ipv4:              "",
			ipv6:              "2001:db8::68",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "vlanInterfaceIP 'ipv6Addresses' is an invalid ipv6 address",
			sriovInterface:    "ens1",
			vlanID:            10,
			ipv4:              "10.10.10.10",
			ipv6:              "",
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.WithVlanInterfaceIP(testCase.sriovInterface,
			testCase.ipv4, testCase.ipv6, testCase.vlanID)
		assert.Equal(t, testCase.expectedError, testPolicy.errorMsg)

		desireState := &DesiredState{}
		if testCase.expectedError == "" {
			_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)
			assert.Equal(t, &DesiredState{
				Interfaces: []NetworkInterface{
					{
						Name:  fmt.Sprintf("%s.%d", testCase.sriovInterface, testCase.vlanID),
						Type:  "vlan",
						State: "up",
						Vlan: Vlan{
							BaseIface: testCase.sriovInterface,
							ID:        int(testCase.vlanID),
						},
						Ipv4: InterfaceIpv4{
							Enabled: true,
							Address: []InterfaceIPAddress{{
								PrefixLen: 24,
								IP:        net.ParseIP(testCase.ipv4),
							}},
						},
						Ipv6: InterfaceIpv6{Enabled: true,
							Address: []InterfaceIPAddress{{
								PrefixLen: 64,
								IP:        net.ParseIP(testCase.ipv6),
							}},
						},
					},
				},
			}, desireState)
		}
	}
}

func TestPolicyWithAbsentInterface(t *testing.T) {
	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		expectedError     string
		baseInterface     string
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			baseInterface:     "ens1",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "nodenetworkconfigurationpolicy 'interfaceName' cannot be empty",
			baseInterface:     "",
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.WithAbsentInterface(testCase.baseInterface)
		assert.Equal(t, testCase.expectedError, testPolicy.errorMsg)

		desireState := &DesiredState{}
		if testCase.expectedError == "" {
			_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)
			assert.Equal(t, desireState, &DesiredState{
				Interfaces: []NetworkInterface{{Name: testCase.baseInterface, State: "absent"}}})
		}
	}
}

func TestPolicyWithWithOptions(t *testing.T) {
	testSettings := buildTestClientWithDummyPolicyObject()
	testBuilder := buildValidPolicyTestBuilder(testSettings).WithOptions(
		func(builder *PolicyBuilder) (*PolicyBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidPolicyTestBuilder(testSettings).WithOptions(
		func(builder *PolicyBuilder) (*PolicyBuilder, error) {
			return builder, fmt.Errorf("error")
		})

	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestPolicyWaitUntilCondition(t *testing.T) {
	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		expectedError     error
		condition         shared.ConditionType
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nil,
			condition:         shared.NodeNetworkStateConditionAvailable,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     context.DeadlineExceeded,
			condition:         shared.NodeNetworkConfigurationEnactmentConditionFailing,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:     fmt.Errorf("cannot wait for NodeNetworkConfigurationPolicy condition because it does not exist"),
			condition:         shared.NodeNetworkStateConditionAvailable,
		},
	}
	for _, testCase := range testCases {
		err := testCase.testNMStatePolicy.WaitUntilCondition(testCase.condition, 2*time.Second)
		assert.Equal(t, testCase.expectedError, err)
	}
}

// buildValidTestBuilder returns a valid Builder for testing purposes.
func buildValidPolicyTestBuilder(apiClient *clients.Settings) *PolicyBuilder {
	return NewPolicyBuilder(apiClient, defaultPolicyName, map[string]string{"test": "test"})
}

// buildValidTestBuilder returns a valid Builder for testing purposes.
func buildInValidPolicyTestBuilder(apiClient *clients.Settings) *PolicyBuilder {
	return NewPolicyBuilder(apiClient, "test", map[string]string{})
}

func buildTestClientWithDummyPolicyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyPolicyObject(),
		SchemeAttachers: v1TestSchemes,
	})
}

func buildDummyPolicyObject() []runtime.Object {
	return append([]runtime.Object{}, &nmstatev1.NodeNetworkConfigurationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultPolicyName,
		},
		Spec: shared.NodeNetworkConfigurationPolicySpec{},
		Status: shared.NodeNetworkConfigurationPolicyStatus{
			Conditions: shared.ConditionList{
				shared.Condition{
					Type:   shared.NodeNetworkStateConditionAvailable,
					Status: corev1.ConditionTrue,
				},
			},
		},
	})
}
