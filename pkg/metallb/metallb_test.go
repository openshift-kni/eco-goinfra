package metallb

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	mlbtypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlboperator"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	mlbTestSchemes = []clients.SchemeAttacher{
		mlbtypes.AddToScheme,
	}
	defaultMetalLbName         = "metallbio"
	defaultMetalLbNsName       = "test-namespace"
	defaultMetalLbNodeSelector = map[string]string{"test": "test"}
)

func TestMetalLbPull(t *testing.T) {
	generateMetalLb := func(name, namespace string) *mlbtypes.MetalLB {
		return &mlbtypes.MetalLB{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: mlbtypes.MetalLBSpec{},
		}
	}

	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                "metallbio",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("metallb 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "metallbio",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("metallb 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "metallbio",
			namespace:           "test-namespace",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("metallb object metallbio does not exist in namespace test-namespace"),
			client:              true,
		},
		{
			name:                "metallbio",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("metallb 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testMetalLb := generateMetalLb(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testMetalLb)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: mlbTestSchemes,
			})
		}

		builderResult, err := Pull(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

func TestMetalLbNewBuilder(t *testing.T) {
	generateMetalLb := NewBuilder

	testCases := []struct {
		name          string
		namespace     string
		label         map[string]string
		expectedError string
	}{
		{
			name:          "metallbio",
			namespace:     "test-namespace",
			label:         map[string]string{"test": "test"},
			expectedError: "",
		},
		{
			name:          "",
			namespace:     "test-namespace",
			label:         map[string]string{"test": "test"},
			expectedError: "metallb 'name' cannot be empty",
		},
		{
			name:          "metallbio",
			namespace:     "",
			label:         map[string]string{"test": "test"},
			expectedError: "metallb 'nsname' cannot be empty",
		},
		{
			name:          "metallbio",
			namespace:     "test-namespace",
			label:         map[string]string{},
			expectedError: "metallb 'nodeSelector' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{
			SchemeAttachers: mlbTestSchemes,
		})
		testMetalLbBuilder := generateMetalLb(testSettings, testCase.name, testCase.namespace, testCase.label)
		assert.Equal(t, testCase.expectedError, testMetalLbBuilder.errorMsg)
		assert.NotNil(t, testMetalLbBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testMetalLbBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testMetalLbBuilder.Definition.Namespace)
		}
	}
}

