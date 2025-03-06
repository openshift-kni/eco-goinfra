package olm

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	oplmV1alpha1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/olm/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPullClusterServiceVersion(t *testing.T) {
	generateClusterService := func(name, namespace string) *oplmV1alpha1.ClusterServiceVersion {
		return &oplmV1alpha1.ClusterServiceVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: oplmV1alpha1.ClusterServiceVersionSpec{
				DisplayName: "test",
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
			name:                "clusterserviceversion",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterserviceversion 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "clusterserviceversion",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterserviceversion 'namespace' cannot be empty"),
			client:              true,
		},
		{
			name:                "clusterserviceversion",
			namespace:           "test-namespace",
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf(
				"clusterserviceversion object clusterserviceversion does not exist in namespace test-namespace"),
			client: true,
		},
		{
			name:                "clusterserviceversion",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterserviceversion 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		clusterServiceVersion := generateClusterService(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, clusterServiceVersion)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects, SchemeAttachers: testSchemes})
		}

		builderResult, err := PullClusterServiceVersion(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

func TestClusterServiceVersionGet(t *testing.T) {
	testCases := []struct {
		clusterService *ClusterServiceVersionBuilder
		expectedError  string
	}{
		{
			clusterService: buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject()),
			expectedError:  "",
		},
		{
			clusterService: buildValidClusterServiceBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: "clusterserviceversions.operators.coreos.com \"clusterservice\" not found",
		},
	}

	for _, testCase := range testCases {
		clusterService, err := testCase.clusterService.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, clusterService.Name, testCase.clusterService.Definition.Name)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestClusterServiceVersionExist(t *testing.T) {
	testCases := []struct {
		clusterService *ClusterServiceVersionBuilder
		expectedStatus bool
	}{
		{
			clusterService: buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject()),
			expectedStatus: true,
		},
		{
			clusterService: buildValidClusterServiceBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.clusterService.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestClusterServiceDelete(t *testing.T) {
	testCases := []struct {
		clusterService *ClusterServiceVersionBuilder
		expectedError  error
	}{
		{
			clusterService: buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject()),
			expectedError:  nil,
		},
		{
			clusterService: buildValidClusterServiceBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.clusterService.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.clusterService.Object)
		}
	}
}

func TestClusterServiceVersionUpdate(t *testing.T) {
	testCases := []struct {
		testBuilder   *ClusterServiceVersionBuilder
		expectedError error
		displayName   string
	}{
		{
			testBuilder:   buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject()),
			expectedError: nil,
			displayName:   "newName",
		},
		{
			testBuilder: buildValidClusterServiceBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: fmt.Errorf("cannot update non-existent ClusterServiceVersion"),
			displayName:   "newName",
		},
		{
			testBuilder:   buildInvalidClusterServiceBuilder(),
			expectedError: fmt.Errorf("ClusterServiceVersion builder cannot have nil apiClient"),
			displayName:   "newName",
		},
	}

	for _, testCase := range testCases {
		assert.NotEqual(t, testCase.testBuilder.Definition.Spec.DisplayName, testCase.displayName)

		testCase.testBuilder.Definition.Spec.DisplayName = testCase.displayName

		testBuilder, err := testCase.testBuilder.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testBuilder.Object.Spec.DisplayName, testCase.displayName)
		}
	}
}

func TestClusterServiceGetAlmExamples(t *testing.T) {
	testCases := []struct {
		clusterService *ClusterServiceVersionBuilder
		almExample     map[string]string
		expectedError  error
	}{
		{
			almExample: map[string]string{"alm-examples": "test"},
			clusterService: buildValidClusterServiceBuilder(
				buildTestClientWithDummyClusterServiceObjectWitAlm(map[string]string{"alm-examples": "test"})),
			expectedError: nil,
		},
		{
			almExample: nil,
			clusterService: buildValidClusterServiceBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: fmt.Errorf("alm-examples not found in given clusterserviceversion named clusterservice"),
		},
		{
			almExample:     map[string]string{"alm-examples": "test"},
			clusterService: buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject()),
			expectedError:  fmt.Errorf("alm-examples not found in given clusterserviceversion named clusterservice"),
		},
	}

	for _, testCase := range testCases {
		almExamples, err := testCase.clusterService.GetAlmExamples()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, almExamples, testCase.almExample["alm-examples"])
		}
	}
}

