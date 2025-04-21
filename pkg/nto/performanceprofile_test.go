package nto //nolint:misspell

import (
	"fmt"
	"testing"

	performanceprofilev2 "github.com/openshift/cluster-node-tuning-operator/pkg/apis/performanceprofile/v2"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultPerformanceProfileName = "default"
	defaultIsolatedCPU            = "2-27,30-55"
	defaultReservedCPU            = "0-1,28-29"
	defaultMCPSelector            = map[string]string{"machineconfiguration.openshift.io/role": "worker"}
	defaultNodeSelector           = map[string]string{"node-role.kubernetes.io/worker": ""}
	defaultNumaTopology           = "single-numa-node"
	defaultHugePagesNodeOne       = int32(0)
	defaultHugePagesNodeTwo       = int32(1)
	defaultHugepageSize           = "2M"
	defaultHugepages              = []performanceprofilev2.HugePage{
		{
			Size:  performanceprofilev2.HugePageSize(defaultHugepageSize),
			Count: 32768,
			Node:  &defaultHugePagesNodeOne,
		},
	}
	defaultHugepagesTwoNumaNodes = []performanceprofilev2.HugePage{
		{
			Size:  performanceprofilev2.HugePageSize(defaultHugepageSize),
			Count: 32768,
			Node:  &defaultHugePagesNodeOne,
		},
		{
			Size:  performanceprofilev2.HugePageSize(defaultHugepageSize),
			Count: 32768,
			Node:  &defaultHugePagesNodeTwo,
		},
	}
	defaultNetInterfaceNameOne = "ens2f(0|1)"
	defaultNetInterfaceNameTwo = "ens3f(0|1)"
	defaultVendorID            = "123456"
	defaultDeviceID            = "7654321"
	emptyString                = ""
	paoTestSchemes             = []clients.SchemeAttacher{
		performanceprofilev2.AddToScheme,
	}
)

func TestPullPerformanceProfile(t *testing.T) {
	generatePerformanceProfile := func(name string) *performanceprofilev2.PerformanceProfile {
		return &performanceprofilev2.PerformanceProfile{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
	}

	testCases := []struct {
		perfProfileName     string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			perfProfileName:     "test",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			perfProfileName:     "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("performanceProfile 'name' cannot be empty"),
			client:              true,
		},
		{
			perfProfileName:     "pptest",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("PerformanceProfile object pptest does not exist"),
			client:              true,
		},
		{
			perfProfileName:     "pptest",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("performanceProfile 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testPerformanceProfile := generatePerformanceProfile(testCase.perfProfileName)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPerformanceProfile)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: paoTestSchemes,
			})
		}

		builderResult, err := Pull(testSettings, testCase.perfProfileName)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testPerformanceProfile.Name, builderResult.Object.Name)
		}
	}
}

func TestNewBuilder(t *testing.T) {
	testCases := []struct {
		perfProfileName string
		cpuIsolated     string
		cpuReserved     string
		nodeSelector    map[string]string
		expectedError   string
	}{
		{
			perfProfileName: defaultPerformanceProfileName,
			cpuIsolated:     defaultIsolatedCPU,
			cpuReserved:     defaultReservedCPU,
			nodeSelector:    defaultNodeSelector,
			expectedError:   "",
		},
		{
			perfProfileName: "",
			cpuIsolated:     defaultIsolatedCPU,
			cpuReserved:     defaultReservedCPU,
			nodeSelector:    defaultNodeSelector,
			expectedError:   "PerformanceProfile's name is empty",
		},
		{
			perfProfileName: defaultPerformanceProfileName,
			cpuIsolated:     "",
			cpuReserved:     defaultReservedCPU,
			nodeSelector:    defaultNodeSelector,
			expectedError:   "PerformanceProfile's 'cpuIsolated' is empty",
		},
		{
			perfProfileName: defaultPerformanceProfileName,
			cpuIsolated:     defaultIsolatedCPU,
			cpuReserved:     "",
			nodeSelector:    defaultNodeSelector,
			expectedError:   "PerformanceProfile's 'cpuReserved' is empty",
		},
		{
			perfProfileName: defaultPerformanceProfileName,
			cpuIsolated:     defaultIsolatedCPU,
			cpuReserved:     defaultReservedCPU,
			nodeSelector:    map[string]string{},
			expectedError:   "PerformanceProfile's 'nodeSelector' is empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testBuilder := NewBuilder(testSettings,
			testCase.perfProfileName,
			testCase.cpuIsolated,
			testCase.cpuReserved,
			testCase.nodeSelector)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)
		assert.NotNil(t, testBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.perfProfileName, testBuilder.Definition.Name)
		}
	}
}

