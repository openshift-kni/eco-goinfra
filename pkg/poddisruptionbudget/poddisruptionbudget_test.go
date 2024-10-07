package poddisruptionbudget

import (
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

func TestPDBNewBuilder(t *testing.T) {
	testCases := []struct {
		apiClientNil    bool
		testName        string
		testNamespace   string
		expectedBuilder Builder
		expectedError   string
	}{
		{
			apiClientNil:  false,
			testName:      "testPDB",
			testNamespace: "testNamespace",
			expectedBuilder: Builder{
				Definition: &policyv1.PodDisruptionBudget{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testPDB",
						Namespace: "testNamespace",
					},
				},
			},
		},
		{
			apiClientNil:  true,
			testName:      "testPDB",
			testNamespace: "testNamespace",
			expectedError: "",
		},
		{
			apiClientNil:  false,
			testName:      "",
			testNamespace: "testNamespace",
			expectedError: "pod disruption budget 'name' cannot be empty",
		},
		{
			apiClientNil:  false,
			testName:      "testPDB",
			testNamespace: "",
			expectedError: "pod disruption budget 'namespace' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: nil,
		})

		if testCase.apiClientNil {
			testSettings = nil
		}

		builderResult := NewBuilder(testSettings, testCase.testName, testCase.testNamespace)

		if testCase.expectedError != "" {
			assert.NotNil(t, builderResult)
		} else {
			if testCase.apiClientNil {
				assert.Nil(t, builderResult)
			} else {
				assert.NotNil(t, builderResult)
				assert.Equal(t, testCase.expectedBuilder.Definition.Name, builderResult.Definition.Name)
				assert.Equal(t, testCase.expectedBuilder.Definition.Namespace, builderResult.Definition.Namespace)
				assert.Equal(t, testCase.expectedError, builderResult.errorMsg)
			}
		}
	}
}

func TestPDBPull(t *testing.T) {
	testCases := []struct {
		testName            string
		testNamespace       string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
		apiClientNil        bool
	}{
		{
			testName:            "test1",
			testNamespace:       "testNamespace",
			expectedError:       false,
			addToRuntimeObjects: true,
			apiClientNil:        false,
		},
		{
			testName:            "test2",
			testNamespace:       "testNamespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "PodDisruptionBudget object test2 does not exist in namespace testNamespace",
			apiClientNil:        false,
		},
		{
			testName:            "",
			testNamespace:       "testNamespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "PodDisruptionBudget 'name' cannot be empty",
			apiClientNil:        false,
		},
		{
			testName:            "test3",
			testNamespace:       "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "PodDisruptionBudget 'namespace' cannot be empty",
			apiClientNil:        false,
		},
		{
			testName:            "test4",
			testNamespace:       "testNamespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "apiClient is nil",
			apiClientNil:        true,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testPDB := generatePDB(testCase.testName, testCase.testNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPDB)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		if testCase.apiClientNil {
			testSettings = nil
		}

		builderResult, err := Pull(testSettings, testCase.testName, testCase.testNamespace)

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testPDB.Name, builderResult.Object.Name)
			assert.Equal(t, testPDB.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestPDBWithPDBSpec(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		expectedError bool
		errorMessage  string
	}{
		{ // Test Case 1 - Happy path, no nils
			builderNil:    false,
			definitionNil: false,
			expectedError: false,
		},
		{ // Test Case 2 - Builder is nil
			builderNil:    true,
			definitionNil: false,
			expectedError: true,
		},
		{ // Test Case 3 - Definition is nil
			builderNil:    false,
			definitionNil: true,
			expectedError: true,
		},
	}

	for _, testCase := range testCases {
		testBuilder := &Builder{
			apiClient: clients.GetTestClients(clients.TestClientParams{}).PolicyV1Interface,
			Definition: &policyv1.PodDisruptionBudget{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testPDB",
					Namespace: "testNamespace",
				},
			},
		}

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		oneInt := intstr.FromInt(1)
		result := testBuilder.WithPDBSpec(policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &oneInt,
		})

		if !testCase.expectedError {
			assert.NotNil(t, result)
			assert.NotNil(t, testBuilder.Definition)
			assert.Equal(t, oneInt, *testBuilder.Definition.Spec.MinAvailable)
		} else {
			if !testCase.builderNil {
				assert.NotNil(t, result)
			} else {
				assert.Nil(t, result)
			}
		}
	}
}

func TestPDBCreate(t *testing.T) {
	testCases := []struct {
		pdbExistsAlready bool
	}{
		{
			pdbExistsAlready: false,
		},
		{
			pdbExistsAlready: true,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		testPDB := generatePDB("testPDB", "testNamespace")

		if testCase.pdbExistsAlready {
			runtimeObjects = append(runtimeObjects, testPDB)
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects, testPDB.Name, testPDB.Namespace)
		builderResult, err := testBuilder.Create()

		assert.Nil(t, err)
		assert.NotNil(t, testBuilder.Object)
		assert.Equal(t, testPDB.Name, builderResult.Object.Name)
		assert.Equal(t, testPDB.Namespace, builderResult.Object.Namespace)
	}
}

func TestPDBDelete(t *testing.T) {
	testCases := []struct {
		pdbExistsAlready bool
	}{
		{
			pdbExistsAlready: false,
		},
		{
			pdbExistsAlready: true,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		testPDB := generatePDB("testPDB", "testNamespace")

		if testCase.pdbExistsAlready {
			runtimeObjects = append(runtimeObjects, testPDB)
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects, testPDB.Name, testPDB.Namespace)
		err := testBuilder.Delete()

		assert.Nil(t, err)
	}
}

func TestPDBExists(t *testing.T) {
	testCases := []struct {
		pdbExistsAlready bool
	}{
		{
			pdbExistsAlready: true,
		},
		{
			pdbExistsAlready: false,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		testPDB := generatePDB("testPDB", "testNamespace")

		if testCase.pdbExistsAlready {
			runtimeObjects = append(runtimeObjects, testPDB)
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects, testPDB.Name, testPDB.Namespace)
		result := testBuilder.Exists()

		assert.Equal(t, testCase.pdbExistsAlready, result)
	}
}

func TestPDBUpdate(t *testing.T) {
	testCases := []struct {
		pdbExistsAlready bool
		force            bool
	}{
		{
			pdbExistsAlready: true,
			force:            false,
		},
		{
			pdbExistsAlready: false,
			force:            false,
		},
		{
			pdbExistsAlready: true,
			force:            true,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		testPDB := generatePDB("testPDB", "testNamespace")

		if testCase.pdbExistsAlready {
			runtimeObjects = append(runtimeObjects, testPDB)
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects, testPDB.Name, testPDB.Namespace)
		builderResult, err := testBuilder.Update(testCase.force)

		if testCase.pdbExistsAlready {
			assert.Nil(t, err)
			assert.NotNil(t, testBuilder.Object)
			assert.Equal(t, testPDB.Name, builderResult.Object.Name)
			assert.Equal(t, testPDB.Namespace, builderResult.Object.Namespace)
		} else {
			assert.NotNil(t, err)
		}
	}
}

func buildTestBuilderWithFakeObjects(runtimeObjects []runtime.Object, name, namespace string) *Builder {
	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: runtimeObjects,
	})

	testBuilder := NewBuilder(testSettings, name, namespace)

	return testBuilder
}

func generatePDB(name, namespace string) *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}
