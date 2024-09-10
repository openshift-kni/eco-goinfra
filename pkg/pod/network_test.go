package pod

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/k8snetworkplumbingwg/multus-cni.v4/pkg/types"
)

func TestPodNetworkStaticAnnotation(t *testing.T) {
	testCases := []struct {
		annotationName     string
		expectedAnnotation *types.NetworkSelectionElement
	}{
		{
			annotationName: "test",
		},
		{
			annotationName: "",
		},
	}
	for _, testCase := range testCases {
		annotation := StaticAnnotation(testCase.annotationName)
		if testCase.annotationName == "" {
			assert.Nil(t, annotation)
		} else {
			assert.NotNil(t, annotation)
		}
	}
}

func TestPodNetworkStaticIPAnnotation(t *testing.T) {
	testCases := []struct {
		annotationName string
		ipAddr         []string
		expectedNil    bool
	}{
		{
			annotationName: "test",
			ipAddr:         []string{"192.168.1.1"},
			expectedNil:    false,
		},
		{
			annotationName: "test",
			ipAddr:         []string{"192.168.1.1/24", "2001:0000:130F:0000:0000:09C0:876A:130B"},
			expectedNil:    false,
		},
		{
			annotationName: "test",
			ipAddr:         []string{"192.168.1.1/24", "2001:0000:130F:0000:0000:09C0:876A:130B/64"},
			expectedNil:    false,
		},
		{
			annotationName: "",
			ipAddr:         []string{"192.168.1.1/24"},
			expectedNil:    true,
		},
		{
			annotationName: "test",
			ipAddr:         []string{},
			expectedNil:    false,
		},
		{
			annotationName: "test",
			ipAddr:         []string{"192.168.1.1/24", "ABC"},
			expectedNil:    true,
		},
	}
	for _, testCase := range testCases {
		annotation := StaticIPAnnotation(testCase.annotationName, testCase.ipAddr)
		if testCase.expectedNil {
			assert.Nil(t, annotation)
		} else {
			assert.NotNil(t, annotation)
		}
	}
}

func TestPodNetworkStaticIPAnnotationWithNamespace(t *testing.T) {
	testCases := []struct {
		annotationName string
		namespace      string
		ipAddr         []string
		expectedNil    bool
	}{
		{
			annotationName: "test",
			namespace:      "test",
			ipAddr:         []string{"192.168.1.1"},
			expectedNil:    false,
		},
		{
			annotationName: "test",
			namespace:      "test",
			ipAddr:         []string{"192.168.1.1/24", "2001:0000:130F:0000:0000:09C0:876A:130B"},
			expectedNil:    false,
		},
		{
			annotationName: "test",
			namespace:      "test",
			ipAddr:         []string{"192.168.1.1/24", "2001:0000:130F:0000:0000:09C0:876A:130B/64"},
			expectedNil:    false,
		},
		{
			annotationName: "",
			namespace:      "test",
			ipAddr:         []string{"192.168.1.1/24"},
			expectedNil:    true,
		},
		{
			annotationName: "test",
			namespace:      "",
			ipAddr:         []string{"192.168.1.1/24"},
			expectedNil:    true,
		},
		{
			annotationName: "test",
			namespace:      "test",
			ipAddr:         []string{},
			expectedNil:    false,
		},
	}
	for _, testCase := range testCases {
		annotation := StaticIPAnnotationWithNamespace(testCase.annotationName, testCase.namespace, testCase.ipAddr)
		if testCase.expectedNil {
			assert.Nil(t, annotation)
		} else {
			assert.NotNil(t, annotation)
		}
	}
}

