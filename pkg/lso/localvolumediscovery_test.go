package lso

import (
	"fmt"
	"testing"
	"time"

	lsov1alpha1 "github.com/openshift/local-storage-operator/api/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultLocalVolumeDiscoveryName      = "auto-discover-devices"
	defaultLocalVolumeDiscoveryNamespace = "test-lvdspace"
	v1alpha1testSchemes                  = []clients.SchemeAttacher{
		lsov1alpha1.AddToScheme,
	}
)

func TestPullLocalVolumeDiscovery(t *testing.T) {
	generateLocalVolumeDiscovery := func(name, namespace string) *lsov1alpha1.LocalVolumeDiscovery {
		return &lsov1alpha1.LocalVolumeDiscovery{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
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
			name:                defaultLocalVolumeDiscoveryName,
			namespace:           defaultLocalVolumeDiscoveryNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           defaultLocalVolumeDiscoveryNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("localVolumeDiscovery 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                defaultLocalVolumeDiscoveryName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("localVolumeDiscovery 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "lvdtest",
			namespace:           defaultLocalVolumeDiscoveryNamespace,
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf("localVolumeDiscovery object lvdtest does not exist " +
				"in namespace test-lvdspace"),
			client: true,
		},
		{
			name:                "lvdtest",
			namespace:           defaultLocalVolumeDiscoveryNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("localVolumeDiscovery 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testLocalVolumeDiscovery := generateLocalVolumeDiscovery(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testLocalVolumeDiscovery)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: v1alpha1testSchemes,
			})
		}

		builderResult, err := PullLocalVolumeDiscovery(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testLocalVolumeDiscovery.Name, builderResult.Object.Name)
			assert.Equal(t, testLocalVolumeDiscovery.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestNewLocalVolumeDiscoveryBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
		client        bool
	}{
		{
			name:          defaultLocalVolumeDiscoveryName,
			namespace:     defaultLocalVolumeDiscoveryNamespace,
			expectedError: "",
			client:        true,
		},
		{
			name:          "",
			namespace:     defaultLocalVolumeDiscoveryNamespace,
			expectedError: "localVolumeDiscovery 'name' cannot be empty",
			client:        true,
		},
		{
			name:          defaultLocalVolumeDiscoveryName,
			namespace:     "",
			expectedError: "localVolumeDiscovery 'nsname' cannot be empty",
			client:        true,
		},
		{
			name:          defaultLocalVolumeDiscoveryName,
			namespace:     defaultLocalVolumeDiscoveryNamespace,
			expectedError: "",
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings
		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testLocalVolumeDiscovery := NewLocalVolumeDiscoveryBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedError == "" {
			if testCase.client {
				assert.Equal(t, testCase.name, testLocalVolumeDiscovery.Definition.Name)
				assert.Equal(t, testCase.namespace, testLocalVolumeDiscovery.Definition.Namespace)
			} else {
				assert.Nil(t, testLocalVolumeDiscovery)
			}
		} else {
			assert.Equal(t, testCase.expectedError, testLocalVolumeDiscovery.errorMsg)
			assert.NotNil(t, testLocalVolumeDiscovery.Definition)
		}
	}
}

func TestLocalVolumeDiscoveryExists(t *testing.T) {
	testCases := []struct {
		testLocalVolumeDiscovery *LocalVolumeDiscoveryBuilder
		expectedStatus           bool
	}{
		{
			testLocalVolumeDiscovery: buildValidLVDObjectBuilder(buildLVDClientWithDummyObject()),
			expectedStatus:           true,
		},
		{
			testLocalVolumeDiscovery: buildInValidLVDObjectBuilder(buildLVDClientWithDummyObject()),
			expectedStatus:           false,
		},
		{
			testLocalVolumeDiscovery: buildValidLVDObjectBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:           false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testLocalVolumeDiscovery.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestLocalVolumeDiscoveryGet(t *testing.T) {
	testCases := []struct {
		testLocalVolumeDiscovery *LocalVolumeDiscoveryBuilder
		expectedError            error
	}{
		{
			testLocalVolumeDiscovery: buildValidLVDObjectBuilder(buildLVDClientWithDummyObject()),
			expectedError:            nil,
		},
		{
			testLocalVolumeDiscovery: buildInValidLVDObjectBuilder(buildLVDClientWithDummyObject()),
			expectedError:            fmt.Errorf("localVolumeDiscovery 'name' cannot be empty"),
		},
		{
			testLocalVolumeDiscovery: buildValidLVDObjectBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("localvolumediscoveries.local.storage.openshift.io " +
				"\"auto-discover-devices\" not found"),
		},
	}

	for _, testCase := range testCases {
		localVolumeDiscoveryObj, err := testCase.testLocalVolumeDiscovery.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, localVolumeDiscoveryObj.Name, testCase.testLocalVolumeDiscovery.Definition.Name)
			assert.Equal(t, localVolumeDiscoveryObj.Namespace, testCase.testLocalVolumeDiscovery.Definition.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestLocalVolumeDiscoveryCreate(t *testing.T) {
	testCases := []struct {
		testLocalVolumeDiscovery *LocalVolumeDiscoveryBuilder
		expectedError            error
	}{
		{
			testLocalVolumeDiscovery: buildValidLVDObjectBuilder(buildLVDClientWithDummyObject()),
			expectedError:            nil,
		},
		{
			testLocalVolumeDiscovery: buildInValidLVDObjectBuilder(buildLVDClientWithDummyObject()),
			expectedError:            fmt.Errorf("localVolumeDiscovery 'name' cannot be empty"),
		},
		{
			testLocalVolumeDiscovery: buildValidLVDObjectBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:            nil,
		},
	}

	for _, testCase := range testCases {
		testLocalVolumeSetBuilder, err := testCase.testLocalVolumeDiscovery.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testLocalVolumeSetBuilder.Definition.Name, testLocalVolumeSetBuilder.Object.Name)
			assert.Equal(t, testLocalVolumeSetBuilder.Definition.Namespace, testLocalVolumeSetBuilder.Object.Namespace)
			assert.Nil(t, err)
		}
	}
}

func TestLocalVolumeDiscoveryDelete(t *testing.T) {
	testCases := []struct {
		testLocalVolumeDiscovery *LocalVolumeDiscoveryBuilder
		expectedError            error
	}{
		{
			testLocalVolumeDiscovery: buildValidLVDObjectBuilder(buildLVDClientWithDummyObject()),
			expectedError:            nil,
		},
		{
			testLocalVolumeDiscovery: buildValidLVDObjectBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:            nil,
		},
		{
			testLocalVolumeDiscovery: buildInValidLVDObjectBuilder(buildLVDClientWithDummyObject()),
			expectedError:            fmt.Errorf("localVolumeDiscovery 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testLocalVolumeDiscovery.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testLocalVolumeDiscovery.Object)
		}
	}
}

func TestLocalVolumeDiscoveryWithNodeSelector(t *testing.T) {
	testCases := []struct {
		testNodeSelector corev1.NodeSelector
		expectedError    error
	}{
		{
			testNodeSelector: corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "cluster.ocs.openshift.io/openshift-storage",
						Operator: "In",
						Values:   []string{""},
					}}},
				},
			},
			expectedError: nil,
		},
		{
			testNodeSelector: corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "cluster.ocs.openshift.io/openshift-storage",
						Operator: "Exists",
					}}},
				},
			},
			expectedError: nil,
		},
		{
			testNodeSelector: corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "cluster.ocs.openshift.io/openshift-storage",
						Operator: "Exists",
					}, {
						Key:      "machineconfiguration.openshift.io/role",
						Operator: "In",
						Values:   []string{"customcnf", "worker"},
					}}},
				},
			},
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLVDObjectBuilder(buildLVDClientWithDummyObject())

		result := testBuilder.WithNodeSelector(testCase.testNodeSelector)

		if testCase.expectedError == nil {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testNodeSelector, *result.Definition.Spec.NodeSelector)
			assert.Equal(t, "", result.errorMsg)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), result.errorMsg)
		}
	}
}

