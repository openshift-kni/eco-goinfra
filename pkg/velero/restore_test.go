package velero

import (
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

func TestNewRestoreBuilder(t *testing.T) {
	testcases := []struct {
		name           string
		namespace      string
		backupName     string
		expectedErrMsg string
	}{
		{
			name:           "restore-test-name-1",
			namespace:      "restore-test-namespace-1",
			backupName:     "backup-test-1",
			expectedErrMsg: "",
		},
		{
			name:           "",
			namespace:      "restore-test-namespace-2",
			backupName:     "backup-test-2",
			expectedErrMsg: "restore name cannot be an empty string",
		},
		{
			name:           "restore-test-name-3",
			namespace:      "",
			backupName:     "backup-test-3",
			expectedErrMsg: "restore namespace cannot be an empty string",
		},
		{
			name:           "restore-test-name-4",
			namespace:      "restore-test-namespace-4",
			backupName:     "",
			expectedErrMsg: "restore backupName cannot be an empty string",
		},
	}

	for _, test := range testcases {
		testBuilder := NewRestoreBuilder(
			clients.GetTestClients(clients.TestClientParams{}), test.name, test.namespace, test.backupName)
		assert.Equal(t, test.expectedErrMsg, testBuilder.errorMsg)
	}
}

func TestPullRestore(t *testing.T) {
	generateRestore := func(name, namespace string) *velerov1.Restore {
		return &velerov1.Restore{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		name                string
		namespace           string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			name:                "restore-test-1",
			namespace:           "restore-test-namespace-1",
			expectedError:       false,
			addToRuntimeObjects: true,
		},
		{
			name:                "restore-test-2",
			namespace:           "restore-test-namespace-2",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "restore object restore-test-2 does not exist in namespace restore-test-namespace-2",
		},
		{
			name:                "",
			namespace:           "restore-test-namespace-3",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "restore name cannot be empty",
		},
		{
			name:                "restore-test-4",
			namespace:           "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "restore namespace cannot be empty",
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testRestore := generateRestore(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testRestore)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: v1TestSchemes,
		})

		// Test the Pull method
		builderResult, err := PullRestore(testSettings, testCase.name, testCase.namespace)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			// Check the error message
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

// buildValidRestoreTestBuilder returns a valid RestoreBuilder for testing purposes.
func buildValidRestoreTestBuilder() *RestoreBuilder {
	return NewRestoreBuilder(clients.GetTestClients(clients.TestClientParams{}),
		"restore-test-name", "restore-test-namespace", "backupName-test-name")
}

func TestRestoreWithStorageLocation(t *testing.T) {
	testCases := []struct {
		location         string
		expectedErrorMsg string
	}{
		{
			location:         "default",
			expectedErrorMsg: "",
		},
		{
			location:         "",
			expectedErrorMsg: "restore storage location cannot be an empty string",
		},
	}

	for _, test := range testCases {
		testBuilder := buildValidRestoreTestBuilder()

		testBuilder.WithStorageLocation(test.location)

		assert.Equal(t, test.expectedErrorMsg, testBuilder.errorMsg)
	}
}
