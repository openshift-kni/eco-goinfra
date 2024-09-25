package velero

import (
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

var (
	v1TestSchemes = []clients.SchemeAttacher{
		velerov1.AddToScheme,
	}
)

func TestNewBackupBuilder(t *testing.T) {
	testcases := []struct {
		name           string
		namespace      string
		targetService  string
		expectedErrMsg string
	}{
		{
			name:           "backup-test-name-1",
			namespace:      "backup-test-namespace-1",
			expectedErrMsg: "",
		},
		{
			name:           "",
			namespace:      "backup-test-namespace-2",
			expectedErrMsg: "backup name cannot be an empty string",
		},
		{
			name:           "backup-test-name-3",
			namespace:      "",
			expectedErrMsg: "backup namespace cannot be an empty string",
		},
	}

	for _, test := range testcases {
		testBuilder := NewBackupBuilder(clients.GetTestClients(clients.TestClientParams{}), test.name, test.namespace)
		assert.Equal(t, test.expectedErrMsg, testBuilder.errorMsg)
	}
}

func TestPullBackup(t *testing.T) {
	generateBackup := func(name, namespace string) *velerov1.Backup {
		return &velerov1.Backup{
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
			name:                "backup-test-1",
			namespace:           "backup-test-namespace-1",
			expectedError:       false,
			addToRuntimeObjects: true,
		},
		{
			name:                "backup-test-2",
			namespace:           "backup-test-namespace-2",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "backup object backup-test-2 does not exist in namespace backup-test-namespace-2",
		},
		{
			name:                "",
			namespace:           "backup-test-namespace-3",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "backup name cannot be empty",
		},
		{
			name:                "backup-test-4",
			namespace:           "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "backup namespace cannot be empty",
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testBackup := generateBackup(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testBackup)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: v1TestSchemes,
		})

		// Test the Pull method
		builderResult, err := PullBackup(testSettings, testCase.name, testCase.namespace)
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

// buildValidBackupTestBuilder returns a valid Builder for testing purposes.
func buildValidBackupTestBuilder() *BackupBuilder {
	return NewBackupBuilder(clients.GetTestClients(clients.TestClientParams{}),
		"backup-test-name", "backup-test-namespace")
}

func TestWithStorageLocation(t *testing.T) {
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
			expectedErrorMsg: "backup storage location cannot be an empty string",
		},
	}

	for _, test := range testCases {
		testBuilder := buildValidBackupTestBuilder()

		testBuilder.WithStorageLocation(test.location)

		assert.Equal(t, test.expectedErrorMsg, testBuilder.errorMsg)
	}
}

func TestWithIncludedNamespace(t *testing.T) {
	testCases := []struct {
		namespaces       []string
		expectedErrorMsg string
	}{
		{
			namespaces:       []string{"includeme", "includeme2"},
			expectedErrorMsg: "",
		},
		{
			namespaces:       []string{"includeme", ""},
			expectedErrorMsg: "backup includedNamespace cannot be an empty string",
		},
	}

	for _, test := range testCases {
		testBuilder := buildValidBackupTestBuilder()

		for _, namespace := range test.namespaces {
			testBuilder.WithIncludedNamespace(namespace)
		}

		assert.Equal(t, test.expectedErrorMsg, testBuilder.errorMsg)

		if test.expectedErrorMsg == "" {
			assert.Len(t, testBuilder.Definition.Spec.IncludedNamespaces, len(test.namespaces))
		}
	}
}

func TestWithIncludedClusterScopedResource(t *testing.T) {
	testCases := []struct {
		clusterScopedResources []string
		expectedErrorMsg       string
	}{
		{
			clusterScopedResources: []string{"clusterroles", "clusterrolebindings"},
			expectedErrorMsg:       "",
		},
		{
			clusterScopedResources: []string{"clusterroles", ""},
			expectedErrorMsg:       "backup includedClusterScopedResource cannot be an empty string",
		},
	}

	for _, test := range testCases {
		testBuilder := buildValidBackupTestBuilder()

		for _, clusterScopedResources := range test.clusterScopedResources {
			testBuilder.WithIncludedClusterScopedResource(clusterScopedResources)
		}

		assert.Equal(t, test.expectedErrorMsg, testBuilder.errorMsg)

		if test.expectedErrorMsg == "" {
			assert.Len(t, testBuilder.Definition.Spec.IncludedClusterScopedResources, len(test.clusterScopedResources))
		}
	}
}

func TestWithIncludedNamespaceScopedResource(t *testing.T) {
	testCases := []struct {
		namespaceResources []string
		expectedErrorMsg   string
	}{
		{
			namespaceResources: []string{"deployments", "services", "secrets"},
			expectedErrorMsg:   "",
		},
		{
			namespaceResources: []string{"configmaps", ""},
			expectedErrorMsg:   "backup includedNamespaceScopedResource cannot be an empty string",
		},
	}

	for _, test := range testCases {
		testBuilder := buildValidBackupTestBuilder()

		for _, namespace := range test.namespaceResources {
			testBuilder.WithIncludedNamespaceScopedResource(namespace)
		}

		assert.Equal(t, test.expectedErrorMsg, testBuilder.errorMsg)

		if test.expectedErrorMsg == "" {
			assert.Len(t, testBuilder.Definition.Spec.IncludedNamespaceScopedResources, len(test.namespaceResources))
		}
	}
}

func TestWithExcludedClusterScopedResources(t *testing.T) {
	testCases := []struct {
		clusterScopedResources []string
		expectedErrorMsg       string
	}{
		{
			clusterScopedResources: []string{"clusterroles", "clusterrolebindings"},
			expectedErrorMsg:       "",
		},
		{
			clusterScopedResources: []string{"", "clusterrolebindings"},
			expectedErrorMsg:       "backup excludedClusterScopedResource cannot be an empty string",
		},
	}

	for _, test := range testCases {
		testBuilder := buildValidBackupTestBuilder()

		for _, clusterScopedResources := range test.clusterScopedResources {
			testBuilder.WithExcludedClusterScopedResource(clusterScopedResources)
		}

		assert.Equal(t, test.expectedErrorMsg, testBuilder.errorMsg)

		if test.expectedErrorMsg == "" {
			assert.Len(t, testBuilder.Definition.Spec.ExcludedClusterScopedResources, len(test.clusterScopedResources))
		}
	}
}

func TestWithExcludedNamespaceScopedResources(t *testing.T) {
	testCases := []struct {
		namespaceResources []string
		expectedErrorMsg   string
	}{
		{
			namespaceResources: []string{"deployments", "services", "secrets"},
			expectedErrorMsg:   "",
		},
		{
			namespaceResources: []string{"", "configmaps"},
			expectedErrorMsg:   "backup excludedNamespaceScopedResource cannot be an empty string",
		},
	}

	for _, test := range testCases {
		testBuilder := buildValidBackupTestBuilder()

		for _, namespaceResources := range test.namespaceResources {
			testBuilder.WithExcludedNamespaceScopedResources(namespaceResources)
		}

		assert.Equal(t, test.expectedErrorMsg, testBuilder.errorMsg)

		if test.expectedErrorMsg == "" {
			assert.Len(t, testBuilder.Definition.Spec.ExcludedNamespaceScopedResources, len(test.namespaceResources))
		}
	}
}