func TestLocalVolumeDiscoveryWithTolerations(t *testing.T) {
	testCases := []struct {
		testTolerations   []corev1.Toleration
		expectedErrorText string
	}{
		{
			testTolerations: []corev1.Toleration{{
				Key:      "node.ocs.openshift.io/storage",
				Operator: "Equal",
				Value:    "true",
				Effect:   "NoSchedule",
			}},
			expectedErrorText: "",
		},
		{
			testTolerations:   []corev1.Toleration{},
			expectedErrorText: "'tolerations' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLVDObjectBuilder(buildLVDClientWithDummyObject())

		result := testBuilder.WithTolerations(testCase.testTolerations)
		assert.Equal(t, testCase.expectedErrorText, result.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testTolerations, result.Definition.Spec.Tolerations)
		}
	}
}

func TestLocalVolumeSetIsDiscovering(t *testing.T) {
	testCases := []struct {
		testLVDBuilder *LocalVolumeDiscoveryBuilder
		testPhase      bool
	}{
		{
			testLVDBuilder: buildValidLVDObjectBuilder(buildLVDClientWithDummyObject(lsov1alpha1.Discovering)),
			testPhase:      true,
		},
		{
			testLVDBuilder: buildValidLVDObjectBuilder(buildLVDClientWithDummyObject(lsov1alpha1.DiscoveryFailed)),
			testPhase:      false,
		},
	}

	for _, testCase := range testCases {
		isDiscoveringResult := testCase.testLVDBuilder.IsDiscovering(2 * time.Second)

		assert.Equal(t, testCase.testPhase, isDiscoveringResult)
	}
}

func buildValidLVDObjectBuilder(apiClient *clients.Settings) *LocalVolumeDiscoveryBuilder {
	lvdBuilder := NewLocalVolumeDiscoveryBuilder(
		apiClient, defaultLocalVolumeDiscoveryName, defaultLocalVolumeDiscoveryNamespace)

	return lvdBuilder
}

func buildInValidLVDObjectBuilder(apiClient *clients.Settings) *LocalVolumeDiscoveryBuilder {
	lvdBuilder := NewLocalVolumeDiscoveryBuilder(
		apiClient, "", defaultLocalVolumeDiscoveryNamespace)

	return lvdBuilder
}

func buildLVDClientWithDummyObject(phase ...lsov1alpha1.DiscoveryPhase) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyLocalVolumeDiscovery(phase...),
		SchemeAttachers: v1alpha1testSchemes,
	})
}

func buildDummyLocalVolumeDiscovery(phase ...lsov1alpha1.DiscoveryPhase) []runtime.Object {
	discoveryPhase := lsov1alpha1.Discovering

	if len(phase) > 0 {
		discoveryPhase = phase[0]
	}

	return append([]runtime.Object{}, &lsov1alpha1.LocalVolumeDiscovery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultLocalVolumeDiscoveryName,
			Namespace: defaultLocalVolumeDiscoveryNamespace,
		},
		Status: lsov1alpha1.LocalVolumeDiscoveryStatus{
			Phase: discoveryPhase,
		},
	})
}