func TestClusterServiceGetPhase(t *testing.T) {
	testCases := []struct {
		clusterService *ClusterServiceVersionBuilder
		expectedError  error
		phase          oplmV1alpha1.ClusterServiceVersionPhase
	}{
		{
			clusterService: buildValidClusterServiceBuilder(
				buildTestClientWithDummyClusterServiceObjectWitPhase(oplmV1alpha1.CSVPhaseSucceeded)),
			phase:         oplmV1alpha1.CSVPhaseSucceeded,
			expectedError: nil,
		},
		{
			clusterService: buildValidClusterServiceBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			phase:         oplmV1alpha1.CSVPhaseSucceeded,
			expectedError: fmt.Errorf("clusterservice clusterserviceversion not found in test-namespace namespace"),
		},
	}

	for _, testCase := range testCases {
		phase, err := testCase.clusterService.GetPhase()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.phase, phase)
		}
	}
}

func TestClusterServiceIsSuccessful(t *testing.T) {
	testCases := []struct {
		clusterService *ClusterServiceVersionBuilder
		expectedError  error
		successful     bool
	}{
		{
			clusterService: buildValidClusterServiceBuilder(
				buildTestClientWithDummyClusterServiceObjectWitPhase(oplmV1alpha1.CSVPhaseSucceeded)),
			successful:    true,
			expectedError: nil,
		},
		{
			clusterService: buildValidClusterServiceBuilder(
				buildTestClientWithDummyClusterServiceObjectWitPhase(oplmV1alpha1.CSVPhaseFailed)),
			successful:    false,
			expectedError: nil,
		},
		{
			clusterService: buildValidClusterServiceBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			successful: false,
			expectedError: fmt.Errorf("failed to get phase value for clusterservice clusterserviceversion in " +
				"test-namespace namespace due to clusterservice clusterserviceversion not found in test-namespace namespace"),
		},
	}

	for _, testCase := range testCases {
		successful, err := testCase.clusterService.IsSuccessful()
		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCase.successful, successful)
		}
	}
}

func buildValidClusterServiceBuilder(apiClient *clients.Settings) *ClusterServiceVersionBuilder {
	return newClusterServiceBuilder(apiClient, "clusterservice", "test-namespace")
}

func buildInvalidClusterServiceBuilder() *ClusterServiceVersionBuilder {
	return invalidClusterServiceBuilder("clusterservice", "test-namespace")
}

func buildTestClientWithDummyClusterServiceObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyClusterService(nil, ""),
		SchemeAttachers: testSchemes,
	})
}

func buildTestClientWithDummyClusterServiceObjectWitAlm(almExaple map[string]string) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyClusterServiceWithAlm(almExaple),
		SchemeAttachers: testSchemes,
	})
}

func buildTestClientWithDummyClusterServiceObjectWitPhase(
	phase oplmV1alpha1.ClusterServiceVersionPhase) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyClusterServiceWithPhase(phase),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyClusterServiceWithAlm(almExamples map[string]string) []runtime.Object {
	return buildDummyClusterService(almExamples, "")
}

func buildDummyClusterServiceWithPhase(phase oplmV1alpha1.ClusterServiceVersionPhase) []runtime.Object {
	return buildDummyClusterService(nil, phase)
}

func buildDummyClusterService(
	almExamples map[string]string, phase oplmV1alpha1.ClusterServiceVersionPhase) []runtime.Object {
	csv := oplmV1alpha1.ClusterServiceVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "clusterservice",
			Namespace: "test-namespace",
		},
		Spec: oplmV1alpha1.ClusterServiceVersionSpec{
			DisplayName: "test",
		},
		Status: oplmV1alpha1.ClusterServiceVersionStatus{
			Phase: oplmV1alpha1.CSVPhaseSucceeded,
		},
	}

	if len(almExamples) > 0 {
		csv.Annotations = almExamples
	}

	if phase != "" {
		csv.Status.Phase = phase
	}

	return append([]runtime.Object{}, &csv)
}

func newClusterServiceBuilder(apiClient *clients.Settings, name, namespace string) *ClusterServiceVersionBuilder {
	return &ClusterServiceVersionBuilder{
		apiClient: apiClient,
		Definition: &oplmV1alpha1.ClusterServiceVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: oplmV1alpha1.ClusterServiceVersionSpec{
				DisplayName: "test",
			}}}
}

func invalidClusterServiceBuilder(name, namespace string) *ClusterServiceVersionBuilder {
	return &ClusterServiceVersionBuilder{
		apiClient: nil,
		Definition: &oplmV1alpha1.ClusterServiceVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: oplmV1alpha1.ClusterServiceVersionSpec{
				DisplayName: "test",
			}}}
}