func TestPodNetworkStaticIPAnnotationWithInterfaceAndNamespace(t *testing.T) {
	testCases := []struct {
		annotationName string
		namespace      string
		intName        string
		ipAddr         []string
		expectedNil    bool
	}{
		{
			annotationName: "test",
			namespace:      "test",
			intName:        "eth1",
			ipAddr:         []string{"192.168.1.1"},
			expectedNil:    false,
		},
		{
			annotationName: "test",
			namespace:      "test",
			intName:        "eth1",
			ipAddr:         []string{"192.168.1.1/24", "2001:0000:130F:0000:0000:09C0:876A:130B"},
			expectedNil:    false,
		},
		{
			annotationName: "test",
			namespace:      "test",
			intName:        "eth1",
			ipAddr:         []string{"192.168.1.1/24", "2001:0000:130F:0000:0000:09C0:876A:130B/64"},
			expectedNil:    false,
		},
		{
			annotationName: "",
			namespace:      "test",
			intName:        "eth1",
			ipAddr:         []string{"192.168.1.1/24"},
			expectedNil:    true,
		},
		{
			annotationName: "test",
			namespace:      "",
			intName:        "eth1",
			ipAddr:         []string{"192.168.1.1/24"},
			expectedNil:    true,
		},
		{
			annotationName: "test",
			namespace:      "test",
			intName:        "eth1",
			ipAddr:         []string{},
			expectedNil:    false,
		},
		{
			annotationName: "test",
			namespace:      "test",
			intName:        "",
			ipAddr:         []string{"192.168.1.1"},
			expectedNil:    true,
		},
	}
	for _, testCase := range testCases {
		annotation := StaticIPAnnotationWithInterfaceAndNamespace(
			testCase.annotationName, testCase.namespace, testCase.intName, testCase.ipAddr)
		if testCase.expectedNil {
			assert.Nil(t, annotation)
		} else {
			assert.NotNil(t, annotation)
		}
	}
}

func TestPodNetworkStaticIPAnnotationWithMacAddress(t *testing.T) {
	testCases := []struct {
		annotationName string
		macAddress     string
		ipAddr         []string
		expectedNil    bool
	}{
		{
			annotationName: "test",
			macAddress:     "00-B0-D0-63-C2-26",
			ipAddr:         []string{"192.168.1.1"},
			expectedNil:    false,
		},
		{
			annotationName: "test",
			macAddress:     "00-B0-D0-63-C2-26",
			ipAddr:         []string{"192.168.1.1/24", "2001:0000:130F:0000:0000:09C0:876A:130B"},
			expectedNil:    false,
		},
		{
			annotationName: "test",
			macAddress:     "00-B0-D0-63-C2-26",
			ipAddr:         []string{"192.168.1.1/24", "2001:0000:130F:0000:0000:09C0:876A:130B/64"},
			expectedNil:    false,
		},
		{
			annotationName: "",
			macAddress:     "00-B0-D0-63-C2-26",
			ipAddr:         []string{"192.168.1.1/24"},
			expectedNil:    true,
		},
		{
			annotationName: "test",
			macAddress:     "",
			ipAddr:         []string{"192.168.1.1/24"},
			expectedNil:    false,
		},
		{
			annotationName: "test",
			macAddress:     "00-B0-D0-63-C2-26",
			ipAddr:         []string{},
			expectedNil:    false,
		},
	}
	for _, testCase := range testCases {
		annotation := StaticIPAnnotationWithMacAddress(testCase.annotationName, testCase.ipAddr, testCase.macAddress)
		if testCase.expectedNil {
			assert.Nil(t, annotation)
		} else {
			assert.NotNil(t, annotation)
		}
	}
}

func TestPodNetworkStaticIPAnnotationWithMacAndNamespace(t *testing.T) {
	testCases := []struct {
		annotationName string
		macAddress     string
		namespace      string
		expectedNil    bool
	}{
		{
			annotationName: "test",
			macAddress:     "00-B0-D0-63-C2-26",
			namespace:      "test",
			expectedNil:    false,
		},
		{
			annotationName: "",
			macAddress:     "00-B0-D0-63-C2-26",
			namespace:      "test",
			expectedNil:    true,
		},
		{
			annotationName: "test",
			macAddress:     "",
			namespace:      "test",
			expectedNil:    false,
		},
		{
			annotationName: "test",
			macAddress:     "00-B0-D0-63-C2-26",
			namespace:      "",
			expectedNil:    true,
		},
	}
	for _, testCase := range testCases {
		annotation := StaticIPAnnotationWithMacAndNamespace(testCase.annotationName, testCase.namespace, testCase.macAddress)
		if testCase.expectedNil {
			assert.Nil(t, annotation)
		} else {
			assert.NotNil(t, annotation)
		}
	}
}

