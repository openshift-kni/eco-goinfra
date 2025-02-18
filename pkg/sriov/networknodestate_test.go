package sriov

import (
	"fmt"
	"testing"
	"time"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultNodeName   = "test1"
	defaultNodeNsName = "testnamespace"
)

func TestNewNetworkNodeStateBuilder(t *testing.T) {
	generateNetworkBuilder := NewNetworkNodeStateBuilder

	testCases := []struct {
		nodeName          string
		nsName            string
		expectedErrorText string
		client            bool
	}{
		{
			nodeName: defaultNodeName,
			nsName:   defaultNodeNsName,
			client:   true,
		},
		{
			nodeName:          "",
			nsName:            defaultNodeNsName,
			expectedErrorText: "SriovNetworkNodeState 'nodeName' is empty",
			client:            true,
		},
		{
			nodeName:          defaultNodeName,
			nsName:            "",
			expectedErrorText: "SriovNetworkNodeState 'nsname' is empty",
			client:            true,
		},
	}
	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testNetworkStructure := generateNetworkBuilder(
			testSettings, testCase.nodeName, testCase.nsName)
		assert.NotNil(t, testNetworkStructure)

		if len(testCase.expectedErrorText) > 0 {
			assert.Equal(t, testCase.expectedErrorText, testNetworkStructure.errorMsg)
		}
	}
}

func TestNetworkNodeStateDiscovery(t *testing.T) {
	generateNetworkNodeState := func(name, namespace string) *srIovV1.SriovNetworkNodeState {
		return &srIovV1.SriovNetworkNodeState{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: srIovV1.SriovNetworkNodeStateSpec{},
		}
	}

	testCases := []struct {
		nodeName            string
		nsName              string
		expectedError       error
		addToRuntimeObjects bool
	}{
		{
			nodeName:            defaultNodeName,
			nsName:              defaultNodeNsName,
			addToRuntimeObjects: true,
		},
		{
			nodeName:            "",
			nsName:              defaultNodeNsName,
			expectedError:       fmt.Errorf("SriovNetworkNodeState 'nodeName' is empty"),
			addToRuntimeObjects: true,
		},
		{
			nodeName:            defaultNodeName,
			nsName:              "",
			expectedError:       fmt.Errorf("SriovNetworkNodeState 'nsname' is empty"),
			addToRuntimeObjects: true,
		},
		{
			nodeName:            defaultNodeName,
			nsName:              defaultNodeNsName,
			expectedError:       fmt.Errorf("not found"),
			addToRuntimeObjects: false,
		},
	}
	for _, testCase := range testCases {
		var (
			testSettings   *clients.Settings
			runtimeObjects []runtime.Object
		)

		if testCase.addToRuntimeObjects {
			networkNodeState := generateNetworkNodeState(testCase.nodeName, testCase.nsName)
			runtimeObjects = append(runtimeObjects, networkNodeState)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: testSchemes,
		})

		networkNodeStateBuilder := NewNetworkNodeStateBuilder(testSettings, testCase.nodeName, testCase.nsName)
		err := networkNodeStateBuilder.Discover()

		if testCase.addToRuntimeObjects {
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.True(t, k8serrors.IsNotFound(err))
		}

		if testCase.expectedError == nil {
			assert.NotNil(t, networkNodeStateBuilder.Objects)
		} else {
			assert.Nil(t, networkNodeStateBuilder.Objects)
		}
	}
}

func TestNetworkNodeStateGetNICs(t *testing.T) {
	testCases := []struct {
		netInterface  srIovV1.InterfaceExts
		nodeName      string
		expectedError error
	}{
		{
			netInterface: []srIovV1.InterfaceExt{{Name: "eth1"}},
			nodeName:     defaultNodeName,
		},
		{
			netInterface: []srIovV1.InterfaceExt{{Name: "eth1"}, {Name: "eth2"}},
			nodeName:     defaultNodeName,
		},
		{
			netInterface: []srIovV1.InterfaceExt{},
			nodeName:     defaultNodeName,
		},
		{
			netInterface:  []srIovV1.InterfaceExt{},
			expectedError: fmt.Errorf("SriovNetworkNodeState 'nodeName' is empty"),
		},
	}

	for _, testCase := range testCases {
		var (
			testSettings   *clients.Settings
			runtimeObjects []runtime.Object
		)

		networkNodeState := buildNodeNetworkStateWithNics(testCase.netInterface)
		runtimeObjects = append(runtimeObjects, networkNodeState)

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: testSchemes,
		})

		networkNodeStateBuilder := NewNetworkNodeStateBuilder(testSettings, testCase.nodeName, defaultNodeNsName)
		nics, err := networkNodeStateBuilder.GetNICs()

		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, networkNodeStateBuilder.Objects)
		}

		if len(testCase.netInterface) > 0 && testCase.expectedError == nil {
			assert.NotNil(t, nics)
			assert.Equal(t, nics, testCase.netInterface)
		}
	}
}

