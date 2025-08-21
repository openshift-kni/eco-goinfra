package clusteroperator

import (
	"fmt"
	"testing"

	"github.com/golang/glog"
	configV1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	clusterOperatorGVK = schema.GroupVersionKind{
		Group:   APIGroup,
		Version: APIVersion,
		Kind:    APIKind,
	}
	defaultClusterOperatorName = "test-co"
)

func TestClusterOperatorPull(t *testing.T) {
	generateClusterOperator := func(name string) *configV1.ClusterOperator {
		return &configV1.ClusterOperator{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: configV1.ClusterOperatorSpec{},
		}
	}

	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                "etcd",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterOperator 'clusterOperatorName' cannot be empty"),
			client:              true,
		},
		{
			name:                "cotest",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("clusterOperator object cotest does not exist"),
			client:              true,
		},
		{
			name:                "cotest",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterOperator 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testClusterOperator := generateClusterOperator(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testClusterOperator)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := Pull(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.Equal(t, testClusterOperator.Name, builderResult.Object.Name)
		}
	}
}

func TestClusterOperatorExist(t *testing.T) {
	testCases := []struct {
		testClusterOperator *Builder
		expectedStatus      bool
	}{
		{
			testClusterOperator: buildValidClusterOperatorBuilder(buildClusterOperatorClientWithDummyObject()),
			expectedStatus:      true,
		},
		{
			testClusterOperator: buildInValidClusterOperatorBuilder(buildClusterOperatorClientWithDummyObject()),
			expectedStatus:      false,
		},
		{
			testClusterOperator: buildValidClusterOperatorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:      false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testClusterOperator.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestClusterOperatorGet(t *testing.T) {
	testCases := []struct {
		testClusterOperator *Builder
		expectedError       error
	}{
		{
			testClusterOperator: buildValidClusterOperatorBuilder(buildClusterOperatorClientWithDummyObject()),
			expectedError:       nil,
		},
		{
			testClusterOperator: buildInValidClusterOperatorBuilder(buildClusterOperatorClientWithDummyObject()),
			expectedError:       fmt.Errorf("the clusterOperator 'name' cannot be empty"),
		},
		{
			testClusterOperator: buildValidClusterOperatorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:       fmt.Errorf("clusteroperators.config.openshift.io \"test-co\" not found"),
		},
	}

	for _, testCase := range testCases {
		clusterOperatorObj, err := testCase.testClusterOperator.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, clusterOperatorObj, testCase.testClusterOperator.Definition)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestClusterOperatorHasDesiredVersion(t *testing.T) {
	testCases := []struct {
		desiredVersion      string
		expectedOutput      bool
		testClusterOperator *Builder
		expectedError       error
	}{
		{
			desiredVersion:      "",
			expectedOutput:      false,
			testClusterOperator: buildValidClusterOperatorBuilder(buildClusterOperatorClientWithDummyObject()),
			expectedError:       fmt.Errorf("desiredVersion can't be empty"),
		},
		{
			desiredVersion: "4.14.0",
			expectedOutput: true,
			testClusterOperator: buildFakeClusterOperatorWithVersions(buildClusterOperatorClientWithDummyObject(),
				[]string{"4.13.2", "4.14.0"}),
			expectedError: nil,
		},
		{
			desiredVersion:      "4.14.0",
			expectedOutput:      false,
			testClusterOperator: nil,
			expectedError:       fmt.Errorf("error: received nil ClusterOperator builder"),
		},
		{
			desiredVersion: "4.14.0",
			expectedOutput: false,
			testClusterOperator: buildFakeClusterOperatorWithVersions(buildClusterOperatorClientWithDummyObject(),
				[]string{"4.13.2", "4.13.0"}),
			expectedError: nil,
		},
		{
			desiredVersion: "4.14.0",
			expectedOutput: false,
			testClusterOperator: buildFakeClusterOperatorWithVersions(buildClusterOperatorClientWithDummyObject(),
				[]string{"4.13.2", "invalid"}),
			expectedError: nil,
		},
		{
			desiredVersion:      "4.14.0",
			expectedOutput:      false,
			testClusterOperator: buildFakeClusterOperatorWithVersions(buildClusterOperatorClientWithDummyObject(), []string{}),
			expectedError:       fmt.Errorf("undefined cluster operator status versions"),
		},
		{
			desiredVersion:      "4.14.0",
			expectedOutput:      false,
			testClusterOperator: buildNilClientClusterOperatorBuilder(),
			expectedError:       fmt.Errorf("ClusterOperator builder cannot have nil apiClient"),
		},
	}
	for _, testCase := range testCases {
		result, err := testCase.testClusterOperator.HasDesiredVersion(testCase.desiredVersion)
		assert.Equal(t, testCase.expectedOutput, result)
		assert.Equal(t, testCase.expectedError, err)
	}
}

func buildValidClusterOperatorBuilder(apiClient *clients.Settings) *Builder {
	return newBuilder(apiClient, defaultClusterOperatorName, configV1.ClusterOperatorStatus{})
}

func buildInValidClusterOperatorBuilder(apiClient *clients.Settings) *Builder {
	return newBuilder(apiClient, "", configV1.ClusterOperatorStatus{})
}

func buildNilClientClusterOperatorBuilder() *Builder {
	return newBuilder(nil, defaultClusterOperatorName, configV1.ClusterOperatorStatus{})
}

func buildClusterOperatorClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyClusterOperatorConfig(),
		GVK:            []schema.GroupVersionKind{clusterOperatorGVK},
	})
}

func buildDummyClusterOperatorConfig() []runtime.Object {
	return append([]runtime.Object{}, &configV1.ClusterOperator{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultClusterOperatorName,
		},
		Spec: configV1.ClusterOperatorSpec{},
	})
}

