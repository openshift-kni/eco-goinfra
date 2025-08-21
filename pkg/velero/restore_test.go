package velero

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
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
		client         bool
		expectedErrMsg string
	}{
		{
			name:           "restore-test-name-1",
			namespace:      "restore-test-namespace-1",
			backupName:     "backup-test-1",
			client:         true,
			expectedErrMsg: "",
		},
		{
			name:           "",
			namespace:      "restore-test-namespace-2",
			backupName:     "backup-test-2",
			client:         true,
			expectedErrMsg: "restore name cannot be an empty string",
		},
		{
			name:           "restore-test-name-3",
			namespace:      "",
			backupName:     "backup-test-3",
			client:         true,
			expectedErrMsg: "restore namespace cannot be an empty string",
		},
		{
			name:           "restore-test-name-4",
			namespace:      "restore-test-namespace-4",
			backupName:     "",
			client:         true,
			expectedErrMsg: "restore backupName cannot be an empty string",
		},
		{
			name:           "restore-test-name-5",
			namespace:      "restore-test-namespace-5",
			backupName:     "backup-test-5",
			client:         false,
			expectedErrMsg: "",
		},
	}

	for _, test := range testcases {
		var testSettings *clients.Settings

		if test.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewRestoreBuilder(testSettings, test.name, test.namespace, test.backupName)

		if test.client {
			assert.Equal(t, test.expectedErrMsg, testBuilder.errorMsg)

			if test.expectedErrMsg == "" {
				assert.Equal(t, test.name, testBuilder.Definition.Name)
				assert.Equal(t, test.namespace, testBuilder.Definition.Namespace)
				assert.Equal(t, test.backupName, testBuilder.Definition.Spec.BackupName)
			}
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestPullRestore(t *testing.T) {
	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                "restore-test-1",
			namespace:           "restore-test-namespace-1",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "restore-test-2",
			namespace:           "restore-test-namespace-2",
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf(
				"restore object restore-test-2 does not exist in namespace restore-test-namespace-2"),
		},
		{
			name:                "",
			namespace:           "restore-test-namespace-3",
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("restore name cannot be empty"),
		},
		{
			name:                "restore-test-4",
			namespace:           "",
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("restore namespace cannot be empty"),
		},
		{
			name:                "restore-test-5",
			namespace:           "restore-test-namespace-5",
			addToRuntimeObjects: false,
			client:              false,
			expectedError:       fmt.Errorf("the apiClient cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, buildDummyRestore(testCase.name, testCase.namespace))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: v1TestSchemes,
			})
		}

		// Test the Pull method
		builderResult, err := PullRestore(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
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
		testBuilder := buildValidRestoreTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

		testBuilder.WithStorageLocation(test.location)

		assert.Equal(t, test.expectedErrorMsg, testBuilder.errorMsg)
	}
}

func TestRestoreGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *RestoreBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidRestoreTestBuilder(buildTestClientWithDummyRestore()),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidRestoreTestBuilder(buildTestClientWithDummyRestore()),
			expectedError: "restore backupName cannot be an empty string",
		},
		{
			testBuilder:   buildValidRestoreTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "restores.velero.io \"restore-test-name\" not found",
		},
	}

	for _, testCase := range testCases {
		restore, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, restore.Name)
			assert.Equal(t, testCase.testBuilder.Definition.Namespace, restore.Namespace)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestRestoreExists(t *testing.T) {
	testCases := []struct {
		testBuilder *RestoreBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidRestoreTestBuilder(buildTestClientWithDummyRestore()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidRestoreTestBuilder(buildTestClientWithDummyRestore()),
			exists:      false,
		},
		{
			testBuilder: buildValidRestoreTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestRestoreCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *RestoreBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidRestoreTestBuilder(buildTestClientWithDummyRestore()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidRestoreTestBuilder(buildTestClientWithDummyRestore()),
			expectedError: fmt.Errorf("restore backupName cannot be an empty string"),
		},
		{
			testBuilder:   buildValidRestoreTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.testBuilder.Definition.Name, testBuilder.Object.Name)
			assert.Equal(t, testCase.testBuilder.Definition.Namespace, testBuilder.Object.Namespace)
		}
	}
}

func TestRestoreUpdate(t *testing.T) {
	testcases := []struct {
		testBuilder   *RestoreBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidRestoreTestBuilder(buildTestClientWithDummyRestore()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidRestoreTestBuilder(buildTestClientWithDummyRestore()),
			expectedError: fmt.Errorf("restore backupName cannot be an empty string"),
		},
		{
			testBuilder:   buildValidRestoreTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("cannot update non-existent restore"),
		},
	}

	for _, testCase := range testcases {
		assert.Empty(t, testCase.testBuilder.Definition.Spec.IncludedNamespaces)

		testCase.testBuilder.Definition.Spec.IncludedNamespaces = []string{"test-namespace"}

		testBuilder, err := testCase.testBuilder.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, []string{"test-namespace"}, testBuilder.Object.Spec.IncludedNamespaces)
		}
	}
}

func TestRestoreDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *RestoreBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidRestoreTestBuilder(buildTestClientWithDummyRestore()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidRestoreTestBuilder(buildTestClientWithDummyRestore()),
			expectedError: fmt.Errorf("restore backupName cannot be an empty string"),
		},
		{
			testBuilder:   buildValidRestoreTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testBuilder.Object)
		}
	}
}

// buildDummyRestore returns a dummy Restore object with the given name and namespace.
func buildDummyRestore(name, nsname string) *velerov1.Restore {
	return &velerov1.Restore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}

// buildTestClientWithDummyRestore returns a test client with a dummy Restore object using the name restore-test-name
// and namespace restore-test-namespace.
func buildTestClientWithDummyRestore() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyRestore("restore-test-name", "restore-test-namespace"),
		},
		SchemeAttachers: v1TestSchemes,
	})
}

// buildValidRestoreTestBuilder returns a valid RestoreBuilder for testing purposes.
func buildValidRestoreTestBuilder(apiClient *clients.Settings) *RestoreBuilder {
	return NewRestoreBuilder(apiClient, "restore-test-name", "restore-test-namespace", "backupName-test-name")
}

// buildInvalidRestoreTestBuilder returns an invalid RestoreBuilder for testing purposes.
func buildInvalidRestoreTestBuilder(apiClient *clients.Settings) *RestoreBuilder {
	return NewRestoreBuilder(apiClient, "restore-test-name", "restore-test-namespace", "")
}