func TestNetworkNodeStateGetUpNICs(t *testing.T) {
	testCases := []struct {
		netInterface  srIovV1.InterfaceExts
		upInterface   srIovV1.InterfaceExts
		nodeName      string
		expectedError error
	}{
		{

			netInterface: []srIovV1.InterfaceExt{
				{Name: "eth1", LinkSpeed: "1000Mb"},
				{Name: "eth2", LinkSpeed: "1000Mb"},
			},
			upInterface: []srIovV1.InterfaceExt{
				{Name: "eth1", LinkSpeed: "1000Mb"},
				{Name: "eth2", LinkSpeed: "1000Mb"},
			},
			nodeName: defaultNodeName,
		},
		{
			netInterface: []srIovV1.InterfaceExt{
				{Name: "eth1", LinkSpeed: "1000Mb"},
				{Name: "eth2"},
			},
			upInterface: []srIovV1.InterfaceExt{
				{Name: "eth1", LinkSpeed: "1000Mb"},
			},
			nodeName: defaultNodeName,
		},
		{
			netInterface: []srIovV1.InterfaceExt{
				{Name: "eth1"},
				{Name: "eth2"},
			},
			upInterface: []srIovV1.InterfaceExt{},
			nodeName:    defaultNodeName,
		},
		{
			netInterface: []srIovV1.InterfaceExt{
				{Name: "eth1", LinkSpeed: "-1 Mb/s"},
				{Name: "eth2", LinkSpeed: "1000Mb"},
			},
			upInterface: []srIovV1.InterfaceExt{
				{Name: "eth2", LinkSpeed: "1000Mb"},
			},
			nodeName: defaultNodeName,
		},
		{
			netInterface: []srIovV1.InterfaceExt{
				{Name: "eth1", LinkSpeed: "1000Mb"},
			},
			expectedError: fmt.Errorf("SriovNetworkNodeState 'nodeName' is empty"),
		},
	}

	for _, testCase := range testCases {
		networkNodeState := buildNodeNetworkStateWithNics(testCase.netInterface)
		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  []runtime.Object{networkNodeState},
			SchemeAttachers: testSchemes,
		})
		networkNodeStateBuilder := NewNetworkNodeStateBuilder(testSettings, testCase.nodeName, defaultNodeNsName)
		nics, err := networkNodeStateBuilder.GetUpNICs()

		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, networkNodeStateBuilder.Objects)
		}

		if len(testCase.netInterface) > 0 && len(testCase.upInterface) > 0 {
			assert.NotNil(t, nics)
		}

		if nics != nil {
			assert.Equal(t, nics, testCase.upInterface)
		}
	}
}

func TestNetworkNodeStateWaitUntilSyncStatus(t *testing.T) {
	testCases := []struct {
		syncStatus    string
		expectedError error
	}{
		{
			syncStatus: "Succeeded",
		},
		{
			syncStatus:    "",
			expectedError: fmt.Errorf("syncStatus cannot be empty"),
		},
	}
	for _, testCase := range testCases {
		networkNodeState := buildNodeNetworkStateSyncStatus(defaultNodeName, defaultNodeNsName, testCase.syncStatus)
		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  []runtime.Object{networkNodeState},
			SchemeAttachers: testSchemes,
		})
		networkNodeStateBuilder := NewNetworkNodeStateBuilder(testSettings, defaultNodeName, defaultNodeNsName)
		err := networkNodeStateBuilder.Discover()
		assert.Nil(t, err)
		err = networkNodeStateBuilder.WaitUntilSyncStatus(testCase.syncStatus, 30*time.Second)
		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestNetworkNodeStateGetNumVFs(t *testing.T) {
	testCases := []struct {
		netInterface srIovV1.InterfaceExts
	}{
		{
			netInterface: []srIovV1.InterfaceExt{
				{Name: "eth1", LinkSpeed: "1000Mb", NumVfs: 5},
			},
		},
		{
			netInterface: []srIovV1.InterfaceExt{
				{Name: "eth1", LinkSpeed: "1000Mb", NumVfs: 5},
				{Name: "eth2", LinkSpeed: "1000Mb", NumVfs: 10},
			},
		},
	}

	for _, testCase := range testCases {
		networkNodeState := buildNodeNetworkStateWithNics(testCase.netInterface)
		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  []runtime.Object{networkNodeState},
			SchemeAttachers: testSchemes,
		})
		networkNodeStateBuilder := NewNetworkNodeStateBuilder(testSettings, defaultNodeName, defaultNodeNsName)
		err := networkNodeStateBuilder.Discover()
		assert.Nil(t, err)

		for _, netInterface := range testCase.netInterface {
			vfNum, err := networkNodeStateBuilder.GetNumVFs(netInterface.Name)
			assert.Nil(t, err)
			assert.Equal(t, netInterface.NumVfs, vfNum)
		}
	}
}