func TestPerformanceProfileExists(t *testing.T) {
	testCases := []struct {
		testPerformanceProfile *Builder
		expectedStatus         bool
	}{
		{
			testPerformanceProfile: buildValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject()),
			expectedStatus:         true,
		},
		{
			testPerformanceProfile: buildInValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject()),
			expectedStatus:         false,
		},
		{
			testPerformanceProfile: buildValidPerformanceProfileBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:         false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testPerformanceProfile.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestPerformanceProfileGet(t *testing.T) {
	testCases := []struct {
		testPerformanceProfile *Builder
		expectedError          error
	}{
		{
			testPerformanceProfile: buildValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject()),
			expectedError:          nil,
		},
		{
			testPerformanceProfile: buildInValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject()),
			expectedError:          fmt.Errorf("PerformanceProfile's name is empty"),
		},
		{
			testPerformanceProfile: buildValidPerformanceProfileBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:          fmt.Errorf("performanceprofiles.performance.openshift.io \"default\" not found"),
		},
	}

	for _, testCase := range testCases {
		performanceProfileObj, err := testCase.testPerformanceProfile.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, performanceProfileObj.Name, testCase.testPerformanceProfile.Definition.Name)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestPerformanceProfileCreate(t *testing.T) {
	testCases := []struct {
		testPerformanceProfile *Builder
		expectedError          string
	}{
		{
			testPerformanceProfile: buildValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject()),
			expectedError:          "",
		},
		{
			testPerformanceProfile: buildInValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject()),
			expectedError:          "PerformanceProfile's name is empty",
		},
		{
			testPerformanceProfile: buildValidPerformanceProfileBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:          "",
		},
	}

	for _, testCase := range testCases {
		testPerformanceProfileBuilder, err := testCase.testPerformanceProfile.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testPerformanceProfileBuilder.Definition.Name, testPerformanceProfileBuilder.Object.Name)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestPerformanceProfileDelete(t *testing.T) {
	testCases := []struct {
		testPerformanceProfile *Builder
		expectedError          error
	}{
		{
			testPerformanceProfile: buildValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject()),
			expectedError:          nil,
		},
		{
			testPerformanceProfile: buildInValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject()),
			expectedError:          fmt.Errorf("PerformanceProfile's name is empty"),
		},
		{
			testPerformanceProfile: buildValidPerformanceProfileBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:          nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testPerformanceProfile.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testPerformanceProfile.Object)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestPerformanceProfileUpdate(t *testing.T) {
	testCases := []struct {
		testPerformanceProfile *Builder
		topologyPolicy         string
		forceFlag              bool
		expectedError          string
	}{
		{
			testPerformanceProfile: buildValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject()),
			topologyPolicy:         defaultNumaTopology,
			forceFlag:              true,
			expectedError:          "",
		},
		{
			testPerformanceProfile: buildValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject()),
			topologyPolicy:         defaultNumaTopology,
			forceFlag:              false,
			expectedError:          "",
		},
		{
			testPerformanceProfile: buildInValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject()),
			expectedError:          "PerformanceProfile's name is empty",
		},
	}

	for _, testCase := range testCases {
		assert.Nil(t, nil, testCase.testPerformanceProfile.Definition.Spec.NUMA)
		assert.Nil(t, nil, testCase.testPerformanceProfile.Object)
		testCase.testPerformanceProfile.WithNumaTopology(testCase.topologyPolicy)
		_, err := testCase.testPerformanceProfile.Update(testCase.forceFlag)

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.topologyPolicy,
				*testCase.testPerformanceProfile.Definition.Spec.NUMA.TopologyPolicy)
		}
	}
}

