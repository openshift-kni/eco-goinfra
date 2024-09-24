package nvidiagpu

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/schemes/nvidiagpu/nvidiagputypes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testSchemes = []clients.SchemeAttacher{
		nvidiagputypes.AddToScheme,
	}
	defaultClusterPolicyName = "default-cluster-policy"
	almExample               = `[
 {
    "apiVersion": "nvidia.com/v1",
    "kind": "ClusterPolicy",
    "metadata": {
      "name": "default-cluster-policy"
    },
    "spec": {
      "driver": {
        "enabled": true
      }
    }
  }
]
`
)

func TestNvidiaGPUNewBuilderFromObjectString(t *testing.T) {
	testCases := []struct {
		almExample    string
		expectedError string
		client        bool
	}{
		{
			almExample:    almExample,
			expectedError: "",
			client:        true,
		},
		{
			almExample:    almExample,
			expectedError: "",
			client:        false,
		},
		{
			almExample:    "{ invalid: data }",
			expectedError: "error initializing ClusterPolicy from alm-examples: ",
			client:        true,
		},
		{
			almExample:    "",
			expectedError: "error initializing ClusterPolicy from alm-examples: almExample is an empty string",
			client:        true,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings
		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		nvidiaGPUNewBuilder := NewBuilderFromObjectString(testSettings, testCase.almExample)

		if testCase.expectedError == "" {
			if testCase.client {
				assert.Empty(t, nvidiaGPUNewBuilder.errorMsg)
				assert.NotNil(t, nvidiaGPUNewBuilder.Definition)
			} else {
				assert.Nil(t, nvidiaGPUNewBuilder)
			}
		} else {
			assert.Contains(t, nvidiaGPUNewBuilder.errorMsg, testCase.expectedError)
			assert.Nil(t, nvidiaGPUNewBuilder.Definition)
		}
	}
}

func TestNvidiaGPUPull(t *testing.T) {
	generateClusterPolicy := func(name string) *nvidiagputypes.ClusterPolicy {
		return &nvidiagputypes.ClusterPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: nvidiagputypes.ClusterPolicySpec{},
		}
	}

	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                "clusterpolicy",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterPolicy 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "clusterpolicy",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("ClusterPolicy object clusterpolicy does not exist"),
			client:              true,
		},
		{
			name:                "clusterpolicy",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterPolicy 'apiClient' cannot be nil"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testClusterPolicy := generateClusterPolicy(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testClusterPolicy)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects, SchemeAttachers: testSchemes})
		}

		builderResult, err := Pull(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
		}
	}
}

func TestNvidiaGPUGet(t *testing.T) {
	testCases := []struct {
		clusterPolicy *Builder
		expectedError error
	}{
		{
			clusterPolicy: buildValidClusterPolicyBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			clusterPolicy: buildInValidClusterPolicyBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("can not redefine the undefined ClusterPolicy"),
		},
		{
			clusterPolicy: buildValidClusterPolicyBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: fmt.Errorf("clusterpolicies.nvidia.com \"default-cluster-policy\" not found"),
		},
	}

	for _, testCase := range testCases {
		clusterPolicy, err := testCase.clusterPolicy.Get()
		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}

		if testCase.expectedError == nil {
			assert.Equal(t, clusterPolicy.Name, testCase.clusterPolicy.Definition.Name)
		}
	}
}

func TestNvidiaGPUExist(t *testing.T) {
	testCases := []struct {
		clusterPolicy  *Builder
		expectedStatus bool
	}{
		{
			clusterPolicy:  buildValidClusterPolicyBuilder(buildTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			clusterPolicy:  buildInValidClusterPolicyBuilder(buildTestClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			clusterPolicy: buildValidClusterPolicyBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.clusterPolicy.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestNvidiaGPUDelete(t *testing.T) {
	testCases := []struct {
		clusterPolicy *Builder
		expectedError error
	}{
		{
			clusterPolicy: buildValidClusterPolicyBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			clusterPolicy: buildValidClusterPolicyBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.clusterPolicy.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.clusterPolicy.Object)
		}
	}
}

func TestNvidiaGPUCreate(t *testing.T) {
	testCases := []struct {
		clusterPolicy *Builder
		expectedError error
	}{
		{
			clusterPolicy: buildValidClusterPolicyBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			clusterPolicy: buildValidClusterPolicyBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: nil,
		},
		{
			clusterPolicy: buildInValidClusterPolicyBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("can not redefine the undefined ClusterPolicy"),
		},
	}

	for _, testCase := range testCases {
		clusterPolicyBuilder, err := testCase.clusterPolicy.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, clusterPolicyBuilder.Definition.Name, clusterPolicyBuilder.Object.Name)
		}
	}
}

func TestNvidiaGPUUpdate(t *testing.T) {
	testCases := []struct {
		clusterPolicy *Builder
		expectedError error
		cdiEnabled    bool
	}{
		{
			clusterPolicy: buildValidClusterPolicyBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
			cdiEnabled:    true,
		},
		{
			clusterPolicy: buildValidClusterPolicyBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
			cdiEnabled:    false,
		},
	}

	for _, testCase := range testCases {
		assert.Nil(t, testCase.clusterPolicy.Definition.Spec.CDI.Enabled)
		assert.Nil(t, nil, testCase.clusterPolicy.Object)
		testCase.clusterPolicy.Definition.Spec.CDI.Enabled = &testCase.cdiEnabled
		testCase.clusterPolicy.Definition.ObjectMeta.ResourceVersion = "999"
		_, err := testCase.clusterPolicy.Update(false)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, &testCase.cdiEnabled, testCase.clusterPolicy.Definition.Spec.CDI.Enabled)
		}
	}
}

func buildInValidClusterPolicyBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilderFromObjectString(apiClient, "")
}

func buildValidClusterPolicyBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilderFromObjectString(apiClient, almExample)
}

func buildTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyClusterPolicy(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyClusterPolicy() []runtime.Object {
	return append([]runtime.Object{}, &nvidiagputypes.ClusterPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultClusterPolicyName,
		},
		Spec: nvidiagputypes.ClusterPolicySpec{},
	})
}
