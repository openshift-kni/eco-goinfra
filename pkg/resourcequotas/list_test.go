package resourcequotas

import (
	"strconv"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestResourceQuotaList(t *testing.T) {
	generateResourceQuota := func(name, namespace string) *corev1.ResourceQuota {
		return &corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		numResourceQuotasToGenerate int
		testNamespace               string
		expectedError               bool
		expectedErrorText           string
		numOptionsToPass            int
		expectedLength              int
	}{
		{ // Test Case 1 - Happy path, one resource quota, one option
			numResourceQuotasToGenerate: 1,
			testNamespace:               "testNamespace",
			expectedError:               false,
			numOptionsToPass:            1,
			expectedLength:              1,
		},
		{ // Test Case 2 - Happy path, two resource quotas, no options
			numResourceQuotasToGenerate: 2,
			testNamespace:               "testNamespace",
			expectedError:               false,
			numOptionsToPass:            0,
			expectedLength:              2,
		},
		{ // Test Case 3 - No resource quotas, one option, empty list
			numResourceQuotasToGenerate: 0,
			testNamespace:               "testNamespace",
			expectedError:               false,
			numOptionsToPass:            1,
			expectedLength:              0,
		},
		{ // Test Case 4 - Two options, invalid
			numResourceQuotasToGenerate: 0,
			testNamespace:               "testNamespace",
			expectedError:               true,
			expectedErrorText:           "error: more than one ListOptions was passed",
			numOptionsToPass:            2,
			expectedLength:              0,
		},
		{ // Test Case 5 - Empty namespace
			numResourceQuotasToGenerate: 0,
			testNamespace:               "",
			expectedError:               true,
			expectedErrorText:           "failed to list resource quotas, 'nsname' parameter is empty",
			numOptionsToPass:            0,
			expectedLength:              0,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		// Generate resource quotas
		for i := 0; i < testCase.numResourceQuotasToGenerate; i++ {
			rqNameNum := strconv.Itoa(i)
			runtimeObjects = append(runtimeObjects, generateResourceQuota("testResourceQuota"+rqNameNum, testCase.testNamespace))
		}

		// Generate fake client
		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		var testOptions []metav1.ListOptions

		// Generate options
		for i := 0; i < testCase.numOptionsToPass; i++ {
			testOptions = append(testOptions, metav1.ListOptions{
				Limit: 10, // arbitrary value
			})
		}

		resourceQuotaList, err := List(testSettings, testCase.testNamespace, testOptions...)

		if testCase.expectedError {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), testCase.expectedErrorText)
		} else {
			assert.NoError(t, err)
			assert.Len(t, resourceQuotaList, testCase.expectedLength)
		}
	}
}