func buildFakeClusterOperatorListWithDesiredVersion(apiClient *clients.Settings) []*Builder {
	desiredVersion := "4.14.0"

	return buildFakeClusterOperatorListWithVersions(apiClient, []string{desiredVersion,
		"4.13.24", "4.13.25"}, []string{"4.13.23", desiredVersion, "4.13.25"})
}

func buildFakeClusterOperatorListWithoutDesiredVersion(apiClient *clients.Settings) []*Builder {
	return buildFakeClusterOperatorListWithVersions(apiClient, []string{"4.13.23",
		"4.13.24", "4.13.25"}, []string{"4.13.23", "", "4.13.25"})
}

func buildFakeClusterOperatorListWithVersions(apiClient *clients.Settings,
	versionList1, versionList2 []string) []*Builder {
	listWithVersions := []*Builder{buildFakeClusterOperatorWithVersions(apiClient, versionList1),
		buildFakeClusterOperatorWithVersions(apiClient, versionList2)}

	return listWithVersions
}

// buildFakeClusterOperatorWithVersions creates a fake clusterOperator with
// a specific list of ClusterOperatorStatusVersions for testing purpose.
func buildFakeClusterOperatorWithVersions(apiClient *clients.Settings, versions []string) *Builder {
	clusterOperatorStatusVersions := createFakeClusterOperatorStatusVersions(versions)

	clusterOperatorStatus := createFakeClusterOperatorStatus(clusterOperatorStatusVersions)

	clusterOperator := newBuilder(apiClient, defaultClusterOperatorName, clusterOperatorStatus)
	clusterOperator.Object = clusterOperator.Definition

	return clusterOperator
}

// createFakeClusterOperatorStatusVersions creates a fake list of ClusterOperatorStatusVersions.
func createFakeClusterOperatorStatusVersions(versions []string) []configV1.OperandVersion {
	var clusterOperatorStatusVersions []configV1.OperandVersion

	for _, version := range versions {
		clusterOperatorStatusVersions = append(clusterOperatorStatusVersions,
			configV1.OperandVersion{Name: defaultClusterOperatorName, Version: version})
	}

	return clusterOperatorStatusVersions
}

// createFakeClusterOperatorStatus creates a fake lusterOperatorStatus containing a list of OperandVersions.
func createFakeClusterOperatorStatus(
	clusterOperatorStatusVersions []configV1.OperandVersion) configV1.ClusterOperatorStatus {
	clusterOperatorStatus := configV1.ClusterOperatorStatus{
		Versions: clusterOperatorStatusVersions,
	}

	return clusterOperatorStatus
}

// newBuilder method creates new instance of builder (for the unit test propose only).
func newBuilder(apiClient *clients.Settings, name string, status configV1.ClusterOperatorStatus) *Builder {
	glog.V(100).Infof("Initializing new Builder structure with the name: %s", name)

	var client runtimeClient.Client

	if apiClient != nil {
		client = apiClient.Client
	}

	builder := &Builder{
		apiClient: client,
		Definition: &configV1.ClusterOperator{
			ObjectMeta: metav1.ObjectMeta{
				Name:            name,
				ResourceVersion: "999",
			},
			Status: status,
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the clusterOperator is empty")

		builder.errorMsg = "the clusterOperator 'name' cannot be empty"

		return builder
	}

	return builder
}