func TestPerformanceProfileWithHugePages(t *testing.T) {
	testCases := []struct {
		testHugePageSize  string
		testPages         []performanceprofilev2.HugePage
		expectedErrorText string
	}{
		{
			testHugePageSize:  defaultHugepageSize,
			testPages:         defaultHugepages,
			expectedErrorText: "",
		},
		{
			testHugePageSize: "1G",
			testPages: []performanceprofilev2.HugePage{
				{
					Size:  "1G",
					Count: 32768,
					Node:  &defaultHugePagesNodeOne,
				},
			},
			expectedErrorText: "",
		},
		{
			testHugePageSize:  defaultHugepageSize,
			testPages:         defaultHugepagesTwoNumaNodes,
			expectedErrorText: "",
		},
		{
			testHugePageSize: defaultHugepageSize,
			testPages: []performanceprofilev2.HugePage{
				{
					Size:  performanceprofilev2.HugePageSize(defaultHugepageSize),
					Count: 32768,
				},
			},
			expectedErrorText: "",
		},
		{
			testHugePageSize:  "2G",
			testPages:         defaultHugepages,
			expectedErrorText: "'hugePageSize' argument is not in allowed list: [2M 1G]",
		},
		{
			testHugePageSize:  "",
			testPages:         defaultHugepages,
			expectedErrorText: "'hugePageSize' argument cannot be empty",
		},
		{
			testHugePageSize:  defaultHugepageSize,
			testPages:         []performanceprofilev2.HugePage{},
			expectedErrorText: "'hugePages' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject())

		result := testBuilder.WithHugePages(testCase.testHugePageSize, testCase.testPages)

		if testCase.expectedErrorText != "" {
			assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, performanceprofilev2.HugePageSize(testCase.testHugePageSize),
				*result.Definition.Spec.HugePages.DefaultHugePagesSize)
			assert.Equal(t, testCase.testPages, result.Definition.Spec.HugePages.Pages)
		}
	}
}

func TestPerformanceProfileWithMachineConfigPoolSelector(t *testing.T) {
	testCases := []struct {
		testMachineConfigPoolSelector map[string]string
		expectedErrorText             string
	}{
		{
			testMachineConfigPoolSelector: defaultMCPSelector,
			expectedErrorText:             "",
		},
		{
			testMachineConfigPoolSelector: map[string]string{},
			expectedErrorText:             "'machineConfigPoolSelector' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject())

		result := testBuilder.WithMachineConfigPoolSelector(testCase.testMachineConfigPoolSelector)

		if testCase.expectedErrorText != "" {
			assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testMachineConfigPoolSelector, result.Definition.Spec.MachineConfigPoolSelector)
		}
	}
}

func TestPerformanceProfileWithNodeSelector(t *testing.T) {
	testCases := []struct {
		testNodeSelector  map[string]string
		expectedErrorText string
	}{
		{
			testNodeSelector:  defaultNodeSelector,
			expectedErrorText: "",
		},
		{
			testNodeSelector:  map[string]string{},
			expectedErrorText: "'nodeSelector' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject())

		result := testBuilder.WithNodeSelector(testCase.testNodeSelector)

		if testCase.expectedErrorText != "" {
			assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testNodeSelector, result.Definition.Spec.NodeSelector)
		}
	}
}

func TestPerformanceProfileWithNumaTopology(t *testing.T) {
	testCases := []struct {
		testNumaTopology  string
		expectedErrorText string
	}{
		{
			testNumaTopology:  "best-effort",
			expectedErrorText: "",
		},
		{
			testNumaTopology:  "restricted",
			expectedErrorText: "",
		},
		{
			testNumaTopology:  "single-numa-node",
			expectedErrorText: "",
		},
		{
			testNumaTopology: "some-not-supported-policy",
			expectedErrorText: "'allowedTopologyPolicies' argument is not in allowed list " +
				"[best-effort restricted single-numa-node]",
		},
		{
			testNumaTopology:  "",
			expectedErrorText: "'topologyPolicy' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject())

		result := testBuilder.WithNumaTopology(testCase.testNumaTopology)

		if testCase.expectedErrorText != "" {
			assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testNumaTopology, *result.Definition.Spec.NUMA.TopologyPolicy)
		}
	}
}

