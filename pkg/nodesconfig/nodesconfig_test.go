package nodesconfig

import (
	"fmt"
	"testing"

	"github.com/golang/glog"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	configV1 "github.com/openshift/api/config/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	nodesConfigGVK = schema.GroupVersionKind{
		Group:   APIGroup,
		Version: APIVersion,
		Kind:    APIKind,
	}
	defaultNodesConfigName = "cluster"
	defaultCGroupMode      = configV1.CgroupModeEmpty
)

func TestNodeConfigPull(t *testing.T) {
	generateNodeConfig := func(name string) *configV1.Node {
		return &configV1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: configV1.NodeSpec{
				CgroupMode: defaultCGroupMode,
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
			name:                "test",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("nodesConfig 'nodesConfigObjName' cannot be empty"),
			client:              true,
		},
		{
			name:                "argocdtest",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("nodesConfig object argocdtest does not exist"),
			client:              true,
		},
		{
			name:                "argocdtest",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("nodesConfig Config 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testNodeConfig := generateNodeConfig(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testNodeConfig)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := Pull(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testNodeConfig.Name, builderResult.Object.Name)
		}
	}
}

func TestNodeConfigExist(t *testing.T) {
	testCases := []struct {
		testNodesConfig *Builder
		expectedStatus  bool
	}{
		{
			testNodesConfig: buildValidNodeConfigBuilder(buildNodeConfigClientWithDummyObject()),
			expectedStatus:  true,
		},
		{
			testNodesConfig: buildInValidNodeConfigBuilder(buildNodeConfigClientWithDummyObject()),
			expectedStatus:  false,
		},
		{
			testNodesConfig: buildInValidNodeConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:  false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testNodesConfig.Exists()
		assert.Equal(t, exist, testCase.expectedStatus)
	}
}

func TestNodesConfigGet(t *testing.T) {
	testCases := []struct {
		testNodesConfig *Builder
		expectedError   error
	}{
		{
			testNodesConfig: buildValidNodeConfigBuilder(buildNodeConfigClientWithDummyObject()),
			expectedError:   nil,
		},
		{
			testNodesConfig: buildInValidNodeConfigBuilder(buildNodeConfigClientWithDummyObject()),
			expectedError:   fmt.Errorf("the nodesConfig 'name' cannot be empty"),
		},
		{
			testNodesConfig: buildInValidNodeConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:   fmt.Errorf("the nodesConfig 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		nodesConfigObj, err := testCase.testNodesConfig.Get()
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.Equal(t, nodesConfigObj, testCase.testNodesConfig.Definition)
		}
	}
}

func TestNodesConfigUpdate(t *testing.T) {
	testCases := []struct {
		testNodesConfig *Builder
		expectedError   error
		cGroupMode      configV1.CgroupMode
	}{
		{
			testNodesConfig: buildValidNodeConfigBuilder(buildNodeConfigClientWithDummyObject()),
			expectedError:   nil,
			cGroupMode:      configV1.CgroupModeV2,
		},
		{
			testNodesConfig: buildInValidNodeConfigBuilder(buildNodeConfigClientWithDummyObject()),
			expectedError:   fmt.Errorf("the nodesConfig 'name' cannot be empty"),
			cGroupMode:      configV1.CgroupModeV2,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, defaultCGroupMode, testCase.testNodesConfig.Definition.Spec.CgroupMode)
		assert.Nil(t, nil, testCase.testNodesConfig.Object)
		testCase.testNodesConfig.WithCGroupMode(testCase.cGroupMode)
		_, err := testCase.testNodesConfig.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.cGroupMode, testCase.testNodesConfig.Definition.Spec.CgroupMode)
		}
	}
}

func TestGetNodesConfigGVR(t *testing.T) {
	assert.Equal(t, GetNodesConfigIoGVR(),
		schema.GroupVersionResource{
			Group: APIGroup, Version: APIVersion, Resource: APIKind,
		})
}

func TestNodesConfigCGroupMode(t *testing.T) {
	testCases := []struct {
		testCGroupMode    configV1.CgroupMode
		expectedError     bool
		expectedErrorText string
	}{
		{ // Test Case 1 - empty cgroup mode
			testCGroupMode:    configV1.CgroupModeEmpty,
			expectedError:     false,
			expectedErrorText: "",
		},
		{ // Test Case 2 - valid cgroup mode
			testCGroupMode:    configV1.CgroupModeV1,
			expectedError:     false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidNodeConfigBuilder(buildNodeConfigClientWithDummyObject())

		result := testBuilder.WithCGroupMode(testCase.testCGroupMode)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			// Assert that the cGroup Mode was added to the Builder
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testCGroupMode, result.Definition.Spec.CgroupMode)
		}
	}
}

func buildValidNodeConfigBuilder(apiClient *clients.Settings) *Builder {
	return newBuilder(apiClient, defaultNodesConfigName, defaultCGroupMode)
}

func buildInValidNodeConfigBuilder(apiClient *clients.Settings) *Builder {
	return newBuilder(apiClient, "", defaultCGroupMode)
}

func buildNodeConfigClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyNodesConfig(),
		GVK:            []schema.GroupVersionKind{nodesConfigGVK},
	})
}

func buildDummyNodesConfig() []runtime.Object {
	return append([]runtime.Object{}, &configV1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultNodesConfigName,
		},
		Spec: configV1.NodeSpec{
			CgroupMode: defaultCGroupMode,
		},
	})
}

// newBuilder method creates new instance of builder (for the unit test propose only).
func newBuilder(apiClient *clients.Settings, name string, cgroupMode configV1.CgroupMode) *Builder {
	glog.V(100).Infof("Initializing new Builder structure with the name: %s", name)

	builder := &Builder{
		apiClient: apiClient.Client,
		Definition: &configV1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:            name,
				ResourceVersion: "999",
			},
			Spec: configV1.NodeSpec{
				CgroupMode: cgroupMode,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the nodesConfig is empty")

		builder.errorMsg = "the nodesConfig 'name' cannot be empty"

		return builder
	}

	return builder
}