func TestNetworkNodeStateGetTotalVFs(t *testing.T) {
	testCases := []struct {
		netInterface srIovV1.InterfaceExts
	}{
		{
			netInterface: []srIovV1.InterfaceExt{
				{Name: "eth1", LinkSpeed: "1000Mb", TotalVfs: 5},
			},
		},
		{
			netInterface: []srIovV1.InterfaceExt{
				{Name: "eth1", LinkSpeed: "1000Mb", TotalVfs: 5},
				{Name: "eth2", LinkSpeed: "1000Mb", TotalVfs: 10},
			},
		},
	}

	for _, testCase := range testCases {
		networkNodeState := buildNodeNetworkStateWithNics(testCase.netInterface)
		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  []runtime.Object{networkNodeState},
			SchemeAttachers: testSchemes,
		})
		networkNodeStateBuilder := NewNetworkNodeStateBuilder(testSettings, defaultNodeName, defaultNodeNsName)
		err := networkNodeStateBuilder.Discover()
		assert.Nil(t, err)

		for _, netInterface := range testCase.netInterface {
			totalVFNum, err := networkNodeStateBuilder.GetTotalVFs(netInterface.Name)
			assert.Nil(t, err)
			assert.Equal(t, netInterface.TotalVfs, totalVFNum)
		}
	}
}

func TestNetworkNodeStateGetDriverName(t *testing.T) {
	testCases := []struct {
		netInterface srIovV1.InterfaceExts
		upInterface  srIovV1.InterfaceExts
	}{
		{
			netInterface: []srIovV1.InterfaceExt{
				{Name: "eth1", Driver: "Mlx"},
			},
		},
		{
			netInterface: []srIovV1.InterfaceExt{
				{Name: "eth1", Driver: "intel"},
				{Name: "eth2", Driver: "mlx"},
			},
		},
	}

	for _, testCase := range testCases {
		networkNodeState := buildNodeNetworkStateWithNics(testCase.netInterface)
		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  []runtime.Object{networkNodeState},
			SchemeAttachers: testSchemes,
		})
		networkNodeStateBuilder := NewNetworkNodeStateBuilder(testSettings, defaultNodeName, defaultNetNsName)
		err := networkNodeStateBuilder.Discover()
		assert.Nil(t, err)

		for _, netInterface := range testCase.netInterface {
			driver, err := networkNodeStateBuilder.GetDriverName(netInterface.Name)
			assert.Nil(t, err)
			assert.Equal(t, netInterface.Driver, driver)
		}
	}
}

func TestNetworkNodeStateGetPciAddress(t *testing.T) {
	testCases := []struct {
		netInterface srIovV1.InterfaceExts
		upInterface  srIovV1.InterfaceExts
	}{
		{
			netInterface: []srIovV1.InterfaceExt{
				{Name: "eth1", PciAddress: "11111"},
			},
		},
		{
			netInterface: []srIovV1.InterfaceExt{
				{Name: "eth1", LinkSpeed: "11111"},
				{Name: "eth2", PciAddress: "22222"},
			},
		},
	}

	for _, testCase := range testCases {
		networkNodeState := buildNodeNetworkStateWithNics(testCase.netInterface)
		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  []runtime.Object{networkNodeState},
			SchemeAttachers: testSchemes,
		})
		networkNodeStateBuilder := NewNetworkNodeStateBuilder(testSettings, defaultNodeName, defaultNodeNsName)
		err := networkNodeStateBuilder.Discover()
		assert.Nil(t, err)

		for _, netInterface := range testCase.netInterface {
			pciAddress, err := networkNodeStateBuilder.GetPciAddress(netInterface.Name)
			assert.Nil(t, err)
			assert.Equal(t, netInterface.PciAddress, pciAddress)
		}
	}
}

func buildNodeNetworkState(name, nsName string) *srIovV1.SriovNetworkNodeState {
	return &srIovV1.SriovNetworkNodeState{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsName,
		},
		Spec:   srIovV1.SriovNetworkNodeStateSpec{},
		Status: srIovV1.SriovNetworkNodeStateStatus{},
	}
}

func buildNodeNetworkStateSyncStatus(name, nsName, syncStatus string) *srIovV1.SriovNetworkNodeState {
	nodeNetworkState := buildNodeNetworkState(name, nsName)

	nodeNetworkState.Status.SyncStatus = syncStatus

	return nodeNetworkState
}

func buildNodeNetworkStateWithNics(sriovInterfaces srIovV1.InterfaceExts) *srIovV1.SriovNetworkNodeState {
	nodeNetworkState := buildNodeNetworkState(defaultNodeName, defaultNodeNsName)

	nodeNetworkState.Status.Interfaces = sriovInterfaces

	return nodeNetworkState
}