func TestPerformanceProfileWithRTKernel(t *testing.T) {
	testCases := []struct {
		expectedErrorText string
	}{
		{
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject())

		result := testBuilder.WithRTKernel()

		if testCase.expectedErrorText != "" {
			assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, true, *result.Definition.Spec.RealTimeKernel.Enabled)
		}
	}
}

func TestPerformanceProfileWithGloballyDisableIrqLoadBalancing(t *testing.T) {
	testCases := []struct {
		expectedErrorText string
	}{
		{
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject())

		result := testBuilder.WithGloballyDisableIrqLoadBalancing()

		if testCase.expectedErrorText != "" {
			assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, true, *result.Definition.Spec.GloballyDisableIrqLoadBalancing)
		}
	}
}

func TestPerformanceProfileWithWorkloadHints(t *testing.T) {
	testCases := []struct {
		rtHint              bool
		perPodPowerMgmtHint bool
		highPowerHint       bool
		expectedErrorText   string
	}{
		{
			rtHint:              true,
			perPodPowerMgmtHint: true,
			highPowerHint:       true,
			expectedErrorText:   "",
		},
		{
			rtHint:              false,
			perPodPowerMgmtHint: true,
			highPowerHint:       true,
			expectedErrorText:   "",
		},
		{
			rtHint:              true,
			perPodPowerMgmtHint: false,
			highPowerHint:       true,
			expectedErrorText:   "",
		},
		{
			rtHint:              true,
			perPodPowerMgmtHint: true,
			highPowerHint:       false,
			expectedErrorText:   "",
		},
		{
			rtHint:              false,
			perPodPowerMgmtHint: false,
			highPowerHint:       false,
			expectedErrorText:   "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject())

		result := testBuilder.WithWorkloadHints(testCase.rtHint,
			testCase.perPodPowerMgmtHint, testCase.highPowerHint)

		if testCase.expectedErrorText != "" {
			assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.rtHint, *result.Definition.Spec.WorkloadHints.RealTime)
			assert.Equal(t, testCase.perPodPowerMgmtHint, *result.Definition.Spec.WorkloadHints.PerPodPowerManagement)
			assert.Equal(t, testCase.highPowerHint, *result.Definition.Spec.WorkloadHints.HighPowerConsumption)
		}
	}
}

func TestPerformanceProfileWithAnnotations(t *testing.T) {
	testCases := []struct {
		testAnnotations   map[string]string
		expectedErrorText string
	}{
		{
			testAnnotations:   map[string]string{"performance.openshift.io/ignore-cgroups-version": "true"},
			expectedErrorText: "",
		},
		{
			testAnnotations: map[string]string{
				"kubeletconfig.experimental": "{'systemReserved':{'cpu':'500m','memory':'28Gi'}}",
			},
			expectedErrorText: "",
		},
		{
			testAnnotations: map[string]string{"performance.openshift.io/ignore-cgroups-version": "true",
				"kubeletconfig.experimental": "{\"systemReserved\":{\"cpu\":\"500m\",\"memory\":\"28Gi\"}}"},
			expectedErrorText: "",
		},
		{
			testAnnotations:   map[string]string{},
			expectedErrorText: "'annotations' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject())

		result := testBuilder.WithAnnotations(testCase.testAnnotations)

		if testCase.expectedErrorText != "" {
			assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testAnnotations, result.Definition.Annotations)
		}
	}
}

