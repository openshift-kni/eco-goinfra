package clusterlogging

import (
	"fmt"
	"testing"

	observabilityv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	apiGroup                            = "observability.openshift.io"
	apiVersion                          = "v1"
	kind                                = "ClusterLogForwarder"
	metaDataNameErrorMgs                = "metadata.name: Required value: name is required"
	defaultClusterLogForwarderName      = "instance"
	defaultClusterLogForwarderNamespace = "openshift-logging"
	defaultOutputs                      = []observabilityv1.OutputSpec(nil)
	defaultPipelines                    = []observabilityv1.PipelineSpec(nil)
	newOutputs                          = observabilityv1.OutputSpec{
		Name: "elasticsearch-external",
		Type: "elasticsearch",
		Kafka: &observabilityv1.Kafka{
			URL: "https://dummy-domain.amazonaws.com:443",
		},
	}
	newPipelines = observabilityv1.PipelineSpec{
		Name:       "",
		OutputRefs: []string{"elasticsearch-external"},
		InputRefs:  []string{"application", "infra"},
	}
	clov1TestSchemes = []clients.SchemeAttacher{
		observabilityv1.AddToScheme,
	}
)

func TestClusterLogForwarderPull(t *testing.T) {
	generateClusterLogForwarder := func(name, namespace string) *observabilityv1.ClusterLogForwarder {
		return &observabilityv1.ClusterLogForwarder{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: observabilityv1.ClusterLogForwarderSpec{
				Outputs:   []observabilityv1.OutputSpec{},
				Pipelines: []observabilityv1.PipelineSpec{},
			},
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
			name:                "test",
			namespace:           "openshift-logging",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "openshift-logging",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterlogforwarder 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "test",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterlogforwarder 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "clftest",
			namespace:           "openshift-logging",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("clusterlogforwarder object clftest does not exist in namespace openshift-logging"),
			client:              true,
		},
		{
			name:                "clftest",
			namespace:           "openshift-logging",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterlogforwarder 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testClusterLogForwarder := generateClusterLogForwarder(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testClusterLogForwarder)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: clov1TestSchemes,
			})
		}

		builderResult, err := PullClusterLogForwarder(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testClusterLogForwarder.Name, builderResult.Object.Name)
		}
	}
}

func TestNewClusterLogForwarderBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
	}{
		{
			name:          defaultClusterLogForwarderName,
			namespace:     defaultClusterLogForwarderNamespace,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     defaultClusterLogForwarderNamespace,
			expectedError: "clusterlogforwarder 'name' cannot be empty",
		},
		{
			name:          defaultClusterLogForwarderName,
			namespace:     "",
			expectedError: "clusterlogforwarder 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testClusterLogForwarderBuilder := NewClusterLogForwarderBuilder(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, testClusterLogForwarderBuilder.errorMsg)
		assert.NotNil(t, testClusterLogForwarderBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testClusterLogForwarderBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testClusterLogForwarderBuilder.Definition.Namespace)
		}
	}
}

