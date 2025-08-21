package nrop

import (
	"fmt"
	"testing"

	nropv1 "github.com/openshift-kni/numaresources-operator/api/numaresourcesoperator/v1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultNROPName              = "numaresourcesoperator"
	infoRefreshPeriodic          = nropv1.InfoRefreshPeriodic
	infoRefreshEvents            = nropv1.InfoRefreshEvents
	infoRefreshPeriodicAndEvents = nropv1.InfoRefreshPeriodicAndEvents
)

func TestPull(t *testing.T) {
	generateNROP := func(name string) *nropv1.NUMAResourcesOperator {
		return &nropv1.NUMAResourcesOperator{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
	}

	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                defaultNROPName,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("NUMAResourcesOperator 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "nroptest",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("NUMAResourcesOperator object nroptest does not exist"),
			client:              true,
		},
		{
			name:                "nroptest",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("NUMAResourcesOperator 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testNROP := generateNROP(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testNROP)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := Pull(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testNROP.Name, builderResult.Object.Name)
			assert.Nil(t, err)
		}
	}
}

func TestNewNROPBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		expectedError string
	}{
		{
			name:          defaultNROPName,
			expectedError: "",
		},
		{
			name:          "",
			expectedError: "NUMAResourcesOperator 'name' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testNROPBuilder := NewBuilder(testSettings, testCase.name)
		assert.NotNil(t, testNROPBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testNROPBuilder.Definition.Name)
			assert.Equal(t, "", testNROPBuilder.errorMsg)
		} else {
			assert.Equal(t, testCase.expectedError, testNROPBuilder.errorMsg)
		}
	}
}