func TestPodNetworkStaticIPAnnotationWithInterfaceMacAndNamespace(t *testing.T) {
	testCases := []struct {
		annotationName string
		namespace      string
		intName        string
		macAddress     string
		expectedNil    bool
	}{
		{
			annotationName: "test",
			namespace:      "test",
			intName:        "eth0",
			macAddress:     "00-B0-D0-63-C2-26",
			expectedNil:    false,
		},
		{
			annotationName: "",
			namespace:      "test",
			intName:        "eth0",
			macAddress:     "00-B0-D0-63-C2-26",
			expectedNil:    true,
		},
		{
			annotationName: "test",
			namespace:      "",
			intName:        "eth0",
			macAddress:     "00-B0-D0-63-C2-26",
			expectedNil:    true,
		},
		{
			annotationName: "test",
			namespace:      "test",
			intName:        "",
			macAddress:     "00-B0-D0-63-C2-26",
			expectedNil:    true,
		},
		{
			annotationName: "test",
			namespace:      "test",
			intName:        "eth0",
			macAddress:     "",
			expectedNil:    false,
		},
	}
	for _, testCase := range testCases {
		annotation := StaticIPAnnotationWithInterfaceMacAndNamespace(
			testCase.annotationName, testCase.namespace, testCase.intName, testCase.macAddress)
		if testCase.expectedNil {
			assert.Nil(t, annotation)
		} else {
			assert.NotNil(t, annotation)
		}
	}
}

func TestPodNetworkStaticIPBondAnnotationWithInterface(t *testing.T) {
	testCases := []struct {
		annotationName string
		namespace      string
		intName        string
		sriovNetName   []string
		ipAddrBond     []string
		expectedNil    bool
	}{
		{
			annotationName: "test",
			sriovNetName:   []string{"net1", "net2"},
			ipAddrBond:     []string{"192.168.1.1/24"},
			intName:        "eth0",
			expectedNil:    false,
		},
		{
			annotationName: "",
			sriovNetName:   []string{"net1", "net2"},
			ipAddrBond:     []string{"192.168.1.1/24"},
			intName:        "eth0",
			expectedNil:    true,
		},
		{
			annotationName: "test",
			sriovNetName:   []string{},
			ipAddrBond:     []string{"192.168.1.1/24"},
			intName:        "eth0",
			expectedNil:    true,
		},
		{
			annotationName: "test",
			sriovNetName:   []string{"net1", "net2"},
			ipAddrBond:     []string{"192.168.1.1/24", "error"},
			intName:        "eth0",
			expectedNil:    true,
		},
		{
			annotationName: "test",
			sriovNetName:   []string{"net1", "net2"},
			ipAddrBond: []string{"2001:0000:130F:0000:0000:09C0:876A:130B/64",
				"2001:0000:130F:0000:0000:09C0:876A:130B"},
			intName:     "eth0",
			expectedNil: false,
		},
		{
			annotationName: "test",
			sriovNetName:   []string{"net1", "net2"},
			ipAddrBond:     []string{"192.168.1.1/24"},
			intName:        "",
			expectedNil:    true,
		},
		{
			annotationName: "test",
			sriovNetName:   []string{"net1", "net2"},
			ipAddrBond:     []string{},
			intName:        "eth0",
			expectedNil:    true,
		},
	}
	for _, testCase := range testCases {
		annotation := StaticIPBondAnnotationWithInterface(
			testCase.annotationName, testCase.intName, testCase.sriovNetName, testCase.ipAddrBond)
		if testCase.expectedNil {
			assert.Nil(t, annotation)
		} else {
			assert.NotNil(t, annotation)
		}
	}
}

func TestPodNetworkStaticIPMultiNetDualStackAnnotation(t *testing.T) {
	testCases := []struct {
		sriovNetName  []string
		ipAddresses   []string
		expectedError error
	}{
		{
			sriovNetName:  []string{"net1", "net2"},
			ipAddresses:   []string{"192.168.1.1/24", "192.168.1.1/24"},
			expectedError: nil,
		},
		{
			sriovNetName:  []string{},
			ipAddresses:   []string{"192.168.1.1/24", "192.168.1.1/24"},
			expectedError: fmt.Errorf("sriovNets []string cannot be empty"),
		},
		{
			sriovNetName:  []string{"net1", "net2"},
			ipAddresses:   []string{},
			expectedError: fmt.Errorf("ipAddr []string cannot be empty or an odd number"),
		},
		{
			sriovNetName:  []string{"net1", "net2"},
			ipAddresses:   []string{"192.168.1.1/24", "abc"},
			expectedError: fmt.Errorf("ipAddr []string contain invalid ip address"),
		},
	}
	for _, testCase := range testCases {
		annotation, err := StaticIPMultiNetDualStackAnnotation(testCase.sriovNetName, testCase.ipAddresses)
		if testCase.expectedError == nil {
			assert.NotNil(t, annotation)
		} else {
			assert.Equal(t, err, testCase.expectedError)
		}
	}
}
