package cgu

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/cluster-group-upgrades-operator/pkg/api/clustergroupupgrades/v1alpha1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultCguName           = "cgu-test"
	defaultCguNsName         = "test-ns"
	defaultCguMaxConcurrency = 1
)

//nolint:funlen
func TestPullCgu(t *testing.T) {
	generateCgu := func(name, namespace string) *v1alpha1.ClusterGroupUpgrade {
		return &v1alpha1.ClusterGroupUpgrade{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: v1alpha1.ClusterGroupUpgradeSpec{},
		}
	}

	testCases := []struct {
		cguName             string
		cguNamespace        string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
		client              bool
	}{
		{
			cguName:             "test1",
			cguNamespace:        "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			cguName:             "test2",
			cguNamespace:        "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "cgu object test2 doesn't exist in namespace test-namespace",
			client:              true,
		},
		{
			cguName:             "",
			cguNamespace:        "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "cgu 'name' cannot be empty",
			client:              true,
		},
		{
			cguName:             "test3",
			cguNamespace:        "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "cgu 'namespace' cannot be empty",
			client:              true,
		},
		{
			cguName:             "test3",
			cguNamespace:        "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "cgu 'apiClient' cannot be empty",
			client:              false,
		},
	}
	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testCgu := generateCgu(testCase.cguName, testCase.cguNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testCgu)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		// Test the Pull method
		builderResult, err := Pull(testSettings, testCgu.Name, testCgu.Namespace)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			// Check the error message
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCgu.Name, builderResult.Object.Name)
			assert.Equal(t, testCgu.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestNewCguBuilder(t *testing.T) {
	generateCguBuilder := NewCguBuilder

	testCases := []struct {
		cguName           string
		cguNamespace      string
		cguMaxConcurrency int
		expectedErrorText string
		client            bool
	}{
		{
			cguName:           "test1",
			cguNamespace:      "test-namespace",
			cguMaxConcurrency: 1,
			expectedErrorText: "",
		},
		{
			cguName:           "",
			cguNamespace:      "test-namespace",
			cguMaxConcurrency: 1,
			expectedErrorText: "CGU 'name' cannot be empty",
		},
		{
			cguName:           "test1",
			cguNamespace:      "",
			cguMaxConcurrency: 1,
			expectedErrorText: "CGU 'nsname' cannot be empty",
		},
		{
			cguName:           "test1",
			cguNamespace:      "test-namespace",
			cguMaxConcurrency: 0,
			expectedErrorText: "CGU 'maxConcurrency' cannot be less than 1",
		},
	}
	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testCguStructure := generateCguBuilder(
			testSettings,
			testCase.cguName,
			testCase.cguNamespace,
			testCase.cguMaxConcurrency)
		assert.NotNil(t, testCguStructure)
		assert.Equal(t, testCguStructure.errorMsg, testCase.expectedErrorText)
	}
}

func TestCguWithCluster(t *testing.T) {
	testCases := []struct {
		cluster           string
		expectedErrorText string
	}{
		{
			cluster:           "test-cluster",
			expectedErrorText: "",
		},
		{
			cluster:           "",
			expectedErrorText: "cluster in CGU cluster spec cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyCguObject()
		cguBuilder := buildValidCguTestBuilder(testSettings).WithCluster(testCase.cluster)
		assert.Equal(t, cguBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, cguBuilder.Definition.Spec.Clusters, []string{testCase.cluster})
		}
	}
}

func TestCguWithManagedPolicy(t *testing.T) {
	testCases := []struct {
		policy            string
		expectedErrorText string
	}{
		{
			policy:            "test-policy",
			expectedErrorText: "",
		},
		{
			policy:            "",
			expectedErrorText: "policy in CGU managedpolicies spec cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyCguObject()
		cguBuilder := buildValidCguTestBuilder(testSettings).WithManagedPolicy(testCase.policy)
		assert.Equal(t, cguBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, cguBuilder.Definition.Spec.ManagedPolicies, []string{testCase.policy})
		}
	}
}

func TestCguWithCanary(t *testing.T) {
	testCases := []struct {
		canary            string
		expectedErrorText string
	}{
		{
			canary:            "test-canary",
			expectedErrorText: "",
		},
		{
			canary:            "",
			expectedErrorText: "canary in CGU remediationstrategy spec cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyCguObject()
		cguBuilder := buildValidCguTestBuilder(testSettings).WithCanary(testCase.canary)
		assert.Equal(t, cguBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, cguBuilder.Definition.Spec.RemediationStrategy.Canaries, []string{testCase.canary})
		}
	}
}

func TestCguCreate(t *testing.T) {
	testCases := []struct {
		testCgu       *CguBuilder
		expectedError error
	}{
		{
			testCgu:       buildValidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedError: nil,
		},
		{
			testCgu:       buildInvalidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedError: fmt.Errorf("CGU 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		cguBuilder, err := testCase.testCgu.Create()
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.Equal(t, cguBuilder.Definition, cguBuilder.Object)
		}
	}
}

func TestCguDelete(t *testing.T) {
	testCases := []struct {
		testCgu       *CguBuilder
		expectedError error
	}{
		{
			testCgu:       buildValidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedError: nil,
		},
		{
			testCgu:       buildInvalidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedError: fmt.Errorf("CGU 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testCgu.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testCgu.Object)
		}
	}
}

func TestCguExist(t *testing.T) {
	testCases := []struct {
		testCgu        *CguBuilder
		expectedStatus bool
	}{
		{
			testCgu:        buildValidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedStatus: true,
		},
		{
			testCgu:        buildInvalidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testCgu.Exists()
		assert.Equal(t, testCase.expectedStatus, exists)
	}
}

func buildTestClientWithDummyCguObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyCguObject(),
	})
}

func buildDummyCguObject() []runtime.Object {
	return append([]runtime.Object{}, buildDummyCgu(defaultCguName, defaultCguNsName, defaultCguMaxConcurrency))
}

func buildDummyCgu(name, namespace string, maxConcurrency int) *v1alpha1.ClusterGroupUpgrade {
	return &v1alpha1.ClusterGroupUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ClusterGroupUpgradeSpec{
			RemediationStrategy: &v1alpha1.RemediationStrategySpec{
				MaxConcurrency: maxConcurrency,
			},
		},
	}
}

// buildValidCguTestBuilder returns a valid CguBuilder for testing purposes.
func buildValidCguTestBuilder(apiClient *clients.Settings) *CguBuilder {
	return NewCguBuilder(
		apiClient,
		defaultCguName,
		defaultCguNsName,
		defaultCguMaxConcurrency)
}

// buildinInvalidCguTestBuilder returns an invalid CguBuilder for testing purposes.
func buildInvalidCguTestBuilder(apiClient *clients.Settings) *CguBuilder {
	return NewCguBuilder(
		apiClient,
		defaultCguName,
		"",
		defaultCguMaxConcurrency)
}