func TestNROPExists(t *testing.T) {
	testCases := []struct {
		testNROP       *Builder
		expectedStatus bool
	}{
		{
			testNROP:       buildValidNROPBuilder(buildNROPClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testNROP:       buildInValidNROPBuilder(buildNROPClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testNROP:       buildValidNROPBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testNROP.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestNROPGet(t *testing.T) {
	testCases := []struct {
		testNROP      *Builder
		expectedError error
	}{
		{
			testNROP:      buildValidNROPBuilder(buildNROPClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testNROP:      buildInValidNROPBuilder(buildNROPClientWithDummyObject()),
			expectedError: fmt.Errorf("NUMAResourcesOperator 'name' cannot be empty"),
		},
		{
			testNROP: buildValidNROPBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("numaresourcesoperators.nodetopology.openshift.io \"numaresourcesoperator\" " +
				"not found"),
		},
	}

	for _, testCase := range testCases {
		nropObj, err := testCase.testNROP.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, nropObj.Name, testCase.testNROP.Definition.Name)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestNROPCreate(t *testing.T) {
	testCases := []struct {
		testNROP      *Builder
		expectedError string
	}{
		{
			testNROP:      buildValidNROPBuilder(buildNROPClientWithDummyObject()),
			expectedError: "",
		},
		{
			testNROP:      buildInValidNROPBuilder(buildNROPClientWithDummyObject()),
			expectedError: "NUMAResourcesOperator 'name' cannot be empty",
		},
		{
			testNROP:      buildValidNROPBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testNROPBuilder, err := testCase.testNROP.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testNROPBuilder.Definition.Name, testNROPBuilder.Object.Name)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestNROPDelete(t *testing.T) {
	testCases := []struct {
		testNROP      *Builder
		expectedError error
	}{
		{
			testNROP:      buildValidNROPBuilder(buildNROPClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testNROP:      buildInValidNROPBuilder(buildNROPClientWithDummyObject()),
			expectedError: fmt.Errorf("NUMAResourcesOperator 'name' cannot be empty"),
		},
		{
			testNROP:      buildValidNROPBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testNROP.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testNROP.Object)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestNROPUpdate(t *testing.T) {
	testCases := []struct {
		testNROP      *Builder
		expectedError string
		mcpSelector   metav1.LabelSelector
	}{
		{
			testNROP:      buildValidNROPBuilder(buildNROPClientWithDummyObject()),
			expectedError: "",
			mcpSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"machineconfiguration.openshift.io/role": "mcp-name"}},
		},
		{
			testNROP:      buildValidNROPBuilder(buildNROPWithMCPSelectorClientWithDummyObject()),
			expectedError: "",
			mcpSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"machineconfiguration.openshift.io/role": "mcp-name"}},
		},
		{
			testNROP:      buildInValidNROPBuilder(buildNROPClientWithDummyObject()),
			expectedError: "NUMAResourcesOperator 'name' cannot be empty",
			mcpSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"machineconfiguration.openshift.io/role": "mcp-name"}},
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, []nropv1.NodeGroup(nil), testCase.testNROP.Definition.Spec.NodeGroups)
		assert.Nil(t, nil, testCase.testNROP.Object)
		testCase.testNROP.WithMCPSelector(nropv1.NodeGroupConfig{}, testCase.mcpSelector)
		_, err := testCase.testNROP.Update()

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.mcpSelector.MatchLabels,
				testCase.testNROP.Definition.Spec.NodeGroups[0].MachineConfigPoolSelector.MatchLabels)
		}
	}
}

//nolint:funlen
func TestNROPWithMCPSelector(t *testing.T) {
	testCases := []struct {
		config                nropv1.NodeGroupConfig
		mcpSelector           metav1.LabelSelector
		expectedErrMsg        string
		predefinedMCPSelector bool
		originalNodeSelector  metav1.LabelSelector
	}{
		{
			config: nropv1.NodeGroupConfig{
				InfoRefreshMode: &infoRefreshPeriodic,
			},
			mcpSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"test-mcp-selector-key": "test-mcp-selector-value"}},
			expectedErrMsg:        "",
			predefinedMCPSelector: false,
			originalNodeSelector:  metav1.LabelSelector{},
		},
		{
			config: nropv1.NodeGroupConfig{
				InfoRefreshMode: &infoRefreshEvents,
			},
			mcpSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"test-mcp-selector-key": "test-mcp-selector-value"}},
			expectedErrMsg:        "",
			predefinedMCPSelector: false,
			originalNodeSelector:  metav1.LabelSelector{},
		},
		{
			config: nropv1.NodeGroupConfig{
				InfoRefreshMode: &infoRefreshPeriodicAndEvents,
			},
			mcpSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"test-mcp-selector-key": "test-mcp-selector-value"}},
			expectedErrMsg:        "",
			predefinedMCPSelector: false,
			originalNodeSelector:  metav1.LabelSelector{},
		},
		{
			config: nropv1.NodeGroupConfig{},
			mcpSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"test-mcp-selector-key": "test-mcp-selector-value"}},
			expectedErrMsg:        "",
			predefinedMCPSelector: false,
			originalNodeSelector:  metav1.LabelSelector{},
		},
		{
			config: nropv1.NodeGroupConfig{
				InfoRefreshMode: &infoRefreshPeriodic,
			},
			mcpSelector:           metav1.LabelSelector{},
			expectedErrMsg:        "NUMAResourcesOperator 'machineConfigPoolSelector' cannot be empty",
			predefinedMCPSelector: false,
			originalNodeSelector:  metav1.LabelSelector{},
		},
		{
			config: nropv1.NodeGroupConfig{
				InfoRefreshMode: &infoRefreshPeriodic,
			},
			mcpSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"test-mcp-selector-key": "test-mcp-selector-value"}},
			expectedErrMsg:        "",
			predefinedMCPSelector: true,
			originalNodeSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"other-mcp-selector-key": "other-mcp-selector-value"}},
		},
		{
			config: nropv1.NodeGroupConfig{
				InfoRefreshMode: &infoRefreshPeriodic,
			},
			mcpSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"test-mcp-selector-key": "test-mcp-selector-value"}},
			expectedErrMsg:        "",
			predefinedMCPSelector: true,
			originalNodeSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"test-node-selector-key": "test-node-selector-value",
					"other-node-selector-key": "other-node-selector-value"}},
		},
		{
			config: nropv1.NodeGroupConfig{
				InfoRefreshMode: &infoRefreshPeriodic,
			},
			mcpSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"test-mcp-selector-key": "test-mcp-selector-value"}},
			expectedErrMsg:        "",
			predefinedMCPSelector: true,
			originalNodeSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"test-node-selector-key": "test-node-selector-value",
					"other-node-selector-key": "other-node-selector-value"}},
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidNROPBuilder(buildNROPClientWithDummyObject())

		if testCase.predefinedMCPSelector {
			testBuilder.Definition.Spec.NodeGroups = []nropv1.NodeGroup{{
				Config: &nropv1.NodeGroupConfig{
					InfoRefreshMode: &infoRefreshPeriodic,
				},
				MachineConfigPoolSelector: &testCase.originalNodeSelector,
			}}
		}

		testBuilder.WithMCPSelector(testCase.config, testCase.mcpSelector)

		if testCase.expectedErrMsg == "" {
			if testCase.predefinedMCPSelector {
				assert.Equal(t, testCase.config.InfoRefreshMode,
					testBuilder.Definition.Spec.NodeGroups[1].Config.InfoRefreshMode)
				assert.Equal(t, testCase.mcpSelector.MatchLabels,
					testBuilder.Definition.Spec.NodeGroups[1].MachineConfigPoolSelector.MatchLabels)
			} else {
				assert.Equal(t, testCase.config.InfoRefreshMode,
					testBuilder.Definition.Spec.NodeGroups[0].Config.InfoRefreshMode)
				assert.Equal(t, testCase.mcpSelector.MatchLabels,
					testBuilder.Definition.Spec.NodeGroups[0].MachineConfigPoolSelector.MatchLabels)
			}
		} else {
			assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)
		}
	}
}

func buildValidNROPBuilder(apiClient *clients.Settings) *Builder {
	nropBuilder := NewBuilder(apiClient, defaultNROPName)

	return nropBuilder
}

func buildInValidNROPBuilder(apiClient *clients.Settings) *Builder {
	nropBuilder := NewBuilder(apiClient, "")

	return nropBuilder
}

func buildNROPClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyNROP(),
	})
}

func buildNROPWithMCPSelectorClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyNROPWithMCPSelector(),
	})
}

func buildDummyNROP() []runtime.Object {
	return append([]runtime.Object{}, &nropv1.NUMAResourcesOperator{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultNROPName,
		},
	})
}

func buildDummyNROPWithMCPSelector() []runtime.Object {
	return append([]runtime.Object{}, &nropv1.NUMAResourcesOperator{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultNROPName,
		},
		Spec: nropv1.NUMAResourcesOperatorSpec{
			NodeGroups: []nropv1.NodeGroup{{
				MachineConfigPoolSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"mcpSelectorKey": "mcpSelectorValue",
					}}}},
		},
	})
}
