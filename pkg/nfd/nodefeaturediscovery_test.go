package nfd

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	nfdv1 "github.com/openshift/cluster-nfd-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
)

var (
	almExampleName = "default-cluster-policy"
	almExample     = fmt.Sprintf(`[
 {
    "apiVersion": "nvidia.com/v1",
    "kind": "ClusterPolicy",
    "metadata": {
      "name": "%s",
	  "namespace": "test-namespace"
    },
    "spec": {
      "driver": {
        "enabled": true
      }
    }
  }
]`, almExampleName)
	nfdTestSchemes = []clients.SchemeAttacher{
		nfdv1.AddToScheme,
	}
)

func TestNFDNewBuilderFromObjectString(t *testing.T) {
	testCases := []struct {
		almString         string
		client            bool
		expectedErrorText string
	}{
		{
			almString:         almExample,
			client:            true,
			expectedErrorText: "",
		},
		{
			almString:         "",
			client:            true,
			expectedErrorText: "error initializing NodeFeatureDiscovery from alm-examples: almExample is an empty string",
		},
		{
			almString: "{invalid}",
			client:    true,
			//nolint:lll
			expectedErrorText: "error initializing NodeFeatureDiscovery from alm-examples: invalid character 'i' looking for beginning of object key string",
		},
		{
			almString:         almExample,
			client:            false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		var client *clients.Settings

		if testCase.client {
			client = buildTestClientWithNFDScheme()
		}

		policyBuilder := NewBuilderFromObjectString(client, testCase.almString)

		if testCase.client {
			assert.Equal(t, testCase.expectedErrorText, policyBuilder.errorMsg)

			if testCase.expectedErrorText == "" {
				assert.Equal(t, almExampleName, policyBuilder.Definition.Name)
			}
		} else {
			assert.Nil(t, policyBuilder)
		}
	}
}

func TestNFDPullPolicy(t *testing.T) {
	testCases := []struct {
		nfdName             string
		nfdNamespace        string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			nfdName:             "test",
			nfdNamespace:        "test-namespace",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			nfdName:             "test",
			nfdNamespace:        "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("nodeFeatureDiscovery 'namespace' cannot be empty"),
		},
		{
			nfdName:             "",
			nfdNamespace:        "test-namespace",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("nodeFeatureDiscovery 'name' cannot be empty"),
		},
		{
			nfdName:             "test",
			nfdNamespace:        "test-namespace",
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("nodeFeatureDiscovery object test does not exist in namespace test-namespace"),
		},
		{
			nfdName:             "test",
			nfdNamespace:        "test-namespace",
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("the apiClient of the Policy is nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testNFD := buildDummyNFD(testCase.nfdName, testCase.nfdNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testNFD)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: nfdTestSchemes,
			})
		}

		policyBuilder, err := Pull(testSettings, testNFD.Name, testNFD.Namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testNFD.Name, policyBuilder.Definition.Name)
			assert.Equal(t, testNFD.Namespace, policyBuilder.Definition.Namespace)
		}
	}
}

func TestNFDGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *Builder
		expectedError error
	}{
		{
			testBuilder:   buildValidNFDTestBuilder(buildTestClientWithDummyNFD()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidNFDTestBuilder(buildTestClientWithDummyNFD()),
			expectedError: fmt.Errorf("can not redefine the undefined NodeFeatureDiscovery"),
		},
		{
			testBuilder: buildValidNFDTestBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: nfdTestSchemes})),
			expectedError: fmt.Errorf("nodefeaturediscoveries.nfd.openshift.io \"default-cluster-policy\" not found"),
		},
	}

	for _, testCase := range testCases {
		nfd, err := testCase.testBuilder.Get()
		if testCase.expectedError == nil {
			assert.NotNil(t, nfd)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestNFDExist(t *testing.T) {
	testCases := []struct {
		nfd            *Builder
		expectedStatus bool
	}{
		{
			nfd:            buildValidNFDTestBuilder(buildTestClientWithDummyNFD()),
			expectedStatus: true,
		},
		{
			nfd:            buildInvalidNFDTestBuilder(buildTestClientWithDummyNFD()),
			expectedStatus: false,
		},
		{
			nfd: buildValidNFDTestBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: nfdTestSchemes})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.nfd.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestNFDDelete(t *testing.T) {
	testCases := []struct {
		nfd           *Builder
		expectedError error
	}{
		{
			nfd:           buildValidNFDTestBuilder(buildTestClientWithDummyNFD()),
			expectedError: nil,
		},
		{
			nfd: buildValidNFDTestBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: nfdTestSchemes})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.nfd.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.nfd.Object)
		}
	}
}

func TestNFDCreate(t *testing.T) {
	testCases := []struct {
		clusterPolicy *Builder
		expectedError error
	}{
		{
			clusterPolicy: buildValidNFDTestBuilder(buildTestClientWithDummyNFD()),
			expectedError: nil,
		},
		{
			clusterPolicy: buildValidNFDTestBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: nfdTestSchemes})),
			expectedError: nil,
		},
		{
			clusterPolicy: buildInvalidNFDTestBuilder(buildTestClientWithDummyNFD()),
			expectedError: fmt.Errorf("can not redefine the undefined NodeFeatureDiscovery"),
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

func TestNFDUpdate(t *testing.T) {
	testCases := []struct {
		nfd           *Builder
		expectedError error
		labelList     string
	}{
		{
			nfd:           buildValidNFDTestBuilder(buildTestClientWithDummyNFD()),
			expectedError: nil,
			labelList:     "test",
		},
	}

	for _, testCase := range testCases {
		assert.Empty(t, testCase.nfd.Definition.Spec.LabelWhiteList)
		assert.Nil(t, nil, testCase.nfd.Object)
		testCase.nfd.Definition.Spec.LabelWhiteList = testCase.labelList
		testCase.nfd.Definition.ObjectMeta.ResourceVersion = "999"
		_, err := testCase.nfd.Update(false)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.labelList, testCase.nfd.Definition.Spec.LabelWhiteList)
		}
	}
}

func buildTestClientWithDummyNFD() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyNFD(almExampleName, "test-namespace"),
		},
		SchemeAttachers: nfdTestSchemes,
	})
}

func buildValidNFDTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilderFromObjectString(apiClient, almExample)
}

func buildInvalidNFDTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilderFromObjectString(apiClient, "{invalid}")
}

func buildDummyNFD(name, nsname string) *nfdv1.NodeFeatureDiscovery {
	return &nfdv1.NodeFeatureDiscovery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
		Spec: nfdv1.NodeFeatureDiscoverySpec{},
	}
}

// buildTestClientWithPolicyScheme returns a client with no objects but the Policy scheme attached.
func buildTestClientWithNFDScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: nfdTestSchemes,
	})
}