func TestMetalLbExist(t *testing.T) {
	testCases := []struct {
		testMetalLb    *Builder
		expectedStatus bool
	}{
		{
			testMetalLb:    buildValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testMetalLb:    buildInValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testMetalLb.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestMetalLbGet(t *testing.T) {
	testCases := []struct {
		testMetalLb   *Builder
		expectedError error
	}{
		{
			testMetalLb:   buildValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testMetalLb:   buildInValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError: fmt.Errorf("metallb 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		metalLb, err := testCase.testMetalLb.Get()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, metalLb.Name, testCase.testMetalLb.Definition.Name)
		}
	}
}

func TestMetalLbCreate(t *testing.T) {
	testCases := []struct {
		testMetalLb   *Builder
		expectedError error
	}{
		{
			testMetalLb:   buildValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testMetalLb:   buildInValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError: fmt.Errorf("metallb 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		testMetalLbBuilder, err := testCase.testMetalLb.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testMetalLbBuilder.Definition.Name, testMetalLbBuilder.Object.Name)
		}
	}
}

func TestMetalLbDelete(t *testing.T) {
	testCases := []struct {
		testMetalLb   *Builder
		expectedError error
	}{
		{
			testMetalLb:   buildValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testMetalLb:   buildInValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError: fmt.Errorf("metallb 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testMetalLb.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testMetalLb.Object)
		}
	}
}

func TestMetalLbUpdate(t *testing.T) {
	testCases := []struct {
		testMetalLb   *Builder
		expectedError error
		nodeSelector  map[string]string
	}{
		{
			testMetalLb:   buildValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError: nil,
			nodeSelector:  map[string]string{"test2": "test2"},
		},
		{
			testMetalLb:   buildInValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError: fmt.Errorf("metallb 'name' cannot be empty"),
			nodeSelector:  map[string]string{"test2": "test2"},
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, defaultMetalLbNodeSelector, testCase.testMetalLb.Definition.Spec.SpeakerNodeSelector)
		assert.Nil(t, nil, testCase.testMetalLb.Object)
		testCase.testMetalLb.WithSpeakerNodeSelector(testCase.nodeSelector)
		_, err := testCase.testMetalLb.Update(false)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.nodeSelector, testCase.testMetalLb.Definition.Spec.SpeakerNodeSelector)
		}
	}
}

func TestMetalLbRemoveLabel(t *testing.T) {
	testCases := []struct {
		testMetalLb   *Builder
		key           string
		expectedError string
	}{
		{
			testMetalLb:   buildValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError: "",
			key:           "test",
		},
		{
			testMetalLb:   buildValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError: "error to remove empty key from metalLbIo",
			key:           "",
		},
		{
			testMetalLb:   buildInValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError: "metallb 'name' cannot be empty",
			key:           "",
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, defaultMetalLbNodeSelector, testCase.testMetalLb.Definition.Spec.SpeakerNodeSelector)
		assert.Nil(t, nil, testCase.testMetalLb.Object)
		metalLbBuilder := testCase.testMetalLb.RemoveLabel(testCase.key)
		assert.Equal(t, testCase.expectedError, metalLbBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, map[string]string{}, metalLbBuilder.Definition.Spec.SpeakerNodeSelector)
		}
	}
}

func TestMetalLbWithSpeakerNodeSelector(t *testing.T) {
	testCases := []struct {
		testMetalLb         *Builder
		speakerNodeSelector map[string]string
		expectedError       string
	}{
		{
			testMetalLb:         buildValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError:       "",
			speakerNodeSelector: map[string]string{"node": "nodes"},
		},
		{
			testMetalLb:         buildValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError:       "can not accept empty label and redefine metallb NodeSelector",
			speakerNodeSelector: map[string]string{},
		},
		{
			testMetalLb:         buildInValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError:       "metallb 'name' cannot be empty",
			speakerNodeSelector: map[string]string{"node": "nodes"},
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, defaultMetalLbNodeSelector, testCase.testMetalLb.Definition.Spec.SpeakerNodeSelector)
		metalLbBuilder := testCase.testMetalLb.WithSpeakerNodeSelector(testCase.speakerNodeSelector)
		assert.Equal(t, testCase.expectedError, metalLbBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, metalLbBuilder.Definition.Spec.SpeakerNodeSelector, testCase.speakerNodeSelector)
		}
	}
}

func TestMetalLbWithOptions(t *testing.T) {
	testSettings := buildMetalLbTestClientWithDummyObject()
	testBuilder := buildValidMetalLbBuilder(testSettings).WithOptions(
		func(builder *Builder) (*Builder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidMetalLbBuilder(testSettings).WithOptions(
		func(builder *Builder) (*Builder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestMetalLbWithFRRConfigAlwaysBlock(t *testing.T) {
	testCases := []struct {
		testMetalLb   *Builder
		prefixes      []string
		expectedError string
	}{
		{
			testMetalLb:   buildValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError: "",
			prefixes:      []string{"192.168.1.0/24"},
		},
		{
			testMetalLb:   buildValidMetalLbBuilder(buildMetalLbTestClientWithDummyObject()),
			expectedError: "the prefix 192.168.1. is not a valid CIDR",
			prefixes:      []string{"192.168.1."},
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, defaultMetalLbNodeSelector, testCase.testMetalLb.Definition.Spec.SpeakerNodeSelector)
		metalLbBuilder := testCase.testMetalLb.WithFRRConfigAlwaysBlock(testCase.prefixes)
		assert.Equal(t, testCase.expectedError, metalLbBuilder.errorMsg)
	}
}

func TestGetMetalLbIoGVR(t *testing.T) {
	assert.Equal(t, GetMetalLbIoGVR(),
		schema.GroupVersionResource{
			Group: APIGroup, Version: APIVersion, Resource: "metallbs",
		})
}

func buildValidMetalLbBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(
		apiClient, defaultMetalLbName, defaultMetalLbNsName, map[string]string{"test": "test"})
}

func buildInValidMetalLbBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(
		apiClient, "", defaultMetalLbNsName, defaultMetalLbNodeSelector)
}

func buildMetalLbTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyMetalLb(),
		SchemeAttachers: mlbTestSchemes,
	})
}

func buildDummyMetalLb() []runtime.Object {
	return append([]runtime.Object{}, &mlbtypes.MetalLB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultMetalLbName,
			Namespace: defaultMetalLbNsName,
		},
		Spec: mlbtypes.MetalLBSpec{
			SpeakerNodeSelector: defaultMetalLbNodeSelector,
		},
	})
}
