package olm

import (
	"fmt"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	operatorsv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/olm/package-server/operators/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	operatorsv1Scheme = []clients.SchemeAttacher{
		operatorsv1.AddToScheme,
	}
)

func TestPullPackageManifest(t *testing.T) {
	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                "packagemanifest",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("packageManifest 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "packagemanifest",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("packageManifest 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "packagemanifest",
			namespace:           "test-namespace",
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf(
				"packageManifest object packagemanifest does not exist in namespace test-namespace"),
			client: true,
		},
		{
			name:                "packagemanifest",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("packagemanifest 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		packageManifest := buildPackageManifestDefinition(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, packageManifest)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: operatorsv1Scheme})
		}

		builderResult, err := PullPackageManifest(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

//nolint:funlen
func TestPullPackageManifestByCatalog(t *testing.T) {
	testCases := []struct {
		name                string
		namespace           string
		catalog             string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                "packagemanifest-test",
			namespace:           "test-namespace",
			catalog:             "test",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "packagemanifest-test",
			namespace:           "test-namespace",
			catalog:             "invalid",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("no matching PackageManifests were found"),
			client:              true,
		},
		{
			name:                "",
			namespace:           "test-namespace",
			catalog:             "test",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("packageManifest 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "packagemanifest",
			namespace:           "",
			catalog:             "test",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("failed to list packagemanifests, 'nsname' parameter is empty"),
			client:              true,
		},
		{
			name:                "packagemanifest",
			namespace:           "test-namespace",
			catalog:             "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("packageManifest 'catalog' cannot be empty"),
			client:              true,
		},
		{
			name:                "packagemanifest",
			namespace:           "test-namespace",
			catalog:             "test",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("no matching PackageManifests were found"),
			client:              true,
		},
		{
			name:                "packagemanifest",
			namespace:           "test-namespace",
			catalog:             "test",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("failed to list packageManifest, 'apiClient' parameter is empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		packageManifest := buildPackageManifestDefinition(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, packageManifest)
		}

		if testCase.client {
			var builder *fake.ClientBuilder

			testSettings, builder = clients.GetModifiableTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: operatorsv1Scheme})

			testSettings.Client = builder.WithIndex(
				&operatorsv1.PackageManifest{}, "metadata.name", nameMetadataIndexer).Build()
		}

		builderResult, err := PullPackageManifestByCatalog(
			testSettings, testCase.name, testCase.namespace, testCase.catalog)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

func TestPackageManifestGet(t *testing.T) {
	testCases := []struct {
		packageManifest *PackageManifestBuilder
		expectedError   string
	}{
		{
			packageManifest: buildValidPackageManifestBuilder(buildPackageManifestTestClientWithDummyObject()),
			expectedError:   "",
		},
		{
			packageManifest: buildValidPackageManifestBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: operatorsv1Scheme})),
			expectedError: "packagemanifests.packages.operators.coreos.com \"packagemanifest\" not found",
		},
	}

	for _, testCase := range testCases {
		packageManifest, err := testCase.packageManifest.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, packageManifest.Name, testCase.packageManifest.Definition.Name)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestPackageManifestExist(t *testing.T) {
	testCases := []struct {
		packageManifest *PackageManifestBuilder
		expectedStatus  bool
	}{
		{
			packageManifest: buildValidPackageManifestBuilder(buildPackageManifestTestClientWithDummyObject()),
			expectedStatus:  true,
		},
		{
			packageManifest: buildValidPackageManifestBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: operatorsv1Scheme})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.packageManifest.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func buildValidPackageManifestBuilder(apiClient *clients.Settings) *PackageManifestBuilder {
	return &PackageManifestBuilder{
		Definition: buildPackageManifestDefinition("packagemanifest", "test-namespace"),
		apiClient:  apiClient.Client,
	}
}

func buildPackageManifestTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyPackageManifest(),
		SchemeAttachers: operatorsv1Scheme,
	})
}

func buildDummyPackageManifest() []runtime.Object {
	return append([]runtime.Object{}, &operatorsv1.PackageManifest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "packagemanifest",
			Namespace: "test-namespace",
		},
	})
}

func buildPackageManifestDefinition(name, nsName string) *operatorsv1.PackageManifest {
	return &operatorsv1.PackageManifest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsName,
			Labels:    map[string]string{"catalog": "test"},
		},
	}
}

func nameMetadataIndexer(obj client.Object) []string {
	packageManifest, found := obj.(*operatorsv1.PackageManifest)
	if !found {
		panic(fmt.Errorf("indexer function for type %T's spec.replicas field received"+
			" object of type %T, this should never happen", operatorsv1.PackageManifest{}, obj))
	}

	return []string{packageManifest.Name}
}