func TestClusterLogForwarderExist(t *testing.T) {
	testCases := []struct {
		testClusterLogForwarder *ClusterLogForwarderBuilder
		expectedStatus          bool
	}{
		{
			testClusterLogForwarder: buildValidClusterLogForwarderBuilder(buildClusterLogForwarderClientWithDummyObject()),
			expectedStatus:          true,
		},
		{
			testClusterLogForwarder: buildInValidClusterLogForwarderBuilder(buildClusterLogForwarderClientWithDummyObject()),
			expectedStatus:          false,
		},
		{
			testClusterLogForwarder: buildValidClusterLogForwarderBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:          false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testClusterLogForwarder.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestClusterLogForwarderGet(t *testing.T) {
	testCases := []struct {
		testClusterLogForwarder *ClusterLogForwarderBuilder
		expectedError           string
	}{
		{
			testClusterLogForwarder: buildValidClusterLogForwarderBuilder(buildClusterLogForwarderClientWithDummyObject()),
			expectedError:           "",
		},
		{
			testClusterLogForwarder: buildInValidClusterLogForwarderBuilder(buildClusterLogForwarderClientWithDummyObject()),
			expectedError:           "clusterlogforwarder 'name' cannot be empty",
		},
		{
			testClusterLogForwarder: buildValidClusterLogForwarderBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:           "clusterlogforwarders.observability.openshift.io \"instance\" not found",
		},
	}

	for _, testCase := range testCases {
		clusterLogForwarderObj, err := testCase.testClusterLogForwarder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testClusterLogForwarder.Definition.Name, clusterLogForwarderObj.Name)
			assert.Equal(t, testCase.testClusterLogForwarder.Definition.Namespace, clusterLogForwarderObj.Namespace)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestClusterLogForwarderCreate(t *testing.T) {
	testCases := []struct {
		testClusterLogForwarder *ClusterLogForwarderBuilder
		expectedError           string
	}{
		{
			testClusterLogForwarder: buildValidClusterLogForwarderBuilder(buildClusterLogForwarderClientWithDummyObject()),
			expectedError:           "",
		},
		{
			testClusterLogForwarder: buildInValidClusterLogForwarderBuilder(buildClusterLogForwarderClientWithDummyObject()),
			expectedError:           "clusterlogforwarder 'name' cannot be empty",
		},
		{
			testClusterLogForwarder: buildValidClusterLogForwarderBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:           "resourceVersion can not be set for Create requests",
		},
	}

	for _, testCase := range testCases {
		testClusterLogForwarderBuilder, err := testCase.testClusterLogForwarder.Create()

		if testCase.expectedError == "" && testClusterLogForwarderBuilder != nil {
			assert.Equal(t, testClusterLogForwarderBuilder.Definition.Name, testClusterLogForwarderBuilder.Object.Name)
			assert.Equal(t, testClusterLogForwarderBuilder.Definition.Namespace, testClusterLogForwarderBuilder.Object.Namespace)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestClusterLogForwarderDelete(t *testing.T) {
	testCases := []struct {
		testClusterLogForwarder *ClusterLogForwarderBuilder
		expectedError           error
	}{
		{
			testClusterLogForwarder: buildValidClusterLogForwarderBuilder(buildClusterLogForwarderClientWithDummyObject()),
			expectedError:           nil,
		},
		{
			testClusterLogForwarder: buildInValidClusterLogForwarderBuilder(buildClusterLogForwarderClientWithDummyObject()),
			expectedError:           nil,
		},
		{
			testClusterLogForwarder: buildValidClusterLogForwarderBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:           nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testClusterLogForwarder.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testClusterLogForwarder.Object)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestClusterLogForwarderUpdate(t *testing.T) {
	testCases := []struct {
		testClusterLogForwarder *ClusterLogForwarderBuilder
		expectedError           string
		outputs                 observabilityv1.OutputSpec
		pipelines               observabilityv1.PipelineSpec
	}{
		{
			testClusterLogForwarder: buildValidClusterLogForwarderBuilder(buildClusterLogForwarderClientWithDummyObject()),
			expectedError:           "",
			outputs:                 newOutputs,
			pipelines:               newPipelines,
		},
		{
			testClusterLogForwarder: buildInValidClusterLogForwarderBuilder(buildClusterLogForwarderClientWithDummyObject()),
			expectedError:           "clusterlogforwarder 'name' cannot be empty",
			outputs:                 newOutputs,
			pipelines:               newPipelines,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, defaultOutputs, testCase.testClusterLogForwarder.Definition.Spec.Outputs)
		assert.Equal(t, defaultPipelines, testCase.testClusterLogForwarder.Definition.Spec.Pipelines)
		assert.Nil(t, nil, testCase.testClusterLogForwarder.Object)
		testCase.testClusterLogForwarder.WithOutput(&testCase.outputs)
		_, err := testCase.testClusterLogForwarder.Update(true)

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, []observabilityv1.OutputSpec{testCase.outputs},
				testCase.testClusterLogForwarder.Definition.Spec.Outputs)
		}

		testCase.testClusterLogForwarder.WithPipeline(&testCase.pipelines)
		_, err = testCase.testClusterLogForwarder.Update(true)

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, []observabilityv1.PipelineSpec{testCase.pipelines},
				testCase.testClusterLogForwarder.Definition.Spec.Pipelines)
		}
	}
}

func TestWithManagementState(t *testing.T) {
	testCases := []struct {
		testManagementState observabilityv1.ManagementState
		expectedError       string
	}{
		{
			testManagementState: observabilityv1.ManagementStateManaged,
			expectedError:       "",
		},
		{
			testManagementState: observabilityv1.ManagementStateUnmanaged,
			expectedError:       "",
		},
		{
			testManagementState: "",
			expectedError: "the management state of the clusterlogforwarder is unsupported: \"\";" +
				"accepted only Managed or Unmanaged states",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidClusterLogForwarderBuilder(buildClusterLogForwarderClientWithDummyObject())

		result := testBuilder.WithManagementState(testCase.testManagementState)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testManagementState, result.Definition.Spec.ManagementState)
		} else {
			assert.Equal(t, testCase.expectedError, result.errorMsg)
		}
	}
}

func TestWithServiceAccount(t *testing.T) {
	testCases := []struct {
		testServiceAccount string
		expectedError      string
	}{
		{
			testServiceAccount: "dummy",
			expectedError:      "",
		},
		{
			testServiceAccount: "",
			expectedError:      "clusterlogforwarder 'serviceAccount' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidClusterLogForwarderBuilder(buildClusterLogForwarderClientWithDummyObject())

		result := testBuilder.WithServiceAccount(testCase.testServiceAccount)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testServiceAccount, result.Definition.Spec.ServiceAccount.Name)
		} else {
			assert.Equal(t, testCase.expectedError, result.errorMsg)
		}
	}
}

func TestWithOutput(t *testing.T) {
	testCases := []struct {
		testOutputs   *observabilityv1.OutputSpec
		expectedError string
	}{
		{
			testOutputs:   &newOutputs,
			expectedError: "",
		},
		{
			testOutputs:   nil,
			expectedError: "'outputSpec' parameter is empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidClusterLogForwarderBuilder(buildClusterLogForwarderClientWithDummyObject())

		result := testBuilder.WithOutput(testCase.testOutputs)

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, []observabilityv1.OutputSpec{*testCase.testOutputs}, result.Definition.Spec.Outputs)
		}
	}
}

func TestWithPipeline(t *testing.T) {
	testCases := []struct {
		testPipeline  *observabilityv1.PipelineSpec
		expectedError string
	}{
		{
			testPipeline:  &newPipelines,
			expectedError: "",
		},
		{
			testPipeline:  nil,
			expectedError: "'pipelineSpec' parameter is empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidClusterLogForwarderBuilder(buildClusterLogForwarderClientWithDummyObject())

		result := testBuilder.WithPipeline(testCase.testPipeline)

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, []observabilityv1.PipelineSpec{*testCase.testPipeline}, result.Definition.Spec.Pipelines)
		}
	}
}

func buildValidClusterLogForwarderBuilder(apiClient *clients.Settings) *ClusterLogForwarderBuilder {
	clusterLogForwarderBuilder := NewClusterLogForwarderBuilder(
		apiClient, defaultClusterLogForwarderName, defaultClusterLogForwarderNamespace)
	clusterLogForwarderBuilder.Definition.TypeMeta = metav1.TypeMeta{
		Kind:       kind,
		APIVersion: fmt.Sprintf("%s/%s", apiGroup, apiVersion),
	}
	clusterLogForwarderBuilder.Definition.ResourceVersion = "999"

	return clusterLogForwarderBuilder
}

func buildInValidClusterLogForwarderBuilder(apiClient *clients.Settings) *ClusterLogForwarderBuilder {
	clusterLogForwarderBuilder := NewClusterLogForwarderBuilder(
		apiClient, "", defaultClusterLogForwarderNamespace)
	clusterLogForwarderBuilder.Definition.TypeMeta = metav1.TypeMeta{
		Kind:       kind,
		APIVersion: fmt.Sprintf("%s/%s", apiGroup, apiVersion),
	}
	clusterLogForwarderBuilder.Definition.ResourceVersion = "999"

	return clusterLogForwarderBuilder
}

func buildClusterLogForwarderClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyClusterLogForwarder(),
		SchemeAttachers: clov1TestSchemes,
	})
}

func buildDummyClusterLogForwarder() []runtime.Object {
	return append([]runtime.Object{}, &observabilityv1.ClusterLogForwarder{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultClusterLogForwarderName,
			Namespace: defaultClusterLogForwarderNamespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       kind,
			APIVersion: fmt.Sprintf("%s/%s", apiGroup, apiVersion),
		},
		Spec: observabilityv1.ClusterLogForwarderSpec{
			Outputs:   []observabilityv1.OutputSpec{},
			Pipelines: []observabilityv1.PipelineSpec{},
		},
	})
}