//nolint:funlen
func TestPerformanceProfileWithNet(t *testing.T) {
	testCases := []struct {
		testUserLevelNet  bool
		testDevices       []performanceprofilev2.Device
		expectedErrorText string
	}{
		{
			testUserLevelNet: true,
			testDevices: []performanceprofilev2.Device{{
				InterfaceName: &defaultNetInterfaceNameOne,
				VendorID:      &defaultVendorID,
				DeviceID:      &defaultDeviceID,
			}},
			expectedErrorText: "",
		},
		{
			testUserLevelNet: false,
			testDevices: []performanceprofilev2.Device{{
				InterfaceName: &defaultNetInterfaceNameOne,
				VendorID:      &defaultVendorID,
				DeviceID:      &defaultDeviceID,
			}},
			expectedErrorText: "",
		},
		{
			testUserLevelNet: true,
			testDevices: []performanceprofilev2.Device{{
				InterfaceName: &defaultNetInterfaceNameOne,
				VendorID:      &defaultVendorID,
				DeviceID:      &emptyString,
			}},
			expectedErrorText: "",
		},
		{
			testUserLevelNet: true,
			testDevices: []performanceprofilev2.Device{{
				InterfaceName: &defaultNetInterfaceNameOne,
				VendorID:      &emptyString,
				DeviceID:      &defaultDeviceID,
			}},
			expectedErrorText: "",
		},
		{
			testUserLevelNet: true,
			testDevices: []performanceprofilev2.Device{{
				InterfaceName: &defaultNetInterfaceNameOne,
			}},
			expectedErrorText: "",
		},
		{
			testUserLevelNet: true,
			testDevices: []performanceprofilev2.Device{{
				InterfaceName: &emptyString,
				VendorID:      &defaultVendorID,
				DeviceID:      &emptyString,
			}},
			expectedErrorText: "",
		},
		{
			testUserLevelNet: true,
			testDevices: []performanceprofilev2.Device{{
				InterfaceName: &emptyString,
				VendorID:      &emptyString,
				DeviceID:      &defaultDeviceID,
			}},
			expectedErrorText: "",
		},
		{
			testUserLevelNet: true,
			testDevices: []performanceprofilev2.Device{{
				InterfaceName: &defaultNetInterfaceNameOne,
				VendorID:      &emptyString,
				DeviceID:      &emptyString,
			}},
			expectedErrorText: "",
		},
		{
			testUserLevelNet: true,
			testDevices: []performanceprofilev2.Device{{
				InterfaceName: &defaultNetInterfaceNameOne,
			}, {
				InterfaceName: &defaultNetInterfaceNameTwo,
			}},
			expectedErrorText: "",
		},
		{
			testUserLevelNet:  true,
			testDevices:       []performanceprofilev2.Device{},
			expectedErrorText: "'net' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject())

		result := testBuilder.WithNet(testCase.testUserLevelNet, testCase.testDevices)

		if testCase.expectedErrorText != "" {
			assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, &testCase.testUserLevelNet, result.Definition.Spec.Net.UserLevelNetworking)
			assert.Equal(t, testCase.testDevices, result.Definition.Spec.Net.Devices)
		}
	}
}

func TestPerformanceProfileWithAdditionalKernelArgs(t *testing.T) {
	testCases := []struct {
		testAdditionalKernelArgs []string
		expectedErrorText        string
	}{
		{
			testAdditionalKernelArgs: []string{"nohz_full=2-27,30-55"},
			expectedErrorText:        "",
		},
		{
			testAdditionalKernelArgs: []string{"nohz_full=2-27,30-55", "intel_idle.max_cstate=0", "audit=0"},
			expectedErrorText:        "",
		},
		{
			testAdditionalKernelArgs: []string{},
			expectedErrorText:        "'additionalKernelArgs' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPerformanceProfileBuilder(buildPerformanceProfileWithDummyObject())

		result := testBuilder.WithAdditionalKernelArgs(testCase.testAdditionalKernelArgs)

		if testCase.expectedErrorText != "" {
			assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testAdditionalKernelArgs, result.Definition.Spec.AdditionalKernelArgs)
		}
	}
}

func buildValidPerformanceProfileBuilder(apiClient *clients.Settings) *Builder {
	performanceProfileBuilder := NewBuilder(
		apiClient,
		defaultPerformanceProfileName,
		defaultIsolatedCPU,
		defaultReservedCPU,
		defaultNodeSelector)

	return performanceProfileBuilder
}

func buildInValidPerformanceProfileBuilder(apiClient *clients.Settings) *Builder {
	performanceProfileBuilder := NewBuilder(apiClient,
		"",
		defaultIsolatedCPU,
		defaultReservedCPU,
		defaultNodeSelector)

	return performanceProfileBuilder
}

func buildPerformanceProfileWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyPerformanceProfile(),
		SchemeAttachers: paoTestSchemes,
	})
}

func buildDummyPerformanceProfile() []runtime.Object {
	return append([]runtime.Object{}, &performanceprofilev2.PerformanceProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultPerformanceProfileName,
		},
	})
}
